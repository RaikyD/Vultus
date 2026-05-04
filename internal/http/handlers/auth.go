package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"sequi/internal/auth"
	db "sequi/internal/db/gen"
	"sequi/internal/respond"
	"sequi/internal/stream"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

// AuthHandler holds dependencies for auth endpoints.
type AuthHandler struct {
	queries *db.Queries
	store   *sessions.CookieStore
}

// NewAuthHandler creates an AuthHandler wired to the given pool and session store.
func NewAuthHandler(pool *pgxpool.Pool, store *sessions.CookieStore) *AuthHandler {
	return &AuthHandler{
		queries: db.New(pool),
		store:   store,
	}
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "validation_error", "invalid JSON")
		return
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.Username = strings.TrimSpace(body.Username)

	if err := validateRegister(body.Email, body.Username, body.Password); err != nil {
		respond.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not hash password")
		return
	}

	user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		Email:        body.Email,
		PasswordHash: hash,
		Username:     body.Username,
	})
	if err != nil {
		if isDuplicateError(err) {
			respond.Error(w, http.StatusConflict, "conflict", "email or username already taken")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not create user")
		return
	}

	// Create stream key and stream entity for the new user
	streamKey, err := stream.GenerateKey()
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not generate stream key")
		return
	}
	if _, err := h.queries.CreateStreamKey(r.Context(), db.CreateStreamKeyParams{
		UserID: user.ID,
		Key:    streamKey,
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not create stream key")
		return
	}
	if _, err := h.queries.CreateStream(r.Context(), user.ID); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not create stream entity")
		return
	}

	if err := auth.SaveUserID(w, r, h.store, uuid.UUID(user.ID.Bytes)); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not create session")
		return
	}

	auth.SetCSRFToken(w)
	respond.JSON(w, http.StatusCreated, userResponse(user))
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "validation_error", "invalid JSON")
		return
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))

	user, err := h.queries.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return
	}

	if !auth.CheckPassword(user.PasswordHash, body.Password) {
		respond.Error(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return
	}

	if err := auth.SaveUserID(w, r, h.store, uuid.UUID(user.ID.Bytes)); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not create session")
		return
	}

	auth.SetCSRFToken(w)
	respond.JSON(w, http.StatusOK, userResponse(user))
}

// Logout handles POST /api/auth/logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := auth.Clear(w, r, h.store); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not clear session")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// Me handles GET /api/me.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	pgID := pgtype.UUID{Bytes: userID, Valid: true}
	user, err := h.queries.GetUserByID(r.Context(), pgID)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	respond.JSON(w, http.StatusOK, userResponse(user))
}

// --- helpers ---

type userResp struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
	Bio       *string `json:"bio"`
	CreatedAt string  `json:"created_at"`
}

func userResponse(u db.User) userResp {
	var avatarURL *string
	if u.AvatarUrl.Valid {
		avatarURL = &u.AvatarUrl.String
	}
	var bio *string
	if u.Bio.Valid {
		bio = &u.Bio.String
	}
	return userResp{
		ID:        uuid.UUID(u.ID.Bytes).String(),
		Email:     u.Email,
		Username:  u.Username,
		AvatarURL: avatarURL,
		Bio:       bio,
		CreatedAt: u.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func validateRegister(email, username, password string) error {
	if !strings.Contains(email, "@") || len(email) < 3 {
		return errorStr("invalid email")
	}
	if !usernameRe.MatchString(username) {
		return errorStr("username must be 3-30 alphanumeric characters or underscores")
	}
	if len(password) < 8 {
		return errorStr("password must be at least 8 characters")
	}
	return nil
}

func isDuplicateError(err error) bool {
	s := err.Error()
	return strings.Contains(s, "unique constraint") || strings.Contains(s, "duplicate key")
}

type errorStr string

func (e errorStr) Error() string { return string(e) }
