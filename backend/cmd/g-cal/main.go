package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
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

	geocodeCache, err := scraper.GetGeocodeCache(db)
	if err != nil {
		log.Fatalf("unable to load geocode cache: %v", err)
	}

	allEvents := []calendarEvent{}

	timeStart := time.Now()
	for _, calendarConfig := range calendars {
		events := getCalendarEvents(calendarService, calendarConfig)
		allEvents = append(allEvents, events...)
	}
	newLocations := geocodeCalendarEvents(allEvents, geocodeCache, scraper.GeocodeQuery)
	if err := scraper.BulkUpsertGeocodeData(db, newLocations); err != nil {
		log.Fatalf("unable to persist geocode cache: %v", err)
	}
	if err := syncCalendarEvents(db, allEvents, calendars); err != nil {
		log.Fatalf("unable to sync Google Calendar events: %v", err)
	}

	allEvents = sortEventsChronlogically(allEvents)

	for _, event := range allEvents {
		date := event.Event.Start.DateTime
		if date == "" {
			date = event.Event.Start.Date
		}
		if event.GeocodeError != nil {
			fmt.Printf("[%s] %s (%s) - where: %s [geocoding failed: %v]\n", event.Calendar.City, event.Event.Summary, date, event.Event.Location, event.GeocodeError)
			continue
		}
		if !event.HasCoordinates {
			fmt.Printf("[%s] %s (%s) - where: %s [no location to geocode]\n", event.Calendar.City, event.Event.Summary, date, event.Event.Location)
			continue
		}
		fmt.Printf("[%s] %s (%s) - where: %s [%.6f, %.6f]\n", event.Calendar.City, event.Event.Summary, date, event.Event.Location, event.Latitude, event.Longitude)
	}

	fmt.Printf("Total events: %d Duration: %s\n", len(allEvents), time.Since(timeStart))
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

// geocodeCalendarEvents resolves calendar locations once per city/location
// pair. The results are attached to the event for the later database sync.
func geocodeCalendarEvents(events []calendarEvent, geocodeCache map[string]scraper.GeoCodeCached, geocode geocodeFunc) []scraper.Location {
	type geocodeResult struct {
		latitude  float64
		longitude float64
		err       error
	}

	seen := make(map[string]geocodeResult)
	newLocations := make([]scraper.Location, 0)
	for i := range events {
		location := strings.TrimSpace(events[i].Event.Location)
		if location == "" {
			continue
		}

		key := calendarGeocodeCacheKey(events[i].Calendar.City, location)
		if cached, found := geocodeCache[key]; found {
			events[i].Latitude = cached.Latitude
			events[i].Longitude = cached.Longitude
			events[i].HasCoordinates = true
			continue
		}

		result, found := seen[key]
		if !found {
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
			log.Printf("unable to geocode calendar event location: city=%s location=%q error=%v", events[i].Calendar.City, location, result.err)
			continue
		}

		events[i].Latitude = result.latitude
		events[i].Longitude = result.longitude
		events[i].HasCoordinates = true
	}

	return newLocations
}

func calendarGeocodeCacheKey(city, location string) string {
	normalizedLocation := strings.Join(strings.Fields(strings.ToLower(location)), " ")
	return "gcal:" + city + ":" + normalizedLocation
}

func sortEventsChronlogically(events []calendarEvent) []calendarEvent {
	sort.Slice(events, func(i, j int) bool {
		timeI := getEventStartTime(events[i].Event)
		timeJ := getEventStartTime(events[j].Event)
		return timeI.Before(timeJ)

	})
	return events
}

func getEventStartTime(event *calendar.Event) time.Time {
	startStr := event.Start.DateTime
	if startStr == "" {
		startStr = event.Start.Date
		t, _ := time.Parse("2006-01-02", startStr)
		return t
	}
	t, _ := time.Parse(time.RFC3339, startStr)
	return t
}

func getCalendarEvents(service *calendar.Service, calendarConfig CalendarConfig) []calendarEvent {
	now := time.Now()
	timeMin := now.Format(time.RFC3339)
	timeMax := now.AddDate(0, 100, 0).Format(time.RFC3339)
	allEvents := []calendarEvent{}
	pageToken := ""
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
		for _, event := range events.Items {
			allEvents = append(allEvents, calendarEvent{Event: event, Calendar: calendarConfig})
		}
		pageToken = events.NextPageToken
		if pageToken == "" {
			break
		}

	}

	return allEvents
}
