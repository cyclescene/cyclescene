package imageprocessing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log/slog"
	"math"
	"sort"
	"strings"
)

const (
	markerSize = 64
	padding    = 2
)

// MarkerInfo represents a marker's metadata
type MarkerInfo struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Path   string `json:"path"` // Path to individual marker image in GCS
}

// SpritesheetMetadata represents the complete spritesheet metadata
type SpritesheetMetadata struct {
	Markers map[string]MarkerInfo `json:"markers"`
}

// RegenerateSpritesheet regenerates the spritesheet for a city after a new marker is added
// Steps:
// 1. Query database for all active groups in the city
// 2. Download individual marker images from GCS using paths from metadata
// 3. Save the new marker image individually
// 4. Regenerate spritesheet by compositing all individual markers
// 5. Upload new spritesheet PNG and updated metadata JSON
func (p *ImageProcessor) RegenerateSpritesheet(ctx context.Context, cityCode string, newMarkerID string, newMarkerImg image.Image) error {
	slog.Info("regenerating spritesheet", "city", cityCode, "newMarkerID", newMarkerID)

	// Collection of all markers: ID -> image
	markers := make(map[string]image.Image)

	// Try to download existing metadata
	existingMetadata, err := p.downloadExistingMetadata(ctx, cityCode)
	if err != nil {
		slog.Warn("no existing metadata found, starting fresh", "city", cityCode, "error", err)
		existingMetadata = &SpritesheetMetadata{Markers: make(map[string]MarkerInfo)}
	}

	// If database is available, query for all active groups in this city
	var dbGroups map[string]string // markerID -> markerPath
	if p.db != nil {
		dbGroups, err = p.getGroupMarkersPathsFromDB(ctx, cityCode)
		if err != nil {
			slog.Warn("failed to query groups from database, will use metadata only", "city", cityCode, "error", err)
			dbGroups = make(map[string]string)
		} else {
			slog.Info("queried database for group markers", "city", cityCode, "groupCount", len(dbGroups))
		}
	}

	// Check which markers actually have image files in GCS
	// This is the source of truth - if the image file exists, include it
	validMarkerIDs := make(map[string]string) // markerID -> markerPath

	// First check markers from existing metadata
	slog.Info("checking markers from metadata", "city", cityCode, "metadataMarkerCount", len(existingMetadata.Markers))
	for markerID, info := range existingMetadata.Markers {
		if info.Path != "" {
			validMarkerIDs[markerID] = info.Path
			slog.Info("added marker from metadata", "markerID", markerID, "path", info.Path)
		}
	}
	slog.Info("after checking metadata", "city", cityCode, "validMarkerCount", len(validMarkerIDs))

	// Then check markers from database by trying to load them from standard path
	slog.Info("checking markers from database", "city", cityCode, "dbMarkerCount", len(dbGroups))
	for markerID := range dbGroups {
		if _, alreadyHave := validMarkerIDs[markerID]; alreadyHave {
			slog.Debug("marker already in collection from metadata", "markerID", markerID)
			continue // Already have this from metadata
		}

		// Try to check if file exists at standard path
		standardPath := fmt.Sprintf("%s/groups/%s/marker.png", cityCode, markerID)
		slog.Info("checking if marker file exists", "markerID", markerID, "path", standardPath, "bucket", p.optimizedBucket)
		exists, err := p.objectExists(ctx, p.optimizedBucket, standardPath)
		if err != nil {
			slog.Warn("failed to check if marker exists", "markerID", markerID, "path", standardPath, "error", err)
			continue
		}

		if exists {
			validMarkerIDs[markerID] = standardPath
			slog.Info("added marker from database (file exists)", "markerID", markerID, "bucket", p.optimizedBucket, "path", standardPath)
		} else {
			slog.Warn("marker in database but file not found", "markerID", markerID, "bucket", p.optimizedBucket, "path", standardPath)
		}
	}
	slog.Info("after checking database", "city", cityCode, "validMarkerCount", len(validMarkerIDs))

	slog.Info("markers found with actual image files", "city", cityCode, "totalValid", len(validMarkerIDs), "fromMetadata", len(existingMetadata.Markers), "fromDB", len(dbGroups))

	// Download marker images from validated paths
	slog.Info("loading marker images from GCS", "totalMarkers", len(validMarkerIDs))
	for markerID, markerPath := range validMarkerIDs {
		markerData, err := p.downloadFromGCS(ctx, p.optimizedBucket, markerPath)
		if err != nil {
			slog.Warn("failed to download marker image", "markerID", markerID, "path", markerPath, "error", err)
			continue
		}

		markerImg, _, err := image.Decode(bytes.NewReader(markerData))
		if err != nil {
			slog.Warn("failed to decode marker image", "markerID", markerID, "path", markerPath, "error", err)
			continue
		}

		markers[markerID] = markerImg
		slog.Info("loaded marker image", "markerID", markerID, "path", markerPath)
	}
	slog.Info("finished loading markers", "loaded", len(markers), "total", len(validMarkerIDs))

	// Save and add the new marker to the collection
	newMarkerPath := fmt.Sprintf("%s/groups/%s/marker.png", cityCode, newMarkerID)
	if err := p.saveMarkerImage(ctx, newMarkerPath, newMarkerImg); err != nil {
		return fmt.Errorf("failed to save new marker image: %v", err)
	}

	// Verify the marker was saved successfully
	exists, err := p.objectExists(ctx, p.optimizedBucket, newMarkerPath)
	if err != nil || !exists {
		return fmt.Errorf("failed to verify marker image was saved: path=%s, exists=%v, error=%v", newMarkerPath, exists, err)
	}
	slog.Info("verified marker image saved successfully", "markerID", newMarkerID, "path", newMarkerPath)

	markers[newMarkerID] = newMarkerImg
	slog.Info("added new marker to collection", "markerID", newMarkerID, "path", newMarkerPath, "totalMarkers", len(markers))

	// Sort marker IDs for consistent spritesheet layout
	// Only include markers that we actually have images for
	markerIDs := make([]string, 0, len(markers))
	for id := range markers {
		markerIDs = append(markerIDs, id)
	}
	sort.Strings(markerIDs)

	slog.Info("markers to composite", "city", cityCode, "count", len(markerIDs), "validated", len(validMarkerIDs))

	// Calculate grid dimensions
	cols := int(math.Ceil(math.Sqrt(float64(len(markers)))))
	rows := (len(markers) + cols - 1) / cols
	sheetWidth := cols * (markerSize + padding)
	sheetHeight := rows * (markerSize + padding)

	slog.Info("spritesheet layout", "cols", cols, "rows", rows, "width", sheetWidth, "height", sheetHeight)

	// Create new blank spritesheet
	spritesheet := image.NewRGBA(image.Rect(0, 0, sheetWidth, sheetHeight))

	// Composite all markers and build metadata
	newMetadata := &SpritesheetMetadata{
		Markers: make(map[string]MarkerInfo),
	}

	for i, markerID := range markerIDs {
		markerImg := markers[markerID]

		// Ensure marker is in RGBA format for proper compositing
		var rgbaMarker *image.RGBA
		if rgba, ok := markerImg.(*image.RGBA); ok {
			rgbaMarker = rgba
		} else {
			// Convert to RGBA if needed
			bounds := markerImg.Bounds()
			rgbaMarker = image.NewRGBA(bounds)
			draw.Draw(rgbaMarker, bounds, markerImg, image.Pt(0, 0), draw.Over)
			slog.Info("converted marker image to RGBA", "markerID", markerID, "originalType", fmt.Sprintf("%T", markerImg))
		}

		col := i % cols
		row := i / cols
		x := col * (markerSize + padding)
		y := row * (markerSize + padding)

		// Draw marker onto spritesheet
		dstRect := image.Rect(x, y, x+markerSize, y+markerSize)
		draw.Draw(spritesheet, dstRect, rgbaMarker, image.Pt(0, 0), draw.Over)
		slog.Info("composited marker to spritesheet", "markerID", markerID, "dstRect", dstRect.String())

		// Get path from existing metadata or use new marker path
		markerPath := ""
		if info, exists := existingMetadata.Markers[markerID]; exists {
			markerPath = info.Path
			// Normalize old path format (slc/frog/marker.png) to new format (slc/groups/frog/marker.png)
			if !strings.Contains(markerPath, "/groups/") {
				markerPath = fmt.Sprintf("%s/groups/%s/marker.png", cityCode, markerID)
				slog.Info("normalized old marker path format", "markerID", markerID, "newPath", markerPath)
			}
		} else {
			// This is the new marker
			markerPath = fmt.Sprintf("%s/groups/%s/marker.png", cityCode, markerID)
		}

		// Store metadata with path reference - ensure width/height are always set
		info := MarkerInfo{
			X:      x,
			Y:      y,
			Width:  markerSize,
			Height: markerSize,
			Path:   markerPath,
		}
		newMetadata.Markers[markerID] = info
		slog.Info("marker metadata created", "markerID", markerID, "x", x, "y", y, "width", markerSize, "height", markerSize, "path", markerPath)

		slog.Debug("composited marker", "markerID", markerID, "position", fmt.Sprintf("(%d,%d)", x, y), "path", markerPath)
	}

	// Encode spritesheet to PNG
	pngBuf := new(bytes.Buffer)
	if err := png.Encode(pngBuf, spritesheet); err != nil {
		return fmt.Errorf("failed to encode spritesheet: %v", err)
	}

	// Encode metadata to JSON
	metadataBuf := new(bytes.Buffer)
	encoder := json.NewEncoder(metadataBuf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(newMetadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %v", err)
	}

	// Upload spritesheet PNG
	pngPath := fmt.Sprintf("sprites/%s/markers.png", cityCode)
	slog.Info("uploading spritesheet PNG", "bucket", p.optimizedBucket, "path", pngPath, "size", len(pngBuf.Bytes()))
	if err := p.uploadToGCSWithContentType(ctx, p.optimizedBucket, pngPath, pngBuf.Bytes(), "image/png"); err != nil {
		return fmt.Errorf("failed to upload spritesheet PNG: %v", err)
	}

	// Upload metadata JSON
	jsonPath := fmt.Sprintf("sprites/%s/markers.json", cityCode)
	slog.Info("uploading spritesheet metadata", "bucket", p.optimizedBucket, "path", jsonPath, "markerCount", len(newMetadata.Markers))
	if err := p.uploadToGCSWithContentType(ctx, p.optimizedBucket, jsonPath, metadataBuf.Bytes(), "application/json"); err != nil {
		return fmt.Errorf("failed to upload spritesheet metadata: %v", err)
	}

	slog.Info("spritesheet regeneration complete", "city", cityCode, "totalMarkers", len(newMetadata.Markers), "pngSize", len(pngBuf.Bytes()))
	return nil
}

// downloadExistingMetadata downloads the existing spritesheet metadata JSON
func (p *ImageProcessor) downloadExistingMetadata(ctx context.Context, cityCode string) (*SpritesheetMetadata, error) {
	metadataPath := fmt.Sprintf("sprites/%s/markers.json", cityCode)
	slog.Info("attempting to download existing metadata", "bucket", p.optimizedBucket, "path", metadataPath)

	data, err := p.downloadFromGCS(ctx, p.optimizedBucket, metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata SpritesheetMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %v", err)
	}

	slog.Info("loaded existing metadata", "markerCount", len(metadata.Markers))
	return &metadata, nil
}

// saveMarkerImage encodes and uploads an individual marker image to GCS
func (p *ImageProcessor) saveMarkerImage(ctx context.Context, path string, img image.Image) error {
	// Encode image to PNG
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return fmt.Errorf("failed to encode marker image: %v", err)
	}

	slog.Info("uploading individual marker image", "bucket", p.optimizedBucket, "path", path, "size", len(buf.Bytes()))
	return p.uploadToGCSWithContentType(ctx, p.optimizedBucket, path, buf.Bytes(), "image/png")
}

