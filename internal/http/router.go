package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"

	"sequi/internal/auth"
	"sequi/internal/chat"
	"sequi/internal/http/handlers"
	"sequi/internal/respond"
)

// NewRouter builds the chi router and registers all routes.
func NewRouter(pool *pgxpool.Pool, store *sessions.CookieStore, cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Serve uploaded files (avatars, etc.)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.MediaRoot))))

	hub := chat.NewHub()
	limiter := chat.NewRateLimiter()

	authH := handlers.NewAuthHandler(pool, store)
	categoriesH := handlers.NewCategoriesHandler(pool)
	streamKeysH := handlers.NewStreamKeysHandler(pool, store)
	streamsH := handlers.NewStreamsHandler(pool, cfg.HLSBaseURL)
	hooksH := handlers.NewHooksHandler(pool, cfg.HookSecret)
	chatH := handlers.NewChatHandler(pool, store, hub, limiter)
	subsH := handlers.NewSubscriptionsHandler(pool, store)
	uploadH := handlers.NewUploadHandler(pool, store, cfg.MediaRoot, cfg.PublicBaseURL)

	r.Route("/api", func(r chi.Router) {
		// Public
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/logout", authH.Logout)
		r.Get("/categories", categoriesH.List)
		r.Get("/streams", streamsH.List)
		r.Get("/streams/{id}", streamsH.GetByID)
		r.Get("/streams/{id}/hls", streamsH.GetHLS)
		r.Get("/users/{username}", subsH.GetProfile)

		// SRS hooks (validated by secret query param)
		r.Post("/hooks/stream/start", hooksH.Start)
		r.Post("/hooks/stream/stop", hooksH.Stop)

		// Authenticated
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(store))
			r.Get("/me", authH.Me)
			r.Get("/me/subscriptions", subsH.ListMySubscriptions)
			r.Post("/me/avatar", uploadH.UploadAvatar)
			r.Get("/stream-keys", streamKeysH.Get)
			r.Post("/stream-keys/regenerate", streamKeysH.Regenerate)
			r.Post("/users/{username}/subscribe", subsH.Subscribe)
			r.Delete("/users/{username}/subscribe", subsH.Unsubscribe)
		})
	})

	// WebSocket
	r.Get("/ws/streams/{id}/chat", chatH.ServeWS)

	return r
}

// RouterConfig holds all router-level dependencies.
type RouterConfig struct {
	HLSBaseURL    string
	HookSecret    string
	MediaRoot     string
	PublicBaseURL string
}
