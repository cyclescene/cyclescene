package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spacesedan/cyclescene/backend/internal/scraper"
	"google.golang.org/api/calendar/v3"
)

// CalendarConfig associates a Google Calendar with the Cycle Scene city where
// its rides should appear. Add any number of entries through GOOGLE_CALENDARS.
type CalendarConfig struct {
	ID   string `json:"id"`
	City string `json:"city"`
}

type calendarEvent struct {
	Event          *calendar.Event
	Calendar       CalendarConfig
	Latitude       float64
	Longitude      float64
	HasCoordinates bool
	GeocodeError   error
}

func loadCalendars(raw string) ([]CalendarConfig, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("must be a JSON array, for example [{\"id\":\"calendar@example.com\",\"city\":\"la\"}]")
	}

	var calendars []CalendarConfig
	if err := json.Unmarshal([]byte(raw), &calendars); err != nil {
		return nil, fmt.Errorf("must be valid JSON: %w", err)
	}
	if len(calendars) == 0 {
		return nil, fmt.Errorf("must contain at least one calendar")
	}

	seen := make(map[string]struct{}, len(calendars))
	for i := range calendars {
		calendars[i].ID = strings.TrimSpace(calendars[i].ID)
		calendars[i].City = strings.TrimSpace(calendars[i].City)
		if calendars[i].ID == "" || calendars[i].City == "" {
			return nil, fmt.Errorf("entry %d must include non-empty id and city", i)
		}
		if !scraper.HasCity(calendars[i].City) {
			return nil, fmt.Errorf("entry %d has unsupported city code %q", i, calendars[i].City)
		}
		if _, exists := seen[calendars[i].ID]; exists {
			return nil, fmt.Errorf("calendar ID %q appears more than once", calendars[i].ID)
		}
		seen[calendars[i].ID] = struct{}{}
	}

	return calendars, nil
}
