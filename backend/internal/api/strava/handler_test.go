package strava

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spacesedan/cyclescene/backend/internal/strava"
)

// TestHandler tests for HTTP handlers
// These tests use mock services to test handler logic

func setupTestHandler() (*Handler, *strava.Service, *strava.SessionStore) {
	// Create test config
	config := &strava.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		CallbackPath: "/v1/strava/auth/callback",
		Debug:        true,
	}

	// Create mock HTTP server for Strava API
	mockStravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add rate limit headers to all responses
		w.Header().Set("X-Ratelimit-Limit", "200,2000")
		w.Header().Set("X-Ratelimit-Usage", "1,1")
		w.Header().Set("X-Readratelimit-Limit", "100,1000")
		w.Header().Set("X-Readratelimit-Usage", "1,1")

		switch r.URL.Path {
		case "/oauth/token":
			// Mock token exchange
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "test-access-token",
				"refresh_token": "test-refresh-token",
				"expires_at":    1893456000, // Far future
				"expires_in":    21600,
				"token_type":    "Bearer",
				"athlete": map[string]interface{}{
					"id":        12345,
					"firstname": "Test",
					"lastname":  "User",
					"city":      "Portland",
					"state":     "Oregon",
					"country":   "United States",
				},
			})
		case "/athlete/clubs":
			// Mock clubs response (without /api/v3 prefix since it's in baseURL)
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id":           123,
					"name":         "Portland Bike Club",
					"city":         "Portland",
					"state":        "Oregon",
					"country":      "United States",
					"member_count": 100,
					"sport_type":   "cycling",
				},
			})
		case "/clubs/123":
			// Mock club details (admin)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           123,
				"name":         "Portland Bike Club",
				"city":         "Portland",
				"state":        "Oregon",
				"admin":        true,
				"owner":        false,
				"member_count": 100,
				"sport_type":   "cycling",
			})
		case "/clubs/123/group_events":
			// Mock events response
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id":                   789,
					"title":                "Tuesday Night Ride",
					"description":          "Weekly social ride",
					"activity_type":        "Ride",
					"upcoming_occurrences": []string{"2026-03-01T18:00:00Z"},
					"zone":                 "America/Los_Angeles",
					"address":              "123 Main St, Portland, OR",
					"start_latlng":         []float64{45.5152, -122.6784},
					"club_id":              123,
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))

	// Create client with mock server URL
	client := strava.NewClientWithBaseURL(config, mockStravaServer.URL)
	sessionStore := strava.NewSessionStore()
	service := strava.NewService(client, sessionStore, nil, nil, "http://localhost:3000/callback")

	handler := NewHandler(service)

	return handler, service, sessionStore
}

func TestInitiateOAuth(t *testing.T) {
	handler, _, _ := setupTestHandler()

	// Create test request
	req := httptest.NewRequest("GET", "/v1/strava/auth/initiate?city=pdx", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.InitiateOAuth(w, req)

	// Should redirect to Strava
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Error("Expected Location header to be set")
	}

	// Should contain Strava OAuth URL
	if !contains(location, "strava.com/oauth/authorize") {
		t.Errorf("Expected redirect to Strava OAuth, got: %s", location)
	}

	// Should contain client_id
	if !contains(location, "client_id=test-client-id") {
		t.Errorf("Expected client_id in URL, got: %s", location)
	}

	// Should contain state parameter
	if !contains(location, "state=") {
		t.Errorf("Expected state parameter in URL, got: %s", location)
	}
}

func TestInitiateOAuth_DefaultCity(t *testing.T) {
	handler, _, _ := setupTestHandler()

	// Create test request without city parameter
	req := httptest.NewRequest("GET", "/v1/strava/auth/initiate", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.InitiateOAuth(w, req)

	// Should still redirect (default to pdx)
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}
}

