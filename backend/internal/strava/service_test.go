package strava

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestClub_IsCyclingClub(t *testing.T) {
	tests := []struct {
		name     string
		club     Club
		expected bool
	}{
		{
			name: "cycling via activity_types",
			club: Club{
				ID:            1,
				Name:          "Portland Riders",
				ActivityTypes: []string{"Ride", "VirtualRide"},
			},
			expected: true,
		},
		{
			name: "cycling via sport_type",
			club: Club{
				ID:        2,
				Name:      "SLC Cycling",
				SportType: "cycling",
			},
			expected: true,
		},
		{
			name: "ebike club",
			club: Club{
				ID:            3,
				Name:          "EBike Enthusiasts",
				ActivityTypes: []string{"EBikeRide"},
			},
			expected: true,
		},
		{
			name: "running club",
			club: Club{
				ID:            4,
				Name:          "Portland Runners",
				ActivityTypes: []string{"Run", "TrailRun"},
				SportType:     "running",
			},
			expected: false,
		},
		{
			name: "mixed activity club with cycling",
			club: Club{
				ID:            5,
				Name:          "Triathletes",
				ActivityTypes: []string{"Run", "Swim", "Ride"},
			},
			expected: true,
		},
		{
			name: "empty activity types, non-cycling sport_type",
			club: Club{
				ID:            6,
				Name:          "Swimmers",
				SportType:     "swimming",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.club.IsCyclingClub()
			AssertEqual(t, tt.expected, result)
		})
	}
}

func TestClub_MatchesCity(t *testing.T) {
	tests := []struct {
		name     string
		club     Club
		cityCode string
		expected bool
	}{
		{
			name: "portland matches pdx",
			club: Club{
				ID:   1,
				Name: "Portland Riders",
				City: "Portland",
			},
			cityCode: "pdx",
			expected: true,
		},
		{
			name: "Portland (case insensitive) matches pdx",
			club: Club{
				ID:   2,
				Name: "PORTLAND Cycling",
				City: "PORTLAND",
			},
			cityCode: "pdx",
			expected: true,
		},
		{
			name: "Salt Lake City matches slc",
			club: Club{
				ID:   3,
				Name: "SLC Riders",
				City: "Salt Lake City",
			},
			cityCode: "slc",
			expected: true,
		},
		{
			name: "Seattle does not match pdx",
			club: Club{
				ID:   4,
				Name: "Seattle Cyclists",
				City: "Seattle",
			},
			cityCode: "pdx",
			expected: false,
		},
		{
			name: "Denver does not match slc",
			club: Club{
				ID:   5,
				Name: "Denver Riders",
				City: "Denver",
			},
			cityCode: "slc",
			expected: false,
		},
		{
			name: "unknown city code",
			club: Club{
				ID:   6,
				Name: "Portland Riders",
				City: "Portland",
			},
			cityCode: "xyz",
			expected: false,
		},
		{
			name: "empty city in club",
			club: Club{
				ID:   7,
				Name: "Mystery Riders",
				City: "",
			},
			cityCode: "pdx",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.club.MatchesCity(tt.cityCode)
			AssertEqual(t, tt.expected, result)
		})
	}
}

func TestClubDetail_IsAdminOrOwner(t *testing.T) {
	tests := []struct {
		name     string
		detail   ClubDetail
		expected bool
	}{
		{
			name: "admin true, owner false",
			detail: ClubDetail{
				Admin: true,
				Owner: false,
			},
			expected: true,
		},
		{
			name: "admin false, owner true",
			detail: ClubDetail{
				Admin: false,
				Owner: true,
			},
			expected: true,
		},
		{
			name: "both admin and owner",
			detail: ClubDetail{
				Admin: true,
				Owner: true,
			},
			expected: true,
		},
		{
			name: "neither admin nor owner",
			detail: ClubDetail{
				Admin: false,
				Owner: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.detail.IsAdminOrOwner()
			AssertEqual(t, tt.expected, result)
		})
	}
}

