package routes

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// GetAllRoutes returns all routes as GeoJSON features
// GET /v1/routes
func (h *Handler) GetAllRoutes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slog.Info("[Routes] Fetching all routes")

	routes, err := h.repo.GetAllRoutes(ctx)
	if err != nil {
		slog.Error("[Routes] Failed to fetch routes", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch routes"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(routes); err != nil {
		slog.Error("[Routes] Failed to encode response", "error", err)
	}
}

// RegisterRoutes registers all route handlers
func (h *Handler) RegisterRoutes(r interface {
	Get(string, http.HandlerFunc)
}) {
	slog.Info("[Routes] Registering routes handler")
	r.Get("/routes", h.GetAllRoutes)
}
