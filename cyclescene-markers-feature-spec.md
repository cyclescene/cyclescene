# CycleScene Ride Group Markers & Spritesheet Feature Spec

## Overview

Display custom ride group markers on the map using dynamically generated spritesheets. Each city has its own spritesheet that grows as new ride groups are added. Markers are extracted from the spritesheet at app load and MapLibre handles rendering/visibility.

## Database Schema

### Ride Groups Table

```sql
CREATE TABLE ride_groups (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 public_id TEXT UNIQUE NOT NULL, -- slug/public identifier
 name TEXT NOT NULL,
 city TEXT NOT NULL,
 marker TEXT NOT NULL, -- sprite metadata key (matches public_id)
 created_at TIMESTAMP DEFAULT NOW()
);
```

### Rides Table (updated)

```sql
CREATE TABLE rides (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 route_id UUID REFERENCES routes(id),
 group_id UUID REFERENCES ride_groups(id),
 title TEXT NOT NULL,
 description TEXT,
 source TEXT, -- 'shift2bikes', 'user_submission', etc
 source_id TEXT,
 created_by UUID REFERENCES users(id),
 created_at TIMESTAMP DEFAULT NOW()
);
```

## GCP Bucket Structure

```
your-bucket/
  sprites/
    portland/
      markers.png       (dynamic spritesheet)
      markers.json      (metadata)
    salt-lake-city/
      markers.png
      markers.json
```

## Backend Implementation

### 1. Image Optimization Service Integration

**File: `internal/services/image-optimizer.go`**

Add handler for `type: "groups"` EventArc messages:

```go
type OptimizationPayload struct {
 Bucket  string `json:"bucket"`
 File    string `json:"file"`
 Type    string `json:"type"` // "ride" or "groups"
 City    string `json:"city"`
 GroupID string `json:"groupID"`
}

func HandleImageOptimization(ctx context.Context, e PubSubMessage) error {
 payload := parsePayload(e.Data)

 switch payload.Type {
 case "ride":
  // existing ride image processing
  return handleRideImage(ctx, payload)

 case "groups":
  // Download marker from staging
  img, err := downloadFromBucket(ctx, payload.Bucket, payload.File)
  if err != nil {
   return err
  }

  // Process to 40x40 PNG
  processed, err := uploads.ProcessMarkerImage(payload.City, payload.GroupID, img)
  if err != nil {
   return err
  }

  // Save locally for spritesheet generation
  markerPath := fmt.Sprintf("public/markers/%s/%s.png", payload.City, payload.GroupID)
  if err := os.MkdirAll(filepath.Dir(markerPath), 0755); err != nil {
   return err
  }
  if err := os.WriteFile(markerPath, processed, 0644); err != nil {
   return err
  }

  // Regenerate spritesheet for city
  if err := regenerateSpritesheet(ctx, payload.City); err != nil {
   return fmt.Errorf("failed to regenerate spritesheet: %w", err)
  }

  // Upload new spritesheet to GCP
  if err := uploadSpritesheetToGCP(ctx, payload.City); err != nil {
   return fmt.Errorf("failed to upload spritesheet: %w", err)
  }

  // Delete staging file
  if err := deleteFromBucket(ctx, payload.Bucket, payload.File); err != nil {
   log.Printf("Warning: Failed to delete staging file: %v", err)
  }

  return nil
 }

 return nil
}

func regenerateSpritesheet(ctx context.Context, city string) error {
 cmd := exec.CommandContext(ctx, "go", "run", "./cmd/build-markers/main.go",
  "--city", city)
 return cmd.Run()
}

func uploadSpritesheetToGCP(ctx context.Context, city string) error {
 uploader, err := uploads.NewGCPUploader(ctx, os.Getenv("GCP_BUCKET"))
 if err != nil {
  return err
 }

 pngPath := fmt.Sprintf("public/sprites/%s/markers.png", city)
 jsonPath := fmt.Sprintf("public/sprites/%s/markers.json", city)
 remotePath := fmt.Sprintf("sprites/%s", city)

 return uploader.UploadSpritesheet(ctx, pngPath, jsonPath, remotePath)
}
```

