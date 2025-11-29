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

// MarkerInfo represents a marker's metadata
type MarkerInfo struct {
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Path string `json:"path"` // Path to individual marker image in GCS
}

// SpritesheetMetadata represents the complete spritesheet metadata
type SpritesheetMetadata struct {
	Markers map[string]MarkerInfo `json:"markers"`
}

// RegenerateSpritesheet regenerates the spritesheet for a city after a new marker is added
// Steps:
// 1. Download existing metadata (if exists)
// 2. Download individual marker images from GCS
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

	// Save and add the new marker to the collection
	newMarkerPath := fmt.Sprintf("%s/%s/marker.png", cityCode, newMarkerID)
	if err := p.saveMarkerImage(ctx, newMarkerPath, newMarkerImg); err != nil {
		return fmt.Errorf("failed to save new marker image: %v", err)
	}
	markers[newMarkerID] = newMarkerImg
	slog.Info("added new marker to collection", "markerID", newMarkerID, "path", newMarkerPath, "totalMarkers", len(markers))

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
		dstRect := image.Rect(x, y, x+markerSize, y+markerSize)
		draw.Draw(spritesheet, dstRect, markerImg, image.Pt(0, 0), draw.Over)

		// Get path from existing metadata or use new marker path
		markerPath := ""
		if info, exists := existingMetadata.Markers[markerID]; exists {
			markerPath = info.Path
		} else {
			// This is the new marker
			markerPath = fmt.Sprintf("%s/%s/marker.png", cityCode, markerID)
		}

		// Store metadata with path reference
		newMetadata.Markers[markerID] = MarkerInfo{
			X:    x,
			Y:    y,
			Path: markerPath,
		}

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
