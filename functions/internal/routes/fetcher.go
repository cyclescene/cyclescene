package routes

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

// RouteFetcher handles fetching routes from external sources
type RouteFetcher struct {
	httpClient        *http.Client
	stravaAccessToken string
	rwgpsAuthToken    string
	rwgpsAPIKey       string
}

// NewRouteFetcher creates a new RouteFetcher instance
func NewRouteFetcher(httpClient *http.Client, stravaToken, rwgpsAuthToken, rwgpsAPIKey string) *RouteFetcher {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &RouteFetcher{
		httpClient:        httpClient,
		stravaAccessToken: stravaToken,
		rwgpsAuthToken:    rwgpsAuthToken,
		rwgpsAPIKey:       rwgpsAPIKey,
	}
}

// FetchAndConvert determines the route source from URL and fetches/converts the route
func (f *RouteFetcher) FetchAndConvert(url string) (GeoJSONFeature, error) {
	source, sourceID, err := ParseRouteURL(url)
	if err != nil {
		return GeoJSONFeature{}, err
	}

	switch source {
	case "ridewithgps":
		// For RideWithGPS, pass the base URL without the .gpx extension
		baseURL := fmt.Sprintf("https://ridewithgps.com/routes/%s", sourceID)
		return f.fetchRideWithGPSRoute(baseURL)
	case "strava":
		return f.fetchStravaRoute(sourceID)
	default:
		return GeoJSONFeature{}, fmt.Errorf("unsupported route source: %s", source)
	}
}

// fetchRideWithGPSRoute fetches and converts a RideWithGPS route
func (f *RouteFetcher) fetchRideWithGPSRoute(routeURL string) (GeoJSONFeature, error) {
	// Construct GPX URL - append .gpx?sub_format=track if not already present
	gpxURL := routeURL
	if !contains(gpxURL, ".gpx") {
		gpxURL = fmt.Sprintf("%s.gpx?sub_format=track", routeURL)
	} else if !contains(gpxURL, "sub_format") {
		gpxURL = fmt.Sprintf("%s?sub_format=track", routeURL)
	}

	req, err := http.NewRequest("GET", gpxURL, nil)
	if err != nil {
		return GeoJSONFeature{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Add User-Agent to avoid being redirected to signup
	req.Header.Set("User-Agent", "CycleScene/1.0 (+https://cyclescene.cc)")

	// Add RideWithGPS authentication headers
	if f.rwgpsAPIKey != "" {
		req.Header.Set("x-rwgps-api-key", f.rwgpsAPIKey)
	}
	if f.rwgpsAuthToken != "" {
		req.Header.Set("x-rwgps-auth-token", f.rwgpsAuthToken)
	}

	// Create a client that doesn't follow redirects past a certain point
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return GeoJSONFeature{}, fmt.Errorf("failed to fetch RideWithGPS route: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GeoJSONFeature{}, fmt.Errorf("RideWithGPS returned status %d", resp.StatusCode)
	}

	return ConvertGPXtoGeoJSON(resp.Body)
}

// fetchStravaRoute fetches and converts a Strava route
func (f *RouteFetcher) fetchStravaRoute(routeID string) (GeoJSONFeature, error) {
	gpxURL := fmt.Sprintf("https://www.strava.com/api/v3/routes/%s/export_gpx", routeID)

	req, err := http.NewRequest("GET", gpxURL, nil)
	if err != nil {
		return GeoJSONFeature{}, fmt.Errorf("failed to create request: %w", err)
	}

	if f.stravaAccessToken == "" {
		return GeoJSONFeature{}, fmt.Errorf("Strava access token not configured")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.stravaAccessToken))

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return GeoJSONFeature{}, fmt.Errorf("failed to fetch Strava route: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return GeoJSONFeature{}, fmt.Errorf("Strava returned status %d: %s", resp.StatusCode, string(body))
	}

	return ConvertGPXtoGeoJSON(resp.Body)
}

// ParseRouteURL extracts the source type and ID from a route URL
func ParseRouteURL(url string) (source, sourceID string, err error) {
	// RideWithGPS pattern: https://ridewithgps.com/routes/{id}
	rwgpsPattern := regexp.MustCompile(`ridewithgps\.com/routes/(\d+)`)
	if matches := rwgpsPattern.FindStringSubmatch(url); matches != nil {
		return "ridewithgps", matches[1], nil
	}

	// Strava route pattern: https://www.strava.com/api/v3/routes/{id}/export_gpx
	stravaRoutePattern := regexp.MustCompile(`strava\.com/api/v3/routes/(\d+)/export_gpx`)
	if matches := stravaRoutePattern.FindStringSubmatch(url); matches != nil {
		return "strava", matches[1], nil
	}

	return "", "", fmt.Errorf("unable to parse route URL: %s", url)
}

// ExtractRouteURLFromDescription finds Strava or RideWithGPS links in text
func ExtractRouteURLFromDescription(description string) string {
	// Strava route export pattern
	stravaRoutePattern := regexp.MustCompile(`https://(?:www\.)?strava\.com/api/v3/routes/\d+/export_gpx`)
	if match := stravaRoutePattern.FindString(description); match != "" {
		return match
	}

	// RideWithGPS pattern
	rwgpsPattern := regexp.MustCompile(`https://ridewithgps\.com/routes/\d+`)
	if match := rwgpsPattern.FindString(description); match != "" {
		return match
	}

	// RideWithGPS GPX pattern
	rwgpsGPXPattern := regexp.MustCompile(`https://ridewithgps\.com/routes/\d+\.gpx`)
	if match := rwgpsGPXPattern.FindString(description); match != "" {
		return match
	}

	return ""
}

// contains is a simple helper to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
