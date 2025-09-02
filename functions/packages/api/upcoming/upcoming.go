package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type ResponseHeaders struct {
	ContentType string `json:"Content-Type"`
}

type Response struct {
	Body       []Ride          `json:"body"`
	StatusCode int             `json:"statusCode"`
	Headers    ResponseHeaders `json:"Headers"`
}

func Main(ctx context.Context) Response {
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_DB_RO_TOKEN") == "" {
		log.Fatal("FATAL: TURSO_DB_URL or TURSO_DB_RO_TOKEN not set")
	}

	db, err := connectToDB()
	if err != nil {
		log.Printf("FATAL: failed to connect to Turso: %v\n", err)
		return Response{
			Body:       []Ride{},
			StatusCode: http.StatusInternalServerError,
			Headers: ResponseHeaders{
				ContentType: "application/json",
			},
		}
	}

	rides, err := getUpcomingRides(db)
	if err != nil {
		log.Printf("Error: failed to retrieve upcoming rides: %v\n", err)
		return Response{
			Body:       []Ride{},
			StatusCode: http.StatusInternalServerError,
			Headers: ResponseHeaders{
				ContentType: "application/json",
			},
		}
	}

	return Response{
		Body:       rides,
		StatusCode: http.StatusOK,
		Headers: ResponseHeaders{
			ContentType: "application/json",
		},
	}
}

func connectToDB() (*sql.DB, error) {
	dbURL := os.Getenv("TURSO_DB_URL")
	authToken := os.Getenv("TURSO_DB_RO_TOKEN")

	if dbURL == "" || authToken == "" {
		return nil, fmt.Errorf("TURSO environment variables are not set")
	}

	fullURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	return sql.Open("libsql", fullURL)
}

type Ride struct {
	ID            string          `json:"id"`
	Address       string          `json:"address"`
	Audience      string          `json:"audience"`
	Cancelled     int             `json:"cancelled"`
	Date          string          `json:"date"`
	Details       sql.NullString  `json:"details,omitempty"`
	EndTime       sql.NullString  `json:"endtime,omitempty"`
	EventDuration sql.NullInt32   `json:"eventduration,omitempty"`
	Image         sql.NullString  `json:"image,omitempty"`
	Lat           sql.NullFloat64 `json:"lat,omitempty"`
	Lon           sql.NullFloat64 `json:"lon,omitempty"`
	LocDetails    sql.NullString  `json:"locdetails,omitempty"`
	LocEnd        sql.NullString  `json:"locend,omitempty"`
	LoopRide      int             `json:"loopride"`
	NewsFlash     sql.NullString  `json:"newsflash,omitempty"`
	Organizer     sql.NullString  `json:"organizer,omitempty"`
	SafetyPlan    int             `json:"safetyplan"`
	Shareable     string          `json:"shareable"`
	StartTime     string          `json:"starttime"`
	TimeDetails   sql.NullString  `json:"timedetails,omitempty"`
	Title         string          `json:"title"`
	Venue         sql.NullString  `json:"venue,omitempty"`
	SourceData    string          `json:"source_data"`
}

func scanRides(db *sql.DB, query string, args ...interface{}) ([]Ride, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []Ride
	for rows.Next() {
		var r Ride
		if err := rows.Scan(
			&r.ID,
			&r.Address,
			&r.Audience,
			&r.Cancelled,
			&r.Date,
			&r.Details,
			&r.EndTime,
			&r.EventDuration,
			&r.Image,
			&r.Lat,
			&r.Lon,
			&r.LocDetails,
			&r.LocEnd,
			&r.LoopRide,
			&r.NewsFlash,
			&r.Organizer,
			&r.SafetyPlan,
			&r.Shareable,
			&r.StartTime,
			&r.TimeDetails,
			&r.Title,
			&r.Venue); err != nil {
			return nil, err
		}
		rides = append(rides, r)
	}
	return rides, nil
}

func getUpcomingRides(db *sql.DB) ([]Ride, error) {
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return nil, fmt.Errorf("could not load timezone: %w", err)
	}

	nowInPortland := time.Now().In(location)

	todayStr := nowInPortland.Format("2006-01-02")

	query := `
    SELECT id, address, audience, cancelled, date, details, endtime, eventduration, image, lat, lon, locdetails, locend, loopride, newsflash, organizer, safetyplan, shareable, starttime, timedetails, title, venue
    FROM rides
    WHERE date >= ?
    ORDER BY date ASC, starttime ASC;`
	return scanRides(db, query, todayStr)
}
