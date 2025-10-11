package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

func addBFFRoutes(r chi.Router, db *sql.DB) {
	// token generation and validation
	r.Route("/v1/tokens", func(r chi.Router) {
		r.Post("/submission", generateSubmissionTokenHandler(db))
		r.Post("/validate", validateSubmissionTokenHandler(db))
	})

	// Group management
	r.Route("/v1/groups", func(r chi.Router) {
		r.Get("/validate/{code}", validateGroupCodeHandler(db))
		r.Post("/register", registerGroupHandler(db))
		r.Get("/edit/{token}", getGroupByEditTokenHandler(db))
		r.Put("/edit/{token}", updateGroupHandler(db))
		r.Post("/check-code", checkGroupCodeAvailabilityHandler(db))
	})

	// Ride Submission and editing (BFF Protected)
	r.Route("/v1/rides", func(r chi.Router) {
		r.Post("/submit", submitRideHandler(db))
		r.Get("/edit/{token}", getRideByEditTokenHandler(db))
		r.Put("/edit/{token}", updateRideHandler(db))
	})

}

// Token Generation & Validation

type TokenRequest struct {
	City string `json:"city"`
}

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// generateSubmissionTokenHandler creates a short lived token for form access
func generateSubmissionTokenHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req TokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.City == "" {
			http.Error(w, "City is required", http.StatusBadRequest)
			return
		}

		// generate the token used to verify the submission source
		token, err := generateSecureToken(32)
		if err != nil {
			slog.Error("Failed to generate token", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// allow organizers 30 mins to fill out ride details
		expiresAt := time.Now().Add(30 * time.Minute).Format(time.RFC3339)

		if err = storeSubmissionToken(db, token, req.City, expiresAt); err != nil {
			slog.Error("Failed to store token", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		slog.Info("Generated submission token", "city", req.City, "expiresAt", expiresAt)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			Token:     token,
			ExpiresAt: expiresAt,
		})

	}
}

type ValidateTokenRequest struct {
	Token string `json:"token"`
	City  string `json:"city"`
}

type ValidateTokenResponse struct {
	Valid bool   `json:"valid"`
	City  string `json:"city,omitempty"`
}

// validateSubmissionTokenHandler checks if a token is valid and not expired
func validateSubmissionTokenHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ValidateTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var city string
		var expiresAt string
		var used int

		err := db.QueryRow(`
			SELECT city, expires_at, used 
			FROM submission_tokens 
			WHERE token = ?
		`, req.Token).Scan(&city, &expiresAt, &used)

		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ValidateTokenResponse{Valid: false})
			return
		}

		if err != nil {
			slog.Error("Failed to query token", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Check if already used
		if used == 1 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ValidateTokenResponse{Valid: false})
			return
		}

		// Check if expired
		expires, err := time.Parse(time.RFC3339, expiresAt)
		if err != nil || time.Now().After(expires) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ValidateTokenResponse{Valid: false})
			return
		}

		// Check city matches if provided
		if req.City != "" && city != req.City {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ValidateTokenResponse{Valid: false})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ValidateTokenResponse{
			Valid: true,
			City:  city,
		})
	}
}

// ============================================================================
// GROUP REGISTRATION & MANAGEMENT
// ============================================================================

type GroupValidationResponse struct {
	Valid bool   `json:"valid"`
	Name  string `json:"name,omitempty"`
}

func validateGroupCodeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")

		// Validate format: 4 characters
		if len(code) != 4 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(GroupValidationResponse{Valid: false})
			return
		}

		var name string
		err := db.QueryRow(`
			SELECT name FROM ride_groups WHERE code = ? AND is_active = 1
		`, strings.ToUpper(code)).Scan(&name)

		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(GroupValidationResponse{Valid: false})
			return
		}

		if err != nil {
			slog.Error("Failed to validate group code", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GroupValidationResponse{
			Valid: true,
			Name:  name,
		})
	}
}

type GroupRegistration struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	City        string `json:"city"`
	IconURL     string `json:"icon_url"`
	WebURL      string `json:"web_url"`
}

type GroupResponse struct {
	Success   bool   `json:"success"`
	Code      string `json:"code,omitempty"`
	EditToken string `json:"edit_token,omitempty"`
	Message   string `json:"message,omitempty"`
}

// checkGroupCodeAvailabilityHandler checks if a 4-char code is available
func checkGroupCodeAvailabilityHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		code := strings.ToUpper(strings.TrimSpace(req.Code))

		// Validate format: exactly 4 alphanumeric characters
		if len(code) != 4 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"available": false,
				"message":   "Code must be exactly 4 characters",
			})
			return
		}

		var exists int
		err := db.QueryRow(`SELECT COUNT(*) FROM ride_groups WHERE code = ?`, code).Scan(&exists)
		if err != nil {
			slog.Error("Failed to check code availability", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"available": exists == 0,
			"code":      code,
		})
	}
}

func registerGroupHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate BFF token from header
		bffToken := r.Header.Get("X-BFF-Token")
		if bffToken == "" {
			http.Error(w, "Missing BFF token", http.StatusUnauthorized)
			return
		}

		// Validate origin
		origin := r.Header.Get("Origin")
		if !strings.HasSuffix(origin, "form.cyclescene.cc") {
			slog.Warn("Invalid origin for group registration", "origin", origin)
			http.Error(w, "Unauthorized origin", http.StatusForbidden)
			return
		}

		var registration GroupRegistration
		if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if registration.Code == "" || registration.Name == "" || registration.City == "" {
			http.Error(w, "Missing required fields (code, name, city)", http.StatusBadRequest)
			return
		}

		// Normalize code to uppercase
		registration.Code = strings.ToUpper(strings.TrimSpace(registration.Code))

		// Validate code format
		if len(registration.Code) != 4 {
			http.Error(w, "Code must be exactly 4 characters", http.StatusBadRequest)
			return
		}

		// Verify and mark token as used
		var tokenCity string
		var used int
		err := db.QueryRow(`
			SELECT city, used FROM submission_tokens WHERE token = ?
		`, bffToken).Scan(&tokenCity, &used)

		if err == sql.ErrNoRows || used == 1 {
			http.Error(w, "Invalid or already used token", http.StatusUnauthorized)
			return
		}

		if tokenCity != registration.City {
			http.Error(w, "Token city mismatch", http.StatusBadRequest)
			return
		}

		// Mark token as used
		_, err = db.Exec(`UPDATE submission_tokens SET used = 1 WHERE token = ?`, bffToken)
		if err != nil {
			slog.Error("Failed to mark token as used", "error", err)
		}

		// Generate edit token for the group
		editToken, err := generateSecureToken(32)
		if err != nil {
			slog.Error("Failed to generate edit token", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Insert group (add is_active and edit_token columns)
		_, err = db.Exec(`
			INSERT INTO ride_groups (code, name, description, city, icon_url, web_url, edit_token, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, 1)
		`, registration.Code, registration.Name, registration.Description, registration.City,
			registration.IconURL, registration.WebURL, editToken)

		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				http.Error(w, "Group code already exists", http.StatusConflict)
				return
			}
			slog.Error("Failed to insert group", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		slog.Info("Group registered successfully",
			"code", registration.Code,
			"name", registration.Name,
			"city", registration.City,
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GroupResponse{
			Success:   true,
			Code:      registration.Code,
			EditToken: editToken,
			Message:   "Group registered successfully",
		})
	}
}

func getGroupByEditTokenHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")

		var group GroupRegistration
		err := db.QueryRow(`
			SELECT code, name, description, city, icon_url, web_url
			FROM ride_groups WHERE edit_token = ?
		`, token).Scan(&group.Code, &group.Name, &group.Description, &group.City, &group.IconURL, &group.WebURL)

		if err == sql.ErrNoRows {
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}

		if err != nil {
			slog.Error("Failed to query group", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(group)
	}
}

func updateGroupHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")

		var registration GroupRegistration
		if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Update group (code cannot be changed)
		result, err := db.Exec(`
			UPDATE ride_groups SET
				name = ?, description = ?, icon_url = ?, web_url = ?
			WHERE edit_token = ?
		`, registration.Name, registration.Description, registration.IconURL, registration.WebURL, token)

		if err != nil {
			slog.Error("Failed to update group", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}

		slog.Info("Group updated successfully", "token", token)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GroupResponse{
			Success: true,
			Message: "Group updated successfully",
		})
	}
}

// ============================================================================
// RIDE SUBMISSION
// ============================================================================

