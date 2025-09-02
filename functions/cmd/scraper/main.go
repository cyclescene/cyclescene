package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func buildShift2BikesURL() (string, error) {
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return "", fmt.Errorf("failed to load timezone location: %w", err)
	}

	nowInPortland := time.Now().In(location)

	year, month, day := nowInPortland.Date()
	startDate := time.Date(year, month, day, 0, 0, 0, 0, location)

	endDate := startDate.AddDate(0, 0, 7)

	formattedStartDate := startDate.Format(time.RFC3339)
	formattedEndDate := endDate.Format(time.RFC3339)

	baseURL := "https://www.shift2bikes.org/api/events.php"

	finalURL, _ := url.Parse(baseURL)

	params := url.Values{}
	params.Set("startdate", formattedStartDate)
	params.Set("enddate", formattedEndDate)

	finalURL.RawQuery = params.Encode()

	return finalURL.String(), nil
}

type Shift2BikeEvent struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Venue         string `json:"venue"`
	Address       string `json:"address"`
	Organizer     string `json:"organizer"`
	Details       string `json:"details"`
	Time          string `json:"time"`
	Hideemail     bool   `json:"hideemail"`
	Hidephone     bool   `json:"hidephone"`
	Hidecontact   bool   `json:"hidecontact"`
	Length        any    `json:"length"`
	Timedetails   any    `json:"timedetails"`
	Locdetails    string `json:"locdetails"`
	Loopride      bool   `json:"loopride"`
	Locend        any    `json:"locend"`
	Eventduration int    `json:"eventduration"`
	Weburl        string `json:"weburl"`
	Webname       string `json:"webname"`
	Image         string `json:"image"`
	Audience      string `json:"audience"`
	Tinytitle     string `json:"tinytitle"`
	Printdescr    any    `json:"printdescr"`
	Datestype     string `json:"datestype"`
	Area          string `json:"area"`
	Featured      bool   `json:"featured"`
	Printemail    bool   `json:"printemail"`
	Printphone    bool   `json:"printphone"`
	Printweburl   bool   `json:"printweburl"`
	Printcontact  bool   `json:"printcontact"`
	Published     bool   `json:"published"`
	Safetyplan    bool   `json:"safetyplan"`
	Email         any    `json:"email"`
	Phone         any    `json:"phone"`
	Contact       string `json:"contact"`
	Date          string `json:"date"`
	CaldailyID    string `json:"caldaily_id"`
	Shareable     string `json:"shareable"`
	Cancelled     bool   `json:"cancelled"`
	Newsflash     any    `json:"newsflash"`
	Status        string `json:"status"`
	Endtime       string `json:"endtime"`
}

