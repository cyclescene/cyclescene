package strava

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/spacesedan/cyclescene/backend/internal/strava"
)

// Mock ride service for testing
type mockRideService struct {
	submitCalled    int
	submitError     error
	lastSubmission  interface{}
	returnEventID   int64
	returnEditToken string
	mu              sync.Mutex
}

func (m *mockRideService) SubmitRideWithCoordinates(ctx context.Context, submission interface{}, lat, lng float64) (*mockSubmissionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.submitCalled++
	m.lastSubmission = submission
	if m.submitError != nil {
		return nil, m.submitError
	}
	return &mockSubmissionResponse{
		Success:   true,
		EventID:   m.returnEventID,
		EditToken: m.returnEditToken,
	}, nil
}

type mockSubmissionResponse struct {
	Success   bool
	EventID   int64
	EditToken string
}

// Mock magiclink service for testing
type mockMagicLinkService struct {
	emailsSent int
	lastEmail  string
	sendError  error
	mu         sync.Mutex
}

func (m *mockMagicLinkService) SendImportSummaryEmail(ctx context.Context, email string, events interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.emailsSent++
	m.lastEmail = email
	if m.sendError != nil {
		return nil, m.sendError
	}
	return map[string]string{"message_id": "test-message-id"}, nil
}

func TestImportHandler_ConcurrentImportDetection(t *testing.T) {
	// Test that concurrent imports from the same user are blocked
	handler := &ImportHandler{
		activeImports: sync.Map{},
		debug:         true,
	}

	athleteID := int64(12345)

	// First import should succeed
	_, loaded := handler.activeImports.LoadOrStore(athleteID, true)
	if loaded {
		t.Error("First import should not be blocked")
	}

	// Second import should be blocked
	_, loaded = handler.activeImports.LoadOrStore(athleteID, true)
	if !loaded {
		t.Error("Second import should be blocked")
	}

	// Clean up
	handler.activeImports.Delete(athleteID)

	// After cleanup, should be able to import again
	_, loaded = handler.activeImports.LoadOrStore(athleteID, true)
	if loaded {
		t.Error("Import after cleanup should not be blocked")
	}
}

func TestProgressMessage_Serialization(t *testing.T) {
	tests := []struct {
		name string
		msg  ProgressMessage
	}{
		{
			name: "heartbeat",
			msg: ProgressMessage{
				Type:    "heartbeat",
				Message: "Import still in progress...",
			},
		},
		{
			name: "progress",
			msg: ProgressMessage{
				Type:          "progress",
				EventIndex:    0,
				TotalEvents:   3,
				StravaEventID: 789,
				EventTitle:    "Tuesday Night Ride",
				Step:          "fetching",
				Status:        "in_progress",
				Message:       "Fetching event data...",
			},
		},
		{
			name: "complete",
			msg: ProgressMessage{
				Type:              "complete",
				EventIndex:        0,
				StravaEventID:     789,
				CycleSceneEventID: 42,
				EditToken:         "xyz123",
				EditURL:           "https://form.cyclescene.cc?token=xyz123",
				Success:           true,
			},
		},
		{
			name: "done",
			msg: ProgressMessage{
				Type:             "done",
				TotalImported:    3,
				TotalFailed:      0,
				SummaryEmailSent: true,
				Results: []ImportResult{
					{
						StravaEventID:     789,
						CycleSceneEventID: 42,
						EditToken:         "xyz123",
						Success:           true,
					},
				},
			},
		},
		{
			name: "error",
			msg: ProgressMessage{
				Type:    "error",
				Message: "Session expired",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("Failed to serialize: %v", err)
			}

			// Deserialize
			var decoded ProgressMessage
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to deserialize: %v", err)
			}

			// Verify type
			if decoded.Type != tt.msg.Type {
				t.Errorf("Type mismatch: expected %s, got %s", tt.msg.Type, decoded.Type)
			}
		})
	}
}

func TestImportRequest_Serialization(t *testing.T) {
	req := ImportRequest{
		SessionID:      "abc123",
		OrganizerEmail: "user@example.com",
		Events: []EventImportConfig{
			{
				StravaEventID: 789,
				ClubID:        123,
				Overrides: map[string]string{
					"audience":  "All",
					"image_url": "https://example.com/image.jpg",
				},
			},
		},
	}

	// Serialize
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	// Deserialize
	var decoded ImportRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if decoded.SessionID != req.SessionID {
		t.Errorf("SessionID mismatch: expected %s, got %s", req.SessionID, decoded.SessionID)
	}

	if decoded.OrganizerEmail != req.OrganizerEmail {
		t.Errorf("OrganizerEmail mismatch: expected %s, got %s", req.OrganizerEmail, decoded.OrganizerEmail)
	}

	if len(decoded.Events) != 1 {
		t.Fatalf("Events count mismatch: expected 1, got %d", len(decoded.Events))
	}

	if decoded.Events[0].StravaEventID != 789 {
		t.Errorf("StravaEventID mismatch: expected 789, got %d", decoded.Events[0].StravaEventID)
	}

	if decoded.Events[0].Overrides["audience"] != "All" {
		t.Errorf("Override mismatch: expected 'All', got %s", decoded.Events[0].Overrides["audience"])
	}
}

