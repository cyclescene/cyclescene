# Strava Background Sync Service

**Status:** Production Ready
**Schedule:** Every 3 days at 2am PST (`0 9 */3 * *` UTC)
**Runtime:** Cloud Run Job triggered by Cloud Scheduler

---

## Overview

GCP Cloud Run Job that automatically refreshes Strava event data to maintain 7-day API compliance. The service syncs **existing imported events only** - it does not auto-import new events.

### Key Behaviors

- **Sync existing events** - Refreshes events organizers chose to import
- **Delete stale events** - Removes events deleted on Strava
- **Detach on edit** - Events edited via CycleScene become native and stop syncing
- **Error isolation** - Failures for one athlete don't affect others
- **Rate limit aware** - Conservative limits (90/15min, 900/day)

---

## Architecture

```
Cloud Scheduler (2am PST every 3 days)
         │
         ▼
Cloud Run Job (strava-sync)
         │
         ├─► Turso Main DB (events, connections, metadata)
         ├─► Turso Monitoring DB (API logs)
         ├─► Strava API (token refresh, clubs, events)
         └─► Alert Service (ntfy.sh + Resend email)
```

### Files Structure

```
backend/
  cmd/strava-sync/
    main.go                 # Entry point
    Dockerfile              # Multi-stage Alpine build
    test_sync.sh            # Local testing script
    infra/                  # Terraform (Cloud Run Job + Scheduler)
  internal/strava/
    sync_service.go         # Core sync orchestration
    sync_config.go          # Configuration from env vars
    sync_models.go          # SyncResult, AthleteSync models
    encryption.go           # AES-256-GCM token encryption
    connection_repo.go      # Connection queries
    event_metadata_repo.go  # Event metadata queries
  internal/alerts/
    notifier.go             # ntfy.sh + Resend notifications
```

---

## Sync Flow

For each athlete connection:

1. **Decrypt** refresh token (AES-256-GCM)
2. **Refresh** access token from Strava API
3. **Fetch** athlete's clubs, filter to admin + cycling + city match
4. **Fetch** club events for each matching club
5. **Compare** with database:
   - Existing events → update `last_refreshed_at`, increment `refresh_count`
   - Missing events → DELETE (no longer on Strava)
   - New events → skip (only sync existing imports)
6. **Update** `last_synced_at` for the connection
7. **Continue** to next athlete (errors are isolated)

---

## Key Design Decisions

### Detach on Edit

When an organizer edits a Strava-imported event:
- `source` set to NULL (detached from Strava)
- Row deleted from `strava_event_metadata` (stops sync)
- Event becomes native CycleScene event
- Complies with Strava's "no modification" rule

### Token Revocation (401)

When Strava returns 401 Unauthorized:
- Skip that athlete, continue to next
- Log to monitoring database
- No email to organizer, no cleanup
- Events remain visible (they chose to share them)

### Rate Limiting

Strava limits: 100/15min, 1000/day
Our limits: 90/15min, 900/day (10% buffer)

- Sync stops early if approaching limits
- API usage tracked in monitoring database
- 429 responses handled with backoff

---

## Configuration

### Environment Variables

```bash
# Required
TURSO_MAIN_URL=libsql://...
TURSO_MAIN_AUTH_TOKEN=...
TURSO_MONITORING_URL=libsql://...
TURSO_MONITORING_AUTH_TOKEN=...
STRAVA_CLIENT_ID=...
STRAVA_CLIENT_SECRET=...
STRAVA_TOKEN_ENCRYPTION_KEY=...  # 32-byte base64 key

# Optional
SYNC_MAX_CONNECTIONS=100         # Connections per run
SYNC_MAX_REQUESTS_15MIN=90       # API limit per 15min
SYNC_MAX_REQUESTS_DAY=900        # API limit per day
SYNC_FORCE=false                 # Bypass 3-day interval
STRAVA_DEBUG=false               # Verbose logging

# Alerts
NTFY_TOPIC=cyclescene-alerts
RESEND_API_KEY=...
ADMIN_EMAIL=admin@example.com
```

### Encryption Key

