package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "sequi/internal/db/gen"
	"sequi/internal/respond"
)

type StreamsHandler struct {
	queries    *db.Queries
	hlsBaseURL string
}

func NewStreamsHandler(pool *pgxpool.Pool, hlsBaseURL string) *StreamsHandler {
	return &StreamsHandler{queries: db.New(pool), hlsBaseURL: hlsBaseURL}
}

// List handles GET /api/streams.
// Query params: status, category_id, search.
func (h *StreamsHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	params := db.ListStreamsParams{
		Column1: q.Get("status"),
		Column3: q.Get("search"),
	}

	if catStr := q.Get("category_id"); catStr != "" {
		id, err := uuid.Parse(catStr)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, "validation_error", "invalid category_id")
			return
		}
		params.Column2 = pgtype.UUID{Bytes: id, Valid: true}
	}

	rows, err := h.queries.ListStreams(r.Context(), params)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not fetch streams")
		return
	}

	out := make([]streamListItem, 0, len(rows))
	for _, s := range rows {
		out = append(out, toStreamListItem(s))
	}
	respond.JSON(w, http.StatusOK, out)
}

// GetByID handles GET /api/streams/:id.
func (h *StreamsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "validation_error", "invalid stream id")
		return
	}

	pgID := pgtype.UUID{Bytes: id, Valid: true}
	s, err := h.queries.GetStreamByID(r.Context(), pgID)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "not_found", "stream not found")
		return
	}

	respond.JSON(w, http.StatusOK, toStreamDetail(s))
}

// GetHLS handles GET /api/streams/:id/hls.
func (h *StreamsHandler) GetHLS(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "validation_error", "invalid stream id")
		return
	}

	pgID := pgtype.UUID{Bytes: id, Valid: true}
	s, err := h.queries.GetStreamByID(r.Context(), pgID)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "not_found", "stream not found")
		return
	}

	// To build the HLS URL we need the stream key for this user.
	// We query the active key for the stream's user.
	sk, err := h.queries.GetActiveStreamKey(r.Context(), s.UserID)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "not_found", "no active stream key")
		return
	}

	hlsURL := h.hlsBaseURL + "/" + sk.Key + "/index.m3u8"
	respond.JSON(w, http.StatusOK, map[string]string{"url": hlsURL})
}

// --- response types ---

type streamListItem struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Status      string  `json:"status"`
	ViewerCount int32   `json:"viewer_count"`
	Username    string  `json:"username"`
	AvatarURL   *string `json:"avatar_url"`
	CategoryID  *string `json:"category_id"`
}

type streamDetail struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Status      string  `json:"status"`
	ViewerCount int32   `json:"viewer_count"`
	Username    string  `json:"username"`
	UserID      string  `json:"user_id"`
	AvatarURL   *string `json:"avatar_url"`
	CategoryID  *string `json:"category_id"`
	StartedAt   *string `json:"started_at"`
}

func toStreamListItem(s db.ListStreamsRow) streamListItem {
	item := streamListItem{
		ID:          uuid.UUID(s.ID.Bytes).String(),
		Title:       s.Title,
		Status:      s.Status,
		ViewerCount: s.ViewerCount,
		Username:    s.Username,
	}
	if s.UserAvatarUrl.Valid {
		item.AvatarURL = &s.UserAvatarUrl.String
	}
	if s.CategoryID.Valid {
		catStr := uuid.UUID(s.CategoryID.Bytes).String()
		item.CategoryID = &catStr
	}
	return item
}

func toStreamDetail(s db.GetStreamByIDRow) streamDetail {
	d := streamDetail{
		ID:          uuid.UUID(s.ID.Bytes).String(),
		Title:       s.Title,
		Status:      s.Status,
		ViewerCount: s.ViewerCount,
		Username:    s.Username,
		UserID:      uuid.UUID(s.UserID.Bytes).String(),
	}
	if s.UserAvatarUrl.Valid {
		d.AvatarURL = &s.UserAvatarUrl.String
	}
	if s.CategoryID.Valid {
		catStr := uuid.UUID(s.CategoryID.Bytes).String()
		d.CategoryID = &catStr
	}
	if s.StartedAt.Valid {
		t := s.StartedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		d.StartedAt = &t
	}
	return d
}
