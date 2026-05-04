package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"sequi/internal/auth"
	"sequi/internal/chat"
	db "sequi/internal/db/gen"
	"sequi/internal/respond"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins in dev; restrict in production via config
	CheckOrigin: func(r *http.Request) bool { return true },
}

type ChatHandler struct {
	queries *db.Queries
	store   *sessions.CookieStore
	hub     *chat.Hub
	limiter *chat.RateLimiter
}

func NewChatHandler(pool *pgxpool.Pool, store *sessions.CookieStore, hub *chat.Hub, limiter *chat.RateLimiter) *ChatHandler {
	return &ChatHandler{
		queries: db.New(pool),
		store:   store,
		hub:     hub,
		limiter: limiter,
	}
}

// ServeWS handles GET /ws/streams/{id}/chat.
func (h *ChatHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	streamID, err := uuid.Parse(idStr)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "validation_error", "invalid stream id")
		return
	}

	// Verify stream exists
	pgStreamID := pgtype.UUID{Bytes: streamID, Valid: true}
	if _, err := h.queries.GetStreamByID(r.Context(), pgStreamID); err != nil {
		respond.Error(w, http.StatusNotFound, "not_found", "stream not found")
		return
	}

	// Try to read user from session (optional — unauthenticated users can view)
	var userID uuid.UUID
	var username string
	if uid, ok := auth.GetUserID(r, h.store); ok {
		userID = uid
		if u, err := h.queries.GetUserByID(r.Context(), pgtype.UUID{Bytes: uid, Valid: true}); err == nil {
			username = u.Username
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}

	client := chat.NewClient(h.hub, conn, h.queries, h.limiter, streamID, userID, username)
	h.hub.Register(client)

	// Send recent message history
	h.sendHistory(r, client, pgStreamID)

	go client.WritePump()
	go client.ReadPump(context.Background())
}

func (h *ChatHandler) sendHistory(r *http.Request, client *chat.Client, pgStreamID pgtype.UUID) {
	rows, err := h.queries.ListRecentMessages(r.Context(), db.ListRecentMessagesParams{
		StreamID: pgStreamID,
		Limit:    50,
	})
	if err != nil {
		return
	}

	// Reverse so oldest message comes first
	msgs := make([]chat.OutboundMessage, len(rows))
	for i, row := range rows {
		msgs[len(rows)-1-i] = chat.OutboundMessage{
			Type:      chat.TypeMessage,
			ID:        uuid.UUID(row.ID.Bytes).String(),
			UserID:    uuid.UUID(row.UserID.Bytes).String(),
			Username:  row.Username,
			Content:   row.Content,
			CreatedAt: row.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	payload, _ := json.Marshal(chat.HistoryMessage{
		Type:     chat.TypeHistory,
		Messages: msgs,
	})
	select {
	case client.Send() <- payload:
	default:
	}
}
