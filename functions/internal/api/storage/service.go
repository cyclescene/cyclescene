package storage

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

type Service struct {
	stagingBucketName      string
	optimizedBucketName    string
	projectID              string
	client                 *storage.Client
	signedURLDuration      time.Duration
	serviceAccountEmail    string
}

// NewService creates a new storage service
// Uses Application Default Credentials (will work on Cloud Run with service account)
func NewService() (*Service, error) {
	stagingBucketName := os.Getenv("STAGING_BUCKET_NAME")
	optimizedBucketName := os.Getenv("OPTIMIZED_BUCKET_NAME")
	projectID := os.Getenv("GCP_PROJECT")
	serviceAccountEmail := os.Getenv("SERVICE_ACCOUNT_EMAIL")

	if stagingBucketName == "" {
		return nil, fmt.Errorf("STAGING_BUCKET_NAME environment variable not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("GCP_PROJECT environment variable not set")
	}
	if serviceAccountEmail == "" {
		return nil, fmt.Errorf("SERVICE_ACCOUNT_EMAIL environment variable not set")
	}

	// OPTIMIZED_BUCKET_NAME is optional, only needed for image view URLs
	if optimizedBucketName == "" {
		slog.Warn("OPTIMIZED_BUCKET_NAME environment variable not set, image view URLs will not be available")
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	// Default signed URL duration: 15 minutes for uploads, 7 days for viewing
	duration := 15 * time.Minute
	if d := os.Getenv("SIGNED_URL_DURATION_MINUTES"); d != "" {
		if minutes, err := time.ParseDuration(d + "m"); err == nil {
			duration = minutes
		}
	}

	return &Service{
		stagingBucketName:      stagingBucketName,
		optimizedBucketName:    optimizedBucketName,
		projectID:              projectID,
		client:                 client,
		signedURLDuration:      duration,
		serviceAccountEmail:    serviceAccountEmail,
	}, nil
}

// GenerateSignedURL creates a signed URL that allows the frontend to upload a file
func (s *Service) GenerateSignedURL(ctx context.Context, req *SignedURLRequest) (*SignedURLResponse, error) {
	if req.FileName == "" || req.FileType == "" {
		return nil, fmt.Errorf("file_name and file_type are required")
	}

	// Validate metadata fields are provided
	if req.CityCode == "" || req.EntityType == "" {
		return nil, fmt.Errorf("city_code and entity_type are required")
	}

	// Generate a UUID for the image
	imageUUID := uuid.New().String()

	// Determine file extension from MIME type
	ext := getExtensionFromMimeType(req.FileType)

	// Create object name: {uuid}.{ext}
	objectName := fmt.Sprintf("%s%s", imageUUID, ext)

	slog.Info("generating signed URL", "object", objectName, "bucket", s.stagingBucketName, "duration", s.signedURLDuration, "imageUUID", imageUUID, "cityCode", req.CityCode, "entityType", req.EntityType)

	// Generate signed URL with metadata
	signedURL, err := s.generateSignedURLWithMetadata(ctx, objectName, req.FileType, req.CityCode, req.EntityType)
	if err != nil {
		slog.Error("failed to generate signed URL", "error", err, "object", objectName)
		return nil, fmt.Errorf("failed to generate signed URL: %v", err)
	}

	expiresAt := time.Now().Add(s.signedURLDuration)

	slog.Info("signed URL generated successfully", "imageUUID", imageUUID, "expiresAt", expiresAt)

	return &SignedURLResponse{
		Success:    true,
		SignedURL:  signedURL,
		ObjectName: objectName,
		ImageUUID:  imageUUID,
		ExpiresAt:  expiresAt,
		BucketName: s.stagingBucketName,
	}, nil
}

// generateSignedURLWithMetadata creates a signed URL using the storage client
// The client automatically signs URLs using the service account's credentials
// Works on Cloud Run with Application Default Credentials (service account)
func (s *Service) generateSignedURLWithMetadata(ctx context.Context, objectName, contentType, cityCode, entityType string) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:      storage.SigningSchemeV4,
		Method:      "PUT",
		Expires:     time.Now().Add(s.signedURLDuration),
		ContentType: contentType,
	}

	// Use the storage client's built-in signing with the service account credentials
	// The client was initialized with Application Default Credentials on Cloud Run
	signedURL, err := s.client.Bucket(s.stagingBucketName).SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %v", err)
	}

	slog.Info("generated signed URL", "object", objectName, "cityCode", cityCode, "entityType", entityType)
	return signedURL, nil
}

// GenerateImageViewURL creates a signed URL for viewing an optimized image
// The URL is valid for 7 days (Google Cloud Storage maximum signed URL expiration)
func (s *Service) GenerateImageViewURL(ctx context.Context, req *ImageViewURLRequest) (*ImageViewURLResponse, error) {
	if req.ObjectPath == "" {
		return nil, fmt.Errorf("object_path is required")
	}

	if s.optimizedBucketName == "" {
		return nil, fmt.Errorf("optimized bucket not configured")
	}

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(7 * 24 * time.Hour), // 7 days (GCS max)
	}

	signedURL, err := s.client.Bucket(s.optimizedBucketName).SignedURL(req.ObjectPath, opts)
	if err != nil {
		slog.Error("failed to generate image view URL", "error", err, "object", req.ObjectPath)
		return nil, fmt.Errorf("failed to generate signed URL: %v", err)
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	slog.Info("image view URL generated", "object", req.ObjectPath, "expiresAt", expiresAt)

	return &ImageViewURLResponse{
		Success:   true,
		SignedURL: signedURL,
		ExpiresAt: expiresAt,
	}, nil
}

// Close closes the storage client
func (s *Service) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// getExtensionFromMimeType returns the file extension for a given MIME type
func getExtensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "image/svg+xml":
		return ".svg"
	default:
		return ".jpg" // Default to JPEG
	}
}
