package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	db "sequi/internal/db/gen"
	"sequi/internal/respond"
)

// HooksHandler handles SRS HTTP hook callbacks.
type HooksHandler struct {
	queries    *db.Queries
	hookSecret string
}

func NewHooksHandler(pool *pgxpool.Pool, hookSecret string) *HooksHandler {
	return &HooksHandler{queries: db.New(pool), hookSecret: hookSecret}
}

// srsHookPayload is the JSON body SRS sends on on_publish / on_unpublish.
type srsHookPayload struct {
	Action   string `json:"action"`
	ClientID string `json:"client_id"`
	Stream   string `json:"stream"` // contains the stream key
	App      string `json:"app"`
	Vhost    string `json:"vhost"`
}

// srsOK writes the response SRS hooks require: {"code":0}
// SRS checks for this exact field — any other format causes stream rejection.
func srsOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"code":0}`))
}

// Start handles POST /api/hooks/stream/start (on_publish).
// SRS expects HTTP 200 with {"code":0} to allow the stream; anything else rejects it.
func (h *HooksHandler) Start(w http.ResponseWriter, r *http.Request) {
	if !h.validateSecret(r) {
		respond.Error(w, http.StatusForbidden, "forbidden", "invalid hook secret")
		return
	}

	var payload srsHookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respond.Error(w, http.StatusBadRequest, "validation_error", "invalid payload")
		return
	}

	if payload.Stream == "" {
		respond.Error(w, http.StatusBadRequest, "validation_error", "missing stream key")
		return
	}

	// Look up the stream key
	sk, err := h.queries.GetStreamKeyByKey(r.Context(), payload.Stream)
	if err != nil {
		// Unknown or revoked key — reject the stream
		respond.Error(w, http.StatusForbidden, "forbidden", "unknown stream key")
		return
	}

	// Set stream status to live (idempotent — if already live, no harm)
	if _, err := h.queries.UpdateStreamStatus(r.Context(), db.UpdateStreamStatusParams{
		UserID: sk.UserID,
		Status: "live",
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not update stream status")
		return
	}

	srsOK(w)
}

// Stop handles POST /api/hooks/stream/stop (on_unpublish).
func (h *HooksHandler) Stop(w http.ResponseWriter, r *http.Request) {
	if !h.validateSecret(r) {
		respond.Error(w, http.StatusForbidden, "forbidden", "invalid hook secret")
		return
	}

	var payload srsHookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respond.Error(w, http.StatusBadRequest, "validation_error", "invalid payload")
		return
	}

	if payload.Stream == "" {
		srsOK(w)
		return
	}

	sk, err := h.queries.GetStreamKeyByKey(r.Context(), payload.Stream)
	if err != nil {
		// Key not found — nothing to stop, not an error
		srsOK(w)
		return
	}

	// Set stream status to offline (idempotent)
	if _, err := h.queries.UpdateStreamStatus(r.Context(), db.UpdateStreamStatusParams{
		UserID: sk.UserID,
		Status: "offline",
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not update stream status")
		return
	}

	srsOK(w)
}

// validateSecret checks the ?secret= query param against the configured hook secret.
// If no secret is configured, all requests are allowed (dev convenience).
func (h *HooksHandler) validateSecret(r *http.Request) bool {
	if h.hookSecret == "" {
		return true
	}
	return r.URL.Query().Get("secret") == h.hookSecret
}
