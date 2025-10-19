# Image Optimizer Implementation

This document describes the image optimizer service and how it integrates with your CycleScene API.

## Architecture Overview

```
┌─────────────────────────────────────────┐
│  API (cyclescene-api-gateway)           │
│                                         │
│  - Accepts image uploads → staging      │
│  - Creates ride/group                   │
│  - Calls optimizer service              │
└──────────────┬──────────────────────────┘
               │
               ├─── HTTP POST /optimize ───────┐
               │                               │
        ┌──────▼──────────────────────────┐    │
        │  Image Optimizer Service         │◄───┘
        │  (Cloud Run)                     │
        │                                  │
        │  Processes & optimizes images    │
        │  Updates DB with image_url       │
        │                                  │
        └──────┬──────────────┬────────────┘
               │              │
        ┌──────▼──┐    ┌──────▼──────────────┐
        │ Staging │    │ Optimized Bucket    │
        │ Bucket  │    │ (Final Images)      │
        └─────────┘    └─────────────────────┘
```

## Directory Structure

### API Client (`internal/api/imageoptimizer/`)
Used by the API service to call the image optimizer.

```
internal/api/imageoptimizer/
├── models.go          # OptimizeRequest, OptimizeResponse
└── service.go         # Service with Optimize() and OptimizeAsync()
```

**Usage in API:**
```go
import "github.com/spacesedan/cyclescene/functions/internal/api/imageoptimizer"

// After creating a ride/group:
optimizerService := imageoptimizer.NewService()
optimizerService.OptimizeAsync(&imageoptimizer.OptimizeRequest{
    ImageUUID:  submission.ImageUUID,
    CityCode:   submission.City,
    EntityID:   fmt.Sprintf("%d", eventID),
    EntityType: "ride", // or "group"
})
```

### Image Processing (`internal/imageprocessing/`)
Core image processing logic used by the optimizer service.

```
internal/imageprocessing/
├── processor.go       # ImageProcessor - handles GCS and image optimization
└── db.go             # DBConnector - manages database updates
```

**Responsibilities:**
- `processor.go`: Download from GCS, optimize image, upload to final bucket
- `db.go`: Update `image_url` field in rides/groups table

### Optimizer Service (`cmd/image-optimizer/`)
Cloud Run service that processes images.

```
cmd/image-optimizer/
├── main.go            # Server setup, logger, database connection
├── routes.go          # HTTP handlers for /health and /optimize
├── Dockerfile         # Multi-stage build for Cloud Run
└── infra/             # Terraform deployment
    ├── main.tf        # Service account, buckets, Cloud Run service
    ├── variables.tf   # Configuration variables
    ├── outputs.tf     # Service URL and bucket names
    └── terraform.tfvars.example
```

## Data Flow

### 1. Image Upload (User → API)
```
User uploads image
  ↓
API generates UUID: abc123def456
  ↓
Image saved to GCS: gs://staging-bucket/abc123def456.jpg
  ↓
API stores image_uuid in submission form
```

### 2. Ride/Group Creation (API)
```
User submits ride form with image_uuid: "abc123def456"
  ↓
API creates ride record in DB
  ↓
API gets rideID from DB insert: 12345
```

### 3. Image Optimization (API → Optimizer Service)
```
API calls image optimizer:
POST https://optimizer-service.run.app/optimize
{
  "imageUUID": "abc123def456",
  "cityCode": "PDX",
  "entityID": "12345",
  "entityType": "ride"
}
  ↓
Optimizer processes request asynchronously
```

### 4. Image Processing (Optimizer Service)
```
Download image from staging:
  gs://staging-bucket/abc123def456.jpg
  ↓
Optimize (compress + convert):
  - Decode image
  - Compress with JPEG quality 85%
  - Output as WebP format
  ↓
Upload to final bucket:
  gs://optimized-bucket/PDX/rides/12345/12345_optimized.webp
  ↓
Update database:
  UPDATE rides SET image_url = "https://storage.googleapis.com/..."
  WHERE id = 12345
  ↓
Delete staging file:
  gs://staging-bucket/abc123def456.jpg
```

## File Paths

### Staging Bucket
```
gs://{project-id}-user-media-staging/
├── {imageUUID}.jpg        # User uploads any JPEG/PNG
└── ...
```

### Optimized Bucket
```
gs://{project-id}-user-media-optimized/
├── {cityCode}/
│   ├── rides/
│   │   ├── {rideID}/
│   │   │   └── {rideID}_optimized.webp
│   │   └── ...
│   └── groups/
│       ├── {groupID}/
│       │   └── {groupID}_optimized.webp
│       └── ...
└── ...
```

**Examples:**
- Ride: `gs://project-user-media-optimized/PDX/rides/12345/12345_optimized.webp`
- Group: `gs://project-user-media-optimized/NYC/groups/67890/67890_optimized.webp`

**Public URL:**
```
https://storage.googleapis.com/{project-id}-user-media-optimized/PDX/rides/12345/12345_optimized.webp
```

## Database Schema

Your rides and groups tables need these columns:

```sql
ALTER TABLE rides ADD COLUMN IF NOT EXISTS image_uuid VARCHAR(255);
ALTER TABLE rides ADD COLUMN IF NOT EXISTS image_url TEXT;

ALTER TABLE groups ADD COLUMN IF NOT EXISTS image_uuid VARCHAR(255);
ALTER TABLE groups ADD COLUMN IF NOT EXISTS image_url TEXT;
```

## Integration with API

### Step 1: Update Submission Models
Add `image_uuid` field:

```go
type Submission struct {
    ImageUUID  string `json:"image_uuid"`
    Title      string `json:"title"`
    // ... other fields ...
}
```

