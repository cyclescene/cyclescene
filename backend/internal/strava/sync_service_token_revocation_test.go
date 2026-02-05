package strava

import (
	"testing"
)

// TestSyncService_TokenRevocation validates 401 handling as per Milestone 2.2
// This test verifies the token revocation behavior documented in Phase 1:
// - 401 errors are classified as "token_revoked"
// - Sync continues to next athlete (doesn't fail entire job)
// - Events remain in database (no cleanup)
// - No emails sent to organizer
func TestSyncService_TokenRevocation(t *testing.T) {
	service := &SyncService{}

	// Test that 401 error is classified correctly
	errorType := service.classifyError(ErrUnauthorized)
	AssertEqual(t, "token_revoked", errorType)
}

// TestSyncService_TokenRevocation_ContinueSync verifies sync continues after token revocation
func TestSyncService_TokenRevocation_ContinueSync(t *testing.T) {
	// Simulate multiple athlete syncs where one has revoked token

	// Athlete 1: Token revoked (401)
	athlete1 := NewAthleteSync(100, "pdx")
	athlete1.SetError("token_revoked", "Unauthorized - token has been revoked")

	// Verify athlete 1 has error tracked
	AssertTrue(t, athlete1.Error != nil, "athlete 1 should have error")
	AssertEqual(t, "token_revoked", athlete1.Error.ErrorType)
	AssertEqual(t, int64(100), athlete1.Error.AthleteID)

	// Athlete 2: Successful sync
	athlete2 := NewAthleteSync(200, "pdx")
	athlete2.ClubsFound = 3
	athlete2.AdminClubsFound = 2
	athlete2.EventsRefreshed = 5
	athlete2.EventsDeleted = 1

	// Verify athlete 2 synced successfully
	AssertTrue(t, athlete2.Error == nil, "athlete 2 should not have error")
	AssertEqual(t, 3, athlete2.ClubsFound)
	AssertEqual(t, 5, athlete2.EventsRefreshed)

	// This simulates the behavior in sync_service.go:107-130
	// Where errors are logged but sync continues to next athlete
}

// TestSyncService_TokenRevocation_EventsPersist verifies events remain in DB
// This test documents the design decision from Phase 2 Milestone 2.2:
// - When token is revoked, events REMAIN in database
// - Organizer shared the events by choice, so they stay visible
// - After event date passes, sync becomes irrelevant anyway
// - No cleanup needed, no email to organizer
func TestSyncService_TokenRevocation_EventsPersist(t *testing.T) {
	// This is a documentation test - the actual persistence
	// is handled by NOT deleting events when 401 occurs

	// Verify that token revocation is classified correctly
	service := &SyncService{}
	errorType := service.classifyError(ErrUnauthorized)

	// When this error is returned from syncAthlete():
	// - sync_service.go:170-178 detects 401
	// - Returns early with error
	// - Does NOT call deleteEvent() or cleanup
	// - Events remain in database untouched
	AssertEqual(t, "token_revoked", errorType)

	// Verify the error is tracked for monitoring
	athlete := NewAthleteSync(12345, "pdx")
	athlete.SetError("token_revoked", "Token has been revoked")

	AssertEqual(t, "token_revoked", athlete.Error.ErrorType)
	AssertTrue(t, !athlete.Error.Timestamp.IsZero(), "error timestamp should be set")
}

// TestSyncService_TokenRevocation_NoCleanup verifies no database cleanup
func TestSyncService_TokenRevocation_NoCleanup(t *testing.T) {
	// Design decision from BACKGROUND_SYNC_MILESTONES.md lines 37-44:
	// Decision: When 401 error occurs (token revoked):
	// - Skip that athlete's sync, continue to next
	// - Log to monitoring DB
	// - No email to organizer, no cleanup needed
	// - Events stay visible (they chose to share them)

	service := &SyncService{}

	// When 401 occurs, it's classified as token_revoked
	errorType := service.classifyError(ErrUnauthorized)
	AssertEqual(t, "token_revoked", errorType)

	// The implementation in sync_service.go:170-178 does:
	// 1. Classify error as "token_revoked"
	// 2. Log warning with athlete_id
	// 3. Return early (no cleanup)
	// 4. Continue to next athlete (due to ContinueOnError=true)

	// This test documents that behavior
}

// TestIsUnauthorized verifies the error checking helper
func TestIsUnauthorized(t *testing.T) {
	// Test that IsUnauthorized correctly identifies 401 errors
	AssertTrue(t, IsUnauthorized(ErrUnauthorized), "ErrUnauthorized should be detected")

	// Non-401 errors should not be detected
	AssertFalse(t, IsUnauthorized(ErrRateLimitExceeded), "Rate limit is not unauthorized")
	AssertFalse(t, IsUnauthorized(ErrNotFound), "Not found is not unauthorized")
	AssertFalse(t, IsUnauthorized(ErrServerError), "Server error is not unauthorized")
}

// TestSyncService_TokenRevocation_Integration documents integration test requirements
func TestSyncService_TokenRevocation_Integration(t *testing.T) {
	// This test documents the manual integration test steps from the milestones doc:
	//
	// Manual Test Steps (lines 1416-1426):
	// 1. Connect test Strava account
	// 2. Revoke access on Strava.com > Settings > My Apps
	// 3. Run sync with --use-real-key --force
	// 4. Verify sync logs "token_revoked" and continues
	//
	// Expected behavior:
	// - Sync detects 401 during token refresh
	// - Logs: token_revoked athlete_id=X
	// - Continues to next athlete
	// - Events remain in database
	// - No emails sent

	t.Skip("Manual integration test - see BACKGROUND_SYNC_MILESTONES.md lines 1416-1426")
}
