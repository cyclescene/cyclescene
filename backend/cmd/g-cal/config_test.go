package main

import (
	"testing"

	"github.com/spacesedan/cyclescene/backend/internal/scraper"
	"google.golang.org/api/calendar/v3"
)

func TestLoadCalendars(t *testing.T) {
	calendars, err := loadCalendars(`[
        {"id": "la-rides@example.com", "city": "la"},
        {"id": "pdx-rides@example.com", "city": "pdx"}
    ]`)
	if err != nil {
		t.Fatalf("loadCalendars() error = %v", err)
	}
	if len(calendars) != 2 || calendars[0].City != "la" || calendars[1].ID != "pdx-rides@example.com" {
		t.Fatalf("loadCalendars() = %#v, want both configured calendars", calendars)
	}
}

func TestLoadCalendarsRejectsInvalidConfiguration(t *testing.T) {
	tests := []string{
		"",
		"[]",
		`[{"id":"calendar@example.com"}]`,
		`[{"id":"calendar@example.com","city":"unknown"}]`,
		`[{"id":"calendar@example.com","city":"la"},{"id":"calendar@example.com","city":"pdx"}]`,
	}

	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			if _, err := loadCalendars(raw); err == nil {
				t.Fatal("loadCalendars() error = nil, want configuration error")
			}
		})
	}
}

func TestGeocodeCalendarEventsCachesLocations(t *testing.T) {
	events := []calendarEvent{
		{Event: &calendar.Event{Location: "Griffith Park"}, Calendar: CalendarConfig{City: "la"}},
		{Event: &calendar.Event{Location: " griffith park "}, Calendar: CalendarConfig{City: "la"}},
		{Event: &calendar.Event{}, Calendar: CalendarConfig{City: "la"}},
	}

	calls := 0
	newLocations := geocodeCalendarEvents(events, map[string]scraper.GeoCodeCached{}, func(query, city string) (float64, float64, error) {
		calls++
		if query != "Griffith Park" || city != "la" {
			t.Fatalf("geocode called with (%q, %q)", query, city)
		}
		return 34.1367, -118.2942, nil
	})

	if calls != 1 {
		t.Fatalf("geocode calls = %d, want 1", calls)
	}
	if len(newLocations) != 1 || newLocations[0].Query != "gcal:la:griffith park" {
		t.Fatalf("new geocode locations = %#v", newLocations)
	}
	if !events[0].HasCoordinates || !events[1].HasCoordinates {
		t.Fatal("expected locations to be geocoded")
	}
	if events[0].Latitude != 34.1367 || events[1].Longitude != -118.2942 {
		t.Fatalf("unexpected coordinates: %#v %#v", events[0], events[1])
	}
	if events[2].HasCoordinates || events[2].GeocodeError != nil {
		t.Fatalf("empty location should be skipped: %#v", events[2])
	}
}

func TestGeocodeCalendarEventsUsesPersistentCache(t *testing.T) {
	events := []calendarEvent{
		{Event: &calendar.Event{Location: "Griffith Park"}, Calendar: CalendarConfig{City: "la"}},
	}
	cache := map[string]scraper.GeoCodeCached{
		"gcal:la:griffith park": {Latitude: 34.1367, Longitude: -118.2942},
	}

	newLocations := geocodeCalendarEvents(events, cache, func(string, string) (float64, float64, error) {
		t.Fatal("geocoder should not be called for a cached location")
		return 0, 0, nil
	})

	if len(newLocations) != 0 || !events[0].HasCoordinates {
		t.Fatalf("expected coordinates from persistent cache: locations=%#v event=%#v", newLocations, events[0])
	}
}
