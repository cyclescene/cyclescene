# Signed URL Upload Guide

This guide explains how to use the signed URL feature to upload images directly from the browser to Google Cloud Storage.

## Overview

The API now provides a signed URL endpoint that allows your frontend to:
1. Request a signed URL from the API
2. Upload files directly to the staging bucket using the signed URL
3. Get a UUID for the uploaded file to include in form submissions

This eliminates the need for files to pass through your API server, reducing bandwidth and latency.

## Architecture

```
┌─────────────────────────────────────┐
│  Frontend (PWA)                     │
│                                     │
│  1. Request signed URL              │
│  2. Upload file directly to GCS     │
│  3. Submit form with image_uuid     │
└────────┬────────────────────────────┘
         │
         ├─── POST /v1/storage/upload-url ────────┐
         │                                        │
    ┌────▼────────────────────────────────┐       │
    │  API Gateway                        │◄──────┘
    │  (Generates signed URL)             │
    │                                     │
    └─────────────────────────────────────┘
                    │
                    └─── Returns signed URL ─────┐
                                                 │
    ┌────────────────────────────────────────────▼─┐
    │  Google Cloud Storage                        │
    │  Staging Bucket                              │
    │                                              │
    │  (Frontend uploads file directly using URL) │
    └──────────────────────────────────────────────┘
```

## Frontend Implementation

### Step 1: Request a Signed URL

```javascript
async function getSignedUploadURL(fileName, fileType) {
  const response = await fetch('/v1/storage/upload-url', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-BFF-Token': bffToken, // Your auth token
    },
    body: JSON.stringify({
      file_name: fileName,
      file_type: fileType, // e.g., "image/jpeg"
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to get signed URL: ${response.statusText}`);
  }

  return await response.json();
}
```

### Step 2: Upload File Using Signed URL

```javascript
async function uploadImageFile(file) {
  // Get signed URL from API
  const urlResponse = await getSignedUploadURL(file.name, file.type);

  if (!urlResponse.success) {
    throw new Error(`Failed to generate signed URL: ${urlResponse.error}`);
  }

  const { signed_url, image_uuid, expires_at } = urlResponse;

  // Upload file directly to GCS using the signed URL
  const uploadResponse = await fetch(signed_url, {
    method: 'PUT',
    headers: {
      'Content-Type': file.type,
    },
    body: file,
  });

  if (!uploadResponse.ok) {
    throw new Error(`Failed to upload file: ${uploadResponse.statusText}`);
  }

  console.log(`File uploaded successfully!`);
  console.log(`Image UUID: ${image_uuid}`);
  console.log(`Expires at: ${expires_at}`);

  return {
    imageUUID: image_uuid,
    bucketName: urlResponse.bucket_name,
  };
}
```

### Step 3: Submit Form with Image UUID

After uploading the image, include the `image_uuid` in your ride/group submission:

```javascript
async function submitRide(rideData, imageUUID) {
  const response = await fetch('/v1/rides/submit', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-BFF-Token': bffToken,
    },
    body: JSON.stringify({
      ...rideData,
      image_uuid: imageUUID, // UUID from upload
      title: 'My Ride',
      description: 'Details about the ride',
      city: 'PDX',
      // ... other fields
    }),
  });

  return await response.json();
}
```

### Complete Example

```javascript
async function handleImageUpload(event) {
  const file = event.target.files[0];

  try {
    // Step 1: Get signed URL
    const urlResponse = await fetch('/v1/storage/upload-url', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-BFF-Token': bffToken,
      },
      body: JSON.stringify({
        file_name: file.name,
        file_type: file.type,
      }),
    });

    if (!urlResponse.ok) throw new Error('Failed to get signed URL');
    const urlData = await urlResponse.json();

    // Step 2: Upload file to GCS
    const uploadResponse = await fetch(urlData.signed_url, {
      method: 'PUT',
      headers: { 'Content-Type': file.type },
      body: file,
    });

    if (!uploadResponse.ok) throw new Error('Failed to upload file');

    // Step 3: Store UUID for form submission
    const imageUUID = urlData.image_uuid;
    document.getElementById('image_uuid_input').value = imageUUID;
    document.getElementById('image_preview').src = URL.createObjectURL(file);

    console.log('Image uploaded! UUID:', imageUUID);
  } catch (error) {
    console.error('Error:', error);
    alert('Failed to upload image');
  }
}
```

## API Endpoint Reference

### Generate Signed Upload URL

**Endpoint:** `POST /v1/storage/upload-url`

**Headers:**
```
Content-Type: application/json
X-BFF-Token: {your_auth_token}
```

**Request Body:**
```json
{
  "file_name": "my-photo.jpg",
  "file_type": "image/jpeg"
}
```

**Allowed MIME Types:**
- `image/jpeg`
- `image/png`
- `image/webp`
- `image/gif`

**Response (Success):**
```json
{
  "success": true,
  "signed_url": "https://storage.googleapis.com/bucket-name/...",
  "object_name": "{uuid}.jpg",
  "image_uuid": "{uuid}",
  "expires_at": "2024-10-18T18:15:00Z",
  "bucket_name": "project-user-media-staging"
}
```

**Response (Error):**
```json
{
  "success": false,
  "error": "file_type and file_name are required"
}
```

## Security Considerations

1. **Signed URLs are time-limited**: Default 15 minutes, configurable via `SIGNED_URL_DURATION_MINUTES`
2. **Only image files allowed**: MIME type validation on backend
3. **CORS enabled**: Browser uploads work from allowed origins
4. **Service account permissions**: API service account needs `roles/iam.serviceAccountTokenCreator`
5. **File size limits**: GCS enforces limits (default 5TB, but can be configured)

## Practical Workflow

1. **User selects image** → Frontend requests signed URL
2. **API validates request** → Generates UUID and signed URL
3. **Frontend uploads directly to GCS** → No server-side processing needed
4. **Image waits in staging** → Ready for optimization
5. **User fills form and includes UUID** → Submits to API
6. **API creates ride/group** → References image_uuid
7. **Image optimizer processes** → Moves to optimized bucket, updates DB

## Database Schema Update

Add these columns to your rides and groups tables:

```sql
ALTER TABLE rides ADD COLUMN IF NOT EXISTS image_uuid VARCHAR(255);
ALTER TABLE rides ADD COLUMN IF NOT EXISTS image_url TEXT;

