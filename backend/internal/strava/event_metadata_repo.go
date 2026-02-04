package strava

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// EventMetadataRepository manages Strava event metadata for background sync
type EventMetadataRepository struct {
	db *sql.DB
}

// NewEventMetadataRepository creates a new event metadata repository
func NewEventMetadataRepository(db *sql.DB) *EventMetadataRepository {
	return &EventMetadataRepository{db: db}
}

// EventMetadata represents Strava metadata for an imported event
type EventMetadata struct {
	EventID             int64
	StravaEventID       int64
	StravaClubID        int64
	ImportedByAthleteID int64
	ImportedAt          time.Time
	LastRefreshedAt     time.Time
	RefreshCount        int
}

// SaveEventMetadata stores Strava metadata for a CycleScene event
func (r *EventMetadataRepository) SaveEventMetadata(
	ctx context.Context,
	cycleSceneEventID int64,
	stravaEventID int64,
	stravaClubID int64,
	athleteID int64,
) error {
	query := `
		INSERT INTO strava_event_metadata (
			event_id,
			strava_event_id,
			strava_club_id,
			imported_by_athlete_id,
			imported_at,
			last_refreshed_at
		) VALUES (?, ?, ?, ?, STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'), STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
	`

	_, err := r.db.ExecContext(ctx, query, cycleSceneEventID, stravaEventID, stravaClubID, athleteID)
	if err != nil {
		return fmt.Errorf("failed to save event metadata: %w", err)
	}

	return nil
}

// GetEventMetadata retrieves metadata for a CycleScene event
func (r *EventMetadataRepository) GetEventMetadata(ctx context.Context, cycleSceneEventID int64) (*EventMetadata, error) {
	query := `
		SELECT
			event_id,
			strava_event_id,
			strava_club_id,
			imported_by_athlete_id,
			imported_at,
			last_refreshed_at,
			refresh_count
		FROM strava_event_metadata
		WHERE event_id = ?
	`

	var meta EventMetadata
	var importedAtStr, lastRefreshedAtStr string

	err := r.db.QueryRowContext(ctx, query, cycleSceneEventID).Scan(
		&meta.EventID,
		&meta.StravaEventID,
		&meta.StravaClubID,
		&meta.ImportedByAthleteID,
		&importedAtStr,
		&lastRefreshedAtStr,
		&meta.RefreshCount,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not a Strava event
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query event metadata: %w", err)
	}

	// Parse timestamps
	meta.ImportedAt, _ = time.Parse("2006-01-02 15:04:05.000", importedAtStr)
	meta.LastRefreshedAt, _ = time.Parse("2006-01-02 15:04:05.000", lastRefreshedAtStr)

	return &meta, nil
}

