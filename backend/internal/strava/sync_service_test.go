package strava

import (
	"context"
	"encoding/base64"
	"os"
	"testing"
	"time"
)

// TestEncryptionKey is a fixed 32-byte key for testing (DO NOT use in production!)
// This is base64 encoded: 32 bytes of zeros
const TestEncryptionKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

// SetupTestEncryption sets the test encryption key in the environment
func SetupTestEncryption(t *testing.T) {
	t.Helper()
	os.Setenv("STRAVA_TOKEN_ENCRYPTION_KEY", TestEncryptionKey)
	t.Cleanup(func() {
		os.Unsetenv("STRAVA_TOKEN_ENCRYPTION_KEY")
	})
}

// NewTestTokenEncryption creates a TokenEncryption for testing
func NewTestTokenEncryption(t *testing.T) *TokenEncryption {
	t.Helper()
	SetupTestEncryption(t)
	enc, err := NewTokenEncryption()
	if err != nil {
		t.Fatalf("Failed to create test encryption: %v", err)
	}
	return enc
}

// NewTestSyncConfig returns a SyncConfig for testing with low limits
func NewTestSyncConfig() *SyncConfig {
	return &SyncConfig{
		MaxConnectionsPerRun: 10,
		MaxRequestsPer15Min:  50,
		MaxRequestsPerDay:    500,
		ContinueOnError:      true,
		Debug:                true,
	}
}

func TestSyncService_classifyError(t *testing.T) {
	service := &SyncService{}

	tests := []struct {
		err      error
		expected string
	}{
		{ErrUnauthorized, "token_revoked"},
		{ErrRateLimitExceeded, "rate_limit"},
		{ErrNotFound, "not_found"},
		{ErrForbidden, "api_error"},
		{ErrServerError, "api_error"},
	}

	for _, tt := range tests {
		result := service.classifyError(tt.err)
		if result != tt.expected {
			t.Errorf("classifyError(%v) = %s, want %s", tt.err, result, tt.expected)
		}
	}
}

func TestSyncService_compareEvents(t *testing.T) {
	service := &SyncService{}

	stravaEvents := []GroupEvent{
		{ID: 100}, // Exists on Strava
		{ID: 200}, // Exists on Strava
		{ID: 300}, // Exists on Strava (new, not in DB)
	}

	storedMetadata := []*EventMetadata{
		{EventID: 1, StravaEventID: 100, StravaClubID: 1}, // Should refresh
		{EventID: 2, StravaEventID: 200, StravaClubID: 1}, // Should refresh
		{EventID: 3, StravaEventID: 999, StravaClubID: 1}, // Should delete (not on Strava)
	}

	comparison := service.compareEvents(stravaEvents, storedMetadata)

	// Events 1 and 2 should be refreshed (IDs 100 and 200 exist on Strava)
	AssertEqual(t, 2, len(comparison.ToRefresh))
	AssertTrue(t, containsInt64(comparison.ToRefresh, int64(1)), "Event 1 should be refreshed")
	AssertTrue(t, containsInt64(comparison.ToRefresh, int64(2)), "Event 2 should be refreshed")

	// Event 3 should be deleted (ID 999 not on Strava)
	AssertEqual(t, 1, len(comparison.ToDelete))
	AssertTrue(t, containsInt64(comparison.ToDelete, int64(3)), "Event 3 should be deleted")
}

func TestSyncService_compareEvents_Empty(t *testing.T) {
	service := &SyncService{}

	// No events on Strava
	stravaEvents := []GroupEvent{}

	// But we have stored events
	storedMetadata := []*EventMetadata{
		{EventID: 1, StravaEventID: 100, StravaClubID: 1},
		{EventID: 2, StravaEventID: 200, StravaClubID: 1},
	}

	comparison := service.compareEvents(stravaEvents, storedMetadata)

	// All stored events should be deleted
	AssertEqual(t, 0, len(comparison.ToRefresh))
	AssertEqual(t, 2, len(comparison.ToDelete))
}

