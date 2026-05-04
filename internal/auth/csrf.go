package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const csrfCookieName = "csrf_token"

// SetCSRFToken generates a random CSRF token, sets it as a cookie, and returns the token value.
func SetCSRFToken(w http.ResponseWriter) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)

	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // must be readable by JS
		SameSite: http.SameSiteLaxMode,
	})
	return token
}

// ValidateCSRFToken compares the X-CSRF-Token header against the csrf_token cookie.
// Returns true only when both are present and identical.
func ValidateCSRFToken(r *http.Request) bool {
	header := r.Header.Get("X-CSRF-Token")
	if header == "" {
		return false
	}
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil {
		return false
	}
	return header == cookie.Value
}
