package clienterror

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// LogError handles POST /v1/client-errors
// Records a client error from the PWA
func (h *Handler) LogError(w http.ResponseWriter, r *http.Request) {
	var errLog ClientError
	if err := json.NewDecoder(r.Body).Decode(&errLog); err != nil {
		slog.Warn("Invalid client error request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if errLog.ClientID == "" || errLog.ErrorType == "" || errLog.ErrorMsg == "" {
		slog.Warn("Missing required fields in client error", "client_id", errLog.ClientID)
		http.Error(w, "Missing required fields: client_id, error_type, error_msg", http.StatusBadRequest)
		return
	}

	// Set timestamp if not provided
	if errLog.Timestamp.IsZero() {
		errLog.Timestamp = time.Now()
	}

	if err := h.repo.LogError(&errLog); err != nil {
		slog.Error("Failed to log client error", "error", err)
		http.Error(w, "Failed to log client error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      errLog.ID,
		"message": "Client error logged successfully",
	})
}
