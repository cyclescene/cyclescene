package scraper

import (
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var coordinatesRegex = regexp.MustCompile(
	`(-?\d+\.\d+)\s*[°\s]*[NnSs]?[,\s]+(-?\d+\.\d+)\s*[°\s]*[EeWw]?`,
)

const KeyPrecision = 6

func isValidPortlandCoordinate(lat float64, lng float64) bool {
	const minLat = 45.0
	const maxLat = 46.0
	const minLng = -123.0
	const maxLng = -121.5

	return lat >= minLat && lat <= maxLat && lng >= minLng && lng <= maxLng
}

func processGps(source string, loc *Location) bool {
	matches := coordinatesRegex.FindStringSubmatch(source)

	if len(matches) == 3 {
		lat, latErr := strconv.ParseFloat(matches[1], 64)
		lng, lngErr := strconv.ParseFloat(matches[2], 64)

		if latErr == nil && lngErr == nil {

			lat = math.Abs(lat)

			lng = -math.Abs(lng)

			if isValidPortlandCoordinate(lat, lng) {
				loc.Latitude = lat
				loc.Longitude = lng
				loc.NeedsGeocoding = false
				return true
			} else {
				slog.Warn("GPS coordinates extracted but out of Portland Range", "raw_source", source, "lat", lat, "lng", lng)
			}
		}
	}
	return false
}

func CreateLocationFromEvent(event *Shift2BikeEvent) Location {
	loc := Location{
		Address:        strings.TrimSpace(event.Address),
		Venue:          strings.TrimSpace(event.Venue),
		Details:        strings.TrimSpace(event.Details),
		NeedsGeocoding: true,
	}

	if processGps(event.Locdetails, &loc) {
		goto cleanup
	}

	if processGps(event.Details, &loc) {
		goto cleanup
	}

	if processGps(event.Address, &loc) {
		goto cleanup
	}

	if processGps(event.Venue, &loc) {
		goto cleanup
	}

	if loc.NeedsGeocoding {
		addressLower := strings.ToLower(loc.Address)
		if strings.EqualFold(addressLower, "tba") ||
			strings.EqualFold(addressLower, "tbd") ||
			strings.Contains(addressLower, "maps.app.goo") ||
			strings.Contains(addressLower, "http") ||
			loc.Address == "" {
			loc.Address = ""
			loc.Venue = ""
			loc.Details = ""
		}
	}

cleanup:
	return loc
}

func CreateGeoCodingQuery(loc *Location) string {
	address := loc.Address
	venue := loc.Venue

	var baseQuery string
	if address != "" && venue != "" {
		baseQuery = fmt.Sprintf("%s, %s", address, venue)
	} else if address != "" {
		baseQuery = address
	} else if venue != "" {
		baseQuery = venue
	} else {
		return ""
	}

	lowerQuery := strings.ToLower(baseQuery)

	if strings.Contains(lowerQuery, ", or") || strings.Contains(lowerQuery, ", wa") {
		return baseQuery
	}

	return fmt.Sprintf("%s, Portland, OR", baseQuery)
}

func CreateCanonicalCoordKey(lat float64, lng float64) string {
	factor := math.Pow(10, KeyPrecision)

	rLat := math.Round(lat*factor) / factor
	rLng := math.Round(lng*factor) / factor

	format := fmt.Sprintf("%%.%df,%%.%df", KeyPrecision, KeyPrecision)

	return fmt.Sprintf(format, rLat, rLng)
}
