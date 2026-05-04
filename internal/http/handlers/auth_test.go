package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"

	"sequi/internal/http/handlers"
)

func authRouter() http.Handler {
	r := chi.NewRouter()
	store := sessions.NewCookieStore([]byte("test-secret"))
	h := handlers.NewAuthHandler(nil, store)
	r.Post("/api/auth/register", h.Register)
	r.Post("/api/auth/login", h.Login)
	r.Post("/api/auth/logout", h.Logout)
	return r
}

func postJSON(t *testing.T, router http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func TestRegister_InvalidEmail(t *testing.T) {
	rr := postJSON(t, authRouter(), "/api/auth/register", map[string]string{
		"email": "not-an-email", "username": "alice", "password": "password123",
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid email, got %d", rr.Code)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	rr := postJSON(t, authRouter(), "/api/auth/register", map[string]string{
		"email": "alice@example.com", "username": "alice", "password": "short",
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for short password, got %d", rr.Code)
	}
}

func TestRegister_InvalidUsername(t *testing.T) {
	cases := []string{"ab", "has space", "has-dash", "toolongusernamethatexceedsthirtychars"}
	for _, username := range cases {
		rr := postJSON(t, authRouter(), "/api/auth/register", map[string]string{
			"email": "alice@example.com", "username": username, "password": "password123",
		})
		if rr.Code != http.StatusBadRequest {
			t.Errorf("username %q: expected 400, got %d", username, rr.Code)
		}
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	authRouter().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString("!!!"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	authRouter().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}
