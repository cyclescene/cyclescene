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

	"cloud.google.com/go/storage"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
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
	// Try to find the image file with common extensions since we don't know the exact format
	extensions := []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}
	var imageData []byte
	var stagingObjectName string
	var err error

	for _, ext := range extensions {
		stagingObjectName = fmt.Sprintf("%s%s", imageUUID, ext)
		slog.Info("attempting to download image from staging", "bucket", p.stagingBucket, "object", stagingObjectName)

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

	// Optimize image and convert to WebP format
	slog.Info("optimizing image", "imageUUID", imageUUID)
	optimizedImageData, err := p.optimizeImage(imageData)
	if err != nil {
		return "", fmt.Errorf("failed to optimize image: %v", err)
	}

	// Build optimized bucket path: {cityCode}/{entityType}s/{entityID}/{entityID}_optimized.webp
	entityTypePlural := entityType + "s"
	optimizedObjectName := filepath.Join(cityCode, entityTypePlural, entityID, fmt.Sprintf("%s_optimized.webp", entityID))

	// Upload to optimized bucket
	slog.Info("uploading optimized image", "bucket", p.optimizedBucket, "object", optimizedObjectName)
	if err := p.uploadToGCS(ctx, p.optimizedBucket, optimizedObjectName, optimizedImageData); err != nil {
		return "", fmt.Errorf("failed to upload optimized image: %v", err)
	}

	// Build public URL
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", p.optimizedBucket, optimizedObjectName)

	// Delete staging file
	slog.Info("deleting staging file", "bucket", p.stagingBucket, "object", stagingObjectName)
	if err := p.deleteFromGCS(ctx, p.stagingBucket, stagingObjectName); err != nil {
		slog.Warn("failed to delete staging file, continuing anyway", "error", err)
		// Don't fail the operation if deletion fails
	}

	return publicURL, nil
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

// optimizeImage compresses and optimizes an image to WebP format
func (p *ImageProcessor) optimizeImage(imageData []byte) ([]byte, error) {
	slog.Info("starting image optimization", "dataSize", len(imageData))

	// Decode image config to get format info
	img, format, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err != nil {
		slog.Error("decode config failed", "error", err, "dataSize", len(imageData))
		return nil, fmt.Errorf("failed to decode image config: %v", err)
	}

	slog.Info("image info", "format", format, "width", img.Width, "height", img.Height)

	// Decode the full image
	fullImg, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode full image: %v", err)
	}

	// Create a buffer to write WebP data
	var webpBuffer bytes.Buffer

	// Create WebP encoder options
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, float32(webpQuality))
	if err != nil {
		return nil, fmt.Errorf("failed to create WebP encoder options: %v", err)
	}

	// Encode to WebP and write to buffer
	if err := webp.Encode(&webpBuffer, fullImg, options); err != nil {
		return nil, fmt.Errorf("failed to encode image to WebP: %v", err)
	}

	webpData := webpBuffer.Bytes()

	slog.Info("image compression complete", "originalSize", len(imageData), "compressedSize", len(webpData), "ratio", float64(len(webpData))/float64(len(imageData)), "format", "webp")

	return webpData, nil
}

// Close closes the storage client
func (p *ImageProcessor) Close() error {
	return p.storageClient.Close()
}