func TestGroupEvent_GetLatLong(t *testing.T) {
	tests := []struct {
		name        string
		event       GroupEvent
		expectedLat float64
		expectedLng float64
		hasLocation bool
	}{
		{
			name: "valid coordinates",
			event: GroupEvent{
				StartLatLng: []float64{45.5152, -122.6784},
			},
			expectedLat: 45.5152,
			expectedLng: -122.6784,
			hasLocation: true,
		},
		{
			name: "empty coordinates",
			event: GroupEvent{
				StartLatLng: []float64{},
			},
			expectedLat: 0,
			expectedLng: 0,
			hasLocation: false,
		},
		{
			name: "nil coordinates",
			event: GroupEvent{
				StartLatLng: nil,
			},
			expectedLat: 0,
			expectedLng: 0,
			hasLocation: false,
		},
		{
			name: "only one coordinate",
			event: GroupEvent{
				StartLatLng: []float64{45.5152},
			},
			expectedLat: 0,
			expectedLng: 0,
			hasLocation: false,
		},
		{
			name: "zero coordinates",
			event: GroupEvent{
				StartLatLng: []float64{0, 0},
			},
			expectedLat: 0,
			expectedLng: 0,
			hasLocation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertEqual(t, tt.expectedLat, tt.event.GetLatitude())
			AssertEqual(t, tt.expectedLng, tt.event.GetLongitude())
			AssertEqual(t, tt.hasLocation, tt.event.HasLocation())
		})
	}
}

func TestGroupEvent_GetFirstOccurrence(t *testing.T) {
	event := GroupEvent{
		UpcomingOccurrences: []string{"2026-02-15T18:00:00Z", "2026-02-22T18:00:00Z"},
	}

	eventTime, err := event.GetFirstOccurrence()
	AssertNoError(t, err)
	AssertEqual(t, 2026, eventTime.Year())
	AssertEqual(t, 2, int(eventTime.Month()))
	AssertEqual(t, 15, eventTime.Day())
}

func TestGroupEvent_GetFirstOccurrence_Empty(t *testing.T) {
	event := GroupEvent{
		UpcomingOccurrences: []string{},
	}

	_, err := event.GetFirstOccurrence()
	AssertError(t, err)
}

func TestGroupEvent_GetTimezone(t *testing.T) {
	tests := []struct {
		name     string
		zone     string
		expected string
	}{
		{
			name:     "Los Angeles",
			zone:     "America/Los_Angeles",
			expected: "America/Los_Angeles",
		},
		{
			name:     "Denver",
			zone:     "America/Denver",
			expected: "America/Denver",
		},
		{
			name:     "empty zone defaults to UTC",
			zone:     "",
			expected: "UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := GroupEvent{Zone: tt.zone}
			tz, err := event.GetTimezone()
			AssertNoError(t, err)
			AssertEqual(t, tt.expected, tz.String())
		})
	}
}

