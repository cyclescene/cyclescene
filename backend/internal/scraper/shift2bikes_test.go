package scraper

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestBuildShift2BikesURLUpcomingUsesDocumentedDateRange(t *testing.T) {
	rawURL, err := buildShift2BikesURLUpcoming()
	if err != nil {
		t.Fatalf("build upcoming URL: %v", err)
	}

	startDate, endDate := parseShift2BikesRange(t, rawURL)
	if startDate.After(endDate) {
		t.Fatalf("start date must not be after end date: start=%s end=%s", startDate, endDate)
	}

	assertInclusiveDayRange(t, startDate, endDate, 100)
	assertDateOnlyQuery(t, rawURL)
}

func TestBuildShift2BikesURLPastUsesAscendingDateRange(t *testing.T) {
	rawURL, err := buildShift2BikesURLPast()
	if err != nil {
		t.Fatalf("build past URL: %v", err)
	}

	startDate, endDate := parseShift2BikesRange(t, rawURL)
	if startDate.After(endDate) {
		t.Fatalf("start date must not be after end date: start=%s end=%s", startDate, endDate)
	}

	assertInclusiveDayRange(t, startDate, endDate, 100)
	assertDateOnlyQuery(t, rawURL)
}

func parseShift2BikesRange(t *testing.T, rawURL string) (time.Time, time.Time) {
	t.Helper()

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}

	params := parsedURL.Query()
	startValue := params.Get("startdate")
	endValue := params.Get("enddate")

	startDate, err := time.Parse(time.DateOnly, startValue)
	if err != nil {
		t.Fatalf("parse startdate %q: %v", startValue, err)
	}

	endDate, err := time.Parse(time.DateOnly, endValue)
	if err != nil {
		t.Fatalf("parse enddate %q: %v", endValue, err)
	}

	return startDate, endDate
}

func assertInclusiveDayRange(t *testing.T, startDate, endDate time.Time, expectedDays int) {
	t.Helper()

	days := int(endDate.Sub(startDate).Hours()/24) + 1
	if days != expectedDays {
		t.Fatalf("expected %d inclusive days, got %d: start=%s end=%s", expectedDays, days, startDate, endDate)
	}
}

func assertDateOnlyQuery(t *testing.T, rawURL string) {
	t.Helper()

	if strings.Contains(rawURL, "T") || strings.Contains(rawURL, "%3A") {
		t.Fatalf("expected YYYY-MM-DD date query, got %s", rawURL)
	}
}
