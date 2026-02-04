package strava

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/spacesedan/cyclescene/backend/internal/api/magiclink"
	"github.com/spacesedan/cyclescene/backend/internal/api/ride"
	"github.com/spacesedan/cyclescene/backend/internal/config"
	"github.com/spacesedan/cyclescene/backend/internal/strava"
)

// ImportHandler handles WebSocket connections for Strava event imports
type ImportHandler struct {
	stravaService  *strava.Service
	rideService    *ride.Service
	magicLinkSvc   *magiclink.Service
	editLinkBase   string
	appAccessToken string
	debug          bool

	// Track active imports to prevent concurrent imports per user
	activeImports sync.Map // map[int64]bool - athleteID -> importing
}

// NewImportHandler creates a new WebSocket import handler
func NewImportHandler(
	stravaService *strava.Service,
	rideService *ride.Service,
	magicLinkSvc *magiclink.Service,
	editLinkBase string,
) *ImportHandler {
	return &ImportHandler{
		stravaService:  stravaService,
		rideService:    rideService,
		magicLinkSvc:   magicLinkSvc,
		editLinkBase:   editLinkBase,
		appAccessToken: os.Getenv("STRAVA_ACCESS_TOKEN"),
		debug:          os.Getenv("STRAVA_DEBUG") == "true",
	}
}

// ImportRequest is the initial message from client to start import
type ImportRequest struct {
	SessionID      string              `json:"session_id"`
	OrganizerEmail string              `json:"organizer_email"`
	GroupCode      string              `json:"group_code,omitempty"` // Optional 4-char group code to associate rides with
	Events         []EventImportConfig `json:"events"`
}

// EventImportConfig specifies which event to import and any overrides
// Note: StravaEventID uses json:",string" to avoid JavaScript precision loss with large int64 values
// ClubID is smaller and doesn't need string encoding
type EventImportConfig struct {
	StravaEventID int64             `json:"strava_event_id,string"`
	ClubID        int64             `json:"club_id"`
	Overrides     map[string]string `json:"overrides,omitempty"` // e.g., {"audience": "All", "image_url": "..."}
}

// ProgressMessage is sent to the client during import
type ProgressMessage struct {
	Type          string `json:"type"`                              // "heartbeat", "progress", "complete", "done", "error"
	EventIndex    int    `json:"event_index,omitempty"`             // 0-indexed
	TotalEvents   int    `json:"total_events,omitempty"`            // Total number of events being imported
	StravaEventID int64  `json:"strava_event_id,string,omitempty"`  // String to avoid JS precision loss
	EventTitle    string `json:"event_title,omitempty"`
	Step          string `json:"step,omitempty"`                    // "fetching", "coordinates", "route", "database"
	Status        string `json:"status,omitempty"`                  // "in_progress", "success", "error"
	Message       string `json:"message,omitempty"`

	// Complete message fields
	CycleSceneEventID int64  `json:"cyclescene_event_id,omitempty"`
	EditToken         string `json:"edit_token,omitempty"`
	EditURL           string `json:"edit_url,omitempty"`
	Success           bool   `json:"success,omitempty"`
	Error             string `json:"error,omitempty"`

	// Done message fields (no omitempty so 0 values are included)
	TotalImported    int            `json:"total_imported,omitempty"`
	TotalFailed      int            `json:"total_failed,omitempty"`
	SummaryEmailSent bool           `json:"summary_email_sent,omitempty"`
	Results          []ImportResult `json:"results,omitempty"`
}

// ImportResult is the result of importing a single event
type ImportResult struct {
	StravaEventID     int64  `json:"strava_event_id,string"` // String to avoid JS precision loss
	CycleSceneEventID int64  `json:"cyclescene_event_id,omitempty"`
	EditToken         string `json:"edit_token,omitempty"`
	EditURL           string `json:"edit_url,omitempty"`
	Success           bool   `json:"success"`
	Error             string `json:"error,omitempty"`
	Title             string `json:"title,omitempty"`
}

