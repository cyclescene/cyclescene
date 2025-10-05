package scraperhelpers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2/google"
)

const saCredentialsKey = "GOOGLE_SA_CREDENTIALS"
const addressScope = "https://www.googleapis.com/auth/maps-platform.geocode.address"
const (
	// Portland Bounding Box
	// SW Corner
	PDX_SW_LAT = 45.0
	PDX_SW_LNG = -123.0
	// NE Corner
	PDX_NE_LAT = 46.0
	PDX_NE_LNG = -121.5
)

var authenticatedClient *http.Client

func getAuthenticatedClient(ctx context.Context) (*http.Client, error) {
	if authenticatedClient != nil {
		return authenticatedClient, nil
	}

	credsB64 := os.Getenv(saCredentialsKey)
	if credsB64 == "" {
		return nil, fmt.Errorf("FATAL: Google Service Account credentials not found is %s", saCredentialsKey)
	}

	credsJSON, err := base64.StdEncoding.DecodeString(credsB64)
	if err != nil {
		return nil, fmt.Errorf("FATAL: failed to decode Base64 credentials %w", err)
	}
	creds, err := google.JWTConfigFromJSON([]byte(credsJSON), addressScope)
	if err != nil {

		return nil, fmt.Errorf("failed to create credentials from JSON: %w", err)
	}

	client := creds.Client(ctx)
	client.Timeout = 15 * time.Second
	authenticatedClient = client

	return authenticatedClient, nil

}

func GeocodeQuery(query string) (float64, float64, error) {
	ctx := context.Background()
	client, err := getAuthenticatedClient(ctx)

	baseURL := "https://geocode.googleapis.com/v4beta/geocode/address/"

	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, baseURL+query, nil)
	if err != nil {
		return 0.0, 0.0, err
	}

	q := req.URL.Query()
	q.Add("regionCode", "US")

	q.Add("locationBias.rectangle.low.latitude", strconv.FormatFloat(PDX_SW_LAT, 'f', -1, 64))
	q.Add("locationBias.rectangle.low.longitude", strconv.FormatFloat(PDX_SW_LNG, 'f', -1, 64))
	q.Add("locationBias.rectangle.high.latitude", strconv.FormatFloat(PDX_NE_LAT, 'f', -1, 64))
	q.Add("locationBias.rectangle.high.longitude", strconv.FormatFloat(PDX_NE_LNG, 'f', -1, 64))

	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		return 0.0, 0.0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0.0, 0.0, fmt.Errorf("Google Geocoding API returned non-OK status code %d", res.StatusCode)
	}

	var googleResponse GoogleGeocodeResponse
	if err := json.NewDecoder(res.Body).Decode(&googleResponse); err != nil {
		return 0.0, 0.0, fmt.Errorf("failed to decode Google geocoding response: %v", err)
	}

	// Using v4 the response no longer provides a status
	// if googleResponse.Status != "OK" {
	// 	if googleResponse.Status == "ZERO_RESULTS" {
	// 		fmt.Println(string(bodyBytes))
	// 		return 0.0, 0.0, fmt.Errorf("Google Geocoding API found no results for address: '%s'", query)
	// 	}
	// 	fmt.Println(string(bodyBytes))
	// 	return 0.0, 0.0, fmt.Errorf("Google Geocoding API returned status: %s for (for address: '%s')", googleResponse.Status, query)
	// }

	if len(googleResponse.Results) == 0 {
		return 0.0, 0.0, fmt.Errorf("no results found for address: '%s' (GOOGLE API 'OK' status but empty results", query)
	}

	lat := googleResponse.Results[0].Location.Latitude
	lng := googleResponse.Results[0].Location.Longitude

	return lat, lng, nil
}
