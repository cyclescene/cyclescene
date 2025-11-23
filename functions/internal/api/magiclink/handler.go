package magiclink

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetService() *Service {
	return h.service
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Magic links are sent automatically on ride submission, no public endpoint needed
}

type SendMagicLinkReq struct {
	Email       string `json:"email"`
	RedirectURL string `json:"redirect_url"`
}

// SendMagicLink sends a magic link email to the ride organizer
// The redirect_url should include the edit_token query parameter
func (h *Handler) SendMagicLink(w http.ResponseWriter, r *http.Request) {
	var req SendMagicLinkReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if req.RedirectURL == "" {
		http.Error(w, "Redirect URL is required", http.StatusBadRequest)
		return
	}

	// Get client IP for security tracking
	clientIP := getClientIP(r)

	res, err := h.service.SendMagicLink(r.Context(), SendMagicLinkRequest{
		Email:       req.Email,
		RedirectURL: req.RedirectURL,
		IPAddress:   clientIP,
	})

	if err != nil {
		slog.Error("Failed to send magic link", "error", err)
		http.Error(w, "Failed to send magic link", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// getClientIP extracts the client IP from the request, respecting X-Forwarded-For header
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// Get the first IP in the list
		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}
