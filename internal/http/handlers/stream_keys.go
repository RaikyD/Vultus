package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"sequi/internal/auth"
	db "sequi/internal/db/gen"
	"sequi/internal/respond"
	"sequi/internal/stream"
)

type StreamKeysHandler struct {
	queries *db.Queries
	store   *sessions.CookieStore
}

func NewStreamKeysHandler(pool *pgxpool.Pool, store *sessions.CookieStore) *StreamKeysHandler {
	return &StreamKeysHandler{queries: db.New(pool), store: store}
}

// Get handles GET /api/stream-keys.
func (h *StreamKeysHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	pgID := pgtype.UUID{Bytes: userID, Valid: true}
	sk, err := h.queries.GetActiveStreamKey(r.Context(), pgID)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "not_found", "no active stream key")
		return
	}

	respond.JSON(w, http.StatusOK, streamKeyResp(sk))
}

// Regenerate handles POST /api/stream-keys/regenerate.
func (h *StreamKeysHandler) Regenerate(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	pgID := pgtype.UUID{Bytes: userID, Valid: true}

	newKey, err := stream.GenerateKey()
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not generate key")
		return
	}

	// Revoke existing key
	if err := h.queries.RevokeStreamKey(r.Context(), pgID); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not revoke old key")
		return
	}

	sk, err := h.queries.CreateStreamKey(r.Context(), db.CreateStreamKeyParams{
		UserID: pgID,
		Key:    newKey,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not create stream key")
		return
	}

	respond.JSON(w, http.StatusOK, streamKeyResp(sk))
}

type streamKeyResponse struct {
	ID        string  `json:"id"`
	Key       string  `json:"key"`
	CreatedAt string  `json:"created_at"`
	RevokedAt *string `json:"revoked_at"`
}

func streamKeyResp(sk db.StreamKey) streamKeyResponse {
	var revokedAt *string
	if sk.RevokedAt.Valid {
		s := sk.RevokedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		revokedAt = &s
	}
	return streamKeyResponse{
		ID:        uuid.UUID(sk.ID.Bytes).String(),
		Key:       sk.Key,
		CreatedAt: sk.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		RevokedAt: revokedAt,
	}
}
