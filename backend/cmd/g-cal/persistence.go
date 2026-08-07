package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const googleCalendarSource = "google-calendar"

// syncCalendarEvents writes all fetched Calendar instances in a single
// transaction. Each Calendar instance is represented by one public event and
// one occurrence so the existing ride endpoints can return it unchanged.
func syncCalendarEvents(db *sql.DB, calendarEvents []calendarEvent, calendars []CalendarConfig, lookaheadDays int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin Calendar sync: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	eventStmt, err := tx.Prepare(`
		INSERT INTO events (
			title, tinytitle, description, venue_name, address, is_loop_ride, city,
			organizer_name, web_url, is_published, latitude, longitude, source, source_id
		) VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?, 1, ?, ?, ?, ?)
		ON CONFLICT(source, source_id) WHERE source IS NOT NULL AND source_id IS NOT NULL DO UPDATE SET
			title = excluded.title,
			tinytitle = excluded.tinytitle,
			description = excluded.description,
			venue_name = excluded.venue_name,
			address = excluded.address,
			city = excluded.city,
			organizer_name = excluded.organizer_name,
			web_url = excluded.web_url,
			is_published = excluded.is_published,
			latitude = excluded.latitude,
			longitude = excluded.longitude,
			updated_at = STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')
		RETURNING id
	`)
	if err != nil {
		return fmt.Errorf("prepare event upsert: %w", err)
	}
	defer eventStmt.Close()

	deleteOccurrenceStmt, err := tx.Prepare(`DELETE FROM event_occurrences WHERE event_id = ?`)
	if err != nil {
		return fmt.Errorf("prepare occurrence delete: %w", err)
	}
	defer deleteOccurrenceStmt.Close()

	occurrenceStmt, err := tx.Prepare(`
		INSERT INTO event_occurrences (
			event_id, start_date, start_time, start_datetime,
			event_duration_minutes, event_time_details, is_cancelled
		) VALUES (?, ?, ?, ?, ?, '', ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare occurrence insert: %w", err)
	}
	defer occurrenceStmt.Close()

	syncStmt, err := tx.Prepare(`
		INSERT INTO google_calendar_sync (
			calendar_id, google_event_id, event_id, google_updated_at, last_seen_at
		) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(calendar_id, google_event_id) DO UPDATE SET
			event_id = excluded.event_id,
			google_updated_at = excluded.google_updated_at,
			last_seen_at = excluded.last_seen_at
	`)
	if err != nil {
		return fmt.Errorf("prepare Calendar sync upsert: %w", err)
	}
	defer syncStmt.Close()

	seen := make(map[string]struct{}, len(calendarEvents))
	calendarCities := make(map[string]string, len(calendars))
	for _, calendar := range calendars {
		calendarCities[calendar.ID] = calendar.City
	}
	now := time.Now().UTC().Format(time.RFC3339)
	syncedEvents := 0
	skippedEvents := 0
	for _, calendarEvent := range calendarEvents {
		if calendarEvent.Event == nil || calendarEvent.Event.Id == "" {
			skippedEvents++
			continue
		}
		sourceID := calendarSourceID(calendarEvent.Calendar.ID, calendarEvent.Event.Id)
		if _, duplicate := seen[sourceID]; duplicate {
			skippedEvents++
			continue
		}
		seen[sourceID] = struct{}{}

		startDate, startTime, durationMinutes, err := occurrenceDetails(calendarEvent)
		if err != nil {
			return fmt.Errorf("Calendar event %q: %w", sourceID, err)
		}

		var latitude, longitude any
		if calendarEvent.HasCoordinates {
			latitude = calendarEvent.Latitude
			longitude = calendarEvent.Longitude
		}

		var eventID int64
		organizerName := ""
		if calendarEvent.Event.Organizer != nil {
			organizerName = calendarEvent.Event.Organizer.DisplayName
		}

		err = eventStmt.QueryRow(
			calendarTitle(calendarEvent),
			nilIfEmpty(calendarEvent.Event.Summary),
			calendarEvent.Event.Description,
			nilIfEmpty(calendarEvent.Event.Location),
			nilIfEmpty(calendarEvent.Event.Location),
			calendarEvent.Calendar.City,
			nilIfEmpty(organizerName),
			nilIfEmpty(calendarEvent.Event.HtmlLink),
			latitude,
			longitude,
			googleCalendarSource,
			sourceID,
		).Scan(&eventID)
		if err != nil {
			return fmt.Errorf("upsert Calendar event %q: %w", sourceID, err)
		}

		if _, err := deleteOccurrenceStmt.Exec(eventID); err != nil {
			return fmt.Errorf("replace Calendar occurrence for %q: %w", sourceID, err)
		}
		if _, err := occurrenceStmt.Exec(eventID, startDate, startTime, startDate+" "+startTime, durationMinutes, boolToInt(calendarEvent.Event.Status == "cancelled")); err != nil {
			return fmt.Errorf("insert Calendar occurrence for %q: %w", sourceID, err)
		}
		if _, err := syncStmt.Exec(calendarEvent.Calendar.ID, calendarEvent.Event.Id, eventID, calendarEvent.Event.Updated, now); err != nil {
			return fmt.Errorf("upsert Calendar sync metadata for %q: %w", sourceID, err)
		}
		syncedEvents++
		if syncedEvents%1000 == 0 {
			slog.Info("Turso Google Calendar sync progress", "synced_event_count", syncedEvents, "total_event_count", len(calendarEvents))
		}
	}

	cancelledEvents := int64(0)
	for calendarID, city := range calendarCities {
		startDate, endDate := importDateRangeForCity(city, lookaheadDays)
		cancelledCount, err := cancelUnseenOccurrences(tx, calendarID, now, startDate, endDate)
		if err != nil {
			return fmt.Errorf("reconcile missing Calendar events for %q: %w", calendarID, err)
		}
		if cancelledCount > 0 {
			cancelledEvents += cancelledCount
			slog.Info("marked Calendar events cancelled because they were absent from a completed sync", "calendar_id", calendarID, "count", cancelledCount)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit Calendar sync: %w", err)
	}
	slog.Info("completed Turso Google Calendar sync", "synced_event_count", syncedEvents, "skipped_event_count", skippedEvents, "cancelled_occurrence_count", cancelledEvents)
	return nil
}

// cancelUnseenOccurrences handles an event deleted from Google Calendar within
// the import window. It is called only after all pages have been fetched
// successfully: if Calendar retrieval fails, main exits before
// syncCalendarEvents starts a transaction.
func cancelUnseenOccurrences(tx *sql.Tx, calendarID, syncTime, startDate, endDate string) (int64, error) {
	result, err := tx.Exec(`
		UPDATE event_occurrences
		SET is_cancelled = 1
		WHERE is_cancelled = 0
		  AND start_date BETWEEN ? AND ?
		  AND event_id IN (
			SELECT event_id
			FROM google_calendar_sync
			WHERE calendar_id = ? AND last_seen_at <> ?
		  )
	`, startDate, endDate, calendarID, syncTime)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func importDateRangeForCity(city string, lookaheadDays int) (string, string) {
	timeZone := "America/Los_Angeles"
	if city == "slc" {
		timeZone = "America/Denver"
	}
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		now := time.Now().UTC()
		return now.Format("2006-01-02"), now.AddDate(0, 0, lookaheadDays).Format("2006-01-02")
	}
	now := time.Now().In(location)
	return now.Format("2006-01-02"), now.AddDate(0, 0, lookaheadDays).Format("2006-01-02")
}

func occurrenceDetails(calendarEvent calendarEvent) (string, string, int, error) {
	if calendarEvent.Event.Start == nil {
		return "", "", 0, fmt.Errorf("is missing a start time")
	}
	if calendarEvent.Event.Start.Date != "" {
		return calendarEvent.Event.Start.Date, "00:00", 24 * 60, nil
	}

	start, err := time.Parse(time.RFC3339, calendarEvent.Event.Start.DateTime)
	if err != nil {
		return "", "", 0, fmt.Errorf("has invalid start time %q: %w", calendarEvent.Event.Start.DateTime, err)
	}
	duration := 0
	if calendarEvent.Event.End != nil && calendarEvent.Event.End.DateTime != "" {
		if end, err := time.Parse(time.RFC3339, calendarEvent.Event.End.DateTime); err == nil && end.After(start) {
			duration = int(end.Sub(start).Minutes())
		}
	}
	return start.Format("2006-01-02"), start.Format("15:04"), duration, nil
}

func calendarTitle(event calendarEvent) string {
	if title := strings.TrimSpace(event.Event.Summary); title != "" {
		return title
	}
	return "Untitled ride"
}

func calendarSourceID(calendarID, eventID string) string {
	return calendarID + "|" + eventID
}

func nilIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
