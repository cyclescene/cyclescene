# Backend Services

Cycle Scene backend is built with Go and consists of several microservices running on Google Cloud Run, plus scheduled jobs for data processing and maintenance.

## Services Overview

### API Service
REST API for ride management, submissions, and authentication.

**Location**: `/functions/cmd/api/`
**Technology**: Go, Chi router, TursoDB
**Deployment**: Cloud Run
**Port**: 8080

[Read API README](cmd/api/README.md)

### Scraper v2
Automated scraper that fetches bike events from Shift2Bikes and geocodes their locations.

**Location**: `/functions/cmd/scraperv2/`
**Technology**: Go, Google Geocoding API
**Deployment**: Cloud Run Job (scheduled via Cloud Scheduler)
**Frequency**: Twice daily

[Read Scraper README](cmd/scraperv2/README.md)

### Image Optimizer
Processes and optimizes images uploaded to rides, converting to WebP format and storing in Google Cloud Storage.

**Location**: `/functions/cmd/image-optimizer/`
**Technology**: Go, WebP conversion
**Deployment**: Cloud Run
**Port**: 8080

[Read Image Optimizer README](cmd/image-optimizer/README.md)

### Token Cleaner
Maintenance job that removes expired submission tokens from the database.

**Location**: `/functions/cmd/token-cleaner/`
**Technology**: Go, TursoDB
**Deployment**: Cloud Run Job (scheduled via Cloud Scheduler)
**Frequency**: Daily

### DB Backups
Creates database dumps and backs them up to Google Cloud Storage.

**Location**: `/functions/cmd/db-backups/`
**Technology**: Go, Google Cloud Storage
**Deployment**: Cloud Run Job (scheduled via Cloud Scheduler)
**Frequency**: Daily

## Getting Started

### Prerequisites
- Go 1.24+
- Docker
- GCP Account with:
  - Cloud Run enabled
  - Cloud Scheduler enabled
  - Google Cloud Storage enabled

### Local Setup

1. Install Go dependencies:
```bash
cd functions
go mod download
```

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Run tests:
```bash
go test ./...
```

4. Build a service locally:
```bash
cd cmd/api
go build -o api .
./api
```

## Project Structure

```
functions/
├── cmd/
│   ├── api/               # REST API service
│   ├── scraperv2/         # Event scraper
│   ├── image-optimizer/   # Image processing
│   ├── token-cleaner/     # Database maintenance
│   └── db-backups/        # Backup utility
├── internal/
│   ├── api/               # API handlers and middleware
│   │   ├── auth/          # Authentication logic
│   │   ├── ride/          # Ride handlers
│   │   ├── group/         # Group handlers
│   │   └── middleware/    # HTTP middleware
│   ├── scraper/           # Scraping utilities
│   └── imageprocessing/   # Image utilities
├── go.mod
└── go.sum
```

## Database

All services use **TursoDB** (SQLite-compatible serverless database).

### Connecting to Database

Services authenticate to TursoDB using:
- `TURSO_DB_URL` - Database connection string
- `TURSO_DB_AUTH_TOKEN` - Authentication token

### Migrations

Database migrations are located in `/db/` and use SQL format.

Run migrations:
```bash
./migrate-backend.sh
```

### Database Schema

Key tables:
- `rides` - User-submitted bike rides
- `shift2bikes_events` - Events scraped from Shift2Bikes
- `groups` - Ride organizer groups
- `submission_tokens` - Time-limited tokens for form submissions
- `geocache` - Cached geocoding results
- `admin_api_keys` - API authentication keys

## Deployment

### Deploy All Services
```bash
make deploy-all
```

### Deploy Individual Service
```bash
cd cmd/{service-name}
make deploy
```

### Infrastructure

Infrastructure is defined in `/infrastructure/` using Terraform/OpenTofu.

Bootstrap infrastructure:
```bash
make bootstrap-infra
```

## Environment Variables

Each service uses different environment variables. See individual service READMEs for specific requirements.

Common variables:
- `TURSO_DB_URL` - TursoDB connection string
- `TURSO_DB_AUTH_TOKEN` - TursoDB authentication
- `GCP_PROJECT_ID` - Google Cloud Project ID
- `GCS_BUCKET` - Google Cloud Storage bucket name

## API Documentation

See [API README](cmd/api/README.md) for detailed API endpoint documentation.

## Troubleshooting

### Database Connection Issues
- Verify `TURSO_DB_URL` and `TURSO_DB_AUTH_TOKEN` are set correctly
- Check TursoDB console for connection status
- Ensure GCP service account has database access

### Image Upload Failures
- Verify `GCS_BUCKET` exists and is accessible
- Check service account has Storage permissions
- Review Cloud Logs for detailed error messages

### Scheduled Jobs Not Running
- Check Cloud Scheduler jobs are enabled
- Verify service account has Cloud Run Invoker role
- Review Cloud Logs for execution status

## Contributing

When adding new services or endpoints:
1. Create a new directory under `/cmd/` for new services
2. Add database migrations to `/db/` if schema changes
3. Update this README with service details
4. Add Terraform configuration in `/infrastructure/`
5. Update root Makefile with deployment commands

## Testing

Run all tests:
```bash
go test ./...
```

Run tests for specific service:
```bash
go test ./cmd/api/...
```

## Logging

Services use structured logging. Configure log levels via environment variables where applicable.

View logs in GCP:
```bash
gcloud functions logs read api --limit 50
```

## Performance

Services are deployed to Cloud Run with:
- 2 vCPU default
- Autoscaling enabled
- Minimum instances: 0 (cold start acceptable for scheduled jobs)

See individual service READMEs for performance tuning details.
