package strava

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/spacesedan/cyclescene/backend/internal/api/magiclink"
	"github.com/spacesedan/cyclescene/backend/internal/api/ride"
	"github.com/spacesedan/cyclescene/backend/internal/config"
	"github.com/spacesedan/cyclescene/backend/internal/strava"
)

// GroupEventResponse is a frontend-safe version of GroupEvent with ID as string
// This avoids JavaScript precision loss with large int64 event IDs
type GroupEventResponse struct {
	ID                  string    `json:"id"` // String to avoid JS precision loss
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	ActivityType        string    `json:"activity_type"`
	UpcomingOccurrences []string  `json:"upcoming_occurrences"`
	Zone                string    `json:"zone"`
	Address             string    `json:"address"`
	StartLatLng         []float64 `json:"start_latlng"`
	RouteID             *int64    `json:"route_id"`
	Route               any       `json:"route"`
	SkillLevels         *string   `json:"skill_levels"`
	Terrain             *string   `json:"terrain"`
	WomenOnly           bool      `json:"women_only"`
	Private             bool      `json:"private"`
	Joined              bool      `json:"joined"`
	ClubID              int64     `json:"club_id"`
	Club                any       `json:"club,omitempty"`
	OrganizingAthlete   any       `json:"organizing_athlete,omitempty"`
}

// toGroupEventResponse converts a GroupEvent to frontend-safe response
func toGroupEventResponse(e strava.GroupEvent) GroupEventResponse {
	return GroupEventResponse{
		ID:                  strconv.FormatInt(e.ID, 10),
		Title:               e.Title,
		Description:         e.Description,
		ActivityType:        e.ActivityType,
		UpcomingOccurrences: e.UpcomingOccurrences,
		Zone:                e.Zone,
		Address:             e.Address,
		StartLatLng:         e.StartLatLng,
		RouteID:             e.RouteID,
		Route:               e.Route,
		SkillLevels:         e.SkillLevels,
		Terrain:             e.Terrain,
		WomenOnly:           e.WomenOnly,
		Private:             e.Private,
		Joined:              e.Joined,
		ClubID:              e.ClubID,
		Club:                e.Club,
		OrganizingAthlete:   e.OrganizingAthlete,
	}
}

// Handler handles Strava OAuth and import HTTP requests
type Handler struct {
	stravaService *strava.Service
	importHandler *ImportHandler
	formURL       string // Frontend form app URL for OAuth redirects
	debug         bool
}

// NewHandler creates a new Strava HTTP handler (without import support)
func NewHandler(stravaService *strava.Service) *Handler {
	return &Handler{
		stravaService: stravaService,
		formURL:       getFormURL(),
		debug:         os.Getenv("STRAVA_DEBUG") == "true",
	}
}

// NewHandlerWithImport creates a new Strava HTTP handler with import support
func NewHandlerWithImport(
	stravaService *strava.Service,
	rideService *ride.Service,
	magicLinkSvc *magiclink.Service,
	editLinkBase string,
) *Handler {
	return &Handler{
		stravaService: stravaService,
		importHandler: NewImportHandler(stravaService, rideService, magicLinkSvc, editLinkBase),
		formURL:       getFormURL(),
		debug:         os.Getenv("STRAVA_DEBUG") == "true",
	}
}

// getFormURL returns the frontend form app URL from environment
func getFormURL() string {
	url := os.Getenv("FORM_URL")
	if url == "" {
		// Default to localhost for development
		url = "http://localhost:5173"
	}
	return url
}

// RegisterRoutes registers the Strava routes with the chi router
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/strava", func(r chi.Router) {
		// OAuth flow
		r.Get("/auth/initiate", h.InitiateOAuth)
		r.Get("/auth/callback", h.HandleOAuthCallback)

		// Session management
		r.Post("/logout", h.Logout)

		// Club and event endpoints (require session)
		r.Get("/admin-clubs", h.GetAdminClubs)
		r.Get("/clubs/{clubId}/events", h.GetClubEvents)

		// WebSocket import endpoint (if import handler is configured)
		if h.importHandler != nil {
			r.Get("/import", h.importHandler.HandleImport)
		}
	})
}

