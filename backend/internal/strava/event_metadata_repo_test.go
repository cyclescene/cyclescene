package strava

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// Event Metadata Repository Tests
//
// These tests verify the GetUpcomingStravaEventsByAthlete() functionality
// which is CRITICAL for the sync service to correctly identify which events need syncing.
//
// Key behaviors tested:
// 1. ONLY events with source='strava' are returned (detached events excluded)
// 2. ONLY upcoming events are returned (past events excluded)
// 3. Events filtered by athlete_id (only return events imported by specific athlete)
//
// Run with: RUN_INTEGRATION_TESTS=true CGO_ENABLED=1 go test -v ./internal/strava/... -run TestEventMetadataRepository

// setupEventTestDB creates an in-memory SQLite database with full event schema
func setupEventTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("libsql", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create schema with all required tables
	schema := `
		CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			source TEXT,
			source_id TEXT,
			created_at TEXT DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
			updated_at TEXT DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
		);

		CREATE TABLE IF NOT EXISTS event_occurrences (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_id INTEGER NOT NULL,
			start_date TEXT NOT NULL,
			start_time TEXT,
			start_datetime TEXT,
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

// insertTestEvent inserts a test event and returns its ID
func insertTestEvent(t *testing.T, db *sql.DB, title, source string) int64 {
	t.Helper()

	var sourceVal interface{} = source
	if source == "" {
		sourceVal = nil
	}

	result, err := db.Exec(`
		INSERT INTO events (title, source)
		VALUES (?, ?)
	`, title, sourceVal)
	if err != nil {
		t.Fatalf("Failed to insert test event: %v", err)
	}

	id, _ := result.LastInsertId()
	return id
}

// insertTestOccurrence inserts a test occurrence for an event
func insertTestOccurrence(t *testing.T, db *sql.DB, eventID int64, startDate string) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO event_occurrences (event_id, start_date, start_time, start_datetime)
		VALUES (?, ?, '18:00', ? || ' 18:00')
	`, eventID, startDate, startDate)
	if err != nil {
		t.Fatalf("Failed to insert test occurrence: %v", err)
	}
}