### Step 2: Call Optimizer After Create
In `internal/api/ride/service.go`:

```go
func (s *Service) SubmitRide(submission *Submission) (*SubmissionResponse, error) {
    // ... existing code ...

    eventID, err := s.repo.CreateRide(submission, editToken, lat, lng)
    if err != nil {
        return nil, err
    }

    // Trigger image optimization asynchronously
    if submission.ImageUUID != "" {
        optimizerService := imageoptimizer.NewService()
        optimizerService.OptimizeAsync(&imageoptimizer.OptimizeRequest{
            ImageUUID:  submission.ImageUUID,
            CityCode:   submission.City,
            EntityID:   fmt.Sprintf("%d", eventID),
            EntityType: "ride",
        })
    }

    return &SubmissionResponse{
        Success:   true,
        EventID:   eventID,
        EditToken: editToken,
        Message:   "Ride submitted successfully",
    }, nil
}
```

Do the same for groups in `internal/api/group/service.go`.

## Deployment

### 1. Deploy Image Optimizer Infrastructure
```bash
cd functions/cmd/image-optimizer/infra

# Create terraform.tfvars from example
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values

terraform init
terraform apply
```

### 2. Get Outputs
```bash
terraform output optimizer_service_url
terraform output optimized_bucket_name
```

### 3. Add to API Environment
Add `IMAGE_OPTIMIZER_URL` to your API deployment:

```
IMAGE_OPTIMIZER_URL=https://cyclescene-image-optimizer-{region}-run.app
```

### 4. Update API Deployment
Deploy your updated API code with the optimizer integration.

## Configuration

### Environment Variables

**Image Optimizer Service:**
- `PORT` (default: 8080)
- `APP_ENV` (dev/prod)
- `GCP_PROJECT` - GCP project ID
- `STAGING_BUCKET` - Staging bucket name
- `OPTIMIZED_BUCKET` - Optimized bucket name
- `TURSO_DB_URL` - Database URL
- `TURSO_DB_RW_TOKEN` - Database token

**API Service:**
- `IMAGE_OPTIMIZER_URL` - Optimizer service URL (e.g., https://optimizer.run.app)

### Terraform Variables

```hcl
project_id                = "my-project"
region                    = "us-west1"
staging_bucket_name       = "my-project-user-media-staging"
api_service_account_email = "cyclescene-api@my-project.iam.gserviceaccount.com"
optimizer_cpu_limit       = "2"           # 2 CPUs
optimizer_memory_limit    = "2Gi"         # 2GB RAM
optimizer_max_instances   = 10            # Max 10 concurrent
turso_db_url              = "libsql://..."
turso_db_rw_token         = "..."
```

## Error Handling

### Synchronous Mode
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := optimizerService.Optimize(ctx, &imageoptimizer.OptimizeRequest{...})
if err != nil {
    // Handle error - optimization failed
    slog.Error("optimization failed", "error", err)
}
```

### Asynchronous Mode (Recommended)
```go
optimizerService.OptimizeAsync(&imageoptimizer.OptimizeRequest{...})
// Function returns immediately, optimization happens in background
// Errors are logged but don't affect ride/group creation
```

## Monitoring

### Cloud Run Logs
```bash
gcloud run logs read cyclescene-image-optimizer --limit 100 --region us-west1
```

### Check Image Processing
```bash
# Look for successful optimizations
gcloud run logs read cyclescene-image-optimizer --limit 50 | grep "successfully optimized"

# Look for failures
gcloud run logs read cyclescene-image-optimizer --limit 50 | grep "ERROR"
```

### Database
```sql
-- Check if images are being optimized
SELECT id, image_uuid, image_url FROM rides WHERE image_url IS NOT NULL LIMIT 10;

-- Check pending images (no image_url yet)
SELECT id, image_uuid FROM rides WHERE image_uuid IS NOT NULL AND image_url IS NULL;
```

### Storage
```bash
# List optimized images
gsutil ls -r gs://{project-id}-user-media-optimized/PDX/

# Check file sizes
gsutil du -s gs://{project-id}-user-media-optimized/
```

## Testing Locally

### Start Image Optimizer Service
```bash
cd functions/cmd/image-optimizer

# Set environment variables
export PORT=8080
export APP_ENV=dev
export GCP_PROJECT=my-project
export STAGING_BUCKET=my-project-user-media-staging
export OPTIMIZED_BUCKET=my-project-user-media-optimized
export TURSO_DB_URL=libsql://...
export TURSO_DB_RW_TOKEN=...

go run .
```

### Test Health Endpoint
```bash
curl http://localhost:8080/health
```

### Test Optimization Endpoint
```bash
curl -X POST http://localhost:8080/optimize \
  -H "Content-Type: application/json" \
  -d '{
    "imageUUID": "test-uuid-123",
    "cityCode": "PDX",
    "entityID": "test-ride-1",
    "entityType": "ride"
  }'
```

## Performance Considerations

- **Image size limit**: Default 10MB (can be increased via Cloud Run config)
- **Processing timeout**: 10 minutes
- **Memory**: 2GB (configurable via Terraform)
- **CPU**: 2 CPUs (configurable via Terraform)
- **Concurrent instances**: Max 10 (configurable via Terraform)

## Future Improvements

1. **True WebP conversion**: Add `bimg` or `go-webp` library for better compression
2. **Multiple sizes**: Generate thumbnail, small, medium, large variants
3. **Progressive optimization**: Start with fast compression, improve over time
4. **Image validation**: Detect and reject invalid/malicious images
5. **Retry logic**: Retry failed optimizations with exponential backoff
6. **Metrics**: Track optimization success rates, file sizes, processing times
