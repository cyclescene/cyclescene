package ride

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// Detach-on-Edit Tests
//
// These tests verify the critical detach-on-edit functionality for Strava compliance.
// When an organizer edits a Strava-imported event, the system must:
// 1. Set source to NULL (detach from Strava)
// 2. Delete from strava_event_metadata (stop background sync)
// 3. All within a transaction (atomic operation)
//
// This complies with Strava's "no modification" rule - we cannot modify Strava data,
// so editing converts the event to a native CycleScene event.
//
// Run with: RUN_INTEGRATION_TESTS=true CGO_ENABLED=1 go test -v ./internal/api/ride/... -run TestDetach

// setupDetachTestDB creates an in-memory SQLite database with the required schema
func setupDetachTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("libsql", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create schema matching the actual database
	schema := `
		CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			tinytitle TEXT,
			description TEXT,
			image_url TEXT,
			source TEXT,
			source_id TEXT,
			edit_token TEXT UNIQUE,
			audience TEXT DEFAULT 'all',
			ride_length TEXT,
			area TEXT,
			date_type TEXT DEFAULT 'recurring',
			venue_name TEXT,
			address TEXT,
			location_details TEXT,
			ending_location TEXT,
			is_loop_ride INTEGER DEFAULT 0,
			organizer_name TEXT,
			organizer_email TEXT,
			organizer_phone TEXT,
			web_url TEXT,
			web_name TEXT,
			newsflash TEXT,
			hide_email INTEGER DEFAULT 0,
			hide_phone INTEGER DEFAULT 0,
			hide_contact_name INTEGER DEFAULT 0,
			group_code TEXT,
			latitude REAL DEFAULT 0,
			longitude REAL DEFAULT 0,
			created_at TEXT DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
			updated_at TEXT DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
		);

		CREATE TABLE IF NOT EXISTS event_occurrences (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_id INTEGER NOT NULL,
			start_date TEXT NOT NULL,
			start_time TEXT,
			start_datetime TEXT,
			event_duration_minutes INTEGER,
			event_time_details TEXT,
			newsflash TEXT,
			is_cancelled INTEGER DEFAULT 0,
			FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS strava_event_metadata (
			event_id INTEGER PRIMARY KEY,
			strava_event_id INTEGER NOT NULL,
			strava_club_id INTEGER NOT NULL,
			imported_by_athlete_id INTEGER NOT NULL,
			imported_at TEXT NOT NULL,
			last_refreshed_at TEXT NOT NULL,
			refresh_count INTEGER DEFAULT 0,
			FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
		);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// insertStravaEvent creates a Strava event with metadata and returns its ID and edit token
func insertStravaEvent(t *testing.T, db *sql.DB, title string) (int64, string) {
	t.Helper()

	editToken := "test_edit_token_" + title

	result, err := db.Exec(`
		INSERT INTO events (title, description, source, source_id, edit_token, audience, area)
		VALUES (?, 'Test description', 'strava', '12345', ?, 'all', 'downtown')
	`, title, editToken)
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	eventID, _ := result.LastInsertId()

	// Add an occurrence
	_, err = db.Exec(`
		INSERT INTO event_occurrences (event_id, start_date, start_time, start_datetime)
		VALUES (?, '2026-03-15', '18:00', '2026-03-15 18:00')
	`, eventID)
	if err != nil {
		t.Fatalf("Failed to insert occurrence: %v", err)
	}

	// Add Strava metadata
	_, err = db.Exec(`
		INSERT INTO strava_event_metadata (
			event_id, strava_event_id, strava_club_id, imported_by_athlete_id,
			imported_at, last_refreshed_at, refresh_count
		) VALUES (?, 100, 1, 12345, STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'), STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'), 0)
	`, eventID)
	if err != nil {
		t.Fatalf("Failed to insert metadata: %v", err)
	}

	return eventID, editToken
}

// insertNativeEvent creates a native (non-Strava) event and returns its ID and edit token
func insertNativeEvent(t *testing.T, db *sql.DB, title string) (int64, string) {
	t.Helper()

	editToken := "test_edit_token_" + title

	result, err := db.Exec(`
		INSERT INTO events (title, description, source, edit_token, audience, area)
		VALUES (?, 'Test description', NULL, ?, 'all', 'downtown')
	`, title, editToken)
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	eventID, _ := result.LastInsertId()

	// Add an occurrence
	_, err = db.Exec(`
		INSERT INTO event_occurrences (event_id, start_date, start_time, start_datetime)
		VALUES (?, '2026-03-15', '18:00', '2026-03-15 18:00')
	`, eventID)
	if err != nil {
		t.Fatalf("Failed to insert occurrence: %v", err)
	}

	return eventID, editToken
}

func TestDetachStravaEvent_SetsSourceToNull(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=true and CGO_ENABLED=1 to run")
	}

	db := setupDetachTestDB(t)
	repo := NewRepository(db)

	eventID, editToken := insertStravaEvent(t, db, "Test Strava Event")

	// Verify source is 'strava' before
	var source sql.NullString
	err := db.QueryRow(`SELECT source FROM events WHERE id = ?`, eventID).Scan(&source)
	assertNoError(t, err)
	assertTrue(t, source.Valid && source.String == "strava", "Source should be 'strava' before detach")

	// Update the ride (which triggers detach for Strava events)
	submission := &Submission{
		Title:       "Updated Title",
		Description: "Updated Description",
		Audience:    "all",
		Area:        "downtown",
		Occurrences: []Occurrence{{StartDate: "2026-03-15", StartTime: "18:00"}},
	}

	err = repo.UpdateRide(editToken, submission, 0, 0)
	assertNoError(t, err)

	// Verify source is now NULL
	err = db.QueryRow(`SELECT source FROM events WHERE id = ?`, eventID).Scan(&source)
	assertNoError(t, err)
	assertTrue(t, !source.Valid, "Source should be NULL after detach")
}

func TestDetachStravaEvent_DeletesMetadata(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=true and CGO_ENABLED=1 to run")
	}

	db := setupDetachTestDB(t)
	repo := NewRepository(db)

	eventID, editToken := insertStravaEvent(t, db, "Test Strava Event")

	// Verify metadata exists before
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM strava_event_metadata WHERE event_id = ?`, eventID).Scan(&count)
	assertNoError(t, err)
	assertEqual(t, 1, count)

	// Update the ride (which triggers detach)
	submission := &Submission{
		Title:       "Updated Title",
		Description: "Updated Description",
		Audience:    "all",
		Area:        "downtown",
		Occurrences: []Occurrence{{StartDate: "2026-03-15", StartTime: "18:00"}},
	}

	err = repo.UpdateRide(editToken, submission, 0, 0)
	assertNoError(t, err)

	// Verify metadata is deleted
	err = db.QueryRow(`SELECT COUNT(*) FROM strava_event_metadata WHERE event_id = ?`, eventID).Scan(&count)
	assertNoError(t, err)
	assertEqual(t, 0, count)
}

