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

// GetAllRoutes returns all routes for a specific city as GeoJSON features
// GET /v1/routes?city=pdx
func (h *Handler) GetAllRoutes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slog.Info("[Routes] Request received", "url", r.URL.String(), "query", r.URL.RawQuery)
	city := r.URL.Query().Get("city")
	slog.Info("[Routes] Extracted city from query", "city", city)

	if city == "" {
		slog.Warn("[Routes] Missing city parameter")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "city parameter is required"})
		return
	}

	slog.Info("[Routes] Fetching all routes", "city", city)

	routes, err := h.repo.GetAllRoutes(ctx, city)
	if err != nil {
		slog.Error("[Routes] Failed to fetch routes", "error", err, "city", city)
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
