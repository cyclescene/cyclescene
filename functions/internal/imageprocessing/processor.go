package imageprocessing

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"golang.org/x/image/draw"
)

const (
	webpQuality = 85
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// hexToRGBA converts a hex color string (#RRGGBB) to RGBA color
func hexToRGBA(hexColor string) (color.Color, error) {
	hex := strings.TrimPrefix(hexColor, "#")
	if len(hex) != 6 {
		return nil, fmt.Errorf("invalid hex color: %s", hexColor)
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid hex color: %s", hexColor)
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid hex color: %s", hexColor)
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid hex color: %s", hexColor)
	}

	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}

// createTeardropMarker creates a teardrop-shaped marker with circular user image inside
// The teardrop is colored with the provided markerColor, with padding that forms the teardrop point
func createTeardropMarker(userImg image.Image, size int, markerColor string) image.Image {
	// Parse marker color
	bgColor, err := hexToRGBA(markerColor)
	if err != nil {
		slog.Warn("failed to parse marker color, using default blue", "error", err)
		bgColor = color.RGBA{R: 59, G: 130, B: 246, A: 255} // Default blue (#3B82F6)
	}

	// Create new image for teardrop
	teardrop := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill with transparent background
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			teardrop.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
		}
	}

	// Draw teardrop shape with marker color
	centerX := float64(size) / 2.0
	centerY := float64(size) * 0.38 // Position circle in upper part for teardrop effect

	// Draw teardrop body (circle in upper part)
	circleRadius := float64(size) * 0.32
	// Convert bgColor to RGBA for SetRGBA
	r, g, b, a := bgColor.RGBA()
	rgba := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			// Teardrop body - circle part
			if dist <= circleRadius {
				teardrop.SetRGBA(x, y, rgba)
			}

			// Teardrop point - triangle part (bottom padding)
			if float64(y) > centerY+circleRadius && float64(y) < centerY+circleRadius*1.95 {
				pointY := centerY + circleRadius*1.95
				widthAtY := circleRadius * (pointY - float64(y)) / (pointY - (centerY + circleRadius))
				if math.Abs(dx) <= widthAtY {
					teardrop.SetRGBA(x, y, rgba)
				}
			}
		}
	}

	// Create circular user image with proper sizing
	imageSize := int(circleRadius * 1.9)
	circularUserImg := makeCircularImage(userImg, imageSize)

	// Position circular image in center of teardrop circle
	startX := int(centerX) - imageSize/2
	startY := int(centerY) - imageSize/2

	draw.Draw(teardrop, circularUserImg.Bounds().Add(image.Pt(startX, startY)), circularUserImg, image.Pt(0, 0), draw.Over)

	return teardrop
}

// makeCircularImage creates a circular version of an image with alpha transparency
func makeCircularImage(img image.Image, size int) image.Image {
	// Resize image to desired size
	resized := resizeImage(img, size)

	// Create circular mask
	circular := image.NewRGBA(image.Rect(0, 0, size, size))
	radius := float64(size) / 2.0
	centerX := radius
	centerY := radius

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			// If within circle radius, copy pixel from resized image
			if dist <= radius {
				r, g, b, a := resized.At(x, y).RGBA()
				circular.SetRGBA(x, y, color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)})
			}
			// Otherwise keep transparent
		}
	}

	return circular
}

