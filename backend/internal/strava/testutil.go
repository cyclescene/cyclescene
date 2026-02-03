package strava

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockServer creates a test HTTP server for mocking Strava API
// The server is automatically closed when the test completes
func MockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(func() { server.Close() })
	return server
}

// MockServerWithRateLimits creates a test server that includes rate limit headers
func MockServerWithRateLimits(t *testing.T, handler http.HandlerFunc, usage15min, limit15min, usageDaily, limitDaily int) *httptest.Server {
	t.Helper()
	wrappedHandler := func(w http.ResponseWriter, r *http.Request) {
		// Add rate limit headers to all responses
		w.Header().Set("X-Ratelimit-Limit", formatRateLimitHeader(limit15min, limitDaily))
		w.Header().Set("X-Ratelimit-Usage", formatRateLimitHeader(usage15min, usageDaily))
		w.Header().Set("X-Readratelimit-Limit", formatRateLimitHeader(100, 1000))
		w.Header().Set("X-Readratelimit-Usage", formatRateLimitHeader(usage15min, usageDaily))
		handler(w, r)
	}
	return MockServer(t, wrappedHandler)
}

// formatRateLimitHeader formats rate limit values as "15min,daily"
func formatRateLimitHeader(v1, v2 int) string {
	return formatInt(v1) + "," + formatInt(v2)
}

func formatInt(i int) string {
	return string(rune('0'+i/100%10)) + string(rune('0'+i/10%10)) + string(rune('0'+i%10))
}

// MockTokenResponse returns a mock TokenResponse for testing
func MockTokenResponse() *TokenResponse {
	return &TokenResponse{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    1893456000, // Far future timestamp
		ExpiresIn:    21600,
		TokenType:    "Bearer",
		Athlete: Athlete{
			ID:        12345,
			FirstName: "Test",
			LastName:  "User",
			City:      "Portland",
			State:     "OR",
			Country:   "USA",
		},
	}
}

// MockClub returns a mock Club for testing
func MockClub(id int64, name string) Club {
	return Club{
		ID:          id,
		Name:        name,
		City:        "Portland",
		State:       "OR",
		Country:     "USA",
		MemberCount: 100,
		SportType:   "cycling",
	}
}

// MockClubDetail returns a mock ClubDetail for testing
func MockClubDetail(id int64, name string, admin, owner bool) *ClubDetail {
	return &ClubDetail{
		Club: MockClub(id, name),
		Admin:      admin,
		Owner:      owner,
		Membership: "member",
		ClubType:   "casual_club",
	}
}

// MockGroupEvent returns a mock GroupEvent for testing
func MockGroupEvent(id int64, title string, clubID int64) GroupEvent {
	return GroupEvent{
		ID:                  id,
		Title:               title,
		Description:         "Test event description",
		ActivityType:        "Ride",
		UpcomingOccurrences: []string{"2026-02-15T18:00:00Z"},
		Zone:                "America/Los_Angeles",
		Address:             "123 Main St, Portland, OR",
		StartLatLng:         []float64{45.5152, -122.6784},
		ClubID:              clubID,
		Private:             false,
		WomenOnly:           false,
	}
}

// MockRoute returns a mock Route for testing
func MockRoute(id int64, name string) *Route {
	return &Route{
		ID:            id,
		Name:          name,
		Description:   "Test route description",
		Distance:      25000, // 25km in meters
		ElevationGain: 300,   // 300m
		Type:          "ride",
		SubType:       "road",
	}
}

// WriteJSON writes a JSON response with the given status code
func WriteJSON(t *testing.T, w http.ResponseWriter, status int, data interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		t.Fatalf("Failed to encode JSON response: %v", err)
	}
}

// WriteError writes a Strava-style error response
func WriteError(t *testing.T, w http.ResponseWriter, status int, message, resource, field, code string) {
	t.Helper()
	errResp := map[string]interface{}{
		"message": message,
		"errors": []map[string]string{
			{
				"resource": resource,
				"field":    field,
				"code":     code,
			},
		},
	}
	WriteJSON(t, w, status, errResp)
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// AssertEqual fails the test if expected != actual
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

// AssertTrue fails the test if condition is false
func AssertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Errorf("Expected true: %s", msg)
	}
}

// AssertFalse fails the test if condition is true
func AssertFalse(t *testing.T, condition bool, msg string) {
	t.Helper()
	if condition {
		t.Errorf("Expected false: %s", msg)
	}
}

// TestConfig returns a Config suitable for testing
func TestConfig() *Config {
	return &Config{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		CallbackPath: "/v1/strava/auth/callback",
		Debug:        true,
	}
}

// TestClient creates a Client configured for testing with a mock server URL
func TestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	client := NewClient(TestConfig())
	client.SetBaseURL(serverURL)
	return client
}
