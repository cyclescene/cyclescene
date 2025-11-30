package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spacesedan/cyclescene/functions/internal/routes"
	"github.com/spacesedan/cyclescene/functions/internal/scraper"
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
	if os.Getenv("APP_ENV") == "dev" {
		_ = godotenv.Load()
	}
	// DB Vars
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_DB_RW_TOKEN") == "" {
		log.Fatal("FATAL: Turso env variable not set properly")
	}
	//
	// // set up logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})))
	slog.Info("Starting PDX scraper service")
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

	////// READY TO START (v1.1.0) /////////////////////////

	// Initialize route services
	httpClient := &http.Client{Timeout: 30 * time.Second}
	stravaToken := os.Getenv("STRAVA_ACCESS_TOKEN")
	rwgpsAuthToken := os.Getenv("RWGPS_AUTH_TOKEN")
	rwgpsAPIKey := os.Getenv("RWGPS_API_KEY")

	// Log environment variable status
	slog.Info("Route service environment variables",
		"hasStravaToken", stravaToken != "",
		"hasRwgpsAuthToken", rwgpsAuthToken != "",
		"hasRwgpsAPIKey", rwgpsAPIKey != "",
	)

	if rwgpsAPIKey == "" {
		slog.Warn("RWGPS_API_KEY not set - RideWithGPS route fetching may fail without proper authentication")
	}
	if rwgpsAuthToken == "" {
		slog.Warn("RWGPS_AUTH_TOKEN not set - RideWithGPS route fetching may fail without proper authentication")
	}
	if stravaToken == "" {
		slog.Warn("STRAVA_ACCESS_TOKEN not set - Strava route fetching will not work")
	}

	routeFetcher := routes.NewRouteFetcher(httpClient, stravaToken, rwgpsAuthToken, rwgpsAPIKey)
	routeRepo := routes.NewRepository(db)

	// Load existing routes into cache to avoid reprocessing
	existingRoutes, err := routeRepo.GetAllRoutes(context.Background())
	if err != nil {
		slog.Error("failed to load existing routes", "error", err)
	}
	routeCache := make(map[string]string) // "source:sourceID" -> routeID
	for _, route := range existingRoutes {
		cacheKey := fmt.Sprintf("%s:%s", route.Source, route.SourceID)
		routeCache[cacheKey] = route.ID
		slog.Info("loaded route from cache", "source", route.Source, "sourceID", route.SourceID, "routeID", route.ID)
	}

	// get all previously saved locations from DB
	geocodeCache, err := scraper.GetGeocodeCache(db)
	if err != nil {
		log.Fatalf("something went wrong: %s\n", err.Error())
	}

	// get upcoming Rides
	shift2BikesEvents, err := scraper.GetRides()
	if err != nil {
		fmt.Println("failed to get ride data")
	}

	var rideLocations []scraper.Location
	for i := range shift2BikesEvents.Events {
		event := &shift2BikesEvents.Events[i]
		event.SourcedFrom = EVENT_SOURCE
		event.CityCode = PDX_CITY_CODE

		// Extract and process route if present in event details
		routeURL := routes.ExtractRouteURLFromDescription(event.Details)
		if routeURL != "" {
			slog.Info("route URL found in event description", "routeURL", routeURL, "event", event.Title)
			source, sourceID, err := routes.ParseRouteURL(routeURL)
			if err != nil {
				slog.Warn("failed to parse route URL", "error", err, "routeURL", routeURL, "event", event.Title)
			} else {
				slog.Info("parsed route URL", "source", source, "sourceID", sourceID, "routeURL", routeURL, "event", event.Title)
				// Check if route is already in cache
				cacheKey := fmt.Sprintf("%s:%s", source, sourceID)
				if routeID, found := routeCache[cacheKey]; found {
					slog.Info("route found in cache", "source", source, "sourceID", sourceID, "routeID", routeID, "event", event.Title)
					// TODO: Link route to shift2bikes event when upserted
				} else {
					// Fetch and process new route
					slog.Info("processing new route", "source", source, "sourceID", sourceID, "routeURL", routeURL, "event", event.Title)
					feature, err := routeFetcher.FetchAndConvert(routeURL)
					if err != nil {
						slog.Warn("failed to fetch route", "error", err, "routeURL", routeURL, "source", source, "sourceID", sourceID, "event", event.Title)
					} else {
						// Extract distance from properties
						var distanceKm, distanceMi float64
						if km, ok := feature.Properties["distance_km"].(float64); ok {
							distanceKm = km
						}
						if mi, ok := feature.Properties["distance_mi"].(float64); ok {
							distanceMi = mi
						}

						// Create route in database
						newRouteID, err := routeRepo.CreateRoute(context.Background(), source, sourceID, routeURL, feature, distanceKm, distanceMi)
						if err != nil {
							slog.Error("failed to create route", "error", err, "routeURL", routeURL, "event", event.Title)
						} else {
							slog.Info("route created successfully", "source", source, "sourceID", sourceID, "routeID", newRouteID, "distance_km", distanceKm, "event", event.Title)
							routeCache[cacheKey] = newRouteID
							// TODO: Link route to shift2bikes event when upserted
						}
					}
				}
			}
		}

		// parse Starting location
		location := scraper.CreateLocationFromEvent(event)
		geocodeQuery := scraper.CreateGeoCodingQuery(&location)
		normalizedQuery := strings.ToLower(geocodeQuery)

		location.Query = geocodeQuery
		location.City = PDX_CITY_CODE

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
		var cachedLoc scraper.GeoCodeCached
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

		lat, lng, err := scraper.GeocodeQuery(geocodeQuery, PDX_CITY_CODE)
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
		geocodeCache[normalizedQuery] = scraper.GeoCodeCached{
			Query:     normalizedQuery,
			Latitude:  location.Latitude,
			Longitude: location.Longitude,
		}

		// append any starting location that need to be geocoded
		rideLocations = append(rideLocations, location)
	}

	// Get Locations ready to upsert into db
	if err = scraper.BulkUpsertGeocodeData(db, rideLocations); err != nil {
		slog.Error("unable to bulk upsert ride locations", "locations_len", len(rideLocations), "error", err.Error())
		log.Fatalf("unable to bulk upsert ride locations: %v", err)
	}

	// store ride information
	if err = scraper.BulkUpsertRideData(db, shift2BikesEvents.Events); err != nil {
		slog.Error("unable to bulk upsert ride data", "locations_len", len(rideLocations), "error", err.Error())
		log.Fatalf("unable to bulk upsert ride data: %v", err)

	}

}
