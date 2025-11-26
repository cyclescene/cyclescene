package ride

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// User-submitted rides
func (r *Repository) CreateRide(submission *Submission, editToken string, latitude, longitude float64) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	result, err := tx.Exec(`
		INSERT INTO events (
			title, tinytitle, description, image_url, audience, ride_length, area, date_type,
			venue_name, address, location_details, ending_location, is_loop_ride,
			organizer_name, organizer_email, organizer_phone, web_url, web_name, newsflash,
			hide_email, hide_phone, hide_contact_name, group_code, edit_token, city, is_published,
			latitude, longitude
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
	`,
		submission.Title, submission.TinyTitle, submission.Description, submission.ImageURL,
		submission.Audience, submission.RideLength, submission.Area, submission.DateType,
		submission.VenueName, submission.Address, submission.LocationDetails, submission.EndingLocation,
		boolToInt(submission.IsLoopRide), submission.OrganizerName, submission.OrganizerEmail,
		submission.OrganizerPhone, submission.WebURL, submission.WebName, submission.Newsflash,
		boolToInt(submission.HideEmail), boolToInt(submission.HidePhone), boolToInt(submission.HideContactName),
		nilIfEmpty(submission.GroupCode), editToken, submission.City, latitude, longitude,
	)

	if err != nil {
		return 0, err
	}

	eventID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert occurrences
	for _, occ := range submission.Occurrences {
		startDatetime := fmt.Sprintf("%s %s", occ.StartDate, occ.StartTime)
		_, err = tx.Exec(`
			INSERT INTO event_occurrences (
				event_id, start_date, start_time, start_datetime,
				event_duration_minutes, event_time_details, newsflash
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`, eventID, occ.StartDate, occ.StartTime, startDatetime, occ.EventDurationMinutes, occ.EventTimeDetails, occ.Newsflash)

		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return eventID, nil
}

func (r *Repository) GetRideByEditToken(token string) (*Submission, bool, error) {
	row := r.db.QueryRow(`
		SELECT id, title, tinytitle, description, image_url, audience, ride_length, area, date_type,
			   venue_name, address, location_details, ending_location, is_loop_ride,
			   organizer_name, organizer_email, organizer_phone, web_url, web_name, newsflash,
			   hide_email, hide_phone, hide_contact_name, group_code, city, is_published
		FROM events WHERE edit_token = ?
	`, token)

	var submission Submission
	var id int64
	var isLoopRide, hideEmail, hidePhone, hideContactName, isPublished int
	var groupCode sql.NullString

	err := row.Scan(
		&id, &submission.Title, &submission.TinyTitle, &submission.Description, &submission.ImageURL,
		&submission.Audience, &submission.RideLength, &submission.Area, &submission.DateType,
		&submission.VenueName, &submission.Address, &submission.LocationDetails, &submission.EndingLocation, &isLoopRide,
		&submission.OrganizerName, &submission.OrganizerEmail, &submission.OrganizerPhone,
		&submission.WebURL, &submission.WebName, &submission.Newsflash,
		&hideEmail, &hidePhone, &hideContactName, &groupCode, &submission.City, &isPublished,
	)

	if err != nil {
		return nil, false, err
	}

	submission.IsLoopRide = isLoopRide == 1
	submission.HideEmail = hideEmail == 1
	submission.HidePhone = hidePhone == 1
	submission.HideContactName = hideContactName == 1
	if groupCode.Valid {
		submission.GroupCode = groupCode.String
	}

	// Get occurrences
	rows, err := r.db.Query(`
		SELECT id, start_date, start_time, event_duration_minutes, event_time_details, is_cancelled, newsflash
		FROM event_occurrences WHERE event_id = ?
		ORDER BY start_datetime ASC
	`, id)

	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var occ Occurrence
		var isCancelled int
		if err := rows.Scan(&occ.ID, &occ.StartDate, &occ.StartTime, &occ.EventDurationMinutes, &occ.EventTimeDetails, &isCancelled, &occ.Newsflash); err != nil {
			continue
		}
		occ.IsCancelled = isCancelled == 1
		submission.Occurrences = append(submission.Occurrences, occ)
	}

	// Generate SrcSet if image is present and is an optimized WebP
	if submission.ImageURL != "" && strings.HasSuffix(submission.ImageURL, "_optimized.webp") {
		base := strings.TrimSuffix(submission.ImageURL, "_optimized.webp")
		submission.ImageSrcSet = fmt.Sprintf("%s_400w.webp 400w, %s_800w.webp 800w, %s_1200w.webp 1200w", base, base, base)
	}

	return &submission, isPublished == 1, nil
}

