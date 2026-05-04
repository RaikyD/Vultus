package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"sequi/internal/auth"
	db "sequi/internal/db/gen"
	"sequi/internal/respond"
)

const (
	maxUploadSize = 5 << 20 // 5 MB
)

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type UploadHandler struct {
	queries   *db.Queries
	store     *sessions.CookieStore
	mediaRoot string
	publicURL string
}

func NewUploadHandler(pool *pgxpool.Pool, store *sessions.CookieStore, mediaRoot, publicBaseURL string) *UploadHandler {
	return &UploadHandler{
		queries:   db.New(pool),
		store:     store,
		mediaRoot: mediaRoot,
		publicURL: strings.TrimRight(publicBaseURL, "/"),
	}
}

// UploadAvatar handles POST /api/me/avatar.
func (h *UploadHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respond.Error(w, http.StatusBadRequest, "validation_error", "file too large (max 5MB)")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "validation_error", "avatar field is required")
		return
	}
	defer file.Close()

	// Detect content type from first 512 bytes
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])

	ext, ok := allowedImageTypes[contentType]
	if !ok {
		respond.Error(w, http.StatusBadRequest, "validation_error", "unsupported image type; use JPEG, PNG, WebP or GIF")
		return
	}

	_ = header // suppress unused warning

	// Seek back to start before copying
	if seeker, ok := file.(io.Seeker); ok {
		_, _ = seeker.Seek(0, io.SeekStart)
	}

	// Store under avatars/<userID><ext>
	avatarsDir := filepath.Join(h.mediaRoot, "avatars")
	if err := os.MkdirAll(avatarsDir, 0755); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not create upload directory")
		return
	}

	filename := fmt.Sprintf("%s%s", userID.String(), ext)
	destPath := filepath.Join(avatarsDir, filename)

	dest, err := os.Create(destPath)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not save file")
		return
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not write file")
		return
	}

	avatarURL := fmt.Sprintf("%s/uploads/avatars/%s", h.publicURL, filename)

	user, err := h.queries.UpdateUserAvatar(r.Context(), db.UpdateUserAvatarParams{
		ID:        pgtype.UUID{Bytes: userID, Valid: true},
		AvatarUrl: pgtype.Text{String: avatarURL, Valid: true},
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not update avatar")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{
		"avatar_url": user.AvatarUrl.String,
		"id":         uuid.UUID(user.ID.Bytes).String(),
	})
}
