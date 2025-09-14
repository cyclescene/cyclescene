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
	"strings"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"golang.org/x/exp/slog"
)

func buildShift2BikesURL() (string, error) {
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return "", fmt.Errorf("failed to load timezone location: %w", err)
	}

	nowInPortland := time.Now().In(location)

	year, month, day := nowInPortland.Date()
	startDate := time.Date(year, month, day, 0, 0, 0, 0, location)

	endDate := startDate.AddDate(0, 0, 60)

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

type GoogleGeocodeResponse struct {
	Results []GoogleGeocodeResult `json:"results"`
	Status  string                `json:"status"`
}

type GoogleGeocodeResult struct {
	Geometry GoogleGeometry `json:"geometry"`
}

type GoogleGeometry struct {
	Location GoogleLocation `json:"location"`
}

type GoogleLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lng"`
}

type geocodeCacheEntry struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func Main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})))
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_DB_RW_TOKEN") == "" {
		log.Fatal("FATAL: Turso env variable not set properly")
	}

	if os.Getenv("GOOGLE_GEOCODING_API_KEY") == "" {
		log.Fatal("FATAL: GOOGLE_GEOCODING_API_KEY not properly set")
	}

	url, err := buildShift2BikesURL()
	if err != nil {
		slog.Error("failed to build Shift2bikes event url", slog.String("error", err.Error()))
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

	if err := createTables(db); err != nil {
		log.Fatalf("failed to create rides table: %v", err)
	}

	if err := upsertEvents(db, client, events); err != nil {
		log.Fatalf("something went wrong when upserting events: %v", err)
	}
}

func geocodeAddress(address string, client *http.Client) (float64, float64, error) {
	googleAPIKey := os.Getenv("GOOGLE_GEOCODING_API_KEY")
	baseURL := "https://maps.googleapis.com/maps/api/geocode/json"
	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return 0.0, 0.0, err
	}

	q := req.URL.Query()
	q.Add("address", address)
	q.Add("key", googleAPIKey)
	q.Add("bounds", "45.35,-123.00|45.75,-122.30")
	q.Add("components", "country:US")
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		return 0.0, 0.0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0.0, 0.0, fmt.Errorf("Google Geocoding API returned non-OK status code %d", res.StatusCode)
	}

	var googleResponse GoogleGeocodeResponse
	if err := json.NewDecoder(res.Body).Decode(&googleResponse); err != nil {
		return 0, 0, fmt.Errorf("failed to decode Google geocoding response: %v", err)
	}

	if googleResponse.Status != "OK" {
		if googleResponse.Status == "ZERO_RESULTS" {
			return 0.0, 0.0, fmt.Errorf("Google Geocoding API found no results for address: '%s'", address)
		}
		return 0.0, 0.0, fmt.Errorf("Google Geocoding API returned status: %s for (for address: '%s')", googleResponse.Status, address)
	}

	if len(googleResponse.Results) == 0 {
		return 0.0, 0.0, fmt.Errorf("no results found for address: '%s' (GOOGLE API 'OK' status but empty results", address)
	}

	lat := googleResponse.Results[0].Geometry.Location.Lat
	lon := googleResponse.Results[0].Geometry.Location.Lon

	return lat, lon, nil
}

func cleanAddress(rawinput string) string {
	if rawinput == "" {
		return ""
	}

	clean := strings.ToLower(rawinput)
	spaceRegex := regexp.MustCompile(`\s+`)
	clean = spaceRegex.ReplaceAllLiteralString(clean, " ")
	clean = strings.TrimSpace(clean)

	clean = strings.ReplaceAll(clean, " and ", " & ")
	clean = strings.ReplaceAll(clean, " @ ", " & ")

	suiteRegex := regexp.MustCompile(`(?i)\s+(#|apt|unit|suite)\s*\w+`)
	clean = suiteRegex.ReplaceAllString(clean, "")
	clean = strings.ReplaceAll(clean, "av.", "ave")
	clean = strings.ReplaceAll(clean, "st.", "st")
	clean = strings.ReplaceAll(clean, "pl.", "place")

	hasPortland := strings.Contains(clean, "portland")
	hasOregon := strings.Contains(clean, "or")

	if !hasPortland && !hasOregon {
		clean = clean + ", portland, or"
	} else if hasPortland && !hasOregon {
		clean = clean + ", or"
	}

	return clean
}