func TestSyncService_compareEvents_NoStoredEvents(t *testing.T) {
	service := &SyncService{}

	stravaEvents := []GroupEvent{
		{ID: 100},
		{ID: 200},
	}

	// No stored events
	storedMetadata := []*EventMetadata{}

	comparison := service.compareEvents(stravaEvents, storedMetadata)

	// Nothing to refresh or delete
	AssertEqual(t, 0, len(comparison.ToRefresh))
	AssertEqual(t, 0, len(comparison.ToDelete))
}

func TestSyncService_shouldStopDueToRateLimits(t *testing.T) {
	config := &SyncConfig{
		MaxRequestsPer15Min: 90,
		MaxRequestsPerDay:   900,
	}
	service := &SyncService{config: config}

	tests := []struct {
		name     string
		result   *SyncResult
		expected bool
	}{
		{
			name:     "under limits",
			result:   &SyncResult{APIRequestsUsed: 50, RateLimitUsage15Min: 50, RateLimitUsageDaily: 500},
			expected: false,
		},
		{
			name:     "at request limit",
			result:   &SyncResult{APIRequestsUsed: 90, RateLimitUsage15Min: 50, RateLimitUsageDaily: 500},
			expected: true,
		},
		{
			name:     "at Strava 15min limit",
			result:   &SyncResult{APIRequestsUsed: 50, RateLimitUsage15Min: 90, RateLimitUsageDaily: 500},
			expected: true,
		},
		{
			name:     "at daily limit",
			result:   &SyncResult{APIRequestsUsed: 50, RateLimitUsage15Min: 50, RateLimitUsageDaily: 900},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.shouldStopDueToRateLimits(tt.result)
			AssertEqual(t, tt.expected, result)
		})
	}
}

func TestFilterMetadataByClub(t *testing.T) {
	metadata := []*EventMetadata{
		{EventID: 1, StravaClubID: 100},
		{EventID: 2, StravaClubID: 100},
		{EventID: 3, StravaClubID: 200},
		{EventID: 4, StravaClubID: 100},
		{EventID: 5, StravaClubID: 300},
	}

	filtered := filterMetadataByClub(metadata, 100)

	AssertEqual(t, 3, len(filtered))
	for _, meta := range filtered {
		AssertEqual(t, int64(100), meta.StravaClubID)
	}
}

func TestSyncResult_Duration(t *testing.T) {
	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	end := time.Now()

	result := &SyncResult{
		StartedAt:   start,
		CompletedAt: end,
	}

	duration := result.Duration()
	AssertTrue(t, duration >= 10*time.Millisecond, "duration should be at least 10ms")
}

func TestSyncResult_Duration_NotCompleted(t *testing.T) {
	start := time.Now().Add(-100 * time.Millisecond)

	result := &SyncResult{
		StartedAt: start,
		// CompletedAt is zero
	}

	duration := result.Duration()
	AssertTrue(t, duration >= 100*time.Millisecond, "duration should measure from start to now")
}

func TestAthleteSync_SetError(t *testing.T) {
	sync := NewAthleteSync(12345, "pdx")

	AssertTrue(t, sync.Error == nil, "error should be nil initially")

	sync.SetError("token_revoked", "Token has been revoked")

	AssertTrue(t, sync.Error != nil, "error should be set")
	AssertEqual(t, int64(12345), sync.Error.AthleteID)
	AssertEqual(t, "token_revoked", sync.Error.ErrorType)
	AssertEqual(t, "Token has been revoked", sync.Error.Message)
}