type RideSubmission struct {
	// Core content
	Title       string `json:"title"`
	TinyTitle   string `json:"tinytitle"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Audience    string `json:"audience"`
	RideLength  string `json:"ride_length"`
	Area        string `json:"area"`
	DateType    string `json:"date_type"`

	// Location
	VenueName       string `json:"venue_name"`
	Address         string `json:"address"`
	LocationDetails string `json:"location_details"`
	EndingLocation  string `json:"ending_location"`
	IsLoopRide      bool   `json:"is_loop_ride"`

	// Contact
	OrganizerName   string `json:"organizer_name"`
	OrganizerEmail  string `json:"organizer_email"`
	OrganizerPhone  string `json:"organizer_phone"`
	WebURL          string `json:"web_url"`
	WebName         string `json:"web_name"`
	Newsflash       string `json:"newsflash"`
	HideEmail       bool   `json:"hide_email"`
	HidePhone       bool   `json:"hide_phone"`
	HideContactName bool   `json:"hide_contact_name"`

	// Group
	GroupCode string `json:"group_code"`

	// City
	City string `json:"city"`

	// Occurrences
	Occurrences []RideOccurrence `json:"occurrences"`
}

type RideOccurrence struct {
	StartDate            string `json:"start_date"`
	StartTime            string `json:"start_time"`
	EventDurationMinutes int    `json:"event_duration_minutes"`
	EventTimeDetails     string `json:"event_time_details"`
}

type SubmissionResponse struct {
	Success   bool   `json:"success"`
	EventID   int64  `json:"event_id,omitempty"`
	EditToken string `json:"edit_token,omitempty"`
	Message   string `json:"message,omitempty"`
}

func submitRideHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate BFF token from header
		bffToken := r.Header.Get("X-BFF-Token")
		if bffToken == "" {
			http.Error(w, "Missing BFF token", http.StatusUnauthorized)
			return
		}

		// Validate origin
		origin := r.Header.Get("Origin")
		if !strings.HasSuffix(origin, "form.cyclescene.cc") {
			slog.Warn("Invalid origin for ride submission", "origin", origin)
			http.Error(w, "Unauthorized origin", http.StatusForbidden)
			return
		}

		var submission RideSubmission
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

		// Verify and mark token as used
		var tokenCity string
		var used int
		err := db.QueryRow(`
			SELECT city, used FROM submission_tokens WHERE token = ?
		`, bffToken).Scan(&tokenCity, &used)

		if err == sql.ErrNoRows || used == 1 {
			http.Error(w, "Invalid or already used token", http.StatusUnauthorized)
			return
		}

		if tokenCity != submission.City {
			http.Error(w, "Token city mismatch", http.StatusBadRequest)
			return
		}

		// Mark token as used
		_, err = db.Exec(`UPDATE submission_tokens SET used = 1 WHERE token = ?`, bffToken)
		if err != nil {
			slog.Error("Failed to mark token as used", "error", err)
		}

		// Generate edit token for the ride
		editToken, err := generateSecureToken(32)
		if err != nil {
			slog.Error("Failed to generate edit token", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Begin transaction
		tx, err := db.Begin()
		if err != nil {
			slog.Error("Failed to begin transaction", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Insert event
		result, err := tx.Exec(`
			INSERT INTO events (
				title, tinytitle, description, image_url, audience, ride_length, area, date_type,
				venue_name, address, location_details, ending_location, is_loop_ride,
				organizer_name, organizer_email, organizer_phone, web_url, web_name, newsflash,
				hide_email, hide_phone, hide_contact_name, group_code, edit_token, city, is_published
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
		`,
			submission.Title, submission.TinyTitle, submission.Description, submission.ImageURL,
			submission.Audience, submission.RideLength, submission.Area, submission.DateType,
			submission.VenueName, submission.Address, submission.LocationDetails, submission.EndingLocation,
			boolToInt(submission.IsLoopRide), submission.OrganizerName, submission.OrganizerEmail,
			submission.OrganizerPhone, submission.WebURL, submission.WebName, submission.Newsflash,
			boolToInt(submission.HideEmail), boolToInt(submission.HidePhone), boolToInt(submission.HideContactName),
			nilIfEmpty(submission.GroupCode), editToken, submission.City,
		)

		if err != nil {
			slog.Error("Failed to insert event", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		eventID, err := result.LastInsertId()
		if err != nil {
			slog.Error("Failed to get event ID", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Insert occurrences
		for _, occ := range submission.Occurrences {
			startDatetime := fmt.Sprintf("%s %s", occ.StartDate, occ.StartTime)
			_, err = tx.Exec(`
				INSERT INTO event_occurrences (
					event_id, start_date, start_time, start_datetime, 
					event_duration_minutes, event_time_details
				) VALUES (?, ?, ?, ?, ?, ?)
			`, eventID, occ.StartDate, occ.StartTime, startDatetime, occ.EventDurationMinutes, occ.EventTimeDetails)

			if err != nil {
				slog.Error("Failed to insert occurrence", "error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			slog.Error("Failed to commit transaction", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		slog.Info("Ride submitted successfully",
			"event_id", eventID,
			"city", submission.City,
			"title", submission.Title,
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SubmissionResponse{
			Success:   true,
			EventID:   eventID,
			EditToken: editToken,
			Message:   "Ride submitted successfully and is pending review",
		})
	}
}

// ============================================================================
// RIDE EDITING
// ============================================================================

func getRideByEditTokenHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")

		// Query event by edit token
		row := db.QueryRow(`
			SELECT id, title, tinytitle, description, image_url, audience, ride_length, area, date_type,
				   venue_name, address, location_details, ending_location, is_loop_ride,
				   organizer_name, organizer_email, organizer_phone, web_url, web_name, newsflash,
				   hide_email, hide_phone, hide_contact_name, group_code, city, is_published
			FROM events WHERE edit_token = ?
		`, token)

		var event RideSubmission
		var id int64
		var isLoopRide, hideEmail, hidePhone, hideContactName, isPublished int
		var groupCode sql.NullString

		err := row.Scan(
			&id, &event.Title, &event.TinyTitle, &event.Description, &event.ImageURL,
			&event.Audience, &event.RideLength, &event.Area, &event.DateType,
			&event.VenueName, &event.Address, &event.LocationDetails, &event.EndingLocation, &isLoopRide,
			&event.OrganizerName, &event.OrganizerEmail, &event.OrganizerPhone,
			&event.WebURL, &event.WebName, &event.Newsflash,
			&hideEmail, &hidePhone, &hideContactName, &groupCode, &event.City, &isPublished,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "Ride not found", http.StatusNotFound)
			return
		}

		if err != nil {
			slog.Error("Failed to query ride", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		event.IsLoopRide = isLoopRide == 1
		event.HideEmail = hideEmail == 1
		event.HidePhone = hidePhone == 1
		event.HideContactName = hideContactName == 1
		if groupCode.Valid {
			event.GroupCode = groupCode.String
		}

		// Get occurrences
		rows, err := db.Query(`
			SELECT start_date, start_time, event_duration_minutes, event_time_details
			FROM event_occurrences WHERE event_id = ?
			ORDER BY start_datetime ASC
		`, id)

		if err != nil {
			slog.Error("Failed to query occurrences", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var occ RideOccurrence
			if err := rows.Scan(&occ.StartDate, &occ.StartTime, &occ.EventDurationMinutes, &occ.EventTimeDetails); err != nil {
				slog.Error("Failed to scan occurrence", "error", err)
				continue
			}
			event.Occurrences = append(event.Occurrences, occ)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Event       RideSubmission `json:"event"`
			IsPublished bool           `json:"is_published"`
		}{
			Event:       event,
			IsPublished: isPublished == 1,
		})
	}
}

func updateRideHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")

		var submission RideSubmission
		if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Begin transaction
		tx, err := db.Begin()
		if err != nil {
			slog.Error("Failed to begin transaction", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Update event
		result, err := tx.Exec(`
			UPDATE events SET
				title = ?, tinytitle = ?, description = ?, image_url = ?,
				audience = ?, ride_length = ?, area = ?, date_type = ?,
				venue_name = ?, address = ?, location_details = ?, ending_location = ?, is_loop_ride = ?,
				organizer_name = ?, organizer_email = ?, organizer_phone = ?,
				web_url = ?, web_name = ?, newsflash = ?,
				hide_email = ?, hide_phone = ?, hide_contact_name = ?,
				group_code = ?, updated_at = STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')
			WHERE edit_token = ?
		`,
			submission.Title, submission.TinyTitle, submission.Description, submission.ImageURL,
			submission.Audience, submission.RideLength, submission.Area, submission.DateType,
			submission.VenueName, submission.Address, submission.LocationDetails, submission.EndingLocation,
			boolToInt(submission.IsLoopRide), submission.OrganizerName, submission.OrganizerEmail,
			submission.OrganizerPhone, submission.WebURL, submission.WebName, submission.Newsflash,
			boolToInt(submission.HideEmail), boolToInt(submission.HidePhone), boolToInt(submission.HideContactName),
			nilIfEmpty(submission.GroupCode), token,
		)

		if err != nil {
			slog.Error("Failed to update event", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Ride not found", http.StatusNotFound)
			return
		}

		// Get event ID
		var eventID int64
		err = tx.QueryRow(`SELECT id FROM events WHERE edit_token = ?`, token).Scan(&eventID)
		if err != nil {
			slog.Error("Failed to get event ID", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Delete existing occurrences and re-insert
		_, err = tx.Exec(`DELETE FROM event_occurrences WHERE event_id = ?`, eventID)
		if err != nil {
			slog.Error("Failed to delete occurrences", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Insert new occurrences
		for _, occ := range submission.Occurrences {
			startDatetime := fmt.Sprintf("%s %s", occ.StartDate, occ.StartTime)
			_, err = tx.Exec(`
				INSERT INTO event_occurrences (
					event_id, start_date, start_time, start_datetime,
					event_duration_minutes, event_time_details
				) VALUES (?, ?, ?, ?, ?, ?)
			`, eventID, occ.StartDate, occ.StartTime, startDatetime, occ.EventDurationMinutes, occ.EventTimeDetails)

			if err != nil {
				slog.Error("Failed to insert occurrence", "error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			slog.Error("Failed to commit transaction", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		slog.Info("Ride updated successfully", "event_id", eventID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SubmissionResponse{
			Success: true,
			Message: "Ride updated successfully",
		})
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
