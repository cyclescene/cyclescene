package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	//
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"golang.org/x/exp/slog"
	scraperhelpers "scraperv2/lib/scraper-helpers"
)

const (
	FALLBACK_LAT   = 45.523064
	FALLBACK_LNG   = -122.676483
	FALLBACK_QUERY = "fallback"
	EVENT_SOURCE   = "Shift2Bikes"
)

func Main() {
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
		log.Fatal("failed to create tables check turdo credentials")
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
		fmt.Println("POOP")
	}

	var rideLocations []scraperhelpers.Location
	for i := range shift2BikesEvents.Events {
		event := &shift2BikesEvents.Events[i]
		event.SourcedFrom = EVENT_SOURCE
		// parse Starting location
		location := scraperhelpers.CreateLocationFromEvent(event)
		geocodeQuery := scraperhelpers.CreateGeoCodingQuery(&location)

		location.Query = geocodeQuery

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
		if cachedLoc, found := geocodeCache[strings.ToLower(geocodeQuery)]; found {
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
		rideLocations = append(rideLocations, location)
	}

	// Get Locations ready to upsert into db
	if err = scraperhelpers.BulkUpsertGeocodeData(db, rideLocations); err != nil {
		slog.Error("unable to builk upsert ride locations", "locations_len", len(rideLocations), "error", err.Error())
		panic(err)
	}

	// store ride information
	if err = scraperhelpers.BulkUpsertRideData(db, shift2BikesEvents.Events); err != nil {
		slog.Error("unable to builk upsert ride data", "locations_len", len(rideLocations), "error", err.Error())
		panic(err)

	}

	///// DONE ///////////////////////////

}