// slugify converts a string to a URL-safe slug
func slugify(s string) string {
	// Convert to lowercase
	slug := strings.ToLower(strings.TrimSpace(s))
	// Replace spaces and special characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}

type ImageProcessor struct {
	stagingBucket   string
	optimizedBucket string
	storageClient   *storage.Client
	db              *sql.DB
}

func NewImageProcessor(ctx context.Context, stagingBucket, optimizedBucket string) (*ImageProcessor, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	return &ImageProcessor{
		stagingBucket:   stagingBucket,
		optimizedBucket: optimizedBucket,
		storageClient:   client,
	}, nil
}

// SetDB sets the database connection for the image processor
func (p *ImageProcessor) SetDB(db *sql.DB) {
	p.db = db
}

// ProcessImage handles the complete image optimization workflow
func (p *ImageProcessor) ProcessImage(ctx context.Context, imageUUID, cityCode, entityID, entityType, markerColor string) (string, error) {
	// Check context deadline
	if deadline, ok := ctx.Deadline(); ok {
		slog.Info("context deadline set", "deadline", deadline, "timeUntilDeadline", time.Until(deadline))
	} else {
		slog.Info("no context deadline set")
	}

	// Route to appropriate handler based on entity type
	if entityType == "group" {
		return p.ProcessMarker(ctx, imageUUID, cityCode, entityID, markerColor)
	}

	// Try to find the image file with common extensions since we don't know the exact format
	extensions := []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}
	var imageData []byte
	var stagingObjectName string
	var err error

	for _, ext := range extensions {
		stagingObjectName = fmt.Sprintf("%s%s", imageUUID, ext)
		slog.Info("attempting to download image from staging", "bucket", p.stagingBucket, "object", stagingObjectName)

		// Check if object exists first
		exists, err := p.objectExists(ctx, p.stagingBucket, stagingObjectName)
		if err != nil {
			slog.Warn("failed to check if object exists", "extension", ext, "error", err)
			continue
		}

		if !exists {
			slog.Info("object does not exist in staging bucket", "extension", ext, "object", stagingObjectName)
			continue
		}

		imageData, err = p.downloadFromGCS(ctx, p.stagingBucket, stagingObjectName)
		if err == nil {
			slog.Info("found image with extension", "extension", ext)
			break
		}
		slog.Warn("failed to download with this extension", "extension", ext, "error", err)
	}

	if imageData == nil {
		return "", fmt.Errorf("failed to download image from staging bucket with any common extension")
	}

	// Log file info for debugging
	slog.Info("downloaded image file", "imageUUID", imageUUID, "stagingObjectName", stagingObjectName, "fileSize", len(imageData), "firstBytes", fmt.Sprintf("%x", imageData[:min(8, len(imageData))]))

	// Decode the full image once
	fullImg, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to decode full image: %v", err)
	}

	// Define sizes to generate
	sizes := []int{400, 800, 1200}
	entityTypePlural := entityType + "s"
	var mainPublicURL string

	for _, width := range sizes {
		slog.Info("generating resized image", "width", width)

		resizedImg := resizeImage(fullImg, width)
		optimizedData, err := encodeWebP(resizedImg)
		if err != nil {
			return "", fmt.Errorf("failed to encode resized image (width %d): %v", width, err)
		}

		// Object name: {cityCode}/{entityType}s/{entityID}/{entityID}_{width}w.webp
		objectName := filepath.Join(cityCode, entityTypePlural, entityID, fmt.Sprintf("%s_%dw.webp", entityID, width))

		slog.Info("uploading resized image", "bucket", p.optimizedBucket, "object", objectName)
		if err := p.uploadToGCS(ctx, p.optimizedBucket, objectName, optimizedData); err != nil {
			return "", fmt.Errorf("failed to upload resized image: %v", err)
		}

		// If this is the largest size (1200), also save it as the default "optimized" version for backward compatibility
		if width == 1200 {
			defaultObjectName := filepath.Join(cityCode, entityTypePlural, entityID, fmt.Sprintf("%s_optimized.webp", entityID))
			slog.Info("uploading default optimized image", "bucket", p.optimizedBucket, "object", defaultObjectName)
			if err := p.uploadToGCS(ctx, p.optimizedBucket, defaultObjectName, optimizedData); err != nil {
				return "", fmt.Errorf("failed to upload default optimized image: %v", err)
			}
			mainPublicURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", p.optimizedBucket, defaultObjectName)
		}
	}

	// Delete staging file
	slog.Info("deleting staging file", "bucket", p.stagingBucket, "object", stagingObjectName)
	if err := p.deleteFromGCS(ctx, p.stagingBucket, stagingObjectName); err != nil {
		slog.Warn("failed to delete staging file, continuing anyway", "error", err)
		// Don't fail the operation if deletion fails
	}

	// Return the public URL for the optimized image
	slog.Info("image optimization complete", "publicURL", mainPublicURL)
	return mainPublicURL, nil
}

