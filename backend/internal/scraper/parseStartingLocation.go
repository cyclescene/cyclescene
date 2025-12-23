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

var (
	cardinalDirections = map[string]bool{"n": true, "ne": true, "nw": true, "s": true, "se": true, "sw": true, "e": true, "w": true}
	streetTypes        = map[string]string{
		"avenue":    "Ave",
		"ave":       "Ave",
		"drive":     "Dr",
		"dr":        "Dr",
		"street":    "St",
		"st":        "St",
		"road":      "Rd",
		"rd":        "Rd",
		"boulevard": "Blvd",
		"blvd":      "Blvd",
		"way":       "Way",
		"lane":      "Ln",
		"ln":        "Ln",
		"court":     "Ct",
		"ct":        "Ct",
		"place":     "Pl",
		"pl":        "Pl",
		"circle":    "Cir",
		"cir":       "Cir",
		"terrace":   "Ter",
		"ter":       "Ter",
		"parkway":   "Pkwy",
		"pkwy":      "Pkwy",
	}
	descriptorWords = map[string]bool{
		"trailhead": true, "trail": true, "park": true, "parking": true, "lot": true,
		"entrance": true, "gate": true, "bridge": true, "crossing": true, "greenway": true,
		"neighborhood": true, "area": true, "zone": true, "intersection": true,
		"northern": true, "southern": true, "eastern": true, "western": true,
		"north": true, "south": true, "east": true, "west": true,
	}
)

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
	// Extract city/state from the raw address
	rawAddress := strings.TrimSpace(event.Address)
	extractedCity := extractCityState(rawAddress)

	// Normalize the address
	normalizedAddress := NormalizeAddress(rawAddress)

	loc := Location{
		Address:        normalizedAddress,
		Venue:          strings.TrimSpace(event.Venue),
		Details:        strings.TrimSpace(event.Details),
		CityState:      extractedCity, // Store extracted city/state
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
		}
	}

cleanup:
	return loc
}

