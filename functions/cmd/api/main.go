package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Ride struct {
	ID            string          `json:"id"`
	Address       string          `json:"address"`
	Audience      string          `json:"audience"`
	Cancelled     int             `json:"cancelled"`
	Date          string          `json:"date"`
	Details       sql.NullString  `json:"details,omitempty"`
	EndTime       sql.NullString  `json:"endtime,omitempty"`
	Email         sql.NullString  `json:"email,omitempty"`
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
	WebUrl        sql.NullString  `json:"weburl,omitempty"`
	WebName       sql.NullString  `json:"webname,omitempty"`
	SourceData    string          `json:"source_data"`
}

// Handler logic
func APIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	db, err := connectToDB()
	if err != nil {
		log.Printf("Error: Failed to connect to database: %v", err)
		return
	}
	defer db.Close()

	switch r.URL.Path {
	case "/upcoming":
		handleGetRides(w, r, db, getUpcomingRides)
	case "/past":
		handleGetRides(w, r, db, getPastRides)
	default:
		http.NotFound(w, r)
	}
}

type rideFetcher func(db *sql.DB) ([]Ride, error)

func handleGetRides(w http.ResponseWriter, r *http.Request, db *sql.DB, fetcher rideFetcher) {
	rides, err := fetcher(db)
	if err != nil {
		log.Printf("Error: Database query failed: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if rides == nil {
		rides = []Ride{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rides); err != nil {
		log.Printf("Error: Failed to encode rides to JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// DB LOGIC
func connectToDB() (*sql.DB, error) {
	dbURL := os.Getenv("TURSO_DB_URL")
	authToken := os.Getenv("TURSO_DB_RO_TOKEN")

	if dbURL == "" || authToken == "" {
		return nil, fmt.Errorf("TURSO environment variables are not set")
	}

	fullURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	return sql.Open("libsql", fullURL)
}

func scanRides(db *sql.DB, query string) ([]Ride, error) {
	rows, err := db.Query(query)
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
			&r.Email,
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
			&r.Venue,
			&r.WebName,
			&r.WebUrl); err != nil {
			return nil, err
		}
		rides = append(rides, r)
	}
	return rides, nil
}

func getUpcomingRides(db *sql.DB) ([]Ride, error) {
	query := `
    SELECT id, address, audience, cancelled, date, details, endtime, email, eventduration, image, lat, lon, locdetails, locend, loopride, newsflash, organizer, safetyplan, shareable, starttime, timedetails, title, venue, webname, weburl
    FROM rides
    WHERE date >= date('now', 'localtime')
    ORDER BY date ASC, starttime ASC;`
	return scanRides(db, query)
}

func getPastRides(db *sql.DB) ([]Ride, error) {
	query := `
    SELECT id, address, audience, cancelled, date, details, endtime, email, eventduration, image, lat, lon, locdetails, locend, loopride, newsflash, organizer, safetyplan, shareable, starttime, timedetails, title, venue, webname, weburl
    FROM rides
    WHERE 
        date >= date('now', 'localtime', '-7 days')
        AND
        date < date('now', 'localtime')
    ORDER BY date DESC, starttime DESC;`
	return scanRides(db, query)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	http.HandleFunc("/", APIHandler)
	log.Println("Staring local server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