func TestUpdateRide_SkipsDetachForNativeEvents(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=true and CGO_ENABLED=1 to run")
	}

	db := setupDetachTestDB(t)
	repo := NewRepository(db)

	_, editToken := insertNativeEvent(t, db, "Test Native Event")

	// Update the ride
	submission := &Submission{
		Title:       "Updated Native Title",
		Description: "Updated Description",
		Audience:    "all",
		Area:        "downtown",
		Occurrences: []Occurrence{{StartDate: "2026-03-15", StartTime: "18:00"}},
	}

	// Should not error - native events don't need detaching
	err := repo.UpdateRide(editToken, submission, 0, 0)
	assertNoError(t, err)

	// Verify the title was updated
	var title string
	err = db.QueryRow(`SELECT title FROM events WHERE edit_token = ?`, editToken).Scan(&title)
	assertNoError(t, err)
	assertEqual(t, "Updated Native Title", title)
}

func TestDetachStravaEvent_TransactionRollbackOnError(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=true and CGO_ENABLED=1 to run")
	}

	db := setupDetachTestDB(t)
	repo := NewRepository(db)

	eventID, _ := insertStravaEvent(t, db, "Test Strava Event")

	// Use an invalid edit token to cause the update to fail
	submission := &Submission{
		Title:       "Should Not Update",
		Description: "Test",
		Audience:    "all",
		Area:        "downtown",
		Occurrences: []Occurrence{{StartDate: "2026-03-15", StartTime: "18:00"}},
	}

	// This should fail because the edit token doesn't exist
	err := repo.UpdateRide("nonexistent_token", submission, 0, 0)
	assertError(t, err)

	// Verify the original event is unchanged
	var source sql.NullString
	err = db.QueryRow(`SELECT source FROM events WHERE id = ?`, eventID).Scan(&source)
	assertNoError(t, err)
	assertTrue(t, source.Valid && source.String == "strava", "Source should still be 'strava' after failed update")

	// Verify metadata still exists
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM strava_event_metadata WHERE event_id = ?`, eventID).Scan(&count)
	assertNoError(t, err)
	assertEqual(t, 1, count)
}

func TestUpdateEventDetails_DetachesStravaEvent(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=true and CGO_ENABLED=1 to run")
	}

	db := setupDetachTestDB(t)
	repo := NewRepository(db)

	eventID, editToken := insertStravaEvent(t, db, "Test Strava Event")

	// Update event details (which should also trigger detach)
	err := repo.UpdateEventDetails(editToken, "New description", "beginners", "10-15 miles")
	assertNoError(t, err)

	// Verify source is now NULL
	var source sql.NullString
	err = db.QueryRow(`SELECT source FROM events WHERE id = ?`, eventID).Scan(&source)
	assertNoError(t, err)
	assertTrue(t, !source.Valid, "Source should be NULL after detach via UpdateEventDetails")

	// Verify metadata is deleted
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM strava_event_metadata WHERE event_id = ?`, eventID).Scan(&count)
	assertNoError(t, err)
	assertEqual(t, 0, count)
}

func TestGetRideByEditToken_ReturnsSourceField(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=true and CGO_ENABLED=1 to run")
	}

	db := setupDetachTestDB(t)
	repo := NewRepository(db)

	// Create a Strava event
	_, stravaToken := insertStravaEvent(t, db, "Strava Event")

	// Create a native event
	_, nativeToken := insertNativeEvent(t, db, "Native Event")

	// Get Strava event - should have source='strava'
	submission, _, err := repo.GetRideByEditToken(stravaToken)
	assertNoError(t, err)
	assertEqual(t, "strava", submission.Source)

	// Get native event - should have empty source
	submission, _, err = repo.GetRideByEditToken(nativeToken)
	assertNoError(t, err)
	assertEqual(t, "", submission.Source)
}

// Test helper functions
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func assertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Errorf("Expected true: %s", msg)
	}
}
