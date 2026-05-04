package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	db "sequi/internal/db/gen"
	"sequi/internal/respond"
)

type CategoriesHandler struct {
	queries *db.Queries
}

func NewCategoriesHandler(pool *pgxpool.Pool) *CategoriesHandler {
	return &CategoriesHandler{queries: db.New(pool)}
}

// List handles GET /api/categories.
func (h *CategoriesHandler) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.queries.ListCategories(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal_error", "could not fetch categories")
		return
	}

	type catResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	out := make([]catResp, len(cats))
	for i, c := range cats {
		out[i] = catResp{
			ID:   uuid.UUID(c.ID.Bytes).String(),
			Name: c.Name,
		}
	}
	respond.JSON(w, http.StatusOK, out)
}
