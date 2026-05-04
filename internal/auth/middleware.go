package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"

	"sequi/internal/respond"
)

type contextKey string

const contextKeyUserID contextKey = "userID"

// RequireAuth is middleware that rejects unauthenticated requests with 401.
// On success it stores the user ID in the request context.
func RequireAuth(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := GetUserID(r, store)
			if !ok {
				respond.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
				return
			}
			ctx := context.WithValue(r.Context(), contextKeyUserID, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the user ID placed by RequireAuth.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(contextKeyUserID).(uuid.UUID)
	return id, ok
}