### 2. Spritesheet Generation Tool

**File: `cmd/build-markers/main.go`**

Generate city-specific spritesheets from marker images:

```go
package main

import (
 "encoding/json"
 "flag"
 "fmt"
 "image"
 "image/draw"
 "image/png"
 "log"
 "math"
 "os"
 "path/filepath"
 "sort"

 "github.com/disintegration/imaging"
)

const (
 markerSize = 40
 padding    = 2
)

var city = flag.String("city", "pdx", "City for markers")

type MarkerInfo struct {
 X      int `json:"x"`
 Y      int `json:"y"`
 Width  int `json:"width"`
 Height int `json:"height"`
}

func main() {
 flag.Parse()

 markersDir := fmt.Sprintf("public/markers/%s", *city)
 outputDir := fmt.Sprintf("public/sprites/%s", *city)

 // Read all marker files
 entries, err := os.ReadDir(markersDir)
 if err != nil {
  if os.IsNotExist(err) {
   log.Printf("No markers directory for %s, skipping", *city)
   return
  }
  log.Fatalf("Failed to read markers directory: %v", err)
 }

 var markerFiles []os.DirEntry
 for _, entry := range entries {
  if !entry.IsDir() && isValidImage(entry.Name()) {
   markerFiles = append(markerFiles, entry)
  }
 }

 if len(markerFiles) == 0 {
  log.Printf("No marker images found for %s", *city)
  return
 }

 // Sort for consistent ordering
 sort.Slice(markerFiles, func(i, j int) bool {
  return markerFiles[i].Name() < markerFiles[j].Name()
 })

 // Calculate spritesheet dimensions
 cols := int(math.Ceil(math.Sqrt(float64(len(markerFiles)))))
 rows := (len(markerFiles) + cols - 1) / cols
 sheetWidth := cols * (markerSize + padding)
 sheetHeight := rows * (markerSize + padding)

 // Create blank spritesheet with transparency
 spritesheet := image.NewRGBA(image.Rect(0, 0, sheetWidth, sheetHeight))

 // Load and composite each marker
 metadata := make(map[string]MarkerInfo)
 for i, entry := range markerFiles {
  // Extract marker key from filename (without extension)
  markerKey := entry.Name()[:len(entry.Name())-4]

  markerPath := filepath.Join(markersDir, entry.Name())
  img, err := loadAndResizeImage(markerPath, markerSize)
  if err != nil {
   log.Printf("Warning: Failed to process %s: %v", entry.Name(), err)
   continue
  }

  col := i % cols
  row := i / cols
  x := col * (markerSize + padding)
  y := row * (markerSize + padding)

  // Draw image onto spritesheet
  draw.Draw(spritesheet, img.Bounds().Add(image.Pt(x, y)), img, image.Pt(0, 0), draw.Over)

  metadata[markerKey] = MarkerInfo{
   X:      x,
   Y:      y,
   Width:  markerSize,
   Height: markerSize,
  }
 }

 // Create output directory
 if err := os.MkdirAll(outputDir, 0755); err != nil {
  log.Fatalf("Failed to create output directory: %v", err)
 }

 // Write spritesheet PNG
 spritePath := filepath.Join(outputDir, "markers.png")
 spriteFd, err := os.Create(spritePath)
 if err != nil {
  log.Fatalf("Failed to create spritesheet file: %v", err)
 }
 defer spriteFd.Close()

 if err := png.Encode(spriteFd, spritesheet); err != nil {
  log.Fatalf("Failed to encode spritesheet: %v", err)
 }

 // Write metadata JSON
 metadataPath := filepath.Join(outputDir, "markers.json")
 metadataFd, err := os.Create(metadataPath)
 if err != nil {
  log.Fatalf("Failed to create metadata file: %v", err)
 }
 defer metadataFd.Close()

 metadataWrapper := map[string]map[string]MarkerInfo{
  "markers": metadata,
 }

 encoder := json.NewEncoder(metadataFd)
 encoder.SetIndent("", "  ")
 if err := encoder.Encode(metadataWrapper); err != nil {
  log.Fatalf("Failed to encode metadata: %v", err)
 }

 log.Printf("✓ Generated spritesheet for %s with %d markers", *city, len(metadata))
 log.Printf("  Sheet size: %dx%dpx", sheetWidth, sheetHeight)
 log.Printf("  Output: %s", outputDir)
}

func isValidImage(filename string) bool {
 ext := filepath.Ext(filename)
 validExts := map[string]bool{
  ".png":  true,
  ".jpg":  true,
  ".jpeg": true,
 }
 return validExts[ext]
}

func loadAndResizeImage(path string, size int) (image.Image, error) {
 file, err := os.Open(path)
 if err != nil {
  return nil, err
 }
 defer file.Close()

 img, _, err := image.Decode(file)
 if err != nil {
  return nil, err
 }

 resized := imaging.Fit(img, size, size, imaging.Lanczos)
 return resized, nil
}
```