// HandleImport handles the WebSocket connection for importing events
// WS /strava/import
func (h *ImportHandler) HandleImport(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from cookie before upgrading connection
	// This allows HttpOnly cookies to work (frontend can't read them directly)
	cookieSessionID := ""
	if cookie, err := r.Cookie("strava_session_id"); err == nil {
		cookieSessionID = cookie.Value
	}

	// Accept WebSocket connection with origin validation
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: config.GetWebSocketOrigins(),
	})
	if err != nil {
		slog.Error("Failed to accept WebSocket connection", "error", err)
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute) // Max import time
	defer cancel()

	if h.debug {
		slog.Debug("WebSocket connection accepted for import",
			"has_cookie_session", cookieSessionID != "",
		)
	}

	// Read the initial import request
	var req ImportRequest
	if err := wsjson.Read(ctx, conn, &req); err != nil {
		h.sendError(ctx, conn, "Failed to read import request: "+err.Error())
		return
	}

	// Use cookie session ID if not provided in message (cookie is HttpOnly)
	if req.SessionID == "" {
		req.SessionID = cookieSessionID
	}

	// Validate request
	if req.SessionID == "" {
		h.sendError(ctx, conn, "Missing session - please reconnect to Strava")
		return
	}
	if req.OrganizerEmail == "" {
		h.sendError(ctx, conn, "Missing organizer_email")
		return
	}
	if len(req.Events) == 0 {
		h.sendError(ctx, conn, "No events to import")
		return
	}

	// Get session to validate and get athlete ID
	session, ok := h.stravaService.GetSession(req.SessionID)
	if !ok {
		h.sendError(ctx, conn, "Invalid or expired session - please reconnect to Strava")
		return
	}

	// Check for concurrent imports
	if _, loaded := h.activeImports.LoadOrStore(session.AthleteID, true); loaded {
		h.sendError(ctx, conn, "Import already in progress - please wait for the current import to complete")
		return
	}
	defer h.activeImports.Delete(session.AthleteID)

	if h.debug {
		slog.Debug("Starting import",
			"athlete_id", session.AthleteID,
			"email", req.OrganizerEmail,
			"event_count", len(req.Events),
		)
	}

	// Start heartbeat goroutine
	heartbeatCtx, heartbeatCancel := context.WithCancel(ctx)
	defer heartbeatCancel()
	go h.sendHeartbeats(heartbeatCtx, conn)

	// Process imports
	results := h.processImports(ctx, conn, session, req)

	// Stop heartbeats
	heartbeatCancel()

	// Send summary email if any events were imported successfully
	var summaryEmailSent bool
	successfulEvents := make([]magiclink.ImportedEvent, 0)
	for _, result := range results {
		if result.Success {
			successfulEvents = append(successfulEvents, magiclink.ImportedEvent{
				Title:     result.Title,
				EventID:   result.CycleSceneEventID,
				EditToken: result.EditToken,
				EditURL:   result.EditURL,
			})
		}
	}

	if len(successfulEvents) > 0 && h.magicLinkSvc != nil {
		_, err := h.magicLinkSvc.SendImportSummaryEmail(ctx, req.OrganizerEmail, successfulEvents)
		if err != nil {
			slog.Error("Failed to send import summary email",
				"error", err,
				"email", req.OrganizerEmail,
				"event_count", len(successfulEvents),
			)
		} else {
			summaryEmailSent = true
			if h.debug {
				slog.Debug("Import summary email sent",
					"email", req.OrganizerEmail,
					"event_count", len(successfulEvents),
				)
			}
		}
	}

	// Count successes and failures
	totalImported := 0
	totalFailed := 0
	for _, result := range results {
		if result.Success {
			totalImported++
		} else {
			totalFailed++
		}
	}

	// Send final done message
	doneMsg := ProgressMessage{
		Type:             "done",
		TotalImported:    totalImported,
		TotalFailed:      totalFailed,
		SummaryEmailSent: summaryEmailSent,
		Results:          results,
	}
	wsjson.Write(ctx, conn, doneMsg)

	// Log import metrics (always log at Info level for monitoring)
	slog.Info("strava_import_completed",
		"event", "strava_import_completed",
		"athlete_id", session.AthleteID,
		"total_imported", totalImported,
		"total_failed", totalFailed,
		"total_requested", len(req.Events),
		"summary_email_sent", summaryEmailSent,
		"organizer_email", req.OrganizerEmail,
	)

	if h.debug {
		slog.Debug("Import completed",
			"athlete_id", session.AthleteID,
			"total_imported", totalImported,
			"total_failed", totalFailed,
			"summary_email_sent", summaryEmailSent,
		)
	}

	// Close connection gracefully
	conn.Close(websocket.StatusNormalClosure, "Import completed")
}

