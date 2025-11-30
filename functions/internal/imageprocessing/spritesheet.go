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

	// If database is available, query for all active groups in this city to ensure we have all markers
	var dbGroupMarkers map[string]bool
	if p.db != nil {
		dbGroupMarkers, err = p.getGroupMarkersFromDB(ctx, cityCode)
		if err != nil {
			slog.Warn("failed to query groups from database, will use metadata only", "city", cityCode, "error", err)
			dbGroupMarkers = make(map[string]bool)
		} else {
			slog.Info("queried database for group markers", "city", cityCode, "groupCount", len(dbGroupMarkers))
		}
	}

	// Merge markers from both metadata and database
	// This ensures we include all groups even if metadata was incomplete
	allMarkerIDs := make(map[string]bool)
	for markerID := range existingMetadata.Markers {
		allMarkerIDs[markerID] = true
	}
	for markerID := range dbGroupMarkers {
		allMarkerIDs[markerID] = true
	}

	slog.Info("total markers to include", "city", cityCode, "count", len(allMarkerIDs), "fromMetadata", len(existingMetadata.Markers), "fromDB", len(dbGroupMarkers))

	// Download individual marker images from the paths in metadata
	if len(existingMetadata.Markers) > 0 {
		for markerID, info := range existingMetadata.Markers {
			if info.Path == "" {
				slog.Warn("marker has no path in metadata, skipping", "markerID", markerID)
				continue
			}

			markerData, err := p.downloadFromGCS(ctx, p.optimizedBucket, info.Path)
			if err != nil {
				slog.Warn("failed to download individual marker image", "markerID", markerID, "path", info.Path, "error", err)
				continue
			}

			markerImg, _, err := image.Decode(bytes.NewReader(markerData))
			if err != nil {
				slog.Warn("failed to decode marker image", "markerID", markerID, "error", err)
				continue
			}

			markers[markerID] = markerImg
			slog.Debug("loaded existing marker", "markerID", markerID, "path", info.Path)
		}
	}

	// For markers in the database but not in metadata, try to load from standard path
	slog.Info("attempting to load markers from standard paths", "totalToCheck", len(allMarkerIDs), "alreadyLoaded", len(markers))
	for markerID := range allMarkerIDs {
		if _, alreadyLoaded := markers[markerID]; alreadyLoaded {
			slog.Debug("marker already loaded from metadata", "markerID", markerID)
			continue // Already loaded from metadata
		}

		// Try to load from standard path: {cityCode}/groups/{markerID}/marker.png
		standardPath := fmt.Sprintf("%s/groups/%s/marker.png", cityCode, markerID)
		markerData, err := p.downloadFromGCS(ctx, p.optimizedBucket, standardPath)
		if err != nil {
			slog.Warn("marker image not found at standard path", "markerID", markerID, "path", standardPath, "error", err)
			continue
		}

		markerImg, _, err := image.Decode(bytes.NewReader(markerData))
		if err != nil {
			slog.Warn("failed to decode marker image from standard path", "markerID", markerID, "path", standardPath, "error", err)
			continue
		}

		markers[markerID] = markerImg
		slog.Info("loaded marker from standard path", "markerID", markerID, "path", standardPath)
	}
	slog.Info("finished loading markers from standard paths", "totalLoaded", len(markers))

	// Save and add the new marker to the collection
	newMarkerPath := fmt.Sprintf("%s/groups/%s/marker.png", cityCode, newMarkerID)
	if err := p.saveMarkerImage(ctx, newMarkerPath, newMarkerImg); err != nil {
		return fmt.Errorf("failed to save new marker image: %v", err)
	}
	markers[newMarkerID] = newMarkerImg
	slog.Info("added new marker to collection", "markerID", newMarkerID, "path", newMarkerPath, "totalMarkers", len(markers))

	// Sort marker IDs for consistent spritesheet layout
	// Only include markers that we actually have images for
	markerIDs := make([]string, 0, len(markers))
	for id := range markers {
		markerIDs = append(markerIDs, id)
	}
	sort.Strings(markerIDs)

	slog.Info("markers to composite", "city", cityCode, "count", len(markerIDs), "from", len(allMarkerIDs), "total")

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

	if _, err := io.Copy(writer, bytes.NewReader(data)); err != nil {
		return err
	}

	return writer.Close()
}

// getGroupMarkersFromDB queries the database for all active group markers in a city
// Returns a map of marker IDs (slugified group codes) for groups in the specified city
func (p *ImageProcessor) getGroupMarkersFromDB(ctx context.Context, cityCode string) (map[string]bool, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	markers := make(map[string]bool)

	// Query all active groups in the city
	// If marker field is set, use it; otherwise use slugified code as fallback
	query := `
		SELECT COALESCE(NULLIF(marker, ''), LOWER(code)) as marker_id
		FROM ride_groups
		WHERE city = ? AND is_active = 1
	`

	rows, err := p.db.QueryContext(ctx, query, strings.ToLower(cityCode))
	if err != nil {
		return nil, fmt.Errorf("failed to query group markers: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var marker string
		if err := rows.Scan(&marker); err != nil {
			slog.Warn("failed to scan marker from database", "error", err)
			continue
		}
		if marker != "" {
			markers[marker] = true
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating marker rows: %v", err)
	}

	slog.Info("fetched group markers from database", "city", cityCode, "markerCount", len(markers))
	return markers, nil
}
