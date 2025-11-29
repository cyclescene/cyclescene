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
)

const (
	markerSize = 64
	padding    = 2
)

// MarkerInfo represents a marker's position in the spritesheet
type MarkerInfo struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// SpritesheetMetadata represents the complete spritesheet metadata
type SpritesheetMetadata struct {
	Markers map[string]MarkerInfo `json:"markers"`
}

// RegenerateSpritesheet regenerates the spritesheet for a city after a new marker is added
// Steps:
// 1. Download existing spritesheet + metadata (if exists)
// 2. Extract all existing markers from spritesheet using metadata bounds
// 3. Add the new marker to the collection
// 4. Regenerate spritesheet with all markers
// 5. Upload new spritesheet PNG and metadata JSON
func (p *ImageProcessor) RegenerateSpritesheet(ctx context.Context, cityCode string, newMarkerID string, newMarkerImg image.Image) error {
	slog.Info("regenerating spritesheet", "city", cityCode, "newMarkerID", newMarkerID)

	// Collection of all markers: ID -> image
	markers := make(map[string]image.Image)

	// Try to download existing spritesheet and metadata
	existingMetadata, err := p.downloadExistingMetadata(ctx, cityCode)
	if err != nil {
		slog.Warn("no existing spritesheet found, starting fresh", "city", cityCode, "error", err)
		existingMetadata = &SpritesheetMetadata{Markers: make(map[string]MarkerInfo)}
	}

	// Extract existing markers from spritesheet using metadata bounds
	if len(existingMetadata.Markers) > 0 {
		existingSpritesheetData, err := p.downloadFromGCS(ctx, p.optimizedBucket, fmt.Sprintf("sprites/%s/markers.png", cityCode))
		if err != nil {
			slog.Warn("failed to download existing spritesheet, will use only new marker", "city", cityCode, "error", err)
		} else {
			existingSpritesheet, _, err := image.Decode(bytes.NewReader(existingSpritesheetData))
			if err != nil {
				slog.Warn("failed to decode existing spritesheet", "city", cityCode, "error", err)
			} else {
				// Extract each marker from the spritesheet using metadata bounds
				for markerID, info := range existingMetadata.Markers {
					// Use SubImage to extract the rectangular region
					rect := image.Rect(info.X, info.Y, info.X+info.Width, info.Y+info.Height)
					extractedMarker := existingSpritesheet.(interface {
						SubImage(r image.Rectangle) image.Image
					}).SubImage(rect)
					markers[markerID] = extractedMarker
					slog.Debug("extracted existing marker", "markerID", markerID, "bounds", rect.String())
				}
			}
		}
	}

	// Add the new marker to the collection
	markers[newMarkerID] = newMarkerImg
	slog.Info("added new marker to collection", "markerID", newMarkerID, "totalMarkers", len(markers))

	// Sort marker IDs for consistent spritesheet layout
	markerIDs := make([]string, 0, len(markers))
	for id := range markers {
		markerIDs = append(markerIDs, id)
	}
	sort.Strings(markerIDs)

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

		col := i % cols
		row := i / cols
		x := col * (markerSize + padding)
		y := row * (markerSize + padding)

		// Draw marker onto spritesheet
		draw.Draw(spritesheet, markerImg.Bounds().Add(image.Pt(x, y)), markerImg, image.Pt(0, 0), draw.Over)

		// Store metadata
		newMetadata.Markers[markerID] = MarkerInfo{
			X:      x,
			Y:      y,
			Width:  markerSize,
			Height: markerSize,
		}

		slog.Debug("composited marker", "markerID", markerID, "position", fmt.Sprintf("(%d,%d)", x, y))
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

// uploadToGCSWithContentType uploads data to GCS with a specified content type
func (p *ImageProcessor) uploadToGCSWithContentType(ctx context.Context, bucket, object string, data []byte, contentType string) error {
	writer := p.storageClient.Bucket(bucket).Object(object).NewWriter(ctx)
	writer.ContentType = contentType

	if _, err := io.Copy(writer, bytes.NewReader(data)); err != nil {
		return err
	}

	return writer.Close()
}
