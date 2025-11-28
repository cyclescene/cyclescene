# CycleScene Route Processing & Preview Feature Spec

## Overview

Add route visualization and elevation profiles to CycleScene rides. Support both user-submitted rides and scraped Shift2Bikes events with routes from Strava and RideWithGPS.

## Database Schema

### Routes Table

```sql
CREATE TABLE routes (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 source TEXT NOT NULL, -- 'strava' or 'ridewithgps'
 source_id TEXT NOT NULL,
 source_url TEXT NOT NULL,
 geojson JSONB NOT NULL,
 created_at TIMESTAMP DEFAULT NOW(),
 UNIQUE(source, source_id)
);
```

### Rides Table (updated)

```sql
CREATE TABLE rides (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 route_id UUID REFERENCES routes(id),
 title TEXT NOT NULL,
 description TEXT,
 source TEXT, -- 'shift2bikes', 'user_submission', etc
 source_id TEXT,
 created_by UUID REFERENCES users(id),
 created_at TIMESTAMP DEFAULT NOW()
);
```

## Backend Implementation

### 1. Route Processing Service

**File: `internal/routes/converter.go`**

Parse GPX and convert to GeoJSON with elevation data.

```go
type GeoJSONFeature struct {
 Type       string                 `json:"type"`
 Geometry   GeoJSONGeometry        `json:"geometry"`
 Properties map[string]interface{} `json:"properties"`
}

type GeoJSONGeometry struct {
 Type        string      `json:"type"`
 Coordinates [][]float64 `json:"coordinates"` // [lon, lat, elevation]
}

// ConvertGPXtoGeoJSON parses GPX and returns GeoJSON Feature
func ConvertGPXtoGeoJSON(gpxData io.Reader) (GeoJSONFeature, error)

// ConvertPolylineToGeoJSON converts Strava polyline to GeoJSON
func ConvertPolylineToGeoJSON(polyline string, name string) (GeoJSONFeature, error)
```

**Supported Sources:**

- RideWithGPS: Parse GPX (fetch from `https://ridewithgps.com/routes/{id}.gpx`)
- Strava: Decode polyline from activity data

### 2. Route Fetcher

**File: `internal/routes/fetcher.go`**

Fetch routes from external sources and convert.

```go
type RouteFetcher struct {
 httpClient *http.Client
 stravaClient *StravaClient
}

// FetchAndConvert determines source type and fetches/converts route
func (f *RouteFetcher) FetchAndConvert(url string) (GeoJSONFeature, error)

// parseRouteURL extracts source type and ID from URL
func parseRouteURL(url string) (source, sourceID string, err error)
```

**URL Parsing Rules:**

- RideWithGPS: `https://ridewithgps.com/routes/{id}` → extract `id`
- Strava: `https://www.strava.com/activities/{id}` → extract `id`

### 3. Route Repository

**File: `internal/routes/repository.go`**

Database operations for routes.

```go
type Repository struct {
 db *sql.DB
}

// CreateRoute stores new route, handles UNIQUE constraint
func (r *Repository) CreateRoute(ctx context.Context, route Route) (string, error)

// GetRouteBySourceID retrieves existing route
func (r *Repository) GetRouteBySourceID(ctx context.Context, source, sourceID string) (*Route, error)

// GetAllRoutes returns all routes (for scraper cache initialization)
func (r *Repository) GetAllRoutes(ctx context.Context) ([]Route, error)

// GetRouteByID retrieves single route by ID
func (r *Repository) GetRouteByID(ctx context.Context, id string) (*Route, error)
```

### 4. Ride Submission Handler

**File: `internal/rides/handler.go`**

Handle user-submitted rides with route links.

```go
type RideSubmission struct {
 Title       string `json:"title"`
 Description string `json:"description"`
 RouteURL    string `json:"route_url"`
 CreatedBy   string `json:"created_by"`
}

// SubmitRide processes route and creates ride record
func (h *RideHandler) SubmitRide(w http.ResponseWriter, r *http.Request) {
 // 1. Parse request
 // 2. Process route (fetch, convert, deduplicate)
 // 3. Create ride with route_id
 // 4. Return ride with embedded route GeoJSON
}

// processRoute fetches, converts, and deduplicates route
func (h *RideHandler) processRoute(ctx context.Context, url string) (string, error)
```

### 5. Scraper Integration

**File: `internal/scraper/shift2bikes.go`**

Integrate route processing into existing Shift2Bikes scraper job.

```go
type ScrapeJob struct {
 shift2bikesClient *Shift2BikesClient
 routeFetcher      *RouteFetcher
 routeRepo         *Repository
 rideRepo          *RideRepository
 routeCache        map[string]string // "source:sourceID" -> routeID
}

// Run executes scrape job with route caching
func (j *ScrapeJob) Run(ctx context.Context) error {
 // 1. Load existing routes into cache
 // 2. Fetch Shift2Bikes events
 // 3. For each event:
 //    - Extract route URL from description (regex)
 //    - Check cache, if exists use cached routeID
 //    - If new, fetch and convert, cache result
 //    - Create ride with route_id
}

// extractRouteURLFromDescription finds Strava/RideWithGPS links
func extractRouteURLFromDescription(desc string) string
```

**Route URL Extraction:**