// processImports imports each event sequentially
func (h *ImportHandler) processImports(
	ctx context.Context,
	conn *websocket.Conn,
	session *strava.Session,
	req ImportRequest,
) []ImportResult {
	results := make([]ImportResult, 0, len(req.Events))
	totalEvents := len(req.Events)

	for i, eventConfig := range req.Events {
		result := h.importSingleEvent(ctx, conn, session, req.OrganizerEmail, req.GroupCode, eventConfig, i, totalEvents)
		results = append(results, result)

		// Check if context was cancelled (client disconnected, timeout, etc.)
		if ctx.Err() != nil {
			slog.Warn("Import context cancelled", "error", ctx.Err())
			break
		}
	}

	return results
}

// importSingleEvent imports a single event and sends progress updates
func (h *ImportHandler) importSingleEvent(
	ctx context.Context,
	conn *websocket.Conn,
	session *strava.Session,
	organizerEmail string,
	groupCode string,
	eventConfig EventImportConfig,
	index int,
	total int,
) ImportResult {
	result := ImportResult{
		StravaEventID: eventConfig.StravaEventID,
	}

	// Step 1: Fetch event data
	h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, "", "fetching", "in_progress", "Fetching event data from Strava...")

	events, err := h.stravaService.GetClubEvents(ctx, session.SessionID, eventConfig.ClubID)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to fetch club events: %v", err)
		h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, "", "fetching", "error", result.Error)
		return result
	}

	// Find the specific event
	var targetEvent *strava.GroupEvent
	for i := range events {
		if events[i].ID == eventConfig.StravaEventID {
			targetEvent = &events[i]
			break
		}
	}

	if targetEvent == nil {
		result.Error = "Event not found in Strava - it may have been deleted"
		h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, "", "fetching", "error", result.Error)
		return result
	}

	result.Title = targetEvent.Title
	h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, targetEvent.Title, "fetching", "success", "Event data fetched")

	// Step 2: Get coordinates
	h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, targetEvent.Title, "coordinates", "in_progress", "Processing location...")

	submission, lat, lng, err := h.stravaService.ConvertEventToSubmission(ctx, session.SessionID, targetEvent, organizerEmail)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to convert event: %v", err)
		h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, targetEvent.Title, "coordinates", "error", result.Error)
		return result
	}

	// Apply overrides
	h.applyOverrides(submission, eventConfig.Overrides)

	// Set group code if provided (from import request level)
	if groupCode != "" {
		submission.GroupCode = groupCode
	}

	// Set source tracking for deduplication
	// Include occurrence date in SourceID to allow importing future occurrences of recurring events
	// Format: {strava_event_id}_{YYYY-MM-DD} e.g., "3453605542995245000_2026-02-24"
	// This enables series tracking via: WHERE source_id LIKE '{strava_event_id}_%'
	submission.Source = "strava"
	occurrenceDate := ""
	if len(targetEvent.UpcomingOccurrences) > 0 {
		// Extract date portion from ISO 8601 timestamp (first 10 chars: "2026-02-20")
		occurrenceDate = targetEvent.UpcomingOccurrences[0][:10]
	}
	submission.SourceID = fmt.Sprintf("%d_%s", eventConfig.StravaEventID, occurrenceDate)

	h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, targetEvent.Title, "coordinates", "success", "Location processed")

	// Step 3: Process route (optional, non-blocking)
	if targetEvent.RouteID != nil && h.appAccessToken != "" {
		h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, targetEvent.Title, "route", "in_progress", "Processing route...")

		routeID, err := h.stravaService.ProcessRouteWithToken(ctx, h.appAccessToken, *targetEvent.RouteID, session.CityCode, session.AthleteID)
		if err != nil {
			if h.debug {
				slog.Debug("Route processing failed (continuing without route)",
					"error", err,
					"strava_route_id", *targetEvent.RouteID,
				)
			}
			h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, targetEvent.Title, "route", "success", "Route unavailable - continuing without route")
		} else if routeID != nil {
			// Set route URL for linking
			submission.RouteURL = fmt.Sprintf("https://www.strava.com/routes/%d", *targetEvent.RouteID)
			h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, targetEvent.Title, "route", "success", "Route processed")
		}
	} else {
		h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, targetEvent.Title, "route", "success", "No route to process")
	}

	// Step 4: Save to database
	h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, targetEvent.Title, "database", "in_progress", "Saving event to CycleScene...")

	response, err := h.rideService.SubmitRideWithCoordinates(ctx, submission, lat, lng)
	if err != nil {
		// Check for duplicate
		if isDuplicateError(err) {
			result.Error = "This event has already been imported"
		} else {
			result.Error = fmt.Sprintf("Failed to save event: %v", err)
		}
		h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, targetEvent.Title, "database", "error", result.Error)
		return result
	}

	h.sendProgress(ctx, conn, index, total, eventConfig.StravaEventID, targetEvent.Title, "database", "success", "Event saved successfully")

	// Build result
	result.CycleSceneEventID = response.EventID
	result.EditToken = response.EditToken
	result.EditURL = fmt.Sprintf("%s?token=%s", h.editLinkBase, response.EditToken)
	result.Success = true

	// Send complete message for this event
	completeMsg := ProgressMessage{
		Type:              "complete",
		EventIndex:        index,
		StravaEventID:     eventConfig.StravaEventID,
		CycleSceneEventID: response.EventID,
		EditToken:         response.EditToken,
		EditURL:           result.EditURL,
		Success:           true,
	}
	wsjson.Write(ctx, conn, completeMsg)

	return result
}