func TestGroupEvent_IsUpcoming(t *testing.T) {
	tests := []struct {
		name     string
		event    GroupEvent
		expected bool
	}{
		{
			name: "future event",
			event: GroupEvent{
				UpcomingOccurrences: []string{"2030-12-15T18:00:00Z"},
			},
			expected: true,
		},
		{
			name: "past event",
			event: GroupEvent{
				UpcomingOccurrences: []string{"2020-01-01T18:00:00Z"},
			},
			expected: false,
		},
		{
			name: "no occurrences",
			event: GroupEvent{
				UpcomingOccurrences: []string{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.event.IsUpcoming()
			AssertEqual(t, tt.expected, result)
		})
	}
}

func TestFilterUpcomingEvents(t *testing.T) {
	events := []GroupEvent{
		{ID: 1, Title: "Future Event 1", UpcomingOccurrences: []string{"2030-06-15T18:00:00Z"}},
		{ID: 2, Title: "Past Event", UpcomingOccurrences: []string{"2020-01-01T18:00:00Z"}},
		{ID: 3, Title: "Future Event 2", UpcomingOccurrences: []string{"2030-12-25T10:00:00Z"}},
		{ID: 4, Title: "No Occurrences", UpcomingOccurrences: []string{}},
	}

	upcoming := FilterUpcomingEvents(events)

	AssertEqual(t, 2, len(upcoming))
	AssertEqual(t, int64(1), upcoming[0].ID)
	AssertEqual(t, int64(3), upcoming[1].ID)
}

func TestFilterUpcomingEvents_Empty(t *testing.T) {
	events := []GroupEvent{}
	upcoming := FilterUpcomingEvents(events)
	AssertEqual(t, 0, len(upcoming))
}

func TestFilterUpcomingEvents_AllPast(t *testing.T) {
	events := []GroupEvent{
		{ID: 1, UpcomingOccurrences: []string{"2020-01-01T18:00:00Z"}},
		{ID: 2, UpcomingOccurrences: []string{"2019-06-15T18:00:00Z"}},
	}
	upcoming := FilterUpcomingEvents(events)
	AssertEqual(t, 0, len(upcoming))
}

func TestGroupEvent_ToSubmission(t *testing.T) {
	event := GroupEvent{
		ID:                  123,
		Title:               "Tuesday Night Ride",
		Description:         "A fun social ride",
		UpcomingOccurrences: []string{"2026-02-15T01:00:00Z"}, // 5pm PST / 6pm MST
		Zone:                "America/Los_Angeles",
		Address:             "123 Main St, Portland, OR",
		ClubID:              456,
		Route: &Route{
			ID:       789,
			Name:     "Waterfront Loop",
			Distance: 40233.6, // ~25 miles in meters
		},
	}

	submission, err := event.ToSubmission("pdx", "organizer@example.com", 45.5152, -122.6784)
	AssertNoError(t, err)

	AssertEqual(t, "Tuesday Night Ride", submission.Title)
	AssertEqual(t, "A fun social ride", submission.Description)
	AssertEqual(t, "pdx", submission.City)
	AssertEqual(t, "organizer@example.com", submission.OrganizerEmail)
	AssertEqual(t, "123 Main St, Portland, OR", submission.Address)
	AssertEqual(t, "View on Strava", submission.WebName)
	AssertTrue(t, submission.WebURL != "", "WebURL should not be empty")
	AssertEqual(t, "O", submission.DateType) // One-off event
	AssertEqual(t, 1, len(submission.Occurrences))
	AssertEqual(t, "2026-02-14", submission.Occurrences[0].StartDate) // UTC to PST conversion
	AssertEqual(t, "17:00", submission.Occurrences[0].StartTime)      // 5pm PST
	AssertEqual(t, 120, submission.Occurrences[0].EventDurationMinutes)

	// Check ride length calculated from route
	AssertTrue(t, submission.RideLength != "", "RideLength should be calculated from route")
}

func TestGroupEvent_StravaEventURL(t *testing.T) {
	event := GroupEvent{
		ID:     123,
		ClubID: 456,
	}

	url := event.StravaEventURL()
	AssertEqual(t, "https://www.strava.com/clubs/456/group_events/123", url)
}

func TestRoute_DistanceInMiles(t *testing.T) {
	route := Route{
		Distance: 40233.6, // ~25 miles
	}

	miles := route.DistanceInMiles()
	if miles < 24.9 || miles > 25.1 {
		t.Errorf("Expected ~25 miles, got %f", miles)
	}
}

func TestSessionStore_GenerateStateWithCity(t *testing.T) {
	store := NewSessionStore()

	state, err := store.GenerateStateWithCity("pdx")
	AssertNoError(t, err)
	AssertTrue(t, len(state) > 0, "state should not be empty")

	// Validate and get city
	cityCode, valid := store.ValidateStateAndGetCity(state)
	AssertTrue(t, valid, "state should be valid")
	AssertEqual(t, "pdx", cityCode)

	// State should be consumed
	_, valid = store.ValidateStateAndGetCity(state)
	AssertFalse(t, valid, "state should be consumed after validation")
}

func TestSessionStore_GenerateState_BackwardsCompatibility(t *testing.T) {
	store := NewSessionStore()

	// Old method should still work
	state, err := store.GenerateState()
	AssertNoError(t, err)
	AssertTrue(t, len(state) > 0, "state should not be empty")

	// Should validate with empty city code
	cityCode, valid := store.ValidateStateAndGetCity(state)
	AssertTrue(t, valid, "state should be valid")
	AssertEqual(t, "", cityCode)
}

func TestService_InitiateOAuth(t *testing.T) {
	store := NewSessionStore()
	client := NewClient(TestConfig())
	service := NewService(client, store, nil, nil, "https://example.com/callback")

	authURL, err := service.InitiateOAuth(context.Background(), "pdx")
	AssertNoError(t, err)
	AssertTrue(t, len(authURL) > 0, "authURL should not be empty")

	// Verify URL contains required parameters
	AssertTrue(t, contains(authURL, "client_id="), "should contain client_id")
	AssertTrue(t, contains(authURL, "redirect_uri="), "should contain redirect_uri")
	AssertTrue(t, contains(authURL, "state="), "should contain state")
	AssertTrue(t, contains(authURL, "scope="), "should contain scope")
}

func TestService_InitiateOAuth_InvalidCity(t *testing.T) {
	store := NewSessionStore()
	client := NewClient(TestConfig())
	service := NewService(client, store, nil, nil, "https://example.com/callback")

	_, err := service.InitiateOAuth(context.Background(), "invalid_city")
	AssertError(t, err)
}

func TestService_HandleOAuthCallback_InvalidState(t *testing.T) {
	store := NewSessionStore()
	client := NewClient(TestConfig())
	service := NewService(client, store, nil, nil, "https://example.com/callback")

	_, err := service.HandleOAuthCallback(context.Background(), "test_code", "invalid_state")
	AssertError(t, err)
}

func TestService_GetAdminClubs_NoSession(t *testing.T) {
	store := NewSessionStore()
	client := NewClient(TestConfig())
	service := NewService(client, store, nil, nil, "https://example.com/callback")

	_, err := service.GetAdminClubs(context.Background(), "invalid_session")
	AssertError(t, err)
}

func TestService_GetClubEvents_NoSession(t *testing.T) {
	store := NewSessionStore()
	client := NewClient(TestConfig())
	service := NewService(client, store, nil, nil, "https://example.com/callback")

	_, err := service.GetClubEvents(context.Background(), "invalid_session", 123)
	AssertError(t, err)
}

func TestService_GetAdminClubs_FiltersCorrectly(t *testing.T) {
	// Create mock server that returns clubs
	mockClubs := []Club{
		{ID: 1, Name: "Portland Cyclists", City: "Portland", ActivityTypes: []string{"Ride"}},
		{ID: 2, Name: "Seattle Runners", City: "Seattle", SportType: "running"},
		{ID: 3, Name: "Portland Runners", City: "Portland", SportType: "running"},
		{ID: 4, Name: "Denver Cyclists", City: "Denver", ActivityTypes: []string{"Ride"}},
	}

	mockClubDetails := map[int64]*ClubDetail{
		1: {Club: mockClubs[0], Admin: true, Owner: false},
		4: {Club: mockClubs[3], Admin: false, Owner: false},
	}

	callCount := 0
	server := MockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/athlete/clubs" {
			w.Header().Set("X-Ratelimit-Limit", "200,2000")
			w.Header().Set("X-Ratelimit-Usage", "1,1")
			WriteJSON(t, w, http.StatusOK, mockClubs)
			return
		}

		// Club details endpoint
		if r.URL.Path == "/clubs/1" {
			callCount++
			WriteJSON(t, w, http.StatusOK, mockClubDetails[1])
			return
		}

		// Should not be called for non-Portland or non-cycling clubs
		t.Errorf("Unexpected request to %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	})

	// Create session with pdx city code
	store := NewSessionStore()
	session := &Session{
		AccessToken: "test_token",
		AthleteID:   12345,
		CityCode:    "pdx",
		ExpiresAt:   time.Now().Add(1 * time.Hour), // Must set expiry or session is invalid
	}
	sessionID, _ := store.CreateSession(session)

	// Create service
	client := TestClient(t, server.URL)
	service := NewService(client, store, NewMonitoringRepository(nil, true), nil, "https://example.com/callback")

	// Get admin clubs
	adminClubs, err := service.GetAdminClubs(context.Background(), sessionID)
	AssertNoError(t, err)

	// Should only return Portland cycling admin club
	AssertEqual(t, 1, len(adminClubs))
	AssertEqual(t, "Portland Cyclists", adminClubs[0].Name)
	AssertEqual(t, 1, callCount) // Only 1 GetClubDetails call (for Portland Cyclists)
}

func TestService_ProcessRoute_NoSession(t *testing.T) {
	store := NewSessionStore()
	client := NewClient(TestConfig())
	service := NewService(client, store, nil, nil, "https://example.com/callback")

	_, err := service.ProcessRoute(context.Background(), "invalid_session", 123, "pdx")
	AssertError(t, err)
}

func TestService_ProcessRoute_NoRepoConfigured(t *testing.T) {
	store := NewSessionStore()
	session := &Session{
		AccessToken: "test_token",
		AthleteID:   12345,
		CityCode:    "pdx",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	sessionID, _ := store.CreateSession(session)

	client := NewClient(TestConfig())
	service := NewService(client, store, nil, nil, "https://example.com/callback")
	// Note: not calling SetRouteRepository

	_, err := service.ProcessRoute(context.Background(), sessionID, 123, "pdx")
	AssertError(t, err)
	AssertTrue(t, contains(err.Error(), "not configured"), "should error about repo not configured")
}

// Note: Full integration test for ProcessRoute with actual route fetching
// will be done in M3 handler tests with database setup.
// Here we only test the error paths that don't require route repository.

// Helper function for string contains check
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