// InitiateOAuth starts the OAuth flow by redirecting to Strava
// GET /strava/auth/initiate?city=pdx
func (h *Handler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
	cityCode := r.URL.Query().Get("city")
	if cityCode == "" {
		cityCode = "pdx" // Default to Portland
	}

	if h.debug {
		slog.Debug("OAuth initiate request",
			"city_code", cityCode,
			"remote_addr", r.RemoteAddr,
		)
	}

	authURL, err := h.stravaService.InitiateOAuth(r.Context(), cityCode)
	if err != nil {
		slog.Error("Failed to initiate OAuth", "error", err, "city_code", cityCode)
		http.Error(w, "Failed to start OAuth flow", http.StatusInternalServerError)
		return
	}

	// Redirect to Strava authorization
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleOAuthCallback handles the OAuth callback from Strava
// GET /strava/auth/callback?code=xxx&state=xxx
func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// Handle OAuth errors (user denied access, etc.)
	if errorParam != "" {
		slog.Warn("OAuth error from Strava", "error", errorParam)
		// Redirect to frontend callback with error
		redirectURL := h.formURL + "/strava/callback?error=" + errorParam
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	if code == "" || state == "" {
		http.Error(w, "Missing code or state parameter", http.StatusBadRequest)
		return
	}

	if h.debug {
		slog.Debug("OAuth callback received",
			"state", state[:8]+"...",
			"code_length", len(code),
		)
	}

	sessionID, err := h.stravaService.HandleOAuthCallback(r.Context(), code, state)
	if err != nil {
		slog.Error("OAuth callback failed", "error", err)
		// Redirect to frontend callback with error
		redirectURL := h.formURL + "/strava/callback?error=callback_failed"
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Set session cookie using centralized config
	// Dev: HttpOnly=false so frontend can read session for WebSocket auth
	// Prod: HttpOnly=true, Secure=true for security
	http.SetCookie(w, config.NewSessionCookie("strava_session_id", sessionID, 3600))

	if h.debug {
		slog.Debug("OAuth callback successful, session created",
			"session_id", sessionID[:8]+"...",
		)
	}

	// Redirect to frontend callback page (for popup to close)
	// The frontend callback page will send postMessage to parent and close
	redirectURL := h.formURL + "/strava/callback"

	if h.debug {
		slog.Debug("Redirecting to frontend callback",
			"redirect_url", redirectURL,
		)
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// Logout clears the Strava session
// POST /strava/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID := h.getSessionIDFromCookie(r)
	if sessionID != "" {
		h.stravaService.DeleteSession(sessionID)

		if h.debug {
			slog.Debug("Session deleted", "session_id", sessionID[:8]+"...")
		}
	}

	// Clear the session cookie
	h.clearSessionCookie(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// clearSessionCookie clears the Strava session cookie with consistent attributes
// Uses centralized config to ensure attributes match those used when setting
func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, config.ClearSessionCookie("strava_session_id"))
}

// GetAdminClubs returns clubs where the user is an admin or owner
// GET /strava/admin-clubs
func (h *Handler) GetAdminClubs(w http.ResponseWriter, r *http.Request) {
	sessionID := h.getSessionIDFromCookie(r)
	if sessionID == "" {
		http.Error(w, "Unauthorized - no session", http.StatusUnauthorized)
		return
	}

	if h.debug {
		slog.Debug("Fetching admin clubs", "session_id", sessionID[:8]+"...")
	}

	clubs, err := h.stravaService.GetAdminClubs(r.Context(), sessionID)
	if err != nil {
		if err == strava.ErrUnauthorized {
			// Clear invalid session cookie (must match attributes used when setting)
			h.clearSessionCookie(w)
			http.Error(w, "Session expired - please reconnect to Strava", http.StatusUnauthorized)
			return
		}
		slog.Error("Failed to fetch admin clubs", "error", err)
		http.Error(w, "Failed to fetch clubs", http.StatusInternalServerError)
		return
	}

	if h.debug {
		slog.Debug("Admin clubs fetched", "count", len(clubs))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"clubs": clubs})
}

// GetClubEvents returns events for a specific club
// GET /strava/clubs/{clubId}/events
func (h *Handler) GetClubEvents(w http.ResponseWriter, r *http.Request) {
	sessionID := h.getSessionIDFromCookie(r)
	if sessionID == "" {
		http.Error(w, "Unauthorized - no session", http.StatusUnauthorized)
		return
	}

	clubIDStr := chi.URLParam(r, "clubId")
	if clubIDStr == "" {
		http.Error(w, "Missing club ID", http.StatusBadRequest)
		return
	}

	var clubID int64
	if _, err := json.Number(clubIDStr).Int64(); err != nil {
		http.Error(w, "Invalid club ID", http.StatusBadRequest)
		return
	}
	clubID, _ = json.Number(clubIDStr).Int64()

	if h.debug {
		slog.Debug("Fetching club events",
			"session_id", sessionID[:8]+"...",
			"club_id", clubID,
		)
	}

	events, err := h.stravaService.GetClubEvents(r.Context(), sessionID, clubID)
	if err != nil {
		if err == strava.ErrUnauthorized {
			// Clear invalid session cookie (must match attributes used when setting)
			h.clearSessionCookie(w)
			http.Error(w, "Session expired - please reconnect to Strava", http.StatusUnauthorized)
			return
		}
		slog.Error("Failed to fetch club events", "error", err, "club_id", clubID)
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	if h.debug {
		slog.Debug("Club events fetched", "club_id", clubID, "count", len(events))
	}

	// Convert to frontend-safe response with string IDs
	responseEvents := make([]GroupEventResponse, len(events))
	for i, e := range events {
		responseEvents[i] = toGroupEventResponse(e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"events": responseEvents})
}

// getSessionIDFromCookie extracts the session ID from the cookie
func (h *Handler) getSessionIDFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("strava_session_id")
	if err != nil {
		return ""
	}
	return cookie.Value
}