func createTables(db *sql.DB) error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS rides (
        composite_event_id TEXT PRIMARY KEY,
        id TEXT NOT NULL,
        address TEXT,
        audience TEXT NOT NULL,
        cancelled INTEGER NOT NULL,
        date TEXT NOT NULL,
        details TEXT,
        endtime TEXT,
        email TEXT,
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
        webname TEXT,
        weburl TEXT,
        source_data TEXT NOT NULL
    );`

	createGeocodeCacheTableSQL := `
    CREATE TABLE IF NOT EXISTS geocode_cache (
    address_key TEXT PRIMARY KEY,
    lat REAL NOT NULL,
    lon REAL NOT NULL,
    last_updated TEXT NOT NULL
    );`

	if _, err := db.Exec(createTableSQL); err != nil {
		return err
	}

	if _, err := db.Exec(createGeocodeCacheTableSQL); err != nil {
		return err
	}

	return nil
}

func geocodeAndCache(client *http.Client, addressKey string, inMemoryCache map[string]geocodeCacheEntry, geocodeCacheUpsertStmt *sql.Stmt) (float64, float64, error) {
	if addressKey == "" {
		return 0.0, 0.0, fmt.Errorf("empty address key provided for geocoding")
	}

	if entry, ok := inMemoryCache[addressKey]; ok {
		return entry.Lat, entry.Lon, nil
	}

	lat, lon, geocodeErr := geocodeAddress(addressKey, client)

	if geocodeErr != nil {
		log.Printf("failed to geocode address '%s' via Google API: %v", addressKey, geocodeErr)
		return 0.0, 0.0, geocodeErr
	}

	inMemoryCache[addressKey] = geocodeCacheEntry{Lat: lat, Lon: lon}
	_, err := geocodeCacheUpsertStmt.Exec(addressKey, lat, lon, time.Now().Format(time.RFC3339))
	if err != nil {
		return 0.0, 0.0, fmt.Errorf("failed to upsert new geocode results for '%s' to DB cache: %v", addressKey, err)
	}

	return lat, lon, nil
}

func upsertEvents(db *sql.DB, client *http.Client, events Shift2BikeEvents) error {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	cachedAddresses := make(map[string]geocodeCacheEntry)
	rows, err := tx.Query("SELECT address_key, lat, lon FROM geocode_cache")
	if err != nil {
		log.Printf("Error loading geocode cache from DB: %v. Proceedign with empty cache", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var key string
			var lat, lon float64
			if err := rows.Scan(&key, &lat, &lon); err != nil {
				log.Printf("Error scanning geocode cache row: %v", err)
				continue
			}
			cachedAddresses[key] = geocodeCacheEntry{Lat: lat, Lon: lon}
		}
		if err = rows.Err(); err != nil {
			log.Printf("error after iterating geocode cache rows: %v", err)
		}
		log.Printf("Loaded %d addresses into in-memory geocode cache", len(cachedAddresses))

	}

	ridesUpsertStmt, err := tx.Prepare(`
        INSERT INTO rides (
            composite_event_id,
            id,
            address,
            audience,
            cancelled,
            date,
            details,
            endtime,
            email,
            eventduration,
            image,
            lat,
            lon,
            locdetails,
            locend,
            loopride,
            newsflash,
            organizer,
            safetyplan,
            shareable,
            starttime,
            timedetails,
            title,
            venue,
            webname,
            weburl,
            source_data
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(composite_event_id) DO UPDATE SET
            id=excluded.id,
            address=excluded.address,
            audience=excluded.audience,
            cancelled=excluded.cancelled,
            date=excluded.date,
            details=excluded.details,
            endtime=excluded.endtime,
            email=excluded.email,
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
            webname=excluded.webname,
            weburl=excluded.weburl,
            source_data=excluded.source_data;
        `)
	if err != nil {
		log.Printf("failed to prepare ride upsert statement: %v", err)
		return err
	}
	defer ridesUpsertStmt.Close()

	geocodeCacheUpsertStmt, err := tx.Prepare(`
        INSERT INTO geocode_cache (address_key, lat, lon, last_updated)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(address_key) DO UPDATE SET
            lat=excluded.lat,
            lon=excluded.lon,
            last_updated=excluded.last_updated;
        `)
	if err != nil {
		log.Printf("failed to prepare geocode cache upsert statement: %v", err)
		return err
	}
	defer geocodeCacheUpsertStmt.Close()

	for _, event := range events.Events {
		compositeEventID := fmt.Sprintf("%s_%s", event.ID, event.Date)
		sourceData, marshalErr := json.Marshal(event)
		if marshalErr != nil {
			log.Printf("Warning: failed to marshal source_data for event ID %s (composite ID : %s) %v", event.ID, compositeEventID, marshalErr)
			sourceData = []byte("{}")
		}

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
		var currentGeoCodeErr error

		cleanedAddr := cleanAddress(event.Address)
		if cleanedAddr != "" {
			lat, lon, currentGeoCodeErr = geocodeAndCache(client, cleanedAddr, cachedAddresses, geocodeCacheUpsertStmt)
		} else {
			currentGeoCodeErr = fmt.Errorf("cleaned address was empty for event ID %s (composite ID: %s)", event.ID, compositeEventID)
		}

		if currentGeoCodeErr != nil && event.Venue != "" {
			log.Printf(
				"Geocoding event.Address ('%s' -> cleaned '%s') failed or was empty for (composite ID: %s). Attempting event.Venue ('%s'). Error: %v",
				event.Address, cleanedAddr, compositeEventID, event.Venue, currentGeoCodeErr)
			cleanedVenue := cleanAddress(event.Venue)
			if cleanedVenue != "" {
				lat, lon, currentGeoCodeErr = geocodeAndCache(client, cleanedVenue, cachedAddresses, geocodeCacheUpsertStmt)
			} else {
				currentGeoCodeErr = fmt.Errorf("cleaned venue was exmpty after address failure for event ID: %s (composite event ID: %s)", event.ID, compositeEventID)
			}
		}

		if currentGeoCodeErr != nil {
			log.Printf("Could not get coordinates for event ID %s (composite event ID: %s) after trying both address ('%s') and venue ('%s'). Setting to coords to Portland coords [45.54, -122.65]", event.ID, compositeEventID, event.Address, event.Venue)
			lat, lon = 45.54, -122.65
		}

		_, execErr := ridesUpsertStmt.Exec(
			compositeEventID,
			event.ID,
			event.Address,
			event.Audience,
			cancelledInt,
			event.Date,
			event.Details,
			event.Endtime,
			event.Email,
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
			event.Webname,
			event.Weburl,
			string(sourceData))
		if execErr != nil {
			err = fmt.Errorf("failed to execute statement for event ID %s: %v", event.ID, err)
			return err

		}
	}

	log.Printf("Successfully saved %d records to Turso", len(events.Events))
	return nil
}