// insertTestMetadata inserts test metadata for an event
func insertTestMetadata(t *testing.T, db *sql.DB, eventID, stravaEventID, stravaClubID, athleteID int64) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO strava_event_metadata (
			event_id, strava_event_id, strava_club_id, imported_by_athlete_id,
			imported_at, last_refreshed_at, refresh_count
		) VALUES (?, ?, ?, ?, STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'), STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'), 0)
	`, eventID, stravaEventID, stravaClubID, athleteID)
	if err != nil {
		t.Fatalf("Failed to insert test metadata: %v", err)
	}
}

func TestEventMetadataRepository_GetUpcomingStravaEventsByAthlete_FiltersDetachedEvents(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=true and CGO_ENABLED=1 to run")
	}

	// This test verifies the CRITICAL behavior: events with source != 'strava' are excluded
	db := setupEventTestDB(t)
	repo := NewEventMetadataRepository(db)
	ctx := context.Background()

	athleteID := int64(12345)
	futureDate := time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02")

	// Create a Strava event (should be returned)
	stravaEventID := insertTestEvent(t, db, "Strava Event", "strava")
	insertTestOccurrence(t, db, stravaEventID, futureDate)
	insertTestMetadata(t, db, stravaEventID, 100, 1, athleteID)

	// Create a detached event (source=NULL, should NOT be returned)
	detachedEventID := insertTestEvent(t, db, "Detached Event", "")
	insertTestOccurrence(t, db, detachedEventID, futureDate)
	insertTestMetadata(t, db, detachedEventID, 200, 1, athleteID)

	// Create a native event (source='cyclescene', should NOT be returned)
	nativeEventID := insertTestEvent(t, db, "Native Event", "cyclescene")
	insertTestOccurrence(t, db, nativeEventID, futureDate)
	insertTestMetadata(t, db, nativeEventID, 300, 1, athleteID)

	// Get upcoming events - should only return the Strava event
	events, err := repo.GetUpcomingStravaEventsByAthlete(ctx, athleteID)
	AssertNoError(t, err)

	AssertEqual(t, 1, len(events))
	AssertEqual(t, stravaEventID, events[0].EventID)
	AssertEqual(t, int64(100), events[0].StravaEventID)
}

func TestEventMetadataRepository_GetUpcomingStravaEventsByAthlete_FiltersPastEvents(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=true and CGO_ENABLED=1 to run")
	}

	db := setupEventTestDB(t)
	repo := NewEventMetadataRepository(db)
	ctx := context.Background()

	athleteID := int64(12345)
	futureDate := time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02")
	pastDate := time.Now().Add(-30 * 24 * time.Hour).Format("2006-01-02")

	// Create a future Strava event (should be returned)
	futureEventID := insertTestEvent(t, db, "Future Event", "strava")
	insertTestOccurrence(t, db, futureEventID, futureDate)
	insertTestMetadata(t, db, futureEventID, 100, 1, athleteID)

	// Create a past Strava event (should NOT be returned)
	pastEventID := insertTestEvent(t, db, "Past Event", "strava")
	insertTestOccurrence(t, db, pastEventID, pastDate)
	insertTestMetadata(t, db, pastEventID, 200, 1, athleteID)

	// Get upcoming events - should only return the future event
	events, err := repo.GetUpcomingStravaEventsByAthlete(ctx, athleteID)
	AssertNoError(t, err)

	AssertEqual(t, 1, len(events))
	AssertEqual(t, futureEventID, events[0].EventID)
}

func TestEventMetadataRepository_GetUpcomingStravaEventsByAthlete_FiltersOtherAthletes(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=true and CGO_ENABLED=1 to run")
	}

	db := setupEventTestDB(t)
	repo := NewEventMetadataRepository(db)
	ctx := context.Background()

	athlete1 := int64(12345)
	athlete2 := int64(67890)
	futureDate := time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02")

	// Create an event for athlete 1
	event1ID := insertTestEvent(t, db, "Athlete 1 Event", "strava")
	insertTestOccurrence(t, db, event1ID, futureDate)
	insertTestMetadata(t, db, event1ID, 100, 1, athlete1)

	// Create an event for athlete 2
	event2ID := insertTestEvent(t, db, "Athlete 2 Event", "strava")
	insertTestOccurrence(t, db, event2ID, futureDate)
	insertTestMetadata(t, db, event2ID, 200, 1, athlete2)

	// Get events for athlete 1 - should only return their event
	events, err := repo.GetUpcomingStravaEventsByAthlete(ctx, athlete1)
	AssertNoError(t, err)

	AssertEqual(t, 1, len(events))
	AssertEqual(t, event1ID, events[0].EventID)
	AssertEqual(t, athlete1, events[0].ImportedByAthleteID)
}

func TestEventMetadataRepository_GetUpcomingStravaEventsByAthlete_EmptyResult(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=true and CGO_ENABLED=1 to run")
	}

	db := setupEventTestDB(t)
	repo := NewEventMetadataRepository(db)
	ctx := context.Background()

	// Get events for an athlete with no events
	events, err := repo.GetUpcomingStravaEventsByAthlete(ctx, 99999)
	AssertNoError(t, err)
	AssertEqual(t, 0, len(events))
}

func TestEventMetadataRepository_UpdateLastRefreshed(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=true and CGO_ENABLED=1 to run")
	}

	db := setupEventTestDB(t)
	repo := NewEventMetadataRepository(db)
	ctx := context.Background()

	// Create an event and metadata
	eventID := insertTestEvent(t, db, "Test Event", "strava")
	err := repo.SaveEventMetadata(ctx, eventID, 12345, 100, 999)
	AssertNoError(t, err)

	// Initial refresh count should be 0
	meta, _ := repo.GetEventMetadata(ctx, eventID)
	AssertEqual(t, 0, meta.RefreshCount)

	// Update last refreshed
	err = repo.UpdateLastRefreshed(ctx, eventID)
	AssertNoError(t, err)

	// Refresh count should now be 1
	meta, _ = repo.GetEventMetadata(ctx, eventID)
	AssertEqual(t, 1, meta.RefreshCount)

	// Update again
	err = repo.UpdateLastRefreshed(ctx, eventID)
	AssertNoError(t, err)

	// Refresh count should now be 2
	meta, _ = repo.GetEventMetadata(ctx, eventID)
	AssertEqual(t, 2, meta.RefreshCount)
}

// Unit tests that work without database

func TestEventMetadata_Fields(t *testing.T) {
	meta := EventMetadata{
		EventID:             1,
		StravaEventID:       100,
		StravaClubID:        10,
		ImportedByAthleteID: 12345,
		RefreshCount:        5,
	}

	AssertEqual(t, int64(1), meta.EventID)
	AssertEqual(t, int64(100), meta.StravaEventID)
	AssertEqual(t, int64(10), meta.StravaClubID)
	AssertEqual(t, int64(12345), meta.ImportedByAthleteID)
	AssertEqual(t, 5, meta.RefreshCount)
}

func TestEventMetadataRepository_Constructor(t *testing.T) {
	repo := NewEventMetadataRepository(nil)
	AssertTrue(t, repo != nil, "Repository should be created")
}
