package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"sequi/internal/auth"
	db "sequi/internal/db/gen"
	"sequi/internal/respond"
)

type SubscriptionsHandler struct {
	queries *db.Queries
	store   *sessions.CookieStore
}

func NewSubscriptionsHandler(pool *pgxpool.Pool, store *sessions.CookieStore) *SubscriptionsHandler {
	return &SubscriptionsHandler{queries: db.New(pool), store: store}
}

// GetProfile handles GET /api/users/{username}.
// Public endpoint — returns streamer profile + stream info.
// If requester is authenticated, includes subscription status.
func (h *SubscriptionsHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		respond.Error(w, http.StatusBadRequest, "validation_error", "username required")
		return
	}

	streamer, err := h.queries.GetStreamerByUsername(r.Context(), username)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	streamerID := uuid.UUID(streamer.ID.Bytes)
	pgStreamerID := pgtype.UUID{Bytes: streamerID, Valid: true}

	subscribers, _ := h.queries.CountSubscribers(r.Context(), pgStreamerID)

	type streamInfo struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Status      string `json:"status"`
		ViewerCount int32  `json:"viewer_count"`
	}

	type profileResp struct {
		ID              string     `json:"id"`
		Username        string     `json:"username"`
		AvatarURL       *string    `json:"avatar_url"`
		Bio             *string    `json:"bio"`
		SubscriberCount int64      `json:"subscriber_count"`
		IsSubscribed    bool       `json:"is_subscribed"`
		Stream          streamInfo `json:"stream"`
	}

	var avatarURL *string
	if streamer.AvatarUrl.Valid {
		avatarURL = &streamer.AvatarUrl.String
	}
	var bio *string
	if streamer.Bio.Valid {
		bio = &streamer.Bio.String
	}

	resp := profileResp{
		ID:              streamerID.String(),
		Username:        streamer.Username,
		AvatarURL:       avatarURL,
		Bio:             bio,
		SubscriberCount: subscribers,
		Stream: streamInfo{
			ID:          uuid.UUID(streamer.StreamID.Bytes).String(),
			Title:       streamer.Title,
			Status:      streamer.Status,
			ViewerCount: streamer.ViewerCount,
		},
	}

	// Check subscription status if authenticated
	if requesterID, ok := auth.GetUserID(r, h.store); ok {
		pgRequesterID := pgtype.UUID{Bytes: requesterID, Valid: true}
		isSubscribed, _ := h.queries.IsSubscribed(r.Context(), db.IsSubscribedParams{
			SubscriberID: pgRequesterID,
			StreamerID:   pgStreamerID,
		})
		resp.IsSubscribed = isSubscribed
	}

	respond.JSON(w, http.StatusOK, resp)
}

// Subscribe handles POST /api/users/{username}/subscribe.
func (h *SubscriptionsHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	username := chi.URLParam(r, "username")
	streamer, err := h.queries.GetStreamerByUsername(r.Context(), username)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	streamerID := uuid.UUID(streamer.ID.Bytes)
	if requesterID == streamerID {
		respond.Error(w, http.StatusBadRequest, "validation_error", "cannot subscribe to yourself")
		return
	}

	if err := h.queries.Subscribe(r.Context(), db.SubscribeParams{
		SubscriberID: pgtype.UUID{Bytes: requesterID, Valid: true},
		StreamerID:   pgtype.UUID{Bytes: streamerID, Valid: true},
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not subscribe")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]bool{"subscribed": true})
}

// Unsubscribe handles DELETE /api/users/{username}/subscribe.
func (h *SubscriptionsHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	username := chi.URLParam(r, "username")
	streamer, err := h.queries.GetStreamerByUsername(r.Context(), username)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	if err := h.queries.Unsubscribe(r.Context(), db.UnsubscribeParams{
		SubscriberID: pgtype.UUID{Bytes: requesterID, Valid: true},
		StreamerID:   pgtype.UUID{Bytes: uuid.UUID(streamer.ID.Bytes), Valid: true},
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not unsubscribe")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]bool{"subscribed": false})
}

// ListMySubscriptions handles GET /api/me/subscriptions.
func (h *SubscriptionsHandler) ListMySubscriptions(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	rows, err := h.queries.ListSubscriptions(r.Context(), pgtype.UUID{Bytes: requesterID, Valid: true})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not fetch subscriptions")
		return
	}

	type subItem struct {
		ID           string  `json:"id"`
		Username     string  `json:"username"`
		AvatarURL    *string `json:"avatar_url"`
		SubscribedAt string  `json:"subscribed_at"`
	}

	result := make([]subItem, 0, len(rows))
	for _, row := range rows {
		var avatarURL *string
		if row.AvatarUrl.Valid {
			avatarURL = &row.AvatarUrl.String
		}
		result = append(result, subItem{
			ID:           uuid.UUID(row.ID.Bytes).String(),
			Username:     row.Username,
			AvatarURL:    avatarURL,
			SubscribedAt: row.SubscribedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	respond.JSON(w, http.StatusOK, result)
}
