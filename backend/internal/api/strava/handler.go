package strava

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/spacesedan/cyclescene/backend/internal/api/magiclink"
	"github.com/spacesedan/cyclescene/backend/internal/api/ride"
	"github.com/spacesedan/cyclescene/backend/internal/strava"
)

// Handler handles Strava OAuth and import HTTP requests
type Handler struct {
	stravaService *strava.Service
	importHandler *ImportHandler
	debug         bool
}

// NewHandler creates a new Strava HTTP handler (without import support)
func NewHandler(stravaService *strava.Service) *Handler {
	return &Handler{
		stravaService: stravaService,
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
		debug:         os.Getenv("STRAVA_DEBUG") == "true",
	}
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
		// Redirect to frontend with error
		http.Redirect(w, r, "/strava/error?error="+errorParam, http.StatusTemporaryRedirect)
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
		// Redirect to frontend with error
		http.Redirect(w, r, "/strava/error?error=callback_failed", http.StatusTemporaryRedirect)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "strava_session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   3600, // 1 hour
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") != "dev", // Secure in production
		SameSite: http.SameSiteLaxMode,
	})

	if h.debug {
		slog.Debug("OAuth callback successful, session created",
			"session_id", sessionID[:8]+"...",
		)
	}

	// Get session to retrieve city code for redirect
	session, ok := h.stravaService.GetSession(sessionID)
	if !ok {
		// Shouldn't happen, but redirect to default
		http.Redirect(w, r, "/strava/success", http.StatusTemporaryRedirect)
		return
	}

	// In dev mode, show success page with session info for testing
	// In production, redirect to frontend
	if os.Getenv("APP_ENV") == "dev" {
		w.Header().Set("Content-Type", "text/html")
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Strava OAuth Success</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .success { background: #e8f5e9; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .session-id { background: #f5f5f5; padding: 10px; border-radius: 5px; font-family: monospace; word-break: break-all; }
        button { background: #FC4C02; color: white; border: none; padding: 10px 20px; cursor: pointer; border-radius: 5px; margin: 5px; }
        button:hover { background: #E34402; }
        h1 { color: #4CAF50; }
    </style>
</head>
<body>
    <h1>✓ OAuth Successful!</h1>
    <div class="success">
        <p><strong>Athlete:</strong> %s</p>
        <p><strong>City:</strong> %s</p>
        <p><strong>Session ID:</strong></p>
        <div class="session-id" id="sessionId">%s</div>
    </div>
    <button onclick="copySessionId()">Copy Session ID</button>
    <button onclick="window.location.href='http://localhost:3000/test-websocket'">Go to WebSocket Test</button>
    <script>
        function copySessionId() {
            navigator.clipboard.writeText('%s');
            alert('Session ID copied!');
        }
    </script>
</body>
</html>`, session.AthleteName, session.CityCode, sessionID, sessionID)
		w.Write([]byte(html))
		return
	}

	// Production: redirect to frontend with city code
	redirectURL := "/strava/success"
	if session.CityCode != "" {
		redirectURL = "/strava/success?city=" + session.CityCode
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
	http.SetCookie(w, &http.Cookie{
		Name:     "strava_session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") != "dev",
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
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
			// Clear invalid session cookie
			http.SetCookie(w, &http.Cookie{
				Name:   "strava_session_id",
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
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
	json.NewEncoder(w).Encode(clubs)
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
			http.SetCookie(w, &http.Cookie{
				Name:   "strava_session_id",
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// getSessionIDFromCookie extracts the session ID from the cookie
func (h *Handler) getSessionIDFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("strava_session_id")
	if err != nil {
		return ""
	}
	return cookie.Value
}
