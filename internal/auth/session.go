package auth

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

const (
	sessionName   = "sequi_session"
	sessionUserID = "user_id"
)

// NewStore creates a new cookie-based session store.
func NewStore(secret string) *sessions.CookieStore {
	store := sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // set to true in production with HTTPS
	}
	return store
}

// SaveUserID stores the user ID in the session cookie.
func SaveUserID(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, userID uuid.UUID) error {
	sess, err := store.Get(r, sessionName)
	if err != nil {
		// Start fresh if session is corrupt
		sess = sessions.NewSession(store, sessionName)
	}
	sess.Values[sessionUserID] = userID.String()
	return store.Save(r, w, sess)
}

// GetUserID retrieves the user ID from the session cookie.
// Returns (uuid.Nil, false) if the session is absent or invalid.
func GetUserID(r *http.Request, store *sessions.CookieStore) (uuid.UUID, bool) {
	sess, err := store.Get(r, sessionName)
	if err != nil {
		return uuid.Nil, false
	}
	raw, ok := sess.Values[sessionUserID]
	if !ok {
		return uuid.Nil, false
	}
	str, ok := raw.(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(str)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// Clear deletes the session cookie.
func Clear(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore) error {
	sess, err := store.Get(r, sessionName)
	if err != nil {
		return nil // nothing to clear
	}
	sess.Options.MaxAge = -1
	return store.Save(r, w, sess)
}
