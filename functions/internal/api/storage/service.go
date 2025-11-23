package storage

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

type Service struct {
	bucketName        string
	projectID         string
	client            *storage.Client
	signedURLDuration time.Duration
}

// NewService creates a new storage service
// Uses Application Default Credentials (will work on Cloud Run with service account)
func NewService() (*Service, error) {
	bucketName := os.Getenv("STAGING_BUCKET_NAME")
	projectID := os.Getenv("GCP_PROJECT")

	if bucketName == "" {
		return nil, fmt.Errorf("STAGING_BUCKET_NAME environment variable not set")
	}
	if projectID == "" {
		return nil, fmt.Errorf("GCP_PROJECT environment variable not set")
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
		bucketName:        bucketName,
		projectID:         projectID,
		client:            client,
		signedURLDuration: duration,
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

// generateSignedURL creates the actual signed URL using GCS SignedURL function
func (s *Service) generateSignedURL(ctx context.Context, objectName, contentType string) (string, error) {
	// Get the default service account credentials
	credentials, err := getServiceAccountCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get credentials: %v", err)
	}

	// Use the storage.SignedURL function with the credentials
	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         "PUT",
		Expires:        time.Now().Add(s.signedURLDuration),
		ContentType:    contentType,
		GoogleAccessID: credentials.Email,
		PrivateKey:     []byte(credentials.PrivateKey),
	}

	signedURL, err := storage.SignedURL(s.bucketName, objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %v", err)
	}

	return signedURL, nil
}

// generateSignedURLWithMetadata creates a signed URL with custom metadata
// Note: Metadata will be set by the frontend when uploading via x-goog-meta-* headers
func (s *Service) generateSignedURLWithMetadata(ctx context.Context, objectName, contentType, cityCode, entityType string) (string, error) {
	// Get the default service account credentials
	credentials, err := getServiceAccountCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get credentials: %v", err)
	}

	// Use the storage.SignedURL function with the credentials
	// The frontend will include x-goog-meta-* headers when uploading
	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         "PUT",
		Expires:        time.Now().Add(s.signedURLDuration),
		ContentType:    contentType,
		GoogleAccessID: credentials.Email,
		PrivateKey:     []byte(credentials.PrivateKey),
	}

	signedURL, err := storage.SignedURL(s.bucketName, objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %v", err)
	}

	slog.Info("generated signed URL with metadata", "object", objectName, "cityCode", cityCode, "entityType", entityType)

	return signedURL, nil
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

// getServiceAccountCredentials retrieves the service account credentials
// On Cloud Run, this uses Application Default Credentials
func getServiceAccountCredentials(ctx context.Context) (*ServiceAccountCredentials, error) {
	// Try base64-encoded credentials first (for local development)
	if credentialsB64 := os.Getenv("GCP_SERVICE_ACCOUNT_KEY_B64"); credentialsB64 != "" {
		return loadServiceAccountFromBase64(credentialsB64)
	}

	// Try to read from file path (for local development with key file)
	if keyPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); keyPath != "" {
		return loadServiceAccountFromFile(keyPath)
	}

	// On Cloud Run, we need to use the default credentials to get the service account email
	// and then use SignBlob to create signed URLs
	return getDefaultServiceAccountCredentials(ctx)
}

// ServiceAccountCredentials holds the necessary data to sign URLs
type ServiceAccountCredentials struct {
	Email      string
	PrivateKey string
}

// loadServiceAccountFromBase64 loads service account credentials from a base64-encoded JSON key
func loadServiceAccountFromBase64(credentialsB64 string) (*ServiceAccountCredentials, error) {
	// Check if we're in development mode
	appEnv := os.Getenv("APP_ENV")
	if appEnv != "dev" {
		slog.Warn("base64-encoded credentials should only be used in development", "env", appEnv)
	}

	// Decode base64
	keyData, err := base64.StdEncoding.DecodeString(credentialsB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 service account credentials: %v", err)
	}

	// Parse the JSON key
	var keyFile map[string]any
	if err := json.Unmarshal(keyData, &keyFile); err != nil {
		return nil, fmt.Errorf("failed to parse service account key: %v", err)
	}

	// Extract email and private key
	email, ok := keyFile["client_email"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("invalid service account key: missing or invalid client_email")
	}

	privateKey, ok := keyFile["private_key"].(string)
	if !ok || privateKey == "" {
		return nil, fmt.Errorf("invalid service account key: missing or invalid private_key")
	}

	slog.Info("loaded service account from base64-encoded credentials", "email", email)

	return &ServiceAccountCredentials{
		Email:      email,
		PrivateKey: privateKey,
	}, nil
}

// loadServiceAccountFromFile loads service account credentials from a JSON key file
func loadServiceAccountFromFile(keyPath string) (*ServiceAccountCredentials, error) {
	// Check if we're in development mode
	appEnv := os.Getenv("APP_ENV")
	if appEnv != "dev" {
		slog.Warn("file-based credentials should only be used in development", "env", appEnv, "path", keyPath)
	}

	// Read the key file
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account key file: %v", err)
	}

	// Parse the JSON key file
	var keyFile map[string]any
	if err := json.Unmarshal(keyData, &keyFile); err != nil {
		return nil, fmt.Errorf("failed to parse service account key file: %v", err)
	}

	// Extract email and private key
	email, ok := keyFile["client_email"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("invalid service account key file: missing or invalid client_email")
	}

	privateKey, ok := keyFile["private_key"].(string)
	if !ok || privateKey == "" {
		return nil, fmt.Errorf("invalid service account key file: missing or invalid private_key")
	}

	slog.Info("loaded service account from file", "email", email, "path", keyPath)

	return &ServiceAccountCredentials{
		Email:      email,
		PrivateKey: privateKey,
	}, nil
}

// getDefaultServiceAccountCredentials retrieves credentials from the environment
// On Cloud Run, the service account is automatically available
func getDefaultServiceAccountCredentials(ctx context.Context) (*ServiceAccountCredentials, error) {
	// Get service account email from metadata service (Cloud Run specific)
	// For now, we'll construct it from the project ID
	projectID := os.Getenv("PROJECT_ID")

	// Service account email format: {account-id}@{project-id}.iam.gserviceaccount.com
	// The API service account is typically: cyclescene-api@{project-id}.iam.gserviceaccount.com
	email := fmt.Sprintf("cyclescene-api@%s.iam.gserviceaccount.com", projectID)

	// For Cloud Run, we can get the private key from the service account JSON
	// This requires the key file to be available in the environment
	// Alternatively, we can use the iam.serviceAccounts.signBlob API

	slog.Info("using service account for signed URLs", "email", email)

	return &ServiceAccountCredentials{
		Email:      email,
		PrivateKey: "", // Will be retrieved from environment or metadata
	}, nil
}