ALTER TABLE groups ADD COLUMN IF NOT EXISTS image_uuid VARCHAR(255);
ALTER TABLE groups ADD COLUMN IF NOT EXISTS image_url TEXT;
```

Then update your submission models to include:

```go
type Submission struct {
    ImageUUID  string `json:"image_uuid"`  // New field
    ImageURL   string `json:"image_url"`   // Populated by optimizer
    Title      string `json:"title"`
    // ... other fields ...
}
```

## Configuration

### Environment Variables

The API automatically uses:
- `MEDIA_BUCKET` - Staging bucket name (set in Terraform)
- `GCP_PROJECT` - GCP project ID (set in Terraform)
- `SIGNED_URL_DURATION_MINUTES` - Optional, defaults to 15

### Terraform Configuration

The staging bucket and service account permissions are already configured in your API Terraform:

```hcl
module "api_service_account" {
  # ...
  roles = [
    "roles/storage.objectCreator",
    "roles/storage.objectViewer",
    "roles/iam.serviceAccountTokenCreator"  # Required for signed URLs
  ]
}

module "user_media_bucket" {
  # ...
  cors_rules = [
    {
      origin          = var.allowed_origins  # Browser origins
      method          = ["GET", "POST", "PUT", "HEAD"]
      response_header = ["Content-Type", "Content-Length"]
      max_age_seconds = 3600
    }
  ]
}
```

## Troubleshooting

### "Failed to get signed URL"
- Check that `MEDIA_BUCKET` environment variable is set
- Verify service account has `roles/iam.serviceAccountTokenCreator`
- Check API logs for detailed error

### CORS errors when uploading
- Verify your frontend origin is in `allowed_origins` in Terraform
- Check GCS CORS configuration with: `gsutil cors get gs://your-bucket`

### "Invalid file type"
- Only JPEG, PNG, WebP, and GIF are allowed
- Verify `file_type` MIME type in request is exactly one of:
  - `image/jpeg`
  - `image/png`
  - `image/webp`
  - `image/gif`

### Signed URL expires before upload completes
- Increase `SIGNED_URL_DURATION_MINUTES` for slower connections
- Default 15 minutes should be sufficient for most uploads

## Monitoring

### Check uploaded files
```bash
gsutil ls -r gs://your-project-user-media-staging/

# Count files
gsutil ls -r gs://your-project-user-media-staging/ | wc -l

# Check file sizes
gsutil du -sh gs://your-project-user-media-staging/
```

### View API logs
```bash
gcloud run logs read cyclescene-api-gateway --limit 50 | grep "signed URL"
```

### Verify signed URL functionality
```bash
# Generate a signed URL via the API
curl -X POST https://your-api.run.app/v1/storage/upload-url \
  -H "Content-Type: application/json" \
  -H "X-BFF-Token: your-token" \
  -d '{"file_name": "test.jpg", "file_type": "image/jpeg"}'

# Should return a signed_url - test uploading a file to it
curl -X PUT "https://storage.googleapis.com/..." \
  -H "Content-Type: image/jpeg" \
  --data-binary @test.jpg
```

## Performance Tips

1. **Compress images before upload**: Reduce file size before sending
2. **Show upload progress**: Use fetch with progress events
3. **Retry on failure**: Implement exponential backoff for retries
4. **Batch uploads**: Process multiple images in parallel (GCS handles it well)
5. **Clean up old files**: Implement a lifecycle rule to delete old staging files

## Next Steps

After setting up signed URLs:
1. Update frontend to use the signed URL endpoint
2. Update form models to include `image_uuid`
3. Deploy API changes
4. Test with actual file uploads
5. Monitor staging bucket for orphaned files
6. Set up cleanup policy for old images in staging bucket