- Strava Pattern: `https://www.strava.com/api/v3/routes/{rideId}/export_gpx`
  - Requires Athorization header
  - env var `STRAVA_ACCESS_TOKEN`
- RideWithGPS Pattern: `https://ridewithgps.com/routes/{id}.gpx?sub_format=track`
  - env var `RWGPS_AUTH_TOKEN`
  - Requires x-rwgps-auth-token header

- Return first match or empty string

## Frontend Implementation

### 1. Ride Detail Component

**File: `src/routes/rides/[id]/+page.svelte`**

Display route map and elevation graph.

```svelte
<script>
 import Map from '$lib/components/Map.svelte';
 import ElevationGraph from '$lib/components/ElevationGraph.svelte';

 export let data;

 $: ride = data.ride;
 $: route = ride.route; // GeoJSON Feature
 $: elevationData = route?.geometry.coordinates.map((coord, i) => ({
  distance: i * 0.01, // approximate distance in km
  elevation: coord[2]
 })) ?? [];
</script>

<div class="ride-detail">
 <h1>{ride.title}</h1>

 {#if route}
  <div class="route-preview">
   <Map geojson={route} center={calculateCenter(route)} zoom={12} />
  </div>

  <div class="elevation-section">
   <h3>Elevation Profile</h3>
   <ElevationGraph data={elevationData} />
  </div>
 {/if}

 <p>{ride.description}</p>
</div>
```

### 2. Map Component

**File: `src/lib/components/Map.svelte`**

Display route with MapLibre, color by elevation.

```svelte
<script>
 import { Map, NavigationControl } from 'maplibre-gl';

 export let geojson;
 export let center = [0, 0];
 export let zoom = 12;

 let container;
 let map;

 onMount(() => {
  map = new Map({
   container,
   style: 'https://basemaps.cartocdn.com/gl/positron-gl-style/style.json',
   center,
   zoom
  });

  map.on('load', () => {
   // Add route source
   map.addSource('route', {
    type: 'geojson',
    data: geojson
   });

   // Add elevation-colored line layer
   map.addLayer({
    id: 'route-line',
    type: 'line',
    source: 'route',
    paint: {
     'line-color': [
      'interpolate',
      ['linear'],
      ['get', 'elevation'],
      27, '#00ff00',   // low = green
      50, '#ffff00',   // mid = yellow
      100, '#ff0000'   // high = red
     ],
     'line-width': 3
    }
   });
  });
 });
</script>

<div bind:this={container} class="map" />

<style>
 .map {
  width: 100%;
  height: 400px;
 }
</style>
```

### 3. Elevation Graph Component

**File: `src/lib/components/ElevationGraph.svelte`**

Display elevation profile using recharts.

```svelte
<script>
 import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

 export let data = [];
</script>

<ResponsiveContainer width="100%" height={300}>
 <LineChart data={data}>
  <CartesianGrid strokeDasharray="3 3" />
  <XAxis dataKey="distance" label={{ value: 'Distance (km)', position: 'insideBottom', offset: -5 }} />
  <YAxis label={{ value: 'Elevation (m)', angle: -90, position: 'insideLeft' }} />
  <Tooltip />
  <Line type="monotone" dataKey="elevation" stroke="#8884d8" dot={false} />
 </LineChart>
</ResponsiveContainer>
```

### 4. Ride Submission Form

Add input for form link in existing ride submission form.

## Implementation Plan

### Phase 1: Core Route Processing (Weekend)

- [x] Design database schema
- [ ] Implement `ConvertGPXtoGeoJSON`
- [ ] Implement `ConvertPolylineToGeoJSON` (or stub for now)
- [ ] Implement `RouteFetcher` for RideWithGPS and Strava
- [ ] Implement `RouteRepository` with deduplication
- [ ] Integrate into ride submission handler

### Phase 2: Scraper Integration

- [ ] Add route cache to `ScrapeJob`
- [ ] Implement `extractRouteURLFromDescription`
- [ ] Load existing routes at job start
- [ ] Process routes during scrape with cache hits

### Phase 3: Frontend Visualization

- [ ] Build Map component with MapLibre
- [ ] Build ElevationGraph component
- [ ] Integrate into ride detail page
- [ ] Style and test responsiveness

## Testing

### Unit Tests

- `ConvertGPXtoGeoJSON`: Valid GPX → valid GeoJSON
- `parseRouteURL`: Extract IDs from URLs correctly
- `extractRouteURLFromDescription`: Find route links in text

### Integration Tests

- Submit ride with route URL → route processed and deduplicated
- Scrape job with duplicate routes → cache prevents reprocessing
- Retrieve ride → includes route GeoJSON

### Manual Testing

- Submit ride with RideWithGPS route
- Submit ride with Strava activity
- Verify route displays on detail page
- Verify elevation graph renders

## Error Handling

- Invalid route URL → return 400 with message
- Route fetch fails → log error, continue without route
- GPX parse error → return 400 with message
- Strava API error → log error, continue without route
- Database constraint violation → use existing route_id

## Notes

- Routes are optional for rides
- Scraper continues even if route processing fails
- GeoJSON coordinates include elevation as third value
- MapLibre color interpolation: green (low) → yellow (mid) → red (high)
- Elevation graph x-axis is approximate distance (point_index \* 0.01)