// ProcessMarker handles marker image processing for group spritesheets
// Steps:
// 1. Download marker from staging bucket
// 2. Decode and validate image
// 3. Create teardrop shape container with marker color
// 4. Draw user's image on teardrop shape
// 5. Resize to 64x64 PNG
// 6. Regenerate spritesheet for the city (extracts old markers + adds new)
// 7. Query group by code and slugify name to create public_id
// 8. Set public_id and marker in database after spritesheet is ready
// 9. Delete staging file
// 10. Return path to spritesheet PNG
func (p *ImageProcessor) ProcessMarker(ctx context.Context, imageUUID, cityCode, groupCode, markerColor string) (string, error) {
	slog.Info("processing marker", "imageUUID", imageUUID, "cityCode", cityCode, "groupCode", groupCode)

	// Try to find the image file with common extensions
	extensions := []string{".png", ".jpg", ".jpeg", ".gif", ".webp"}
	var imageData []byte
	var stagingObjectName string
	var err error

	for _, ext := range extensions {
		stagingObjectName = fmt.Sprintf("%s%s", imageUUID, ext)
		slog.Info("attempting to download marker from staging", "bucket", p.stagingBucket, "object", stagingObjectName)

		// Check if object exists first
		exists, err := p.objectExists(ctx, p.stagingBucket, stagingObjectName)
		if err != nil {
			slog.Warn("failed to check if object exists", "extension", ext, "error", err)
			continue
		}

		if !exists {
			slog.Info("object does not exist in staging bucket", "extension", ext, "object", stagingObjectName)
			continue
		}

		imageData, err = p.downloadFromGCS(ctx, p.stagingBucket, stagingObjectName)
		if err == nil {
			slog.Info("found marker with extension", "extension", ext)
			break
		}
		slog.Warn("failed to download marker with this extension", "extension", ext, "error", err)
	}

	if imageData == nil {
		return "", fmt.Errorf("failed to download marker from staging bucket with any common extension")
	}

	slog.Info("downloaded marker file", "imageUUID", imageUUID, "stagingObjectName", stagingObjectName, "fileSize", len(imageData))

	// Decode the marker image
	markerImg, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to decode marker image: %v", err)
	}

	// Resize marker to 64x64 directly (skipping teardrop generation)
	const markerSize = 64
	resizedMarker := resizeImage(markerImg, markerSize)

	// Use the slugified group code as the marker key in the spritesheet
	markerKey := slugify(groupCode)
	slog.Info("marker key", "markerKey", markerKey)

	// Regenerate spritesheet with new marker
	if err := p.RegenerateSpritesheet(ctx, cityCode, markerKey, resizedMarker); err != nil {
		return "", fmt.Errorf("failed to regenerate spritesheet: %v", err)
	}

	// If database is available, get group details and set public_id and marker
	if p.db != nil {
		dbConnector := &DBConnector{db: p.db}

		// Get group by code to retrieve name for public_id generation
		groupID, groupName, err := dbConnector.GetGroupByCode(groupCode)
		if err != nil {
			slog.Warn("failed to get group by code, marker will not be persisted in database", "groupCode", groupCode, "error", err)
		} else {
			// Generate public_id by slugifying the group name
			publicID := slugify(groupName)
			slog.Info("setting marker for group", "groupCode", groupCode, "publicID", publicID, "markerKey", markerKey, "groupID", groupID)

			// Update group with public_id and marker
			if err := dbConnector.SetGroupMarkerAndPublicID(groupCode, publicID, markerKey); err != nil {
				slog.Warn("failed to set group marker and public_id", "groupCode", groupCode, "error", err)
				// Don't fail the marker processing if DB update fails - spritesheet is already ready
			} else {
				slog.Info("successfully set marker for group", "groupCode", groupCode, "publicID", publicID)
			}
		}
	} else {
		slog.Warn("database connection not available, skipping marker persistence")
	}

	// Delete staging file
	slog.Info("deleting staging file", "bucket", p.stagingBucket, "object", stagingObjectName)
	if err := p.deleteFromGCS(ctx, p.stagingBucket, stagingObjectName); err != nil {
		slog.Warn("failed to delete staging file, continuing anyway", "error", err)
	}

	// Return the spritesheet PNG path
	spritesheetPath := fmt.Sprintf("https://storage.googleapis.com/%s/sprites/%s/markers.png", p.optimizedBucket, cityCode)
	slog.Info("marker processing complete", "spritesheetPath", spritesheetPath)
	return spritesheetPath, nil
}

// objectExists checks if an object exists in Google Cloud Storage
func (p *ImageProcessor) objectExists(ctx context.Context, bucket, object string) (bool, error) {
	_, err := p.storageClient.Bucket(bucket).Object(object).Attrs(ctx)
	if err != nil {
		// Check if error is "not found" (object doesn't exist)
		if err.Error() == "storage: object doesn't exist" {
			return false, nil
		}
		// Return other errors (permissions, network, etc.)
		return false, err
	}
	return true, nil
}

// downloadFromGCS reads a file from Google Cloud Storage
func (p *ImageProcessor) downloadFromGCS(ctx context.Context, bucket, object string) ([]byte, error) {
	reader, err := p.storageClient.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// uploadToGCS writes a file to Google Cloud Storage
func (p *ImageProcessor) uploadToGCS(ctx context.Context, bucket, object string, data []byte) error {
	writer := p.storageClient.Bucket(bucket).Object(object).NewWriter(ctx)
	writer.ContentType = "image/webp"

	if _, err := writer.Write(data); err != nil {
		return err
	}

	return writer.Close()
}

// deleteFromGCS removes a file from Google Cloud Storage
func (p *ImageProcessor) deleteFromGCS(ctx context.Context, bucket, object string) error {
	return p.storageClient.Bucket(bucket).Object(object).Delete(ctx)
}

// resizeImage resizes the image to the specified width while maintaining aspect ratio
func resizeImage(img image.Image, width int) image.Image {
	bounds := img.Bounds()
	ratio := float64(bounds.Dy()) / float64(bounds.Dx())
	height := int(float64(width) * ratio)

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return dst
}

// encodeWebP encodes an image to WebP format
func encodeWebP(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, float32(webpQuality))
	if err != nil {
		return nil, err
	}

	if err := webp.Encode(&buf, img, options); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Close closes the storage client
func (p *ImageProcessor) Close() error {
	return p.storageClient.Close()
}