// GetMetadataByAthleteID retrieves all events imported by a specific athlete
func (r *EventMetadataRepository) GetMetadataByAthleteID(ctx context.Context, athleteID int64) ([]*EventMetadata, error) {
	query := `
		SELECT
			event_id,
			strava_event_id,
			strava_club_id,
			imported_by_athlete_id,
			imported_at,
			last_refreshed_at,
			refresh_count
		FROM strava_event_metadata
		WHERE imported_by_athlete_id = ?
		ORDER BY imported_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, athleteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query metadata by athlete: %w", err)
	}
	defer rows.Close()

	var metadataList []*EventMetadata
	for rows.Next() {
		var meta EventMetadata
		var importedAtStr, lastRefreshedAtStr string

		err := rows.Scan(
			&meta.EventID,
			&meta.StravaEventID,
			&meta.StravaClubID,
			&meta.ImportedByAthleteID,
			&importedAtStr,
			&lastRefreshedAtStr,
			&meta.RefreshCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metadata: %w", err)
		}

		// Parse timestamps
		meta.ImportedAt, _ = time.Parse("2006-01-02 15:04:05.000", importedAtStr)
		meta.LastRefreshedAt, _ = time.Parse("2006-01-02 15:04:05.000", lastRefreshedAtStr)

		metadataList = append(metadataList, &meta)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metadata: %w", err)
	}

	return metadataList, nil
}

// UpdateLastRefreshed updates the refresh timestamp (for background sync)
func (r *EventMetadataRepository) UpdateLastRefreshed(ctx context.Context, cycleSceneEventID int64) error {
	query := `
		UPDATE strava_event_metadata
		SET last_refreshed_at = STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'),
		    refresh_count = refresh_count + 1
		WHERE event_id = ?
	`

	_, err := r.db.ExecContext(ctx, query, cycleSceneEventID)
	if err != nil {
		return fmt.Errorf("failed to update last_refreshed_at: %w", err)
	}

	return nil
}

// DeleteEventMetadata removes metadata for a CycleScene event
func (r *EventMetadataRepository) DeleteEventMetadata(ctx context.Context, cycleSceneEventID int64) error {
	query := `DELETE FROM strava_event_metadata WHERE event_id = ?`

	_, err := r.db.ExecContext(ctx, query, cycleSceneEventID)
	if err != nil {
		return fmt.Errorf("failed to delete event metadata: %w", err)
	}

	return nil
}

// GetUpcomingStravaEventsByAthlete returns upcoming Strava events imported by an athlete
// CRITICAL: Only returns events with source='strava' (not detached events)
func (r *EventMetadataRepository) GetUpcomingStravaEventsByAthlete(ctx context.Context, athleteID int64) ([]*EventMetadata, error) {
	query := `
		SELECT
			sem.event_id,
			sem.strava_event_id,
			sem.strava_club_id,
			sem.imported_by_athlete_id,
			sem.imported_at,
			sem.last_refreshed_at,
			sem.refresh_count
		FROM strava_event_metadata sem
		INNER JOIN events e ON e.id = sem.event_id
		WHERE sem.imported_by_athlete_id = ?
		  AND e.source = 'strava'          -- CRITICAL: Skip detached events
		  AND e.date >= date('now')        -- Only upcoming events
		ORDER BY e.date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, athleteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query upcoming events for athlete: %w", err)
	}
	defer rows.Close()

	var metadataList []*EventMetadata
	for rows.Next() {
		var meta EventMetadata
		var importedAtStr, lastRefreshedAtStr string

		err := rows.Scan(
			&meta.EventID,
			&meta.StravaEventID,
			&meta.StravaClubID,
			&meta.ImportedByAthleteID,
			&importedAtStr,
			&lastRefreshedAtStr,
			&meta.RefreshCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event metadata: %w", err)
		}

		// Parse timestamps
		meta.ImportedAt, _ = time.Parse("2006-01-02 15:04:05.000", importedAtStr)
		meta.LastRefreshedAt, _ = time.Parse("2006-01-02 15:04:05.000", lastRefreshedAtStr)

		metadataList = append(metadataList, &meta)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event metadata: %w", err)
	}

	return metadataList, nil
}

// ListStaleEvents returns events that haven't been refreshed in over 7 days
func (r *EventMetadataRepository) ListStaleEvents(ctx context.Context) ([]*EventMetadata, error) {
	query := `
		SELECT
			event_id,
			strava_event_id,
			strava_club_id,
			imported_by_athlete_id,
			imported_at,
			last_refreshed_at,
			refresh_count
		FROM strava_event_metadata
		WHERE last_refreshed_at < datetime('now', '-7 days')
		ORDER BY last_refreshed_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query stale events: %w", err)
	}
	defer rows.Close()

	var metadataList []*EventMetadata
	for rows.Next() {
		var meta EventMetadata
		var importedAtStr, lastRefreshedAtStr string

		err := rows.Scan(
			&meta.EventID,
			&meta.StravaEventID,
			&meta.StravaClubID,
			&meta.ImportedByAthleteID,
			&importedAtStr,
			&lastRefreshedAtStr,
			&meta.RefreshCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stale event: %w", err)
		}

		// Parse timestamps
		meta.ImportedAt, _ = time.Parse("2006-01-02 15:04:05.000", importedAtStr)
		meta.LastRefreshedAt, _ = time.Parse("2006-01-02 15:04:05.000", lastRefreshedAtStr)

		metadataList = append(metadataList, &meta)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stale events: %w", err)
	}

	return metadataList, nil
}