// uploadToGCSWithContentType uploads data to GCS with a specified content type
func (p *ImageProcessor) uploadToGCSWithContentType(ctx context.Context, bucket, object string, data []byte, contentType string) error {
	writer := p.storageClient.Bucket(bucket).Object(object).NewWriter(ctx)
	writer.ContentType = contentType
	// Set cache-control with shorter expiration to ensure updates propagate quickly
	// public: allow caching by browsers and CDNs
	// max-age=300: cache for 5 minutes before checking for updates
	writer.CacheControl = "public, max-age=300"

	if _, err := io.Copy(writer, bytes.NewReader(data)); err != nil {
		return err
	}

	return writer.Close()
}

// getGroupMarkersPathsFromDB queries the database for all active groups in a city
// Returns a map of marker IDs to their expected storage paths
func (p *ImageProcessor) getGroupMarkersPathsFromDB(ctx context.Context, cityCode string) (map[string]string, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	markers := make(map[string]string)

	// Query all active groups in the city that have markers set
	// The marker field contains the marker ID used to construct the path: {cityCode}/groups/{markerID}/marker.png
	query := `
		SELECT marker FROM ride_groups
		WHERE city = ? AND is_active = 1 AND marker IS NOT NULL AND marker != ''
	`

	rows, err := p.db.QueryContext(ctx, query, cityCode)
	if err != nil {
		return nil, fmt.Errorf("failed to query group markers: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var markerID string
		if err := rows.Scan(&markerID); err != nil {
			slog.Warn("failed to scan marker ID from database", "error", err)
			continue
		}
		if markerID != "" {
			// Store the expected path for this marker
			expectedPath := fmt.Sprintf("%s/groups/%s/marker.png", cityCode, markerID)
			markers[markerID] = expectedPath
			slog.Info("found marker in database", "markerID", markerID, "path", expectedPath, "cityCode", cityCode)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating marker rows: %v", err)
	}

	slog.Info("fetched group markers from database", "city", cityCode, "markerCount", len(markers))
	return markers, nil
}
