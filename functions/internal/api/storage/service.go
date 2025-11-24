package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

type Service struct {
	bucketName             string
	projectID              string
	client                 *storage.Client
	signedURLDuration      time.Duration
	serviceAccountEmail    string
}

// NewService creates a new storage service
// Uses Application Default Credentials (will work on Cloud Run with service account)
func NewService() (*Service, error) {
	bucketName := os.Getenv("STAGING_BUCKET_NAME")
	projectID := os.Getenv("GCP_PROJECT")
	serviceAccountEmail := os.Getenv("SERVICE_ACCOUNT_EMAIL")

	if bucketName == "" {
		return nil, fmt.Errorf("STAGING_BUCKET_NAME environment variable not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("GCP_PROJECT environment variable not set")
	}
	if serviceAccountEmail == "" {
		return nil, fmt.Errorf("SERVICE_ACCOUNT_EMAIL environment variable not set")
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	// Default signed URL duration: 15 minutes
	duration := 15 * time.Minute
	if d := os.Getenv("SIGNED_URL_DURATION_MINUTES"); d != "" {
		if minutes, err := time.ParseDuration(d + "m"); err == nil {
			duration = minutes
		}
	}

	return &Service{
		bucketName:             bucketName,
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

	slog.Info("generating signed URL", "object", objectName, "bucket", s.bucketName, "duration", s.signedURLDuration, "imageUUID", imageUUID, "cityCode", req.CityCode, "entityType", req.EntityType)

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
		BucketName: s.bucketName,
	}, nil
}

// generateSignedURLWithMetadata creates a signed URL using service account credentials
func (s *Service) generateSignedURLWithMetadata(ctx context.Context, objectName, contentType, cityCode, entityType string) (string, error) {
	privateKey, err := s.getPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get private key: %v", err)
	}

	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         "PUT",
		Expires:        time.Now().Add(s.signedURLDuration),
		ContentType:    contentType,
		GoogleAccessID: s.serviceAccountEmail,
		PrivateKey:     []byte(privateKey),
	}

	signedURL, err := storage.SignedURL(s.bucketName, objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %v", err)
	}

	slog.Info("generated signed URL with metadata", "object", objectName, "cityCode", cityCode, "entityType", entityType)
	return signedURL, nil
}

// getPrivateKey retrieves the service account private key from various sources
func (s *Service) getPrivateKey() (string, error) {
	// Try environment variable first (for Cloud Run with injected secret)
	if key := os.Getenv("SERVICE_ACCOUNT_PRIVATE_KEY"); key != "" {
		slog.Info("using SERVICE_ACCOUNT_PRIVATE_KEY from environment")
		return key, nil
	}

	// Try GOOGLE_APPLICATION_CREDENTIALS file (for local development)
	if keyPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); keyPath != "" {
		slog.Info("using GOOGLE_APPLICATION_CREDENTIALS file", "path", keyPath)
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return "", fmt.Errorf("failed to read service account key file: %v", err)
		}

		var keyFile map[string]interface{}
		if err := json.Unmarshal(keyData, &keyFile); err != nil {
			return "", fmt.Errorf("failed to parse service account key: %v", err)
		}

		privateKey, ok := keyFile["private_key"].(string)
		if !ok || privateKey == "" {
			return "", fmt.Errorf("service account key missing private_key field")
		}

		return privateKey, nil
	}

	return "", fmt.Errorf("no private key found - set SERVICE_ACCOUNT_PRIVATE_KEY env var or GOOGLE_APPLICATION_CREDENTIALS")
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
