package imageprocessing

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"path/filepath"
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

type ImageProcessor struct {
	stagingBucket   string
	optimizedBucket string
	storageClient   *storage.Client
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

// ProcessImage handles the complete image optimization workflow
func (p *ImageProcessor) ProcessImage(ctx context.Context, imageUUID, cityCode, entityID, entityType string) (string, error) {
	// Check context deadline
	if deadline, ok := ctx.Deadline(); ok {
		slog.Info("context deadline set", "deadline", deadline, "timeUntilDeadline", time.Until(deadline))
	} else {
		slog.Info("no context deadline set")
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
