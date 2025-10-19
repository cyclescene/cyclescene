package group

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/spacesedan/cyclescene/functions/internal/api/imageoptimizer"
)

type Handler struct {
	service         *Service
	optimizerClient *imageoptimizer.Client
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service:         service,
		optimizerClient: imageoptimizer.NewClient(),
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/groups", func(r chi.Router) {
		r.Get("/validate/{code}", h.ValidateGroupCode)
		r.Post("/check-code", h.CheckCodeAvailability)
		r.Post("/register", h.RegisterGroup)
		r.Get("/edit/{token}", h.GetGroupByEditToken)
		r.Put("/edit/{token}", h.UpdateGroup)
	})
}

func (h *Handler) ValidateGroupCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	validation, err := h.service.ValidateGroupCode(code)
	if err != nil {
		slog.Error("Failed to validate group code", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(validation)
}

func (h *Handler) CheckCodeAvailability(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	availability, err := h.service.CheckCodeAvailability(req.Code)
	if err != nil {
		slog.Error("Failed to check code availability", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availability)
}

func (h *Handler) RegisterGroup(w http.ResponseWriter, r *http.Request) {
	// Validate BFF token
	bffToken := r.Header.Get("X-BFF-Token")
	if bffToken == "" {
		http.Error(w, "Missing BFF token", http.StatusUnauthorized)
		return
	}

	var registration Registration
	if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if registration.Code == "" || registration.Name == "" || registration.City == "" {
		http.Error(w, "Missing required fields (code, name, city)", http.StatusBadRequest)
		return
	}

	response, err := h.service.RegisterGroup(&registration)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, "Group code already exists", http.StatusConflict)
			return
		}
		slog.Error("Failed to register group", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("Group registered successfully",
		"code", response.Code,
		"name", registration.Name,
		"city", registration.City,
	)

	// Trigger image optimization if icon_uuid is provided
	if registration.IconUUID != "" {
		optimizeReq := &imageoptimizer.OptimizeRequest{
			ImageUUID:  registration.IconUUID,
			CityCode:   registration.City,
			EntityID:   registration.Code,
			EntityType: "group",
		}
		h.optimizerClient.TriggerOptimization(context.Background(), optimizeReq)
		slog.Info("Image optimization triggered for group", "code", response.Code, "image_uuid", registration.IconUUID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetGroupByEditToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	group, err := h.service.GetGroupByEditToken(token)
	if err == sql.ErrNoRows {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("Failed to get group", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	var registration Registration
	if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.service.UpdateGroup(token, &registration)
	if err == sql.ErrNoRows {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("Failed to update group", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("Group updated successfully", "token", token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