func (r *Repository) UpdateRide(token string, submission *Submission, latitude, longitude float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	result, err := tx.Exec(`
		UPDATE events SET
			title = ?, tinytitle = ?, description = ?, image_url = ?,
			audience = ?, ride_length = ?, area = ?, date_type = ?,
			venue_name = ?, address = ?, location_details = ?, ending_location = ?, is_loop_ride = ?,
			organizer_name = ?, organizer_email = ?, organizer_phone = ?,
			web_url = ?, web_name = ?, newsflash = ?,
			hide_email = ?, hide_phone = ?, hide_contact_name = ?,
			group_code = ?, latitude = ?, longitude = ?, updated_at = STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')
		WHERE edit_token = ?
	`,
		submission.Title, submission.TinyTitle, submission.Description, submission.ImageURL,
		submission.Audience, submission.RideLength, submission.Area, submission.DateType,
		submission.VenueName, submission.Address, submission.LocationDetails, submission.EndingLocation,
		boolToInt(submission.IsLoopRide), submission.OrganizerName, submission.OrganizerEmail,
		submission.OrganizerPhone, submission.WebURL, submission.WebName, submission.Newsflash,
		boolToInt(submission.HideEmail), boolToInt(submission.HidePhone), boolToInt(submission.HideContactName),
		nilIfEmpty(submission.GroupCode), latitude, longitude, token,
	)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	// Get event ID
	var eventID int64
	err = tx.QueryRow(`SELECT id FROM events WHERE edit_token = ?`, token).Scan(&eventID)
	if err != nil {
		return err
	}

	// Delete existing occurrences
	_, err = tx.Exec(`DELETE FROM event_occurrences WHERE event_id = ?`, eventID)
	if err != nil {
		return err
	}

	// Insert new occurrences
	for _, occ := range submission.Occurrences {
		startDatetime := fmt.Sprintf("%s %s", occ.StartDate, occ.StartTime)
		_, err = tx.Exec(`
			INSERT INTO event_occurrences (
				event_id, start_date, start_time, start_datetime,
				event_duration_minutes, event_time_details, newsflash
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`, eventID, occ.StartDate, occ.StartTime, startDatetime, occ.EventDurationMinutes, occ.EventTimeDetails, occ.Newsflash)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateOccurrence updates a single occurrence's time, details, and newsflash
func (r *Repository) UpdateOccurrence(token string, occurrenceID int64, startTime string, eventDurationMinutes int, eventTimeDetails string, newsflash string, isCancelled bool) error {
	_, err := r.db.Exec(`
		UPDATE event_occurrences
		SET start_time = ?, event_duration_minutes = ?, event_time_details = ?, newsflash = ?, is_cancelled = ?
		WHERE id = ? AND event_id = (SELECT id FROM events WHERE edit_token = ?)
	`, startTime, eventDurationMinutes, eventTimeDetails, newsflash, boolToInt(isCancelled), occurrenceID, token)
	return err
}

// Scraped rides from Shift2Bikes (PDX only) + Published user-submitted rides
func (r *Repository) GetUpcomingRides(city string) ([]ScrapedRideFromDB, error) {
	tzStr := getTimeZone(city)
	tz, err := time.LoadLocation(tzStr)
	if err != nil {
		slog.Error("failed to load timezone", "error", err.Error())
		return nil, err
	}
	now := time.Now().In(tz)
	todayStr := now.Format("2006-01-02")

	var query string
	var args []any

	// For PDX, include both shift2bikes_events and user-submitted events
	// For other cities, only include user-submitted events
	if city == "pdx" {
		query = `
			SELECT composite_event_id, title, lat, lng, address, audience, cancelled, date, starttime,
			       safetyplan, details, venue, organizer, loopride, shareable, ridesource, endtime,
			       email, eventduration, image, locdetails, locend, newsflash, timedetails, webname, weburl
			FROM shift2bikes_events
			WHERE citycode = ? AND date >= ?
			UNION ALL
			SELECT
				CAST(e.id AS TEXT) as composite_event_id,
				e.title,
				e.latitude as lat,
				e.longitude as lng,
				e.address,
				e.audience,
				eo.is_cancelled as cancelled,
				eo.start_date as date,
				eo.start_time as starttime,
				0 as safetyplan,
				e.description as details,
				e.venue_name as venue,
				e.organizer_name as organizer,
				e.is_loop_ride as loopride,
				'' as shareable,
				'user-submitted' as ridesource,
				'' as endtime,
				e.organizer_email as email,
				eo.event_duration_minutes as eventduration,
				e.image_url as image,
				e.location_details as locdetails,
				e.ending_location as locend,
				e.newsflash,
				eo.event_time_details as timedetails,
				e.web_name as webname,
				e.web_url as weburl
			FROM events e
			JOIN event_occurrences eo ON e.id = eo.event_id
			WHERE e.city = ? AND e.is_published = 1 AND eo.start_date >= ?
			ORDER BY date ASC, starttime ASC
		`
		args = []any{city, todayStr, city, todayStr}
	} else {
		// For non-PDX cities, only query user-submitted events
		query = `
			SELECT
				CAST(e.id AS TEXT) as composite_event_id,
				e.title,
				e.latitude as lat,
				e.longitude as lng,
				e.address,
				e.audience,
				eo.is_cancelled as cancelled,
				eo.start_date as date,
				eo.start_time as starttime,
				0 as safetyplan,
				e.description as details,
				e.venue_name as venue,
				e.organizer_name as organizer,
				e.is_loop_ride as loopride,
				'' as shareable,
				'user-submitted' as ridesource,
				'' as endtime,
				e.organizer_email as email,
				eo.event_duration_minutes as eventduration,
				e.image_url as image,
				e.location_details as locdetails,
				e.ending_location as locend,
				e.newsflash,
				eo.event_time_details as timedetails,
				e.web_name as webname,
				e.web_url as weburl
			FROM events e
			JOIN event_occurrences eo ON e.id = eo.event_id
			WHERE e.city = ? AND e.is_published = 1 AND eo.start_date >= ?
			ORDER BY date ASC, starttime ASC
		`
		args = []any{city, todayStr}
	}

	return r.scanScrapedRides(query, args...)
}

func (r *Repository) GetPastRides(city string) ([]ScrapedRideFromDB, error) {
	tzStr := getTimeZone(city)
	tz, err := time.LoadLocation(tzStr)
	if err != nil {
		slog.Error("failed to load timezone", "error", err.Error())
		return nil, err
	}
	now := time.Now().In(tz)
	sevenDaysAgo := now.AddDate(0, 0, -7)
	todayStr := now.Format("2006-01-02")
	sevenDaysAgoStr := sevenDaysAgo.Format("2006-01-02")

	var query string
	var args []any

	// For PDX, include both shift2bikes_events and user-submitted events
	// For other cities, only include user-submitted events
	if city == "pdx" {
		query = `
			SELECT composite_event_id, title, lat, lng, address, audience, cancelled, date, starttime,
			       safetyplan, details, venue, organizer, loopride, shareable, ridesource, endtime,
			       email, eventduration, image, locdetails, locend, newsflash, timedetails, webname, weburl
			FROM shift2bikes_events
			WHERE citycode = ? AND date BETWEEN ? AND ?
			UNION ALL
			SELECT
				CAST(e.id AS TEXT) as composite_event_id,
				e.title,
				e.latitude as lat,
				e.longitude as lng,
				e.address,
				e.audience,
				eo.is_cancelled as cancelled,
				eo.start_date as date,
				eo.start_time as starttime,
				0 as safetyplan,
				e.description as details,
				e.venue_name as venue,
				e.organizer_name as organizer,
				e.is_loop_ride as loopride,
				'' as shareable,
				'user-submitted' as ridesource,
				'' as endtime,
				e.organizer_email as email,
				eo.event_duration_minutes as eventduration,
				e.image_url as image,
				e.location_details as locdetails,
				e.ending_location as locend,
				e.newsflash,
				eo.event_time_details as timedetails,
				e.web_name as webname,
				e.web_url as weburl
			FROM events e
			JOIN event_occurrences eo ON e.id = eo.event_id
			WHERE e.city = ? AND e.is_published = 1 AND eo.start_date BETWEEN ? AND ?
			ORDER BY date DESC, starttime DESC
		`
		args = []any{city, sevenDaysAgoStr, todayStr, city, sevenDaysAgoStr, todayStr}
	} else {
		// For non-PDX cities, only query user-submitted events
		query = `
			SELECT
				CAST(e.id AS TEXT) as composite_event_id,
				e.title,
				e.latitude as lat,
				e.longitude as lng,
				e.address,
				e.audience,
				eo.is_cancelled as cancelled,
				eo.start_date as date,
				eo.start_time as starttime,
				0 as safetyplan,
				e.description as details,
				e.venue_name as venue,
				e.organizer_name as organizer,
				e.is_loop_ride as loopride,
				'' as shareable,
				'user-submitted' as ridesource,
				'' as endtime,
				e.organizer_email as email,
				eo.event_duration_minutes as eventduration,
				e.image_url as image,
				e.location_details as locdetails,
				e.ending_location as locend,
				e.newsflash,
				eo.event_time_details as timedetails,
				e.web_name as webname,
				e.web_url as weburl
			FROM events e
			JOIN event_occurrences eo ON e.id = eo.event_id
			WHERE e.city = ? AND e.is_published = 1 AND eo.start_date BETWEEN ? AND ?
			ORDER BY date DESC, starttime DESC
		`
		args = []any{city, sevenDaysAgoStr, todayStr}
	}

	return r.scanScrapedRides(query, args...)
}

func (r *Repository) GetRide(city, rideID string) ([]ScrapedRideFromDB, error) {
	query := `
		SELECT composite_event_id, title, lat, lng, address, audience, cancelled, date, starttime, 
		       safetyplan, details, venue, organizer, loopride, shareable, ridesource, endtime, 
		       email, eventduration, image, locdetails, locend, newsflash, timedetails, webname, weburl
		FROM shift2bikes_events
		WHERE composite_event_id = ? AND citycode = ?
	`
	return r.scanScrapedRides(query, rideID, city)
}

func (r *Repository) scanScrapedRides(query string, args ...any) ([]ScrapedRideFromDB, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []ScrapedRideFromDB
	for rows.Next() {
		var ride ScrapedRideFromDB
		if err := rows.Scan(
			&ride.ID, &ride.Title, &ride.Lat, &ride.Lng, &ride.Address,
			&ride.Audience, &ride.Cancelled, &ride.Date, &ride.StartTime,
			&ride.SafetyPlan, &ride.Details, &ride.Venue, &ride.Organizer,
			&ride.LoopRide, &ride.Shareable, &ride.RideSource, &ride.EndTime,
			&ride.Email, &ride.EventDuration, &ride.Image, &ride.LocDetails,
			&ride.LocEnd, &ride.NewsFlash, &ride.TimeDetails, &ride.WebName, &ride.WebURL,
		); err != nil {
			return nil, err
		}
		rides = append(rides, ride)
	}
	return rides, nil
}

func getTimeZone(city string) string {
	timezones := map[string]string{
		"pdx": "America/Los_Angeles",
		"slc": "America/Denver",
	}
	if tz, ok := timezones[city]; ok {
		return tz
	}
	return "America/Los_Angeles" // default
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// GetPendingRides returns all rides that are not yet published
func (r *Repository) GetPendingRides() ([]RideForAdmin, error) {
	rows, err := r.db.Query(`
		SELECT id, title, description, venue_name, city, organizer_name,
		       organizer_email, image_url, image_uuid, is_loop_ride, is_published,
		       created_at, moderation_notes
		FROM events
		WHERE is_published = 0
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []RideForAdmin
	for rows.Next() {
		var ride RideForAdmin
		var moderationNotes, imageURL, imageUUID sql.NullString

		if err := rows.Scan(
			&ride.ID, &ride.Title, &ride.Description, &ride.VenueName, &ride.City,
			&ride.OrganizerName, &ride.OrganizerEmail, &imageURL, &imageUUID,
			&ride.IsLoopRide, &ride.IsPublished, &ride.CreatedAt, &moderationNotes,
		); err != nil {
			return nil, err
		}

		if moderationNotes.Valid {
			ride.ModerationNotes = moderationNotes.String
		}
		if imageURL.Valid {
			ride.ImageURL = imageURL.String
		}
		if imageUUID.Valid {
			ride.ImageUUID = imageUUID.String
		}
		rides = append(rides, ride)
	}

	return rides, rows.Err()
}

// PublishRide marks a ride as published
func (r *Repository) PublishRide(rideID int, moderationNotes string) error {
	now := time.Now().Format(time.RFC3339)

	_, err := r.db.Exec(`
		UPDATE events
		SET is_published = 1, moderation_notes = ?, moderated_at = ?
		WHERE id = ?
	`, moderationNotes, now, rideID)

	return err
}

// ValidateAdminKey checks if an API key is valid and not revoked
func (r *Repository) ValidateAdminKey(apiKey string) (bool, error) {
	// Decode the provided key from base64 to get raw bytes
	decodedKey, err := base64.URLEncoding.DecodeString(apiKey)
	if err != nil {
		// If it can't be decoded, it's not a valid key format
		return false, nil
	}

	// Get all active API keys and compare with provided key using bcrypt
	rows, err := r.db.Query(`
		SELECT api_key FROM admin_api_keys WHERE revoked_at IS NULL
	`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var hashedKey string
		if err := rows.Scan(&hashedKey); err != nil {
			continue
		}

		// Compare decoded key bytes with hashed key
		// bcrypt.CompareHashAndPassword expects the raw bytes that were hashed
		if err := bcrypt.CompareHashAndPassword([]byte(hashedKey), decodedKey); err == nil {
			return true, nil
		}
	}

	return false, rows.Err()
}
