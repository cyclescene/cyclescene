package synclog

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/sync-logs", func(r chi.Router) {
		// POST /v1/sync-logs - Log a sync event (public endpoint)
		r.Post("/", h.LogSync)

		// GET /v1/sync-logs - Get recent sync logs (admin only)
		r.Get("/", h.GetRecentLogs)

		// GET /v1/sync-logs/stats - Get sync statistics (admin only)
		r.Get("/stats", h.GetStats)

		// GET /v1/sync-logs/client/{clientId} - Get logs for a specific client
		r.Get("/client/{clientId}", h.GetClientLogs)
	})
}

// LogSync handles POST /v1/sync-logs
// Records a sync event from a client
func (h *Handler) LogSync(w http.ResponseWriter, r *http.Request) {
	var log SyncLog
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		slog.Warn("Invalid sync log request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if log.ClientID == "" || log.SyncType == "" || log.Status == "" {
		slog.Warn("Missing required fields in sync log", "client_id", log.ClientID)
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Set timestamp if not provided
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	if err := h.repo.LogSync(&log); err != nil {
		slog.Error("Failed to log sync event", "error", err)
		http.Error(w, "Failed to log sync event", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      log.ID,
		"message": "Sync event logged successfully",
	})
}

// GetRecentLogs handles GET /v1/sync-logs
// Returns recent sync logs across all clients (admin only)
func (h *Handler) GetRecentLogs(w http.ResponseWriter, r *http.Request) {
	// TODO: Add admin authentication check here
	// For now, this endpoint is available to all clients

	limit := 100 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 500 {
			limit = parsedLimit
		}
	}

	logs, err := h.repo.GetRecentLogs(limit)
	if err != nil {
		slog.Error("Failed to fetch recent logs", "error", err)
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": len(logs),
		"logs":  logs,
	})
}

// GetClientLogs handles GET /v1/sync-logs/client/{clientId}
// Returns sync logs for a specific client
func (h *Handler) GetClientLogs(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientId")
	if clientID == "" {
		http.Error(w, "Missing client ID", http.StatusBadRequest)
		return
	}

	limit := 50 // Default limit for single client
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 500 {
			limit = parsedLimit
		}
	}

	logs, err := h.repo.GetClientLogs(clientID, limit)
	if err != nil {
		slog.Error("Failed to fetch client logs", "error", err, "client_id", clientID)
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"client_id": clientID,
		"count":     len(logs),
		"logs":      logs,
	})
}

// GetStats handles GET /v1/sync-logs/stats
// Returns aggregate statistics for syncs in the last 24 hours
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	// TODO: Add admin authentication check here

	// Default to last 24 hours, allow override via query param
	since := time.Now().Add(-24 * time.Hour)
	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = parsedTime
		}
	}

	stats, err := h.repo.GetSyncStats(since)
	if err != nil {
		slog.Error("Failed to fetch sync stats", "error", err)
		http.Error(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"since": since,
		"stats": stats,
	})
}
