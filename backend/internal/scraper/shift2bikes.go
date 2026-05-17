package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func buildShift2BikesURL(startDate, endDate time.Time) (string, error) {
	baseURL := "https://www.shift2bikes.org/api/events.php"

	finalURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse Shift2Bikes API URL: %w", err)
	}

	params := url.Values{}
	params.Set("startdate", startDate.Format(time.DateOnly))
	params.Set("enddate", endDate.Format(time.DateOnly))

	finalURL.RawQuery = params.Encode()

	return finalURL.String(), nil
}

func shift2BikesToday() (time.Time, error) {
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to load timezone location: %w", err)
	}

	nowInPortland := time.Now().In(location)
	year, month, day := nowInPortland.Date()

	return time.Date(year, month, day, 0, 0, 0, 0, location), nil
}

func buildShift2BikesURLUpcoming() (string, error) {
	startDate, err := shift2BikesToday()
	if err != nil {
		return "", err
	}

	return buildShift2BikesURL(startDate, startDate.AddDate(0, 0, 99))
}

func buildShift2BikesURLPast() (string, error) {
	endDate, err := shift2BikesToday()
	if err != nil {
		return "", err
	}

	startDate := endDate.AddDate(0, 0, -99)

	return buildShift2BikesURL(startDate, endDate)
}

func fetchAndDecode(url string, target any) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "CycleScene")
	req.Header.Set("x-cycle-scene-version", "1.0.0")
	req.Header.Set("Api-Version", "3")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("shift2bikes API returned status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	decoder := json.NewDecoder(res.Body)
	return decoder.Decode(&target)
}

func getPastRides(events *Shift2BikeEvents) error {
	url, err := buildShift2BikesURLPast()
	if err != nil {
		slog.Error("shift2Bikes API request failed", "url", url, "error", err.Error())
		return err
	}
	return fetchAndDecode(url, &events)

}
func getUpcomingRides(events *Shift2BikeEvents) error {
	url, err := buildShift2BikesURLUpcoming()
	if err != nil {
		slog.Error("shift2Bikes API request failed", "url", url, "error", err.Error())
		return err
	}
	return fetchAndDecode(url, &events)
}

func GetRides() (Shift2BikeEvents, error) {
	var allEvents Shift2BikeEvents
	var upcomingEvents Shift2BikeEvents
	var pastEvents Shift2BikeEvents

	if err := getUpcomingRides(&upcomingEvents); err != nil {
		return allEvents, err
	}
	if err := getPastRides(&pastEvents); err != nil {
		return allEvents, err
	}

	allEvents.Events = append(
		upcomingEvents.Events,
		pastEvents.Events...,
	)

	return allEvents, nil
}
