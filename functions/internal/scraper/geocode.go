package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/transport"
)

const addressScope = "https://www.googleapis.com/auth/maps-platform.geocode.address"

var authenticatedClient *http.Client

func getAuthenticatedClient(ctx context.Context) (*http.Client, error) {
	if authenticatedClient != nil {
		return authenticatedClient, nil
	}

	// Use Application Default Credentials (ADC) from the service account running on Cloud Run
	// This will automatically use the credentials of the Cloud Run service account
	clientOption := option.WithScopes(addressScope)

	// Create HTTP client with ADC
	httpClient, _, err := transport.NewHTTPClient(ctx, clientOption)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticated HTTP client with ADC: %w", err)
	}

	httpClient.Timeout = 15 * time.Second
	authenticatedClient = httpClient

	return authenticatedClient, nil

}

type CityDetails struct {
	CityName string
	State    string
	NELat    float64
	NELng    float64
	SWLat    float64
	SWLng    float64
}

var cityMap = map[string]CityDetails{
	"pdx": {CityName: "Portland", State: "OR", SWLat: 45.4325, SWLng: -122.8367, NELat: 46.00, NELng: -121.5},
	"slc": {CityName: "Salt Lake City", State: "UT", SWLat: 40.6307, SWLng: -112.1, NELat: 41.0, NELng: -111.5},
}

func GeocodeQuery(query, cityCode string) (float64, float64, error) {
	ctx := context.Background()
	client, err := getAuthenticatedClient(ctx)

	baseURL := "https://geocode.googleapis.com/v4beta/geocode/address/"

	cityDetails := cityMap[cityCode]

	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return 0.0, 0.0, err
	}

	q := req.URL.Query()
	q.Add("regionCode", "US")

	q.Add("locationBias.rectangle.low.latitude", strconv.FormatFloat(cityDetails.SWLat, 'f', -1, 64))
	q.Add("locationBias.rectangle.low.longitude", strconv.FormatFloat(cityDetails.SWLng, 'f', -1, 64))
	q.Add("locationBias.rectangle.high.latitude", strconv.FormatFloat(cityDetails.NELat, 'f', -1, 64))
	q.Add("locationBias.rectangle.high.longitude", strconv.FormatFloat(cityDetails.NELng, 'f', -1, 64))
	q.Add("address.addressLines", query)
	q.Add("address.administrativeArea", cityDetails.State)
	q.Add("address.locality", cityDetails.CityName)

	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		return 0.0, 0.0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		// bodyBtyes, _ := io.ReadAll(res.Body)
		// fmt.Println(string(bodyBtyes))
		return 0.0, 0.0, fmt.Errorf("Google Geocoding API returned non-OK status code %d", res.StatusCode)
	}

	var googleResponse GoogleGeocodeResponse
	if err := json.NewDecoder(res.Body).Decode(&googleResponse); err != nil {
		return 0.0, 0.0, fmt.Errorf("failed to decode Google geocoding response: %v", err)
	}

	if len(googleResponse.Results) == 0 {
		return 0.0, 0.0, fmt.Errorf("no results found for address: '%s' (GOOGLE API 'OK' status but empty results", query)
	}

	lat := googleResponse.Results[0].Location.Latitude
	lng := googleResponse.Results[0].Location.Longitude

	fmt.Println(lat, lng)

	return lat, lng, nil
}
