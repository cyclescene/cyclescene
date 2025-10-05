package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"
)

var cityTimeZones = map[string]string{
	"pdx": "America/Los_Angeles",
	"slc": "America/Denver",
}

func getTimeZone(cityCode string) string {
	if tz, ok := cityTimeZones[strings.ToLower(cityCode)]; ok {
		return tz
	}
	return "America/Los_Angeles"
}

func ConnectToDB() (*sql.DB, error) {
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_DB_RW_TOKEN") == "" {
		log.Fatal("FATAL: Turso env variable not set properly")
	}

	// GOOGLE Vars
	if os.Getenv("GOOGLE_GEOCODING_API_KEY") == "" {
		log.Fatal("FATAL: GOOGLE_GEOCODING_API_KEY not properly set")
	}
	//
	// // set up logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})))
	//
	// connect to DB(Turso)
	dbURL := os.Getenv("TURSO_DB_URL")
	authToken := os.Getenv("TURSO_DB_RW_TOKEN")

	fullURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	return sql.Open("libsql", fullURL)
}

func scanRides(db *sql.DB, query string, args ...any) ([]RideFromDB, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []RideFromDB
	for rows.Next() {
		var r RideFromDB
		if err := rows.Scan(
			&r.ID,
			&r.Title,
			&r.Lat,
			&r.Lng,
			&r.Address,
			&r.Audience,
			&r.Cancelled,
			&r.Date,
			&r.StartTime,
			&r.SafetyPlan,
			&r.Details,
			&r.Venue,
			&r.Organizer,
			&r.LoopRide,
			&r.Shareable,
			&r.RideSource,
			&r.EndTime,
			&r.Email,
			&r.EventDuration,
			&r.Image,
			&r.LocDetails,
			&r.LocEnd,
			&r.NewsFlash,
			&r.TimeDetails,
			&r.WebName,
			&r.WebUrl); err != nil {
			return nil, err
		}
		rides = append(rides, r)
	}
	return rides, nil
}

func getUpcomingRides(db *sql.DB, cityCode string) ([]RideFromDB, error) {
	tzStr := getTimeZone(cityCode)
	tz, err := time.LoadLocation(tzStr)
	if err != nil {
		slog.Error("failed to load timezone", "error", err.Error())
		return nil, err
	}

	now := time.Now().In(tz)

	todayStr := now.Format("2006-01-02")

	query := `
    SELECT id, title, lat, lng, address, audience, cancelled, date, starttime, safetyplan, details, venue, organizer, loopride, shareable, ridesource, endtime, email, eventduration, image, locdetails, locend, newsflash, timedetails, webname, weburl
    FROM rides
    WHERE 
			cityCode = ?
			AND
			date >= ?
    ORDER BY date ASC, starttime ASC;`
	return scanRides(db, query, cityCode, todayStr)
}

func getPastRides(db *sql.DB, cityCode string) ([]RideFromDB, error) {
	tzStr := getTimeZone(cityCode)
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
    SELECT id, title, lat, lng, address, audience, cancelled, date, starttime, safetyplan, details, venue, organizer, loopride, shareable, ridesource, endtime, email, eventduration, image, locdetails, locend, newsflash, timedetails, webname, weburl
    FROM rides
    WHERE 
				cityCode = ?
				AND
        date BETWEEN ? AND ?
    ORDER BY date DESC, starttime DESC;`
	return scanRides(db, query, cityCode, sevenDaysAgoStr, todayStr)
}