func TestHandleOAuthCallback_MissingParams(t *testing.T) {
	handler, _, _ := setupTestHandler()

	tests := []struct {
		name string
		url  string
	}{
		{"missing code", "/v1/strava/auth/callback?state=test"},
		{"missing state", "/v1/strava/auth/callback?code=test"},
		{"missing both", "/v1/strava/auth/callback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			handler.HandleOAuthCallback(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestHandleOAuthCallback_OAuthError(t *testing.T) {
	handler, _, _ := setupTestHandler()

	// Strava returns error when user denies access
	req := httptest.NewRequest("GET", "/v1/strava/auth/callback?error=access_denied", nil)
	w := httptest.NewRecorder()

	handler.HandleOAuthCallback(w, req)

	// Should redirect to error page
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}

	location := w.Header().Get("Location")
	if !contains(location, "error=access_denied") {
		t.Errorf("Expected error redirect, got: %s", location)
	}
}

func TestLogout_WithSession(t *testing.T) {
	handler, service, _ := setupTestHandler()

	// Create a session first
	session := &strava.Session{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		AthleteID:    12345,
		AthleteName:  "Test User",
		CityCode:     "pdx",
	}
	sessionID, _ := service.GetSessionStore().CreateSession(session)

	// Create request with session cookie
	req := httptest.NewRequest("POST", "/v1/strava/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "strava_session_id",
		Value: sessionID,
	})
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	// Should return success
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Should clear cookie
	cookies := w.Result().Cookies()
	var foundCookie bool
	for _, c := range cookies {
		if c.Name == "strava_session_id" {
			foundCookie = true
			if c.MaxAge != -1 {
				t.Error("Expected cookie to be deleted (MaxAge=-1)")
			}
		}
	}
	if !foundCookie {
		t.Error("Expected strava_session_id cookie to be set")
	}

	// Session should be deleted
	_, ok := service.GetSession(sessionID)
	if ok {
		t.Error("Expected session to be deleted")
	}
}

