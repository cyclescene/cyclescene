package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	// "github.com/joho/godotenv"
	scraperhelpers "github.com/spacesedan/cyclescene/functions/lib/scraper-helpers"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

const (
	FALLBACK_LAT   = 45.523064
	FALLBACK_LNG   = -122.676483
	FALLBACK_QUERY = "fallback"
	EVENT_SOURCE   = "Shift2Bikes"
	PDX_CITY_CODE  = "pdx"
)

func main() {
	// used in development
	// _ = godotenv.Load()
	// Check for ENV variables
	// DB Vars
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

	db, err := sql.Open("libsql", fullURL)
	if err != nil {
		log.Fatalf("failed to open Turso DB connection: %v", err)
	}
	defer db.Close()

	if err := scraperhelpers.CreateTables(db); err != nil {
		log.Fatal("failed to create tables")
	}

	if err := scraperhelpers.CreateTableIndexes(db); err != nil {
		log.Fatal("failed to create table indexes")
	}

	////// READY TO START /////////////////////////

	// get all previously saved locations from DB
	geocodeCache, err := scraperhelpers.GetGeocodeCache(db)
	if err != nil {
		log.Fatalf("something went wrong: %s\n", err.Error())
	}

	// get upcoming Rides
	shift2BikesEvents, err := scraperhelpers.GetRides()
	if err != nil {
		fmt.Println("failed to get ride data")
	}

	var rideLocations []scraperhelpers.Location
	for i := range shift2BikesEvents.Events {
		event := &shift2BikesEvents.Events[i]
		event.SourcedFrom = EVENT_SOURCE
		event.CityCode = PDX_CITY_CODE

		// parse Starting location
		location := scraperhelpers.CreateLocationFromEvent(event)
		geocodeQuery := scraperhelpers.CreateGeoCodingQuery(&location)
		normalizedQuery := strings.ToLower(geocodeQuery)

		location.Query = geocodeQuery

		// locations where coords were avialable in the ride data
		if !location.NeedsGeocoding {
			event.Location = location
			fmt.Printf("SKIP: Lat: %v, Lng: %v\n", location.Latitude, location.Longitude)
			continue
		}

		// Fallback to default Portland Coords if no good address
		if location.Address == "" && location.Venue == "" {
			location.Query = FALLBACK_QUERY
			location.Latitude = FALLBACK_LAT
			location.Longitude = FALLBACK_LNG
			location.NeedsGeocoding = false
			event.Location = location
			fmt.Println("FALLBACK: to default Portland Coords")
			continue
		}

		// check cache for location
		var cachedLoc scraperhelpers.GeoCodeCached
		var found bool
		if cachedLoc, found = geocodeCache[normalizedQuery]; found {
			location.Latitude = cachedLoc.Latitude
			location.Longitude = cachedLoc.Longitude
			location.NeedsGeocoding = false
			event.Location = location
			fmt.Printf("SKIP (CACHE): %s\n", geocodeQuery)
			continue
		}
		// make request to geocode API for location
		fmt.Printf("GEOCODE: %s\n", geocodeQuery)

		lat, lng, err := scraperhelpers.GeocodeQuery(geocodeQuery)
		if err != nil {
			slog.Error("Unable to geocode query, using fall back coords", "error", err.Error(), "query", geocodeQuery)
			location.Query = FALLBACK_QUERY
			location.Latitude = FALLBACK_LAT
			location.Longitude = FALLBACK_LNG
			location.NeedsGeocoding = false
			event.Location = location
			continue
		}

		location.Latitude = lat
		location.Longitude = lng
		location.NeedsGeocoding = false
		event.Location = location

		// prevent from running geocode API twice in the same run
		geocodeCache[normalizedQuery] = scraperhelpers.GeoCodeCached{
			Query:     normalizedQuery,
			Latitude:  location.Latitude,
			Longitude: location.Longitude,
		}

		// append any starting location that need to be geocoded
		rideLocations = append(rideLocations, location)
	}

	// Get Locations ready to upsert into db
	if err = scraperhelpers.BulkUpsertGeocodeData(db, rideLocations); err != nil {
		slog.Error("unable to bulk upsert ride locations", "locations_len", len(rideLocations), "error", err.Error())
		log.Fatalf("unable to bulk upsert ride locations: %v", err)
	}

	// store ride information
	if err = scraperhelpers.BulkUpsertRideData(db, shift2BikesEvents.Events); err != nil {
		slog.Error("unable to bulk upsert ride data", "locations_len", len(rideLocations), "error", err.Error())
		log.Fatalf("unable to bulk upsert ride data: %v", err)

	}

}
