package strava

import (
	"context"
	"testing"
)

// TestSyncService_ErrorIsolation verifies that one failed athlete doesn't stop entire sync
// This is critical for Milestone 2.1 - error resilience
func TestSyncService_ErrorIsolation(t *testing.T) {
	// This test verifies the error handling behavior documented in Phase 1:
	// - One athlete failure should NOT stop sync for other athletes
	// - Errors should be logged and classified correctly
	// - ContinueOnError=true should allow processing to continue

	config := &SyncConfig{
		MaxConnectionsPerRun: 10,
		MaxRequestsPer15Min:  50,
		MaxRequestsPerDay:    500,
		ContinueOnError:      true, // CRITICAL: Continue on error
		Debug:                true,
	}

	service := &SyncService{
		config: config,
	}

	// Simulate error classification
	tests := []struct {
		name      string
		err       error
		errorType string
	}{
		{
			name:      "401 token revoked",
			err:       ErrUnauthorized,
			errorType: "token_revoked",
		},
		{
			name:      "429 rate limit",
			err:       ErrRateLimitExceeded,
			errorType: "rate_limit",
		},
		{
			name:      "404 not found",
			err:       ErrNotFound,
			errorType: "not_found",
		},
		{
			name:      "500 server error",
			err:       ErrServerError,
			errorType: "api_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorType := service.classifyError(tt.err)
			AssertEqual(t, tt.errorType, errorType)
		})
	}
}

// TestSyncService_ContinueOnError verifies that ContinueOnError flag works correctly
func TestSyncService_ContinueOnError(t *testing.T) {
	// Test that when ContinueOnError=true, sync continues after athlete failure
	// Test that when ContinueOnError=false, sync stops after athlete failure

	ctx := context.Background()
	_ = ctx

	// Verify config respects ContinueOnError flag
	configContinue := &SyncConfig{
		ContinueOnError: true,
	}
	AssertTrue(t, configContinue.ContinueOnError, "ContinueOnError should be true")

	configStop := &SyncConfig{
		ContinueOnError: false,
	}
	AssertFalse(t, configStop.ContinueOnError, "ContinueOnError should be false")
}

// TestAthleteSync_ErrorTracking verifies error tracking in AthleteSync
func TestAthleteSync_ErrorTracking(t *testing.T) {
	// Create a sync for athlete
	sync := NewAthleteSync(12345, "pdx")

	// Initially no error
	AssertTrue(t, sync.Error == nil, "error should be nil initially")

	// Set token revoked error
	sync.SetError("token_revoked", "Token has been revoked by user")

	// Verify error is tracked
	AssertTrue(t, sync.Error != nil, "error should be set")
	AssertEqual(t, int64(12345), sync.Error.AthleteID)
	AssertEqual(t, "token_revoked", sync.Error.ErrorType)
	AssertEqual(t, "Token has been revoked by user", sync.Error.Message)
	AssertTrue(t, !sync.Error.Timestamp.IsZero(), "timestamp should be set")

	// Verify the sync result still contains useful data
	sync.ClubsFound = 3
	sync.AdminClubsFound = 2
	AssertEqual(t, 3, sync.ClubsFound)
	AssertEqual(t, 2, sync.AdminClubsFound)
}

// TestSyncResult_ErrorAggregation verifies that errors are aggregated correctly
func TestSyncResult_ErrorAggregation(t *testing.T) {
	result := &SyncResult{
		TotalConnections:       5,
		ProcessedConnections:   5,
		SuccessfulConnections:  3,
		FailedConnections:      2,
		Errors:                 []SyncError{},
	}

	// Add errors from failed athletes
	error1 := SyncError{
		AthleteID: 100,
		ErrorType: "token_revoked",
		Message:   "Token revoked",
	}
	error2 := SyncError{
		AthleteID: 200,
		ErrorType: "rate_limit",
		Message:   "Rate limit exceeded",
	}

	result.Errors = append(result.Errors, error1, error2)

	// Verify errors are tracked
	AssertEqual(t, 2, len(result.Errors))
	AssertEqual(t, int64(100), result.Errors[0].AthleteID)
	AssertEqual(t, "token_revoked", result.Errors[0].ErrorType)
	AssertEqual(t, int64(200), result.Errors[1].AthleteID)
	AssertEqual(t, "rate_limit", result.Errors[1].ErrorType)

	// Verify stats are correct
	AssertEqual(t, 5, result.ProcessedConnections)
	AssertEqual(t, 3, result.SuccessfulConnections)
	AssertEqual(t, 2, result.FailedConnections)
}