func TestLogout_WithoutSession(t *testing.T) {
	handler, _, _ := setupTestHandler()

	// Create request without session cookie
	req := httptest.NewRequest("POST", "/v1/strava/logout", nil)
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	// Should still return success
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetAdminClubs_NoSession(t *testing.T) {
	handler, _, _ := setupTestHandler()

	req := httptest.NewRequest("GET", "/v1/strava/admin-clubs", nil)
	w := httptest.NewRecorder()

	handler.GetAdminClubs(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetAdminClubs_WithSession(t *testing.T) {
	handler, service, _ := setupTestHandler()

	// Create a session with far-future expiry
	session := &strava.Session{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		AthleteID:    12345,
		AthleteName:  "Test User",
		CityCode:     "pdx",
		ExpiresAt:    time.Now().Add(24 * time.Hour), // Far future
	}
	sessionID, _ := service.GetSessionStore().CreateSession(session)

	req := httptest.NewRequest("GET", "/v1/strava/admin-clubs", nil)
	req.AddCookie(&http.Cookie{
		Name:  "strava_session_id",
		Value: sessionID,
	})
	w := httptest.NewRecorder()

	handler.GetAdminClubs(w, req)

	// Should return 200 (mock server returns clubs)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify response structure
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := response["clubs"]; !ok {
		t.Error("Expected 'clubs' field in response")
	}
}

func TestGetClubEvents_NoSession(t *testing.T) {
	handler, _, _ := setupTestHandler()

	// Setup chi router to extract URL params
	r := chi.NewRouter()
	r.Get("/v1/strava/clubs/{clubId}/events", handler.GetClubEvents)

	req := httptest.NewRequest("GET", "/v1/strava/clubs/123/events", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetClubEvents_InvalidClubID(t *testing.T) {
	handler, service, _ := setupTestHandler()

	// Create a session
	session := &strava.Session{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		AthleteID:    12345,
		AthleteName:  "Test User",
		CityCode:     "pdx",
	}
	sessionID, _ := service.GetSessionStore().CreateSession(session)

	// Setup chi router
	r := chi.NewRouter()
	r.Get("/v1/strava/clubs/{clubId}/events", handler.GetClubEvents)

	req := httptest.NewRequest("GET", "/v1/strava/clubs/invalid/events", nil)
	req.AddCookie(&http.Cookie{
		Name:  "strava_session_id",
		Value: sessionID,
	})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRegisterRoutes(t *testing.T) {
	handler, _, _ := setupTestHandler()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	// Verify routes are registered by making requests
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/strava/auth/initiate"},
		{"GET", "/strava/auth/callback"},
		{"POST", "/strava/logout"},
		{"GET", "/strava/admin-clubs"},
		{"GET", "/strava/clubs/123/events"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			// Should not be 404 (route exists)
			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s %s not found", route.method, route.path)
			}
		})
	}
}

func TestCheckSession_Valid(t *testing.T) {
	handler, service, _ := setupTestHandler()

	// Create a session with far-future expiry
	session := &strava.Session{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		AthleteID:    12345,
		AthleteName:  "Test User",
		CityCode:     "pdx",
		ExpiresAt:    time.Now().Add(24 * time.Hour), // Far future
	}
	sessionID, _ := service.GetSessionStore().CreateSession(session)

	req := httptest.NewRequest("GET", "/v1/strava/check-session", nil)
	req.AddCookie(&http.Cookie{
		Name:  "strava_session_id",
		Value: sessionID,
	})
	w := httptest.NewRecorder()

	handler.CheckSession(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]bool
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response["valid"] {
		t.Error("Expected valid: true in response")
	}
}

func TestCheckSession_NoSession(t *testing.T) {
	handler, _, _ := setupTestHandler()

	req := httptest.NewRequest("GET", "/v1/strava/check-session", nil)
	w := httptest.NewRecorder()

	handler.CheckSession(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestCheckSession_InvalidSession(t *testing.T) {
	handler, _, _ := setupTestHandler()

	req := httptest.NewRequest("GET", "/v1/strava/check-session", nil)
	req.AddCookie(&http.Cookie{
		Name:  "strava_session_id",
		Value: "invalid-session-id",
	})
	w := httptest.NewRecorder()

	handler.CheckSession(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Should clear the invalid cookie
	cookies := w.Result().Cookies()
	var clearedCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "strava_session_id" {
			clearedCookie = cookie
			break
		}
	}
	if clearedCookie != nil && clearedCookie.MaxAge >= 0 {
		t.Error("Expected session cookie to be cleared (MaxAge < 0)")
	}
}

func TestGetClubEvents_WithSession(t *testing.T) {
	handler, service, _ := setupTestHandler()

	// Create a session with far-future expiry
	session := &strava.Session{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		AthleteID:    12345,
		AthleteName:  "Test User",
		CityCode:     "pdx",
		ExpiresAt:    time.Now().Add(24 * time.Hour), // Far future
	}
	sessionID, _ := service.GetSessionStore().CreateSession(session)

	// Setup chi router
	r := chi.NewRouter()
	r.Get("/v1/strava/clubs/{clubId}/events", handler.GetClubEvents)

	req := httptest.NewRequest("GET", "/v1/strava/clubs/123/events", nil)
	req.AddCookie(&http.Cookie{
		Name:  "strava_session_id",
		Value: sessionID,
	})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	events, ok := response["events"].([]interface{})
	if !ok {
		t.Fatal("Expected 'events' array in response")
	}

	if len(events) == 0 {
		t.Error("Expected at least one event in response")
	}

	// Verify event ID is a string (not a number) to avoid JavaScript precision loss
	if len(events) > 0 {
		firstEvent := events[0].(map[string]interface{})
		eventID, ok := firstEvent["id"].(string)
		if !ok {
			t.Error("Expected event ID to be a string")
		}
		if eventID == "" {
			t.Error("Expected event ID to be non-empty")
		}
	}
}

func TestToGroupEventResponse_IDConversion(t *testing.T) {
	// Test that large int64 event IDs are converted to strings correctly
	event := strava.GroupEvent{
		ID:                  9007199254740992, // Large int64 that would lose precision in JS
		Title:               "Test Event",
		ActivityType:        "Ride",
		UpcomingOccurrences: []string{"2026-03-01T18:00:00Z"},
		Zone:                "America/Los_Angeles",
		Address:             "123 Main St",
		StartLatLng:         []float64{45.5, -122.6},
		ClubID:              123,
	}

	response := toGroupEventResponse(event)

	if response.ID != "9007199254740992" {
		t.Errorf("Expected ID to be '9007199254740992', got '%s'", response.ID)
	}

	if response.Title != event.Title {
		t.Errorf("Expected Title to be '%s', got '%s'", event.Title, response.Title)
	}
}

// Helper to check if NewClientWithBaseURL exists, if not we need to add it
func init() {
	// Set debug mode for tests
	os.Setenv("STRAVA_DEBUG", "true")
}
