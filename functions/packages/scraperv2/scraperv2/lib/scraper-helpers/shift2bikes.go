package scraperhelpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func buildShift2BikesURLUpcoming() (string, error) {
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return "", fmt.Errorf("failed to load timezone location: %w", err)
	}

	nowInPortland := time.Now().In(location)

	year, month, day := nowInPortland.Date()
	startDate := time.Date(year, month, day, 0, 0, 0, 0, location)

	endDate := startDate.AddDate(0, 0, 90)

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

func buildShift2BikesURLPast() (string, error) {
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return "", fmt.Errorf("failed to load timezone location: %w", err)
	}

	nowInPortland := time.Now().In(location)

	year, month, day := nowInPortland.Date()
	startDate := time.Date(year, month, day, 0, 0, 0, 0, location)

	endDate := startDate.AddDate(0, 0, -90)

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

func fetchAndDecode(url string, target interface{}) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	return decoder.Decode(&target)
}

func getPastRides(events *Shift2BikeEvents) error {
	url, err := buildShift2BikesURLPast()
	if err != nil {
		return err
	}
	return fetchAndDecode(url, &events)

}
func getUpcomingRides(events *Shift2BikeEvents) error {
	url, err := buildShift2BikesURLUpcoming()
	if err != nil {
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