func CreateGeoCodingQuery(loc *Location) string {
	address := loc.Address

	// Use normalized address if available, otherwise use venue as fallback
	var baseQuery string
	if address != "" {
		baseQuery = address
	} else if loc.Venue != "" {
		baseQuery = loc.Venue
	} else {
		return ""
	}

	// Use extracted city/state if available, otherwise default to Portland, OR
	if loc.CityState != "" {
		return fmt.Sprintf("%s, %s", baseQuery, loc.CityState)
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

// extractCityState extracts city and state from an address string
// Returns the extracted city/state (e.g., "Portland, OR") or empty string if not found
func extractCityState(address string) string {
	// Look for patterns like "Portland, OR" or "Vancouver, WA"
	cityStateRegex := regexp.MustCompile(`(?i)(Portland|Vancouver)\s*,\s*(OR|WA|Oregon|Washington)`)
	matches := cityStateRegex.FindStringSubmatch(address)
	if len(matches) >= 3 {
		city := strings.Title(strings.ToLower(matches[1]))
		state := strings.ToUpper(matches[2])
		// Normalize state abbreviations
		if state == "OREGON" {
			state = "OR"
		} else if state == "WASHINGTON" {
			state = "WA"
		}
		return fmt.Sprintf("%s, %s", city, state)
	}
	return ""
}

// tokenizeAddress splits an address into normalized tokens
func tokenizeAddress(address string) []string {
	// Remove extra whitespace and convert to lowercase for processing
	address = strings.TrimSpace(address)
	if address == "" {
		return []string{}
	}

	// Split on common delimiters: commas, ampersands, and whitespace
	// But preserve & and "and" as tokens
	address = strings.ReplaceAll(address, ",", " , ")
	address = strings.ReplaceAll(address, "&", " & ")
	address = regexp.MustCompile(`\s+`).ReplaceAllString(address, " ")

	tokens := strings.Fields(address)
	var normalized []string

	for _, token := range tokens {
		token = strings.ToLower(token)
		// Clean punctuation
		token = strings.Trim(token, ".,;:")
		if token != "" {
			normalized = append(normalized, token)
		}
	}

	return normalized
}

// hasStreetNumber checks if any token is a street number (e.g., "3947", "41st", "12th")
func hasStreetNumber(tokens []string) bool {
	for _, token := range tokens {
		// Check for pure numbers: 3947
		if regexp.MustCompile(`^\d+$`).MatchString(token) {
			return true
		}
		// Check for ordinal numbers: 41st, 12th, 3rd, etc
		if regexp.MustCompile(`^\d+(st|nd|rd|th)$`).MatchString(token) {
			return true
		}
	}
	return false
}

// hasIntersectionMarker checks if address contains & or " and "
func hasIntersectionMarker(tokens []string) bool {
	for _, token := range tokens {
		if token == "&" || token == "and" {
			return true
		}
	}
	return false
}

// parseIntersection extracts and normalizes an intersection address
// Example: "SE 17th Avenue & SE Andrews Drive, Trolley Trail..." → "SE 17th & SE Andrews"
func parseIntersection(tokens []string) string {
	var result []string
	for _, token := range tokens {
		// Stop at comma (indicates descriptor section)
		if token == "," {
			break
		}

		// Stop at descriptor words (unless they're cardinal directions which are valid in intersections)
		if descriptorWords[token] && !cardinalDirections[token] {
			break
		}

		// Abbreviate street types
		if abbrev, ok := streetTypes[token]; ok {
			result = append(result, abbrev)
		} else {
			result = append(result, token)
		}
	}

	return strings.Join(result, " ")
}

// parseFullAddress extracts a street address with number
// Example: "3947 N Williams Ave, Portland, OR 97227" → "3947 N Williams Ave"
func parseFullAddress(tokens []string) string {
	var result []string
	foundNumber := false

	for _, token := range tokens {
		// Stop at comma or state abbreviations
		if token == "," || token == "or" || token == "wa" || token == "portland" || token == "vancouver" {
			break
		}

		// Check if this is a street number
		if (regexp.MustCompile(`^\d+$`).MatchString(token) || regexp.MustCompile(`^\d+(st|nd|rd|th)$`).MatchString(token)) && !foundNumber {
			foundNumber = true
			result = append(result, token)
			continue
		}

		// Only include tokens up to and including the street type
		if foundNumber {
			// Abbreviate street types
			if abbrev, ok := streetTypes[token]; ok {
				result = append(result, abbrev)
				break // Stop after street type
			}
			result = append(result, token)
		}
	}

	return strings.Join(result, " ")
}

// parseLandmark extracts just the first landmark
// Example: "Blumenauer Bridge, Flanders Crossing, Portland, OR" → "Blumenauer Bridge"
func parseLandmark(tokens []string) string {
	var result []string

	for _, token := range tokens {
		// Stop at comma (next landmark or descriptor)
		if token == "," {
			break
		}

		// Stop at city/state
		if token == "portland" || token == "vancouver" || token == "or" || token == "wa" {
			break
		}

		result = append(result, token)
	}

	return strings.Join(result, " ")
}

// NormalizeAddress intelligently parses and normalizes an address using a tiered approach
// Returns the normalized core address (without city/state)
func NormalizeAddress(address string) string {
	if address == "" {
		return ""
	}

	tokens := tokenizeAddress(address)
	if len(tokens) == 0 {
		return ""
	}

	// Tier 1: Intersection detection (strongest signal)
	if hasIntersectionMarker(tokens) {
		normalized := parseIntersection(tokens)
		if normalized != "" {
			return normalized
		}
	}

	// Tier 2: Full address detection (has street numbers)
	if hasStreetNumber(tokens) {
		normalized := parseFullAddress(tokens)
		if normalized != "" {
			return normalized
		}
	}

	// Tier 3: Landmark-based (fallback)
	normalized := parseLandmark(tokens)
	if normalized != "" {
		return normalized
	}

	// Fallback: return original if nothing matched
	return address
}
