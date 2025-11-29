package group

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/spacesedan/cyclescene/functions/internal/api/events"
)

type Handler struct {
	service        *Service
	eventarcClient *events.EventarcClient
}

func NewHandler(service *Service, eventarcClient *events.EventarcClient) *Handler {
	return &Handler{
		service:        service,
		eventarcClient: eventarcClient,
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

	// Trigger marker image optimization if image_uuid is provided
	if registration.ImageUUID != "" && h.eventarcClient != nil {
		event := &events.ImageOptimizationEvent{
			ImageUUID:  registration.ImageUUID,
			CityCode:   registration.City,
			EntityID:   registration.Code,
			EntityType: "group",
		}
		if err := h.eventarcClient.TriggerOptimization(r.Context(), event); err != nil {
			slog.Warn("failed to trigger marker optimization", "error", err, "code", response.Code)
			// Don't fail the request if optimization trigger fails - the image is already uploaded
		}
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

	// Trigger marker image optimization if a new marker image_uuid is provided
	if registration.ImageUUID != "" && h.eventarcClient != nil {
		// Get group info to retrieve city code
		group, err := h.service.GetGroupByEditToken(token)
		if err == nil && group != nil {
			event := &events.ImageOptimizationEvent{
				ImageUUID:  registration.ImageUUID,
				CityCode:   group.City,
				EntityID:   group.Code,
				EntityType: "group",
			}
			if err := h.eventarcClient.TriggerOptimization(r.Context(), event); err != nil {
				slog.Warn("failed to trigger marker optimization on update", "error", err, "code", group.Code)
				// Don't fail the request if optimization trigger fails - the image is already uploaded
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
