package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spacesedan/cyclescene/backend/internal/scraper"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func main() {
	var err error
	if os.Getenv("APP_ENV") == "dev" {
		// Used while in development
		err = godotenv.Load()
		if err != nil {
			log.Fatalf("failed to read environment variables: %v", err)
		}
	}
	ctx := context.Background()

	calendars, err := loadCalendars(os.Getenv("GOOGLE_CALENDARS"))
	if err != nil {
		log.Fatalf("invalid GOOGLE_CALENDARS configuration: %v", err)
	}

	// Use Application Default Credentials. On Cloud Run this resolves to the
	// job's service account; locally it resolves to gcloud ADC credentials.
	calendarService, err := calendar.NewService(ctx, option.WithScopes(calendar.CalendarReadonlyScope))
	if err != nil {
		log.Fatalf("Unable to create calendar service: %v", err)
	}

	db, err := connectToDB()
	if err != nil {
		log.Fatalf("unable to connect to Turso: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("unable to ping Turso: %v", err)
	}
	slog.Info("connected to Turso for Google Calendar import")

	geocodeCache, err := scraper.GetGeocodeCache(db)
	if err != nil {
		log.Fatalf("unable to load geocode cache: %v", err)
	}
	slog.Info("starting Google Calendar import", "calendar_count", len(calendars), "geocode_cache_entries", len(geocodeCache))

	allEvents := []calendarEvent{}

	timeStart := time.Now()
	for _, calendarConfig := range calendars {
		events := getCalendarEvents(calendarService, calendarConfig)
		allEvents = append(allEvents, events...)
		slog.Info("fetched Google Calendar events", "calendar_id", calendarConfig.ID, "city", calendarConfig.City, "event_count", len(events))
	}
	slog.Info("finished fetching Google Calendar events", "event_count", len(allEvents), "duration", time.Since(timeStart))

	newLocations, geocodeStats := geocodeCalendarEvents(allEvents, geocodeCache, scraper.GeocodeQuery)
	slog.Info("finished geocoding Calendar locations",
		"event_count", len(allEvents),
		"persistent_cache_hits", geocodeStats.PersistentCacheHits,
		"in_memory_cache_hits", geocodeStats.InMemoryCacheHits,
		"geocode_requests", geocodeStats.Requests,
		"locations_without_address", geocodeStats.MissingLocationEvents,
		"geocode_failures", geocodeStats.FailedEvents,
		"new_cache_entries", len(newLocations),
		"duration", time.Since(timeStart),
	)
	if err := scraper.BulkUpsertGeocodeData(db, newLocations); err != nil {
		log.Fatalf("unable to persist geocode cache: %v", err)
	}
	slog.Info("persisted Google Calendar geocode cache entries", "entry_count", len(newLocations))
	if err := syncCalendarEvents(db, allEvents, calendars); err != nil {
		log.Fatalf("unable to sync Google Calendar events: %v", err)
	}
	slog.Info("Google Calendar import completed successfully", "event_count", len(allEvents), "duration", time.Since(timeStart))
}

func connectToDB() (*sql.DB, error) {
	dbURL := strings.TrimSpace(os.Getenv("TURSO_DB_URL"))
	authToken := strings.TrimSpace(os.Getenv("TURSO_DB_RW_TOKEN"))
	if dbURL == "" || authToken == "" {
		return nil, fmt.Errorf("TURSO_DB_URL and TURSO_DB_RW_TOKEN are required")
	}
	return sql.Open("libsql", fmt.Sprintf("%s?authToken=%s", dbURL, authToken))
}

type geocodeFunc func(query, cityCode string) (float64, float64, error)

type geocodeStats struct {
	PersistentCacheHits   int
	InMemoryCacheHits     int
	Requests              int
	MissingLocationEvents int
	FailedEvents          int
}

// geocodeCalendarEvents resolves calendar locations once per city/location
// pair. The results are attached to the event for the later database sync.
func geocodeCalendarEvents(events []calendarEvent, geocodeCache map[string]scraper.GeoCodeCached, geocode geocodeFunc) ([]scraper.Location, geocodeStats) {
	type geocodeResult struct {
		latitude  float64
		longitude float64
		err       error
	}

	seen := make(map[string]geocodeResult)
	newLocations := make([]scraper.Location, 0)
	stats := geocodeStats{}
	for i := range events {
		location := strings.TrimSpace(events[i].Event.Location)
		if location == "" {
			stats.MissingLocationEvents++
			continue
		}

		key := calendarGeocodeCacheKey(events[i].Calendar.City, location)
		if cached, found := geocodeCache[key]; found {
			stats.PersistentCacheHits++
			events[i].Latitude = cached.Latitude
			events[i].Longitude = cached.Longitude
			events[i].HasCoordinates = true
			continue
		}

		result, found := seen[key]
		if found {
			stats.InMemoryCacheHits++
		}
		if !found {
			stats.Requests++
			result.latitude, result.longitude, result.err = geocode(location, events[i].Calendar.City)
			seen[key] = result
			if result.err == nil {
				newLocations = append(newLocations, scraper.Location{
					Query:     key,
					Latitude:  result.latitude,
					Longitude: result.longitude,
					City:      events[i].Calendar.City,
				})
			}
		}

		if result.err != nil {
			events[i].GeocodeError = result.err
			stats.FailedEvents++
			slog.Warn("unable to geocode Calendar event location", "city", events[i].Calendar.City, "location", location, "error", result.err)
			continue
		}

		events[i].Latitude = result.latitude
		events[i].Longitude = result.longitude
		events[i].HasCoordinates = true
	}

	return newLocations, stats
}

func calendarGeocodeCacheKey(city, location string) string {
	normalizedLocation := strings.Join(strings.Fields(strings.ToLower(location)), " ")
	return "gcal:" + city + ":" + normalizedLocation
}

func getCalendarEvents(service *calendar.Service, calendarConfig CalendarConfig) []calendarEvent {
	now := time.Now()
	timeMin := now.Format(time.RFC3339)
	timeMax := now.AddDate(0, 100, 0).Format(time.RFC3339)
	allEvents := []calendarEvent{}
	pageToken := ""
	pageCount := 0
	for {
		req := service.Events.List(calendarConfig.ID).
			ShowDeleted(false).
			SingleEvents(true).
			OrderBy("startTime").
			TimeMin(timeMin).
			TimeMax(timeMax)

		if pageToken != "" {
			req.PageToken(pageToken)
		}
		events, err := req.Do()
		if err != nil {
			log.Fatalf("Unable to retrieve events: %v", err)
		}
		pageCount++
		for _, event := range events.Items {
			allEvents = append(allEvents, calendarEvent{Event: event, Calendar: calendarConfig})
		}
		pageToken = events.NextPageToken
		if pageToken == "" {
			break
		}

	}
	slog.Info("completed Google Calendar API pagination", "calendar_id", calendarConfig.ID, "city", calendarConfig.City, "page_count", pageCount, "event_count", len(allEvents))

	return allEvents
}
