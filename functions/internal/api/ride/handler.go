package ride

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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
	r.Route("/rides", func(r chi.Router) {
		// User-submitted rides
		r.Post("/submit", h.SubmitRide)
		r.Get("/edit/{token}", h.GetRideByEditToken)
		r.Put("/edit/{token}", h.UpdateRide)
		r.Patch("/edit/{token}/occurrences/{occurrenceId}", h.UpdateOccurrence)

		// Scraped rides from Shift2Bikes
		r.Get("/upcoming", h.GetUpcomingRides)
		r.Get("/past", h.GetPastRides)
		r.Get("/ics", h.GenerateICS)
	})
}

// ============================================================================
// USER-SUBMITTED RIDES
// ============================================================================

func (h *Handler) SubmitRide(w http.ResponseWriter, r *http.Request) {
	// Validate BFF token
	bffToken := r.Header.Get("X-BFF-Token")
	if bffToken == "" {
		http.Error(w, "Missing BFF token", http.StatusUnauthorized)
		return
	}

	var submission Submission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if submission.Title == "" || submission.Description == "" || submission.City == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if len(submission.Occurrences) == 0 {
		http.Error(w, "At least one occurrence is required", http.StatusBadRequest)
		return
	}

	response, err := h.service.SubmitRide(&submission)
	if err != nil {
		slog.Error("Failed to submit ride", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("Ride submitted successfully",
		"event_id", response.EventID,
		"city", submission.City,
		"title", submission.Title,
	)

	// Trigger image optimization if image_uuid is provided
	if submission.ImageUUID != "" && h.eventarcClient != nil {
		event := &events.ImageOptimizationEvent{
			ImageUUID:  submission.ImageUUID,
			CityCode:   submission.City,
			EntityID:   fmt.Sprintf("%d", response.EventID),
			EntityType: "ride",
		}
		if err := h.eventarcClient.TriggerOptimization(r.Context(), event); err != nil {
			slog.Warn("failed to trigger image optimization", "error", err, "event_id", response.EventID)
			// Don't fail the request if optimization trigger fails - the image is already uploaded
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetRideByEditToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	response, err := h.service.GetRideByEditToken(token)
	if err == sql.ErrNoRows {
		http.Error(w, "Ride not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("Failed to get ride", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) UpdateRide(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	var submission Submission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.service.UpdateRide(token, &submission)
	if err == sql.ErrNoRows {
		http.Error(w, "Ride not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("Failed to update ride", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("Ride updated successfully", "token", token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type UpdateOccurrenceRequest struct {
	StartTime            string `json:"start_time"`
	EventDurationMinutes int    `json:"event_duration_minutes"`
	EventTimeDetails     string `json:"event_time_details"`
	IsCancelled          bool   `json:"is_cancelled"`
}

func (h *Handler) UpdateOccurrence(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	occurrenceIdStr := chi.URLParam(r, "occurrenceId")

	occurrenceId, err := strconv.ParseInt(occurrenceIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid occurrence ID", http.StatusBadRequest)
		return
	}

	var req UpdateOccurrenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateOccurrence(token, occurrenceId, req.StartTime, req.EventDurationMinutes, req.EventTimeDetails, req.IsCancelled); err != nil {
		slog.Error("Failed to update occurrence", "error", err, "token", token, "occurrence_id", occurrenceId)
		http.Error(w, "Failed to update occurrence", http.StatusInternalServerError)
		return
	}

	slog.Info("Occurrence updated successfully", "token", token, "occurrence_id", occurrenceId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// ============================================================================
// SCRAPED RIDES FROM SHIFT2BIKES
// ============================================================================

func (h *Handler) GetUpcomingRides(w http.ResponseWriter, r *http.Request) {
	cityCode := r.URL.Query().Get("city")
	if cityCode == "" {
		cityCode = "pdx"
	}

	rides, err := h.service.GetUpcomingRides(cityCode)
	if err != nil {
		slog.Error("Failed to get upcoming rides", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.Info("Retrieved upcoming rides", "city", cityCode, "count", len(rides))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rides); err != nil {
		slog.Error("Failed to encode rides to JSON", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetPastRides(w http.ResponseWriter, r *http.Request) {
	cityCode := r.URL.Query().Get("city")
	if cityCode == "" {
		cityCode = "pdx"
	}

	rides, err := h.service.GetPastRides(cityCode)
	if err != nil {
		slog.Error("Failed to get past rides", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.Info("Retrieved past rides", "city", cityCode, "count", len(rides))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rides); err != nil {
		slog.Error("Failed to encode rides to JSON", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GenerateICS(w http.ResponseWriter, r *http.Request) {
	rideID := r.URL.Query().Get("id")
	cityCode := r.URL.Query().Get("city")
	if cityCode == "" {
		cityCode = "pdx"
	}

	icsData, err := h.service.GenerateICSFromRide(cityCode, rideID)
	if err != nil {
		slog.Error("Failed to generate ICS", "error", err, "city", cityCode, "rideID", rideID)
		http.Error(w, "Failed to generate ICS data", http.StatusInternalServerError)
		return
	}

	slog.Info("Generated ICS file", "city", cityCode, "rideID", rideID, "filename", icsData.Filename)

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.ics\"", icsData.Filename))
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(icsData.Content))
}