// applyOverrides applies user-specified overrides to the submission
func (h *ImportHandler) applyOverrides(submission *ride.Submission, overrides map[string]string) {
	if overrides == nil {
		return
	}

	if audience, ok := overrides["audience"]; ok {
		submission.Audience = audience
	}
	if imageURL, ok := overrides["image_url"]; ok {
		submission.ImageURL = imageURL
	}
	if rideLength, ok := overrides["ride_length"]; ok {
		submission.RideLength = rideLength
	}
	if area, ok := overrides["area"]; ok {
		submission.Area = area
	}
}

// sendProgress sends a progress message to the client
func (h *ImportHandler) sendProgress(
	ctx context.Context,
	conn *websocket.Conn,
	index int,
	total int,
	stravaEventID int64,
	eventTitle string,
	step string,
	status string,
	message string,
) {
	msg := ProgressMessage{
		Type:          "progress",
		EventIndex:    index,
		TotalEvents:   total,
		StravaEventID: stravaEventID,
		EventTitle:    eventTitle,
		Step:          step,
		Status:        status,
		Message:       message,
	}
	wsjson.Write(ctx, conn, msg)
}

// sendError sends an error message to the client
func (h *ImportHandler) sendError(ctx context.Context, conn *websocket.Conn, message string) {
	msg := ProgressMessage{
		Type:    "error",
		Message: message,
	}
	wsjson.Write(ctx, conn, msg)
	conn.Close(websocket.StatusInternalError, message)
}

// sendHeartbeats sends periodic heartbeat messages to keep the connection alive
func (h *ImportHandler) sendHeartbeats(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msg := ProgressMessage{
				Type:    "heartbeat",
				Message: "Import still in progress...",
			}
			if err := wsjson.Write(ctx, conn, msg); err != nil {
				if h.debug {
					slog.Debug("Failed to send heartbeat", "error", err)
				}
				return
			}
		}
	}
}

// isDuplicateError checks if the error is a duplicate constraint violation
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "UNIQUE constraint failed") ||
		contains(errStr, "duplicate key") ||
		contains(errStr, "idx_events_source_dedup")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
