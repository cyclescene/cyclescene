package storage

import "time"

// SignedURLRequest represents a request for a signed URL
type SignedURLRequest struct {
	FileName   string `json:"file_name"`
	FileType   string `json:"file_type"`   // MIME type, e.g., "image/jpeg"
	CityCode   string `json:"city_code"`   // City code (e.g., "pdx", "slc")
	EntityID   string `json:"entity_id"`   // ID of the ride or group
	EntityType string `json:"entity_type"` // "ride" or "group"
}

// SignedURLResponse represents the response containing a signed URL for upload
type SignedURLResponse struct {
	Success      bool      `json:"success"`
	SignedURL    string    `json:"signed_url"`
	ObjectName   string    `json:"object_name"` // Path in bucket (without gs://)
	ImageUUID    string    `json:"image_uuid"`  // UUID of the uploaded image
	ExpiresAt    time.Time `json:"expires_at"`
	BucketName   string    `json:"bucket_name"`
	Error        string    `json:"error,omitempty"`
}

// UploadedFileMetadata contains metadata about an uploaded file
type UploadedFileMetadata struct {
	ImageUUID  string
	ObjectName string
	BucketName string
	SignedURL  string
}
