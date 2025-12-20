package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/tokens", func(r chi.Router) {
		r.Post("/submission", h.GenerateSubmissionToken)
		r.Post("/validate", h.ValidateSubmissionToken)
	})
}

type TokenRequest struct {
	City string `json:"city"`
}

func (h *Handler) GenerateSubmissionToken(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.City == "" {
		http.Error(w, "City is required", http.StatusBadRequest)
		return
	}

	token, err := h.service.GenerateSubmissionToken(req.City)
	if err != nil {
		slog.Error("Failed to generate token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("Generated submission token", "city", req.City, "expires_at", token.ExpiresAt)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(token); err != nil {
		slog.Error("Failed to encode token response", "error", err)
	}
}

type ValidateTokenRequest struct {
	Token string `json:"token"`
	City  string `json:"city"`
}

func (h *Handler) ValidateSubmissionToken(w http.ResponseWriter, r *http.Request) {
	var req ValidateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validation, err := h.service.ValidateSubmissionToken(req.Token, req.City)
	if err != nil {
		slog.Error("Failed to validate token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(validation); err != nil {
		slog.Error("Failed to encode validation response", "error", err)
	}
}
