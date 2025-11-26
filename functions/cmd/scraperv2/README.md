# Scraper v2

Automated service that scrapes bike events from Shift2Bikes and geocodes their locations for display in Cycle Scene.

## Overview

Scraper v2 runs on a schedule (twice daily) to:
1. Fetch upcoming events from Shift2Bikes website
2. Parse event details (title, date, location, link)
3. Geocode locations to get latitude/longitude
4. Cache geocoding results to avoid redundant API calls
5. Store events in TursoDB for display in the PWA

**Technology**: Go, HTML scraping, Google Geocoding API
**Deployment**: Cloud Run Job (triggered by Cloud Scheduler)
**Schedule**: 8:00 AM and 8:00 PM UTC

## Getting Started

### Prerequisites
- Go 1.24+
- Google Geocoding API key
- TursoDB credentials
- Google Cloud Project

### Local Setup

1. Navigate to scraper directory:
```bash
cd functions/cmd/scraperv2
```

2. Set environment variables:
```bash
export TURSO_DB_URL="libsql://your-database-url"
export TURSO_DB_AUTH_TOKEN="your-auth-token"
export GOOGLE_GEOCODING_API_KEY="your-api-key"
```

3. Run the scraper:
```bash
go run main.go
```

## How It Works

### Data Flow
```
Shift2Bikes Website
        |
        v
HTML Parsing
  (Extract events)
        |
        v
Location Processing
  (Parse addresses)
        |
        v
Geocoding (with caching)
  (Get lat/lng)
        |
        v
Database Storage
  (Save to TursoDB)
```

### Event Parsing

The scraper extracts:
- Event title
- Date and time
- Location/address
- Event link
- Description (if available)

### Geocoding

For each event location:
1. Check if location is in geocache table
2. If cached, use stored lat/lng
3. If not cached, call Google Geocoding API
4. Store result in cache to avoid future API calls

### Error Handling

- Invalid event data is skipped with logging
- Geocoding failures use fallback coordinates (city center)
- Database errors are retried up to 3 times
- All errors are logged for monitoring

## Configuration

### Environment Variables

Required:
- `TURSO_DB_URL` - TursoDB connection string
- `TURSO_DB_AUTH_TOKEN` - TursoDB authentication token
- `GOOGLE_GEOCODING_API_KEY` - Google Geocoding API key

Optional:
- `LOG_LEVEL` - Logging level (default: info)
- `SCRAPE_TIMEOUT` - HTTP timeout in seconds (default: 30)
- `BATCH_SIZE` - Number of events to batch insert (default: 50)

### Google Geocoding API

Setup:
1. Create GCP project if not done
2. Enable Geocoding API: https://console.cloud.google.com/apis/
3. Create API key in Credentials
4. Set `GOOGLE_GEOCODING_API_KEY` environment variable

Free tier: 25,000 requests per day

## Database

### Tables

#### shift2bikes_events
```sql
CREATE TABLE shift2bikes_events (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT,
  start_date DATETIME NOT NULL,
  end_date DATETIME,
  location TEXT,
  latitude REAL,
  longitude REAL,
  event_url TEXT,
  source TEXT DEFAULT 'shift2bikes',
  city TEXT DEFAULT 'pdx',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (city) REFERENCES cities(code)
);
```

#### geocache
Stores geocoding results to avoid repeated API calls:
```sql
CREATE TABLE geocache (
  id TEXT PRIMARY KEY,
  address TEXT NOT NULL,
  latitude REAL,
  longitude REAL,
  cached_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Deployment

### Docker Build
```bash
docker build -t cyclescene-scraper:latest .
docker tag cyclescene-scraper:latest gcr.io/{PROJECT_ID}/cyclescene-scraper:latest
docker push gcr.io/{PROJECT_ID}/cyclescene-scraper:latest
```

### Cloud Run Job Deployment
```bash
gcloud run jobs create cyclescene-scraper \
  --image gcr.io/{PROJECT_ID}/cyclescene-scraper:latest \
  --task-timeout 600s \
  --set-env-vars TURSO_DB_URL={URL},TURSO_DB_AUTH_TOKEN={TOKEN},GOOGLE_GEOCODING_API_KEY={KEY}
```

### Schedule with Cloud Scheduler
```bash
gcloud scheduler jobs create app-engine cyclescene-scraper-schedule \
  --schedule="0 8,20 * * *" \
  --http-method=POST \
  --uri=https://region-project.run.app/scraper/trigger \
  --oidc-service-account-email=runner@project.iam.gserviceaccount.com
```

Or use Makefile:
```bash
make deploy
```

## Monitoring

### View Logs
```bash
gcloud run jobs logs read cyclescene-scraper --limit 50
```

### Key Metrics
- Events scraped: Count of new events found
- Geocoding success rate: % of events successfully geocoded
- Cache hit rate: % of locations found in cache
- Database insert time: Time to store events

### Alerts
Setup alerts for:
- Job execution failures
- Geocoding API quota exceeded
- Database connection errors
- Unusual event count drops

## Troubleshooting

### No Events Found
- Verify Shift2Bikes website is accessible
- Check if website structure has changed (may need scraper updates)
- Review logs: `gcloud run jobs logs read cyclescene-scraper`

### Geocoding Failures
```
Error: "OVER_QUERY_LIMIT"
```
- Google Geocoding API quota exceeded (25k/day limit)
- Consider enabling caching more aggressively
- Wait until next day for quota reset

### Database Connection Issues
```
Error: "cannot connect to database"
```
- Verify `TURSO_DB_URL` and token are correct
- Check TursoDB instance status
- Ensure Cloud Run job has network access

### Slow Geocoding
- Check Google Geocoding API latency
- Verify cache is being used (check `geocache` table size)
- Consider batching more locations

## Data Quality

### Event Validation
All events are validated before insertion:
- Title is not empty
- Start date is in future
- Location is valid
- URL format is correct

### Duplicate Handling
Events are deduplicated by:
- Event URL (primary key in external system)
- Title + date combination

## Maintenance

### Updating Shift2Bikes Scraper

If Shift2Bikes changes their website structure:
1. Inspect the website's HTML
2. Update CSS selectors in `scraper/shift2bikes.go`
3. Test locally with sample HTML
4. Deploy new version
5. Monitor first execution for errors

### Cleaning Old Events

Periodically clean events older than 6 months:
```sql
DELETE FROM shift2bikes_events
WHERE end_date < DATE('now', '-6 months');
```

### Geocache Maintenance

View cache statistics:
```sql
SELECT COUNT(*) as total_cached,
       COUNT(DISTINCT city) as unique_cities
FROM geocache;
```

## Performance

### Optimization Tips
- Geocoding is the bottleneck: ~0.5s per location
- Caching dramatically improves performance
- Batch database inserts in groups of 50+
- Consider regional geocoding if expanding to multiple cities

### Expected Runtime
- 50 events: ~30 seconds (with caching)
- 100 events: ~60 seconds
- 500 events: ~300 seconds

## Contributing

When updating the scraper:
1. Test HTML parsing changes locally first
2. Add unit tests for new parsing logic
3. Update this README if behavior changes
4. Monitor first production run closely
5. Keep error logging clear and actionable

## Related Documentation

- [Backend Services README](../README.md)
- [API Service README](../api/README.md)
- [Project Architecture](../../README.md)