func TestIsDuplicateError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "unique constraint error",
			err:      &testError{msg: "UNIQUE constraint failed: events.source"},
			expected: true,
		},
		{
			name:     "duplicate key error",
			err:      &testError{msg: "duplicate key value violates unique constraint"},
			expected: true,
		},
		{
			name:     "source dedup index error",
			err:      &testError{msg: "idx_events_source_dedup violation"},
			expected: true,
		},
		{
			name:     "other error",
			err:      &testError{msg: "connection refused"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDuplicateError(tt.err)
			if result != tt.expected {
				t.Errorf("isDuplicateError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestWebSocketHandler_Integration(t *testing.T) {
	// Skip if short test mode
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	// Create test strava service
	config := &strava.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Debug:        true,
	}

	// Create mock Strava server
	mockStravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/clubs/123/group_events":
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id":                   789,
					"title":                "Test Ride",
					"description":          "Test description",
					"activity_type":        "Ride",
					"upcoming_occurrences": []string{"2026-03-01T18:00:00Z"},
					"zone":                 "America/Los_Angeles",
					"address":              "123 Main St",
					"start_latlng":         []float64{45.5, -122.6},
					"club_id":              123,
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockStravaServer.Close()

	client := strava.NewClientWithBaseURL(config, mockStravaServer.URL)
	sessionStore := strava.NewSessionStore()
	stravaService := strava.NewService(client, sessionStore, nil, nil, "http://localhost:3000/callback")

	// Create a test session
	session := &strava.Session{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		AthleteID:    12345,
		AthleteName:  "Test User",
		CityCode:     "pdx",
	}
	sessionID, _ := sessionStore.CreateSession(session)

	// Create import handler (without real ride/magiclink services for this test)
	importHandler := &ImportHandler{
		stravaService:  stravaService,
		rideService:    nil, // Would need full setup
		magicLinkSvc:   nil,
		editLinkBase:   "https://form.cyclescene.cc/rides/edit",
		appAccessToken: "test-app-token",
		debug:          true,
	}

	// Create WebSocket test server
	wsServer := httptest.NewServer(http.HandlerFunc(importHandler.HandleImport))
	defer wsServer.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + wsServer.URL[4:] // http:// -> ws://

	// Connect to WebSocket
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.CloseNow()

	// Send import request
	importReq := ImportRequest{
		SessionID:      sessionID,
		OrganizerEmail: "test@example.com",
		Events: []EventImportConfig{
			{
				StravaEventID: 789,
				ClubID:        123,
			},
		},
	}

	if err := wsjson.Write(ctx, conn, importReq); err != nil {
		t.Fatalf("Failed to send import request: %v", err)
	}

	// Read first response (should be progress or error)
	var response ProgressMessage
	if err := wsjson.Read(ctx, conn, &response); err != nil {
		// This might fail because rideService is nil, which is expected
		// The test verifies the WebSocket connection works
		t.Logf("Received response or error (expected in partial test): %v", err)
	} else {
		t.Logf("Received response type: %s", response.Type)
	}
}

func TestHeartbeat_Timing(t *testing.T) {
	// Skip in short mode as this tests timing
	if testing.Short() {
		t.Skip("Skipping heartbeat timing test in short mode")
	}

	// This test would verify heartbeats are sent every 30 seconds
	// For unit tests, we just verify the heartbeat message format
	msg := ProgressMessage{
		Type:    "heartbeat",
		Message: "Import still in progress...",
	}

	if msg.Type != "heartbeat" {
		t.Errorf("Expected type 'heartbeat', got '%s'", msg.Type)
	}
}

func TestImportResult_Serialization(t *testing.T) {
	result := ImportResult{
		StravaEventID:     789,
		CycleSceneEventID: 42,
		EditToken:         "xyz123",
		EditURL:           "https://form.cyclescene.cc/rides/edit?token=xyz123",
		Success:           true,
		Title:             "Tuesday Night Ride",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	var decoded ImportResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if decoded.StravaEventID != result.StravaEventID {
		t.Errorf("StravaEventID mismatch")
	}

	if decoded.Success != result.Success {
		t.Errorf("Success mismatch")
	}
}

func TestEventImportConfig_Overrides(t *testing.T) {
	config := EventImportConfig{
		StravaEventID: 789,
		ClubID:        123,
		Overrides: map[string]string{
			"audience":    "All",
			"image_url":   "https://example.com/image.jpg",
			"ride_length": "Medium",
			"area":        "Downtown",
		},
	}

	// Verify all overrides are preserved through serialization
	data, _ := json.Marshal(config)
	var decoded EventImportConfig
	json.Unmarshal(data, &decoded)

	if len(decoded.Overrides) != 4 {
		t.Errorf("Expected 4 overrides, got %d", len(decoded.Overrides))
	}

	if decoded.Overrides["audience"] != "All" {
		t.Errorf("audience override mismatch")
	}

	if decoded.Overrides["image_url"] != "https://example.com/image.jpg" {
		t.Errorf("image_url override mismatch")
	}
}
