package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"sequi/internal/http/handlers"
)

// hookPayload matches the JSON SRS sends.
type hookPayload struct {
	Action   string `json:"action"`
	ClientID string `json:"client_id"`
	Stream   string `json:"stream"`
	App      string `json:"app"`
	Vhost    string `json:"vhost"`
}

func postHook(t *testing.T, handler http.Handler, path string, body any, secret string) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	url := path
	if secret != "" {
		url += "?secret=" + secret
	}
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// hookRouter builds a minimal chi router with only the hook endpoints,
// using nil pool (no DB) and the given secret.
// This lets us test secret validation without a real database.
func hookRouter(secret string) http.Handler {
	r := chi.NewRouter()
	h := handlers.NewHooksHandler(nil, secret)
	r.Post("/api/hooks/stream/start", h.Start)
	r.Post("/api/hooks/stream/stop", h.Stop)
	return r
}

func TestHookStart_InvalidSecret(t *testing.T) {
	router := hookRouter("correct-secret")
	rr := postHook(t, router, "/api/hooks/stream/start", hookPayload{Stream: "somekey"}, "wrong-secret")
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestHookStart_MissingSecret(t *testing.T) {
	router := hookRouter("correct-secret")
	rr := postHook(t, router, "/api/hooks/stream/start", hookPayload{Stream: "somekey"}, "")
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestHookStart_NoSecretConfigured_AllowsAll(t *testing.T) {
	// When hookSecret is empty, all requests are allowed (dev mode).
	// With nil pool the DB call will panic, so we only test up to the secret check.
	// A non-empty stream is needed to pass the empty-stream guard.
	// We expect NOT 403.
	router := hookRouter("")
	// This will panic on DB access — catch it via recover in a deferred check.
	// Instead, use an empty stream key to hit the "missing stream key" 400 error
	// before any DB access.
	rr := postHook(t, router, "/api/hooks/stream/start", hookPayload{Stream: ""}, "")
	if rr.Code == http.StatusForbidden {
		t.Error("expected non-403 when no secret is configured")
	}
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty stream key, got %d", rr.Code)
	}
}

func TestHookStop_InvalidSecret(t *testing.T) {
	router := hookRouter("my-secret")
	rr := postHook(t, router, "/api/hooks/stream/stop", hookPayload{Stream: "anykey"}, "bad")
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestHookStop_EmptyStream_ReturnsOK(t *testing.T) {
	// Empty stream on stop is not an error — just ignored.
	router := hookRouter("")
	rr := postHook(t, router, "/api/hooks/stream/stop", hookPayload{Stream: ""}, "")
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for empty stream on stop, got %d", rr.Code)
	}
}

func TestHookStart_InvalidJSON(t *testing.T) {
	router := hookRouter("")
	req := httptest.NewRequest(http.MethodPost, "/api/hooks/stream/start", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}