type Shift2BikeEvents struct {
	Events []Shift2BikeEvent `json:"events"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	url, err := buildShift2BikesURL()
	if err != nil {
		log.Fatalf("failed to build Shift2Bikes event url: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, url, nil)

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to get events from Shift2bikes: %v", err)
	}
	defer res.Body.Close()

	var events Shift2BikeEvents

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&events); err != nil {
		log.Fatalf("failed to decode events: %v", err)
	}

	dbURL := os.Getenv("TURSO_DB_URL")
	authToken := os.Getenv("TURSO_DB_RW_TOKEN")

	fullURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	db, err := sql.Open("libsql", fullURL)
	if err != nil {
		log.Fatalf("failed to open Turso DB connection: %v", err)
	}
	defer db.Close()

	if err := createEventTable(db); err != nil {
		log.Fatalf("failed to create rides table: %v", err)
	}

	if err := upsertEvents(db, client, events); err != nil {
		log.Fatalf("something went wrong when upserting events: %v", err)
	}
}

type GeocodeResult struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

func geocodeAddress(address string, client *http.Client) (float64, float64, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"
	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return 0.0, 0.0, err
	}

	q := req.URL.Query()
	q.Add("q", address)
	q.Add("format", "json")
	q.Add("limit", "1")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", "Bike Bae (jduarte0912@gmail.com)")

	res, err := client.Do(req)
	if err != nil {
		return 0.0, 0.0, err
	}
	defer res.Body.Close()

	var results []GeocodeResult
	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		return 0.0, 0.0, fmt.Errorf("failed to decode geocoding response: %v", err)
	}

	if len(results) == 0 {
		return 0.0, 0.0, fmt.Errorf("no results found for address: %v", address)
	}

	lat, _ := strconv.ParseFloat(results[0].Lat, 64)
	lon, _ := strconv.ParseFloat(results[0].Lon, 64)

	return lat, lon, nil
}

func cleanAddress(rawAddress string) string {
	spaceRegex := regexp.MustCompile(`\s+`)
	clean := spaceRegex.ReplaceAllString(rawAddress, " ")
	clean = strings.TrimSpace(clean)

	clean = strings.ReplaceAll(clean, " and ", " & ")
	clean = strings.ReplaceAll(clean, " AND ", " & ")
	clean = strings.ReplaceAll(clean, " @ ", " & ")

	suiteRegex := regexp.MustCompile(`(?i)\s+(#|apt|unit|suite)\s*\w+`)
	clean = suiteRegex.ReplaceAllString(clean, "")

	clean = strings.ReplaceAll(clean, "Av.", "Ave")
	clean = strings.ReplaceAll(clean, "St.", "St")

	clean = strings.ReplaceAll(clean, " United States", "")
	clean = strings.ReplaceAll(clean, " USA", "")

	if !strings.Contains(clean, "Portland, OR") {
		clean = clean + ", Portland, OR"
	}

	return clean
}

func createEventTable(db *sql.DB) error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS rides (
        id TEXT PRIMARY KEY,
        address TEXT,
        audience TEXT NOT NULL,
        cancelled INTEGER NOT NULL,
        date TEXT NOT NULL,
        details TEXT,
        endtime TEXT,
        eventduration INTEGER,
        image TEXT,
        lat REAL,
        lon REAL,
        locdetails TEXT,
        locend TEXT,
        loopride INTEGER NOT NULL,
        newsflash TEXT,
        organizer TEXT,
        safetyplan INTEGER NOT NULL,
        shareable TEXT,
        starttime TEXT NOT NULL,
        timedetails TEXT,
        title TEXT NOT NULL,
        venue TEXT,
        source_data TEXT NOT NULL
    );`

	if _, err := db.Exec(createTableSQL); err != nil {
		return err
	}

	return nil
}

func upsertEvents(db *sql.DB, client *http.Client, events Shift2BikeEvents) error {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
		return err
	}

	stmt, err := tx.Prepare(`
        INSERT INTO rides (id, address, audience, cancelled, date, details, endtime, eventduration, image, lat, lon, locdetails, locend, loopride, newsflash, organizer, safetyplan, shareable, starttime, timedetails, title, venue, source_data)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET
            address=excluded.address,
            audience=excluded.audience,
            cancelled=excluded.cancelled,
            date=excluded.date,
            details=excluded.details,
            endtime=excluded.endtime,
            eventduration=excluded.eventduration,
            image=excluded.image,
            lat=excluded.lat,
            lon=excluded.lon,
            locdetails=excluded.locdetails,
            locend=excluded.locend,
            loopride=excluded.loopride,
            newsflash=excluded.newsflash,
            organizer=excluded.organizer,
            safetyplan=excluded.safetyplan,
            shareable=excluded.shareable,
            starttime=excluded.starttime,
            timedetails=excluded.timedetails,
            title=excluded.title,
            venue=excluded.venue,
            source_data=excluded.source_data;
        `)
	if err != nil {
		log.Printf("failed to prepare statement: %v", err)
		return err
	}
	defer stmt.Close()

	for _, event := range events.Events {
		sourceData, _ := json.Marshal(event)

		cancelledInt := 0
		if event.Cancelled {
			cancelledInt = 1
		}

		loopride := 0
		if event.Loopride {
			loopride = 1
		}

		safetyplan := 0
		if event.Safetyplan {
			safetyplan = 1
		}

		var lat, lon float64
		if event.Address != "" {
			sanitizedAddress := cleanAddress(event.Address)
			lat, lon, _ = geocodeAddress(sanitizedAddress, client)
		}

		_, err := stmt.Exec(
			event.ID,
			event.Address,
			event.Audience,
			cancelledInt,
			event.Date,
			event.Details,
			event.Endtime,
			event.Eventduration,
			event.Image,
			lat,
			lon,
			event.Locdetails,
			event.Locend,
			loopride,
			event.Newsflash,
			event.Organizer,
			safetyplan,
			event.Shareable,
			event.Time,
			event.Timedetails,
			event.Title,
			event.Venue,
			string(sourceData))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute statement for event ID %s: %v", event.ID, err)

		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("failed to commit transaction: %v", err)
		return err
	}

	log.Printf("Successfully saved %d records to Turso", len(events.Events))

	return nil
}
