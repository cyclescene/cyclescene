package routes

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
)

// GeoJSONFeature represents a GeoJSON Feature with optional elevation data
type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// GeoJSONGeometry represents GeoJSON geometry with coordinates
// Coordinates are [lon, lat, elevation]
type GeoJSONGeometry struct {
	Type        string        `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

// GPXTrack represents the structure of a GPX file (simplified)
type GPXTrack struct {
	XMLName xml.Name `xml:"gpx"`
	Tracks  []struct {
		Segments []struct {
			Points []struct {
				Lat float64 `xml:"lat,attr"`
				Lon float64 `xml:"lon,attr"`
				Ele float64 `xml:"ele"`
			} `xml:"trkpt"`
		} `xml:"trkseg"`
	} `xml:"trk"`
	Routes []struct {
		Points []struct {
			Lat float64 `xml:"lat,attr"`
			Lon float64 `xml:"lon,attr"`
			Ele float64 `xml:"ele"`
		} `xml:"rtept"`
	} `xml:"rte"`
}

// ConvertGPXtoGeoJSON parses GPX data and returns a GeoJSON Feature with elevation data
func ConvertGPXtoGeoJSON(gpxData io.Reader) (GeoJSONFeature, error) {
	var gpx GPXTrack
	if err := xml.NewDecoder(gpxData).Decode(&gpx); err != nil {
		return GeoJSONFeature{}, fmt.Errorf("failed to parse GPX: %w", err)
	}

	// Collect all coordinates (tracks take precedence over routes)
	var coords [][][]float64

	// Process tracks first
	for _, track := range gpx.Tracks {
		for _, segment := range track.Segments {
			var segmentCoords [][]float64
			for _, point := range segment.Points {
				segmentCoords = append(segmentCoords, []float64{point.Lon, point.Lat, point.Ele})
			}
			coords = append(coords, segmentCoords)
		}
	}

	// If no tracks, process routes
	if len(coords) == 0 {
		for _, route := range gpx.Routes {
			var routeCoords [][]float64
			for _, point := range route.Points {
				routeCoords = append(routeCoords, []float64{point.Lon, point.Lat, point.Ele})
			}
			if len(routeCoords) > 0 {
				coords = append(coords, routeCoords)
			}
		}
	}

	if len(coords) == 0 {
		return GeoJSONFeature{}, fmt.Errorf("no valid coordinates found in GPX")
	}

	// Flatten coordinates if multiple segments
	var allCoords [][]float64
	for _, segment := range coords {
		allCoords = append(allCoords, segment...)
	}

	// Calculate distance
	distanceKm := calculateDistance(allCoords)
	distanceMi := distanceKm * 0.621371

	feature := GeoJSONFeature{
		Type: "Feature",
		Geometry: GeoJSONGeometry{
			Type:        "LineString",
			Coordinates: allCoords,
		},
		Properties: map[string]interface{}{
			"distance_km": distanceKm,
			"distance_mi": distanceMi,
		},
	}

	return feature, nil
}

// ConvertPolylineToGeoJSON converts a Strava polyline to GeoJSON
// Strava polylines are encoded using the Google polyline algorithm
func ConvertPolylineToGeoJSON(polyline string, name string) (GeoJSONFeature, error) {
	coords, err := decodePolyline(polyline)
	if err != nil {
		return GeoJSONFeature{}, fmt.Errorf("failed to decode polyline: %w", err)
	}

	if len(coords) == 0 {
		return GeoJSONFeature{}, fmt.Errorf("polyline contains no coordinates")
	}

	// Calculate distance
	distanceKm := calculateDistance(coords)
	distanceMi := distanceKm * 0.621371

	feature := GeoJSONFeature{
		Type: "Feature",
		Geometry: GeoJSONGeometry{
			Type:        "LineString",
			Coordinates: coords,
		},
		Properties: map[string]interface{}{
			"name":         name,
			"distance_km": distanceKm,
			"distance_mi": distanceMi,
		},
	}

	return feature, nil
}

// decodePolyline decodes a Google polyline-encoded string
// https://developers.google.com/maps/documentation/utilities/polylinealgorithm
func decodePolyline(polyline string) ([][]float64, error) {
	var coords [][]float64
	var lat, lng int32
	var i int

	for i < len(polyline) {
		var result int32
		var shift uint32
		var b int32

		// Decode latitude
		for {
			if i >= len(polyline) {
				return nil, fmt.Errorf("incomplete latitude in polyline")
			}
			b = int32(polyline[i]) - 63
			i++
			result |= (b & 0x1f) << shift
			shift += 5
			if b < 0x20 {
				break
			}
		}

		dlat := result
		if result&1 != 0 {
			dlat = ^(result >> 1)
		} else {
			dlat = result >> 1
		}
		lat += dlat

		// Decode longitude
		result = 0
		shift = 0
		for {
			if i >= len(polyline) {
				return nil, fmt.Errorf("incomplete longitude in polyline")
			}
			b = int32(polyline[i]) - 63
			i++
			result |= (b & 0x1f) << shift
			shift += 5
			if b < 0x20 {
				break
			}
		}

		dlng := result
		if result&1 != 0 {
			dlng = ^(result >> 1)
		} else {
			dlng = result >> 1
		}
		lng += dlng

		coords = append(coords, []float64{float64(lng) / 1e5, float64(lat) / 1e5, 0})
	}

	return coords, nil
}

// calculateDistance calculates the total distance in kilometers using Haversine formula
func calculateDistance(coords [][]float64) float64 {
	if len(coords) < 2 {
		return 0
	}

	var totalDistance float64
	const earthRadiusKm = 6371.0

	for i := 0; i < len(coords)-1; i++ {
		lat1 := coords[i][1] * math.Pi / 180
		lon1 := coords[i][0] * math.Pi / 180
		lat2 := coords[i+1][1] * math.Pi / 180
		lon2 := coords[i+1][0] * math.Pi / 180

		dlat := lat2 - lat1
		dlon := lon2 - lon1

		a := math.Sin(dlat/2)*math.Sin(dlat/2) +
			math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
		c := 2 * math.Asin(math.Sqrt(a))
		totalDistance += earthRadiusKm * c
	}

	return totalDistance
}