```bash
# Generate once, never change (tokens become unusable if lost)
openssl rand -base64 32

# Store in:
# - GitHub Actions secrets: STRAVA_TOKEN_ENCRYPTION_KEY
# - GCP Secret Manager (optional)
# - Password manager backup
```

---

## Testing

### Local Testing

```bash
cd backend

# Test mode (won't decrypt real tokens)
./cmd/strava-sync/test_sync.sh

# Real mode (decrypts actual connections)
./cmd/strava-sync/test_sync.sh --use-real-key

# Force mode (bypass 3-day interval)
./cmd/strava-sync/test_sync.sh --use-real-key --force
```

### Unit Tests

```bash
# All sync-related tests
go test ./internal/strava/... ./internal/api/ride/... ./internal/alerts/...

# Integration tests (require CGO)
RUN_INTEGRATION_TESTS=true CGO_ENABLED=1 go test -v ./...
```

### Test Coverage

- `encryption_test.go` - Token encryption/decryption
- `sync_service_test.go` - Event comparison, rate limiting
- `sync_service_error_isolation_test.go` - Error isolation
- `sync_service_token_revocation_test.go` - Token revocation
- `connection_repo_test.go` - Connection queries
- `event_metadata_repo_test.go` - Metadata queries, detach filtering
- `detach_on_edit_test.go` - Strava compliance

---

## Deployment

### Infrastructure

- **Docker image:** `gcr.io/cyclescene-prod/strava-sync:latest`
- **Cloud Run Job:** `strava-sync` (us-west1)
- **Cloud Scheduler:** `strava-sync-trigger` (every 3 days at 2am PST)
- **Terraform:** `backend/cmd/strava-sync/infra/`

### CI/CD

GitHub Actions workflow: `.github/workflows/deploy-strava-sync.yml`

- Builds on push to main (when backend/cmd/strava-sync/** changes)
- Multi-stage Docker build
- Pushes to Artifact Registry
- Updates Cloud Run Job

### Manual Trigger

```bash
# Via GCP Console or CLI
gcloud run jobs execute strava-sync --region us-west1
```

---

## Troubleshooting

### Common Issues

| Error | Cause | Solution |
|-------|-------|----------|
| "Failed to decrypt token" | Key mismatch | Verify `STRAVA_TOKEN_ENCRYPTION_KEY` matches encryption key |
| "Token revoked" for all | Athletes disconnected | Expected, sync skips and continues |
| "Rate limit exceeded" | Too many requests | Check limits, verify not already at quota |
| "No events found" | No imported events | Verify `source='strava'` events exist |
| 0 connections synced | Recently synced | Use `--force` flag or wait 3 days |

### Database Gotchas

1. **Timestamps are TEXT** - Must parse with `time.Parse("2006-01-02 15:04:05.000", str)`
2. **Events don't have date** - Must JOIN with `event_occurrences.start_date`
3. **Detached events skipped** - Query filters `source = 'strava'`

### Monitoring Queries

```sql
-- Recent sync activity
SELECT * FROM strava_api_logs
WHERE created_at > datetime('now', '-1 day')
ORDER BY created_at DESC LIMIT 50;

-- Events needing refresh
SELECT event_id, last_refreshed_at, refresh_count
FROM strava_event_metadata
WHERE event_id IN (SELECT id FROM events WHERE source = 'strava');

-- Connection sync status
SELECT athlete_id, city_code, last_synced_at
FROM strava_connections
ORDER BY last_synced_at DESC NULLS LAST;
```

---

## Strava Compliance

- ✅ Refresh data within 7 days (we do 3 days)
- ✅ Delete events within 48 hours if removed from Strava (immediate)
- ✅ Store tokens encrypted at rest (AES-256-GCM)
- ✅ Never log access/refresh tokens
- ✅ Respect rate limits (90/15min, 900/day)
- ✅ Handle token revocation (401)
- ✅ No modification of Strava data (detach on edit)

---

## Alerts

Critical alerts sent via ntfy.sh push + Resend email:
- Sync job failure (exit code != 0)
- Zero athletes synced successfully
- Approaching rate limits (>80%)
