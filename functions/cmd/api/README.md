# API Service

REST API for Cycle Scene. Handles ride management, submissions, authentication, and group operations.

## Overview

The API service is the central backend for Cycle Scene, providing endpoints for:
- Ride submission and retrieval
- Group management
- Magic link authentication
- Image upload coordination
- Event management

**Technology**: Go, Chi HTTP router, TursoDB
**Port**: 8080
**Deployment**: Google Cloud Run

## Getting Started

### Prerequisites
- Go 1.24+
- TursoDB account with connection credentials
- Google Cloud Storage bucket for image uploads
- Google Cloud Project for deployment

### Local Setup

1. Clone and navigate to the API directory:
```bash
cd functions/cmd/api
```

2. Set up environment variables:
```bash
export TURSO_DB_URL="libsql://your-database-url"
export TURSO_DB_AUTH_TOKEN="your-auth-token"
export GCS_BUCKET="your-gcs-bucket"
export GCP_PROJECT_ID="your-project-id"
```

3. Install dependencies:
```bash
cd ../.. # Navigate to functions root
go mod download
```

4. Run the API:
```bash
cd cmd/api
go run main.go
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Rides

#### GET /api/rides
Retrieve rides with optional filters.

Query parameters:
- `city` - Filter by city code (e.g., "pdx")
- `limit` - Number of results (default: 20)
- `offset` - Pagination offset (default: 0)

Example:
```bash
curl "http://localhost:8080/api/rides?city=pdx&limit=10"
```

#### POST /api/rides
Submit a new ride.

Required fields:
```json
{
  "title": "Sunday Morning Ride",
  "description": "Easy-paced ride through the park",
  "start_time": "2025-05-15T09:00:00Z",
  "end_time": "2025-05-15T11:00:00Z",
  "location": {
    "latitude": 45.5152,
    "longitude": -122.6784,
    "address": "Waterfront Park, Portland"
  },
  "group_id": "shift2bikes"
}
```

Returns submission token for tracking.

#### GET /api/rides/:id
Retrieve a specific ride.

#### PATCH /api/rides/:id
Update a ride (requires authentication).

### Groups

#### GET /api/groups
List all ride organizer groups.

#### POST /api/groups
Create a new group (requires admin authentication).

### Authentication

#### POST /api/auth/magic-link
Request a magic link for authentication.

```json
{
  "email": "user@example.com"
}
```

#### GET /api/auth/verify
Verify a magic link token.

Query parameters:
- `token` - The magic link token

### Image Upload

#### POST /api/images/upload
Upload an image for a ride.

Multipart form data:
- `file` - Image file (JPEG, PNG, WebP)
- `ride_id` - Associated ride ID

Returns:
```json
{
  "image_url": "https://cdn.example.com/images/...",
  "image_id": "..."
}
```

## Configuration

### Environment Variables

Required:
- `TURSO_DB_URL` - TursoDB connection string
- `TURSO_DB_AUTH_TOKEN` - TursoDB authentication token
- `GCS_BUCKET` - Google Cloud Storage bucket name

Optional:
- `PORT` - Server port (default: 8080)
- `LOG_LEVEL` - Logging level (default: info)
- `CORS_ORIGIN` - CORS allowed origins (default: *)

### Database Connection

The API uses a connection pool to TursoDB:
- Max connections: 25
- Connection timeout: 10 seconds
- Query timeout: 30 seconds

## Middleware

### Authentication
Magic link-based authentication for protected endpoints.

Protected endpoints require:
- Authorization header: `Authorization: Bearer {token}`

### CORS
Enabled for all origins by default. Configure via `CORS_ORIGIN` environment variable.

### Logging
All requests are logged with:
- Request method and path
- Response status
- Duration
- Errors (if any)

## Error Handling

The API returns consistent error responses:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "status": 400
}
```

Common error codes:
- `VALIDATION_ERROR` - Invalid request data
- `NOT_FOUND` - Resource not found
- `UNAUTHORIZED` - Authentication required
- `INTERNAL_ERROR` - Server error

## Deployment

### Docker Build

```bash
docker build -t cyclescene-api:latest .
docker tag cyclescene-api:latest gcr.io/{PROJECT_ID}/cyclescene-api:latest
docker push gcr.io/{PROJECT_ID}/cyclescene-api:latest
```

### Cloud Run Deployment

```bash
gcloud run deploy cyclescene-api \
  --image gcr.io/{PROJECT_ID}/cyclescene-api:latest \
  --platform managed \
  --region us-central1 \
  --set-env-vars TURSO_DB_URL={URL},TURSO_DB_AUTH_TOKEN={TOKEN},GCS_BUCKET={BUCKET}
```

Or use the Makefile:
```bash
make deploy
```

## Testing

### Run All Tests
```bash
go test ./...
```

### Test Specific Package
```bash
go test ./internal/api/ride
```

### With Coverage
```bash
go test -cover ./...
```

### E2E Testing
Create rides and verify they appear in API responses:

```bash
# Submit a ride
curl -X POST http://localhost:8080/api/rides \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Ride","start_time":"2025-05-15T09:00:00Z"}'

# Retrieve rides
curl http://localhost:8080/api/rides
```

## Performance

### Optimization Tips
- Use pagination with `limit` and `offset` for large result sets
- Filter by city to reduce data transfer
- Cache ride responses on the frontend (rides don't change frequently)

### Database Queries
Key indexes on:
- `rides(city, start_time)` - For city-based filtering
- `shift2bikes_events(city, event_date)` - For event queries
- `groups(id)` - For group lookups

## Troubleshooting

### Database Connection Errors
```
Error: "cannot connect to database"
```
- Verify `TURSO_DB_URL` is correct
- Check `TURSO_DB_AUTH_TOKEN` is valid
- Ensure TursoDB instance is running

### Image Upload Failures
```
Error: "GCS bucket not accessible"
```
- Verify `GCS_BUCKET` exists
- Check service account has Storage permissions
- Ensure bucket is in the same region as Cloud Run

### Slow Response Times
- Check database query times in logs
- Review Cloud Trace for bottlenecks
- Consider enabling query caching for read-heavy endpoints

## Contributing

When adding new endpoints:
1. Define request/response types in appropriate package
2. Implement handler in `internal/api/`
3. Add middleware as needed (auth, validation)
4. Write tests in `*_test.go` files
5. Update this README with endpoint documentation
6. Add database migrations if needed

## Related Documentation

- [Backend Services README](../README.md)
- [Database Schema](../../db/)
- [Project Architecture](../../README.md)