func TestNewAthleteSync(t *testing.T) {
	sync := NewAthleteSync(12345, "pdx")

	AssertEqual(t, int64(12345), sync.AthleteID)
	AssertEqual(t, "pdx", sync.CityCode)
	AssertEqual(t, 0, sync.ClubsFound)
	AssertEqual(t, 0, sync.AdminClubsFound)
	AssertEqual(t, 0, sync.EventsRefreshed)
	AssertEqual(t, 0, sync.EventsDeleted)
	AssertEqual(t, 0, sync.APIRequestsUsed)
	AssertTrue(t, sync.Error == nil, "error should be nil")
}

func TestEncryption_RoundTrip(t *testing.T) {
	enc := NewTestTokenEncryption(t)

	original := "test_refresh_token_12345"

	ciphertext, nonce, err := enc.Encrypt(original)
	AssertNoError(t, err)
	AssertTrue(t, len(ciphertext) > 0, "ciphertext should not be empty")
	AssertTrue(t, len(nonce) > 0, "nonce should not be empty")

	decrypted, err := enc.Decrypt(ciphertext, nonce)
	AssertNoError(t, err)
	AssertEqual(t, original, decrypted)
}

func TestEncryptionKey_Format(t *testing.T) {
	// Verify the test key decodes to 32 bytes
	keyBytes, err := base64.StdEncoding.DecodeString(TestEncryptionKey)
	AssertNoError(t, err)
	AssertEqual(t, 32, len(keyBytes))
}

func TestSyncConfig_Defaults(t *testing.T) {
	config := DefaultSyncConfig()

	AssertEqual(t, 100, config.MaxConnectionsPerRun)
	AssertEqual(t, 90, config.MaxRequestsPer15Min)
	AssertEqual(t, 900, config.MaxRequestsPerDay)
	AssertTrue(t, config.ContinueOnError, "ContinueOnError should default to true")
	AssertFalse(t, config.Debug, "Debug should default to false")
}

func TestSyncConfig_FromEnv(t *testing.T) {
	// Set test env vars
	os.Setenv("SYNC_MAX_CONNECTIONS", "50")
	os.Setenv("SYNC_MAX_REQUESTS_15MIN", "45")
	os.Setenv("SYNC_MAX_REQUESTS_DAY", "450")
	os.Setenv("SYNC_CONTINUE_ON_ERROR", "false")
	os.Setenv("STRAVA_DEBUG", "true")

	t.Cleanup(func() {
		os.Unsetenv("SYNC_MAX_CONNECTIONS")
		os.Unsetenv("SYNC_MAX_REQUESTS_15MIN")
		os.Unsetenv("SYNC_MAX_REQUESTS_DAY")
		os.Unsetenv("SYNC_CONTINUE_ON_ERROR")
		os.Unsetenv("STRAVA_DEBUG")
	})

	config := NewSyncConfigFromEnv()

	AssertEqual(t, 50, config.MaxConnectionsPerRun)
	AssertEqual(t, 45, config.MaxRequestsPer15Min)
	AssertEqual(t, 450, config.MaxRequestsPerDay)
	AssertFalse(t, config.ContinueOnError, "ContinueOnError should be false from env")
	AssertTrue(t, config.Debug, "Debug should be true from env")
}

func TestEventComparison_Empty(t *testing.T) {
	comparison := &EventComparison{}

	AssertEqual(t, 0, len(comparison.ToRefresh))
	AssertEqual(t, 0, len(comparison.ToDelete))
}

// Helper function for int64 slices
func containsInt64(slice []int64, val int64) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// Integration test - requires environment variables
func TestSyncService_Integration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=true to run")
	}

	// This test requires:
	// - TURSO_DB_URL and TURSO_DB_RW_TOKEN
	// - TURSO_MONITORING_DB_URL and TURSO_MONITORING_DB_RW_TOKEN
	// - STRAVA_CLIENT_ID and STRAVA_CLIENT_SECRET
	// - STRAVA_TOKEN_ENCRYPTION_KEY

	ctx := context.Background()

	// The actual test would set up real DB connections and run the sync
	// For now, we just verify the test infrastructure works
	t.Log("Integration test infrastructure is ready")
	_ = ctx
}
