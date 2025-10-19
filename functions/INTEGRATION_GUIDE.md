# Image Optimizer Integration Guide

This guide explains how to integrate the image optimizer service into your API.

## Overview

The image optimizer processes images after rides/groups are created. The flow is:

1. User uploads image → Staging bucket with UUID (e.g., `{uuid}.jpg`)
2. User submits ride/group form → API creates record with `image_uuid`
3. API calls image optimizer → Passes metadata
4. Optimizer processes image → Saves to optimized bucket with final path
5. Optimizer updates database → Sets `image_url` field

## Integration Points

### 1. Image Submission (GCS Metadata)

When users upload images to the staging bucket, add custom GCS metadata:

```go
ctx := context.Background()
client, _ := storage.NewClient(ctx)

// Set custom metadata when uploading to staging bucket
objectHandle := client.Bucket("staging-bucket").Object("image-uuid.jpg")
objectAttrs := &storage.ObjectAttrsToUpdate{
    Metadata: map[string]string{
        "city-code":  "PDX",
        "entity-id":  "ride-123",
        "entity-type": "ride", // or "group"
    },
}
objectHandle.Update(ctx, objectAttrs)
```

### 2. After Creating Ride/Group Record

Update the ride/group service to call the optimizer after successful creation:

```go
// In ride/service.go SubmitRide method
import "github.com/spacesedan/cyclescene/functions/internal/api/imageoptimizer"

func (s *Service) SubmitRide(submission *Submission) (*SubmissionResponse, error) {
    // ... existing code to create ride ...

    eventID, err := s.repo.CreateRide(submission, editToken, lat, lng)
    if err != nil {
        return nil, err
    }

    // Trigger image optimization asynchronously if image UUID provided
    if submission.ImageUUID != "" {
        optimizerClient := imageoptimizer.NewClient()
        optimizerClient.OptimizeAsync(&imageoptimizer.OptimizeRequest{
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
        Message:   "Ride submitted successfully and is pending review",
    }, nil
}
```

### 3. Required Environment Variables

Add these to your Cloud Run Service environment:

```
IMAGE_OPTIMIZER_URL=https://cyclescene-image-optimizer-{region}-run.app
STAGING_BUCKET={project-id}-user-media-staging
OPTIMIZED_BUCKET={project-id}-user-media-optimized
GCP_PROJECT={project-id}
```

## Implementation Options

### Option A: Asynchronous (Recommended)

Use `optimizerClient.OptimizeAsync()` - fire and forget:
- User gets response immediately
- Image optimization happens in the background
- No impact on API response time
- Failures are logged but don't break ride/group creation

```go
optimizerClient.OptimizeAsync(&imageoptimizer.OptimizeRequest{...})
```

### Option B: Synchronous

Use `optimizerClient.Optimize()` with context:
- API waits for optimization to complete
- User gets image URL in response (if provided)
- Slower response times
- Failures prevent ride creation

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := optimizerClient.Optimize(ctx, &imageoptimizer.OptimizeRequest{
    ImageUUID:  submission.ImageUUID,
    CityCode:   submission.City,
    EntityID:   fmt.Sprintf("%d", eventID),
    EntityType: "ride",
})

if err != nil {
    slog.Warn("image optimization failed but ride was created", "error", err)
    // Continue anyway - the image will still process
}
```

## Database Schema Requirements

Your rides and groups tables should have:

```sql
ALTER TABLE rides ADD COLUMN image_uuid VARCHAR(255);
ALTER TABLE rides ADD COLUMN image_url TEXT;

ALTER TABLE groups ADD COLUMN image_uuid VARCHAR(255);
ALTER TABLE groups ADD COLUMN image_url TEXT;
```

## Output Path Format

Optimized images are stored at:

```
gs://{project-id}-user-media-optimized/{cityCode}/{entityType}s/{entityID}/{entityID}_optimized.webp
```

Example:
- Ride: `gs://project-user-media-optimized/PDX/rides/ride-123/ride-123_optimized.webp`
- Group: `gs://project-user-media-optimized/NYC/groups/group-456/group-456_optimized.webp`

Public URL:
```
https://storage.googleapis.com/{project-id}-user-media-optimized/{cityCode}/{entityType}s/{entityID}/{entityID}_optimized.webp
```

## Deployment Steps

### 1. Update Submission Model

Add `ImageUUID` field to the ride/group Submission model:

```go
type Submission struct {
    ImageUUID  string `json:"image_uuid"` // UUID of uploaded image
    Title      string `json:"title"`
    // ... other fields ...
}
```

### 2. Deploy Image Optimizer Infrastructure

```bash
cd functions/cmd/image-optimizer/infra
terraform apply -var-file terraform.tfvars
```

### 3. Capture Terraform Outputs

Save these values from terraform outputs:

```
optimizer_service_url = "https://cyclescene-image-optimizer-{region}-run.app"
optimized_bucket_name = "{project-id}-user-media-optimized"
```

### 4. Update API Deployment

Add to your API terraform/deployment:

```
IMAGE_OPTIMIZER_URL="https://cyclescene-image-optimizer-{region}-run.app"
```

### 5. Update Ride/Group Services

Integrate the optimizer client as shown in "Integration Points" above.

### 6. Deploy Updated API

```bash
cd functions/cmd/api/infra
terraform apply
```

## Error Handling

The optimizer client includes built-in error handling:

- Network failures are retried via context timeout
- Optimizer service failures are logged but don't break ride creation (async mode)
- Invalid requests return detailed error messages
- Database update failures in optimizer are logged separately

## Testing Locally

```bash
# Start optimizer service
cd functions/cmd/image-optimizer
go run .

# In another terminal, test the optimizer
curl -X POST http://localhost:8080/optimize \
  -H "Content-Type: application/json" \
  -d '{
    "imageUUID": "test-uuid-123",
    "cityCode": "PDX",
    "entityID": "test-ride-1",
    "entityType": "ride"
  }'

# Check health
curl http://localhost:8080/health
```

## Monitoring

Monitor the image optimizer service:

1. **Cloud Run logs**: View optimizer processing logs in Cloud Run console
2. **Errors in API**: Search for "image optimization failed" in API logs
3. **Database**: Check if `image_url` fields are being populated
4. **Optimized bucket**: Verify images are appearing in the correct paths

## Notes

- Images are automatically deleted from staging bucket after successful optimization
- If optimization fails, staging image remains and can be retried
- WebP conversion uses JPEG compression algorithm (requires additional library for true WebP)
- Maximum image processing time: 10 minutes (configurable in Cloud Run)
- Service scales to 0 when idle, no cost when not processing
