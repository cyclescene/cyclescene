# Image Optimizer

Service that processes and optimizes images uploaded to rides, converting them to WebP format and storing in Google Cloud Storage.

## Overview

Image Optimizer handles:
1. Image upload validation
2. Format conversion (to WebP)
3. Compression and optimization
4. Storage in Google Cloud Storage
5. Database tracking of processed images

**Technology**: Go, WebP conversion, Google Cloud Storage
**Port**: 8080
**Deployment**: Google Cloud Run

## Getting Started

### Prerequisites
- Go 1.24+
- Google Cloud Storage bucket
- Google Cloud Project

### Local Setup

1. Navigate to directory:
```bash
cd functions/cmd/image-optimizer
```

2. Set environment variables:
```bash
export GCS_BUCKET="your-gcs-bucket"
export GCP_PROJECT_ID="your-project-id"
export TURSO_DB_URL="libsql://your-database-url"
export TURSO_DB_AUTH_TOKEN="your-auth-token"
```

3. Run the service:
```bash
go run main.go
```

Service will be available at `http://localhost:8080`

## API Endpoints

### POST /api/images/upload
Upload and process an image.

Request (multipart/form-data):
- `file` - Image file (JPEG, PNG, WebP)
- `ride_id` - Associated ride ID

Response:
```json
{
  "image_id": "abc123def456",
  "image_url": "https://storage.googleapis.com/bucket/images/abc123def456.webp",
  "size_bytes": 45678,
  "width": 1920,
  "height": 1080
}
```

### GET /api/images/:id
Retrieve image metadata.

Response:
```json
{
  "id": "abc123def456",
  "ride_id": "ride-123",
  "url": "https://storage.googleapis.com/bucket/images/abc123def456.webp",
  "size_bytes": 45678,
  "width": 1920,
  "height": 1080,
  "created_at": "2025-05-15T10:30:00Z"
}
```

## Configuration

### Environment Variables

Required:
- `GCS_BUCKET` - Google Cloud Storage bucket name
- `GCP_PROJECT_ID` - Google Cloud Project ID

Optional:
- `PORT` - Server port (default: 8080)
- `MAX_FILE_SIZE` - Maximum upload size in MB (default: 10)
- `TURSO_DB_URL` - TursoDB connection string
- `TURSO_DB_AUTH_TOKEN` - TursoDB authentication token
- `LOG_LEVEL` - Logging level (default: info)

### Google Cloud Storage

Setup:
1. Create GCS bucket in GCP console
2. Configure CORS if frontend is on different domain:
```json
[
  {
    "origin": ["https://cyclescene.cc"],
    "method": ["GET", "PUT", "POST"],
    "responseHeader": ["Content-Type"],
    "maxAgeSeconds": 3600
  }
]
```

3. Set permissions on service account to write to bucket

## Image Processing

### Supported Formats
Input: JPEG, PNG, WebP, GIF (converts to WebP)
Output: WebP (optimized)

### Optimization

WebP compression settings:
- Quality: 80 (good balance of quality/size)
- Method: 6 (slower but better compression)
- Output: Lossy compression by default

### Size Limits

- Maximum file size: 10 MB (configurable)
- Maximum dimensions: 4000x4000 pixels
- Minimum dimensions: 100x100 pixels

Files exceeding limits are rejected with 400 error.

### Processing Flow

```
Upload File
    |
    v
Validate Format & Size
    |
    v
Decode Image
    |
    v
Resize if needed
    |
    v
Convert to WebP
    |
    v
Upload to GCS
    |
    v
Create Database Record
    |
    v
Return URL & Metadata
```

## Database

### Images Table
```sql
CREATE TABLE images (
  id TEXT PRIMARY KEY,
  ride_id TEXT NOT NULL,
  gcs_path TEXT NOT NULL,
  url TEXT NOT NULL,
  size_bytes INTEGER,
  width INTEGER,
  height INTEGER,
  original_filename TEXT,
  mime_type TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (ride_id) REFERENCES rides(id)
);
```

## Deployment

### Docker Build
```bash
docker build -t cyclescene-image-optimizer:latest .
docker tag cyclescene-image-optimizer:latest gcr.io/{PROJECT_ID}/cyclescene-image-optimizer:latest
docker push gcr.io/{PROJECT_ID}/cyclescene-image-optimizer:latest
```

### Cloud Run Deployment
```bash
gcloud run deploy cyclescene-image-optimizer \
  --image gcr.io/{PROJECT_ID}/cyclescene-image-optimizer:latest \
  --platform managed \
  --region us-central1 \
  --set-env-vars GCS_BUCKET={BUCKET},GCP_PROJECT_ID={PROJECT_ID}
```

Or use Makefile:
```bash
make deploy
```

## Testing

### Upload a Test Image
```bash
curl -X POST http://localhost:8080/api/images/upload \
  -F "file=@test.jpg" \
  -F "ride_id=ride-123"
```

### Retrieve Image Metadata
```bash
curl http://localhost:8080/api/images/abc123def456
```

### Run Tests
```bash
go test ./...
```

## Monitoring

### Key Metrics
- Upload success rate
- Average processing time
- Average file size reduction
- Storage usage growth

### Logs
View recent logs:
```bash
gcloud run logs read cyclescene-image-optimizer --limit 50
```

Look for:
- Upload errors (unsupported format, too large)
- Processing errors (corruption, memory issues)
- Storage errors (permission denied, quota exceeded)

## Troubleshooting

### Upload Fails with "Unsupported Format"
- Verify image is actual image file (not corrupted)
- Check file extension matches content
- Ensure format is JPEG, PNG, WebP, or GIF

### "GCS bucket not accessible" Error
```
Error: "failed to write to GCS"
```
- Verify `GCS_BUCKET` is correct
- Check service account has Storage Object Creator role
- Ensure bucket exists in same region as Cloud Run

### File Size Errors
```
Error: "file exceeds maximum size"
```
- Check max file size setting (default 10MB)
- Consider increasing `MAX_FILE_SIZE` if needed
- Compress image before upload

### Out of Memory Errors
- Reduce maximum image dimensions
- Lower WebP quality setting
- Monitor Cloud Run memory allocation

## Performance Optimization

### Tips
- Use WebP format for input (faster processing)
- Compress images before upload
- Configure reasonable quality/size tradeoff
- Monitor GCS latency

### Caching
Images are permanent in GCS. Consider:
- Setting expiration headers
- Using GCS lifecycle policies for cleanup
- CDN integration for faster delivery

### Batch Processing
Currently processes one image at a time. For bulk uploads:
- Upload multiple images in sequence
- Consider async job queue if needed

## Security

### Validation
- File type validated by magic bytes (not just extension)
- File size checked before processing
- Image dimensions validated
- Database records created for audit trail

### Storage
- Files stored with unique IDs (not user-controllable)
- Original filenames not exposed in URLs
- GCS permissions restrict public access

### API
- Ride ID must be valid (exists in database)
- Consider adding authentication for production

## Contributing

When modifying image processing:
1. Test with various image formats
2. Monitor output quality vs file size
3. Update this README if behavior changes
4. Consider backward compatibility

## Related Documentation

- [Backend Services README](../README.md)
- [API Service README](../api/README.md)
- [Project Architecture](../../README.md)