### 3. Marker Upload Handler

**File: `internal/rides/handler.go`**

Handle ride group creation with marker upload:

```go
type CreateGroupRequest struct {
 Name   string `json:"name"`
 City   string `json:"city"`
 Marker []byte `json:"marker"` // multipart file
}

func (h *RideGroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
 // Parse multipart form
 r.ParseMultipartForm(5 << 20) // 5MB limit

 name := r.FormValue("name")
 city := r.FormValue("city")

 // Validate city
 if !isValidCity(city) {
  http.Error(w, "Invalid city", http.StatusBadRequest)
  return
 }

 // Get marker file
 markerFile, header, err := r.FormFile("marker")
 if err != nil {
  http.Error(w, "No marker image provided", http.StatusBadRequest)
  return
 }
 defer markerFile.Close()

 // Validate file size
 if header.Size > 5*1024*1024 {
  http.Error(w, "Marker too large (max 5MB)", http.StatusBadRequest)
  return
 }

 // Validate image type
 contentType := header.Header.Get("Content-Type")
 if !isValidImageType(contentType) {
  http.Error(w, "Invalid image type", http.StatusBadRequest)
  return
 }

 // Generate public ID (slug from name)
 publicID := slugify(name)

 // Create group in database
 group := RideGroup{
  ID:       generateID(),
  PublicID: publicID,
  Name:     name,
  City:     city,
  Marker:   publicID, // sprite metadata key
 }

 if err := h.db.CreateRideGroup(group); err != nil {
  http.Error(w, "Failed to create group", http.StatusInternalServerError)
  return
 }

 // Upload marker to staging bucket with EventArc trigger
 stagingPath := fmt.Sprintf("markers/%s/%s.png", city, publicID)
 signedURL, err := h.generateSignedURL(stagingPath, "PUT")
 if err != nil {
  http.Error(w, "Failed to generate upload URL", http.StatusInternalServerError)
  return
 }

 // User uploads to signed URL, which triggers EventArc
 // EventArc payload includes:
 // {
 //   "bucket": "staging-bucket",
 //   "file": "markers/{city}/{publicID}.png",
 //   "type": "groups",
 //   "city": "{city}",
 //   "groupID": "{publicID}"
 // }

 w.Header().Set("Content-Type", "application/json")
 json.NewEncoder(w).Encode(map[string]interface{}{
  "id":       group.ID,
  "name":     group.Name,
  "city":     group.City,
  "uploadURL": signedURL,
 })
}

func isValidCity(city string) bool {
 validCities := map[string]bool{
  "portland":       true,
  "salt-lake-city": true,
 }
 return validCities[city]
}

func isValidImageType(contentType string) bool {
 validTypes := map[string]bool{
  "image/png":  true,
  "image/jpeg": true,
  "image/jpg":  true,
 }
 return validTypes[contentType]
}

func slugify(s string) string {
 // Convert name to URL-friendly slug
 // "Shift2Bikes" -> "shift2bikes"
 return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}
```

### 4. Rides Query

**File: `internal/rides/repository.go`**

Query rides with group info for client:

```go
type RideResponse struct {
 ID       string `json:"id"`
 Title    string `json:"title"`
 Description string `json:"description"`
 Latitude float64 `json:"latitude"`
 Longitude float64 `json:"longitude"`
 GroupName string `json:"group_name"`
 MarkerKey string `json:"marker_key"`
 City     string `json:"city"`
}

func (r *Repository) GetRidesByCity(ctx context.Context, city string) ([]RideResponse, error) {
 query := `
  SELECT
   ri.id,
   ri.title,
   ri.description,
   ri.latitude,
   ri.longitude,
   COALESCE(g.name, 'Unknown'),
   COALESCE(g.marker, 'default'),
   COALESCE(g.city, ?)
  FROM rides ri
  LEFT JOIN ride_groups g ON ri.group_id = g.id
  WHERE g.city = ? OR (g.city IS NULL AND ? = ?)
  ORDER BY ri.created_at DESC
 `

 rows, err := r.db.QueryContext(ctx, query, city, city, city, city)
 if err != nil {
  return nil, err
 }
 defer rows.Close()

 var rides []RideResponse
 for rows.Next() {
  var ride RideResponse
  if err := rows.Scan(&ride.ID, &ride.Title, &ride.Description,
   &ride.Latitude, &ride.Longitude, &ride.GroupName, &ride.MarkerKey, &ride.City); err != nil {
   return nil, err
  }
  rides = append(rides, ride)
 }

 return rides, rows.Err()
}
```

## Frontend Implementation

### 1. Marker Loading Component

**File: `src/lib/markers.ts`**

Load and extract markers from spritesheet:

```typescript
const SPRITES_BASE_URL = "https://storage.googleapis.com/your-bucket/sprites";

interface MarkerInfo {
  x: number;
  y: number;
  width: number;
  height: number;
}

interface MarkersData {
  metadata: Record<string, MarkerInfo>;
  spritesheet: HTMLImageElement;
}

export async function loadMarkers(city: string): Promise<MarkersData> {
  const [metadata, spritesheetBlob] = await Promise.all([
    fetch(`${SPRITES_BASE_URL}/${city}/markers.json`).then((r) => r.json()),
    fetch(`${SPRITES_BASE_URL}/${city}/markers.png`).then((r) => r.blob()),
  ]);

  const spritesheet = new Image();
  spritesheet.src = URL.createObjectURL(spritesheetBlob);

  return new Promise((resolve) => {
    spritesheet.onload = () => {
      resolve({
        metadata: metadata.markers,
        spritesheet,
      });
    };
  });
}

export function extractMarkerImage(
  spritesheet: HTMLImageElement,
  metadata: MarkerInfo,
): string {
  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d")!;
  canvas.width = metadata.width;
  canvas.height = metadata.height;

  ctx.drawImage(
    spritesheet,
    metadata.x,
    metadata.y,
    metadata.width,
    metadata.height,
    0,
    0,
    metadata.width,
    metadata.height,
  );

  return canvas.toDataURL("image/png");
}
```

### 2. Map Component

**File: `src/lib/components/Map.svelte`**

Display rides with custom markers:

```svelte
<script>
 import { onMount } from 'svelte';
 import { Map as MapLibreGL } from 'maplibre-gl';
 import { loadMarkers, extractMarkerImage } from '$lib/markers';

 export let city = 'portland';
 export let rides = [];

 let map;
 let container;
 let loading = true;
 let error = null;

 async function initializeMap() {
  try {
   // Load spritesheet and metadata
   const { metadata, spritesheet } = await loadMarkers(city);

   // Initialize map
   map = new MapLibreGL({
    container,
    style: 'https://basemaps.cartocdn.com/gl/positron-gl-style/style.json',
    center: city === 'portland' ? [-122.6765, 45.5152] : [-111.8910, 40.7608],
    zoom: 12
   });

   map.on('load', () => {
    // Add individual marker images from spritesheet
    for (const [key, coords] of Object.entries(metadata)) {
     const markerDataUrl = extractMarkerImage(spritesheet, coords);
     const img = new Image();
     img.src = markerDataUrl;
     img.onload = () => {
      map.addImage(`marker-${key}`, img);
     };
    }

    // Add rides source
    map.addSource('rides', {
     type: 'geojson',
     data: {
      type: 'FeatureCollection',
      features: rides.map(ride => ({
       type: 'Feature',
       geometry: {
        type: 'Point',
        coordinates: [ride.longitude, ride.latitude]
       },
       properties: {
        id: ride.id,
        title: ride.title,
        markerKey: ride.marker_key,
        groupName: ride.group_name
       }
      }))
     }
    });

    // Add marker layer
    map.addLayer({
     id: 'ride-markers',
     type: 'symbol',
     source: 'rides',
     layout: {
      'icon-image': ['concat', 'marker-', ['get', 'markerKey']],
      'icon-size': 1,
      'icon-offset': ['get', 'offset'] // Custom offset logic
     }
    });

    // Handle marker clicks
    map.on('click', 'ride-markers', (e) => {
     const ride = e.features[0].properties;
     onMarkerClick(ride.id);
    });

    map.getCanvas().style.cursor = 'pointer';

    loading = false;
   });
  } catch (err) {
   console.error('Failed to initialize map:', err);
   error = err.message;
  }
 }

 function onMarkerClick(rideId) {
  // Dispatch event or navigate to ride detail
  window.location.href = `/rides/${rideId}`;
 }

 onMount(() => {
  initializeMap();
 });
</script>

<div bind:this={container} class="map">
 {#if loading}
  <p>Loading markers...</p>
 {:else if error}
  <p>Error loading map: {error}</p>
 {/if}
</div>

<style>
 .map {
  width: 100%;
  height: 600px;
  position: relative;
 }
</style>
```

### 3. Ride List Page

**File: `src/routes/rides/+page.svelte`**

Display rides for selected city:

```svelte
<script>
 import Map from '$lib/components/Map.svelte';

 export let data;

 $: city = data.city;
 $: rides = data.rides;
</script>

<div class="page">
 <h1>Bike Rides in {city}</h1>

 <Map {city} {rides} />

 <div class="rides-list">
  {#each rides as ride (ride.id)}
   <RideCard {ride} />
  {/each}
 </div>
</div>
```

## Implementation Plan

### Phase 1: Backend Setup (Weekend)

- [ ] Add `groups` case to image optimization service
- [ ] Implement `cmd/build-markers/main.go`
- [ ] Implement marker upload handler with validation
- [ ] Update rides query to include group info (no internal IDs exposed)

### Phase 2: Frontend Integration

- [ ] Implement marker loading and extraction
- [ ] Build Map component with marker rendering
- [ ] Update Map component to use custom offsets for pointer positioning
- [ ] Integrate click handler with existing ride card display

### Phase 3: Testing & Polish

- [ ] Test marker upload and spritesheet generation
- [ ] Test multi-city spritesheet loading
- [ ] Test marker rendering and interactions
- [ ] Verify no marker overlap with custom offset logic

## Data Flow

1. **Group Creation**
   - Organizer submits marker image via form
   - Frontend generates signed URL for staging bucket
   - User uploads to signed URL
   - EventArc trigger fires with `type: "groups"`

2. **Spritesheet Generation**
   - Image optimization service downloads from staging
   - Processes to 40x40 PNG
   - Saves to `public/markers/{city}/{publicID}.png`
   - Regenerates spritesheet via `cmd/build-markers/main.go`
   - Uploads new spritesheet to GCP
   - Deletes staging file

3. **Ride Display**
   - User opens app, selects city
   - Service worker fetches city-specific spritesheet
   - Frontend extracts individual markers from spritesheet
   - MapLibre renders rides with custom markers
   - Custom offset logic prevents marker overlap

## GCP Bucket Permissions

Bucket should be:

- Public read (anyone can fetch spritesheet)
- Service account write (for image optimization service)
- EventArc can trigger on uploads to staging path

## Notes

- Spritesheet dimensions grow dynamically based on marker count
- Markers are extracted once at app load (no per-marker fetches)
- MapLibre handles all rendering, visibility, and interactions
- Custom offset logic prevents markers from stacking
- City-specific spritesheets keep payload sizes reasonable
- No internal group IDs exposed to frontend (security)
- All permission checks happen on backend
