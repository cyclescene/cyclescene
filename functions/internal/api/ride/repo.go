package ride

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// User-submitted rides
func (r *Repository) CreateRide(submission *Submission, editToken string) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO events (
			title, tinytitle, description, image_url, audience, ride_length, area, date_type,
			venue_name, address, location_details, ending_location, is_loop_ride,
			organizer_name, organizer_email, organizer_phone, web_url, web_name, newsflash,
			hide_email, hide_phone, hide_contact_name, group_code, edit_token, city, is_published
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`,
		submission.Title, submission.TinyTitle, submission.Description, submission.ImageURL,
		submission.Audience, submission.RideLength, submission.Area, submission.DateType,
		submission.VenueName, submission.Address, submission.LocationDetails, submission.EndingLocation,
		boolToInt(submission.IsLoopRide), submission.OrganizerName, submission.OrganizerEmail,
		submission.OrganizerPhone, submission.WebURL, submission.WebName, submission.Newsflash,
		boolToInt(submission.HideEmail), boolToInt(submission.HidePhone), boolToInt(submission.HideContactName),
		nilIfEmpty(submission.GroupCode), editToken, submission.City,
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
				event_duration_minutes, event_time_details
			) VALUES (?, ?, ?, ?, ?, ?)
		`, eventID, occ.StartDate, occ.StartTime, startDatetime, occ.EventDurationMinutes, occ.EventTimeDetails)

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
		SELECT start_date, start_time, event_duration_minutes, event_time_details
		FROM event_occurrences WHERE event_id = ?
		ORDER BY start_datetime ASC
	`, id)

	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var occ Occurrence
		if err := rows.Scan(&occ.StartDate, &occ.StartTime, &occ.EventDurationMinutes, &occ.EventTimeDetails); err != nil {
			continue
		}
		submission.Occurrences = append(submission.Occurrences, occ)
	}

	return &submission, isPublished == 1, nil
}

func (r *Repository) UpdateRide(token string, submission *Submission) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		UPDATE events SET
			title = ?, tinytitle = ?, description = ?, image_url = ?,
			audience = ?, ride_length = ?, area = ?, date_type = ?,
			venue_name = ?, address = ?, location_details = ?, ending_location = ?, is_loop_ride = ?,
			organizer_name = ?, organizer_email = ?, organizer_phone = ?,
			web_url = ?, web_name = ?, newsflash = ?,
			hide_email = ?, hide_phone = ?, hide_contact_name = ?,
			group_code = ?, updated_at = STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')
		WHERE edit_token = ?
	`,
		submission.Title, submission.TinyTitle, submission.Description, submission.ImageURL,
		submission.Audience, submission.RideLength, submission.Area, submission.DateType,
		submission.VenueName, submission.Address, submission.LocationDetails, submission.EndingLocation,
		boolToInt(submission.IsLoopRide), submission.OrganizerName, submission.OrganizerEmail,
		submission.OrganizerPhone, submission.WebURL, submission.WebName, submission.Newsflash,
		boolToInt(submission.HideEmail), boolToInt(submission.HidePhone), boolToInt(submission.HideContactName),
		nilIfEmpty(submission.GroupCode), token,
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
				event_duration_minutes, event_time_details
			) VALUES (?, ?, ?, ?, ?, ?)
		`, eventID, occ.StartDate, occ.StartTime, startDatetime, occ.EventDurationMinutes, occ.EventTimeDetails)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Scraped rides from Shift2Bikes
func (r *Repository) GetUpcomingRides(city string) ([]ScrapedRideFromDB, error) {
	tzStr := getTimeZone(city)
	tz, err := time.LoadLocation(tzStr)
	if err != nil {
		slog.Error("failed to load timezone", "error", err.Error())
		return nil, err
	}
	now := time.Now().In(tz)
	todayStr := now.Format("2006-01-02")

	query := `
		SELECT composite_event_id, title, lat, lng, address, audience, cancelled, date, starttime, 
		       safetyplan, details, venue, organizer, loopride, shareable, ridesource, endtime, 
		       email, eventduration, image, locdetails, locend, newsflash, timedetails, webname, weburl
		FROM shift2bikes_events
		WHERE citycode = ? AND date >= ?
		ORDER BY date ASC, starttime ASC
	`
	return r.scanScrapedRides(query, city, todayStr)
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

	query := `
		SELECT composite_event_id, title, lat, lng, address, audience, cancelled, date, starttime, 
		       safetyplan, details, venue, organizer, loopride, shareable, ridesource, endtime, 
		       email, eventduration, image, locdetails, locend, newsflash, timedetails, webname, weburl
		FROM shift2bikes_events
		WHERE citycode = ? AND date BETWEEN ? AND ?
		ORDER BY date DESC, starttime DESC
	`
	return r.scanScrapedRides(query, city, sevenDaysAgoStr, todayStr)
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
