# Strava Background Sync Service

## Overview

GCP Cloud Run Job that runs every 48 hours to refresh Strava event data and maintain 7-day compliance.

---

## Architecture

### Service Type
**Cloud Run Job** (preferred over Cloud Functions for longer execution time)

### Trigger
**Cloud Scheduler** → every 48 hours (comfortable margin under 7-day requirement)

### Runtime
- Go binary (reuse existing backend code)
- Connects to same Turso database
- Uses existing Strava client/service layer

---

## Sync Flow (Per Connection)

```
For each athlete_id in strava_connections:
  1. Decrypt refresh_token
  2. Get fresh access_token from Strava
  3. Fetch athlete's clubs (GET /api/v3/athlete/clubs)
  4. For each admin club:
       - Fetch club events (GET /api/v3/clubs/{id}/group_events)
       - Filter: cycling clubs, matching city, upcoming events
  5. Compare fetched events to DB (strava_event_metadata):
       - New events → skip (only sync existing imports)
       - Existing events → update if changed, refresh timestamp
       - Missing events → DELETE (no longer on Strava)
  6. Update last_synced_at for athlete
  7. Handle errors gracefully (log, continue to next athlete)
```

### Key Decision: **Sync Only, Don't Auto-Import**
- Sync refreshes *existing* imported events
- Does NOT auto-import new events (requires user action)
- This keeps control with users and reduces surprise events

---

## Encryption/Decryption

### Setup
```go
// Use AES-256-GCM (already in golang.org/x/crypto)
import (
    "crypto/aes"
    "crypto/cipher"
    "encoding/base64"
)

// Encryption key from env var (32 bytes for AES-256)
encryptionKey := os.Getenv("STRAVA_TOKEN_ENCRYPTION_KEY")

func encryptToken(plaintext string, key []byte) (encrypted, nonce []byte, err error) {
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    nonce = make([]byte, gcm.NonceSize())
    rand.Read(nonce)
    encrypted = gcm.Seal(nil, nonce, []byte(plaintext), nil)
    return
}

func decryptToken(encrypted, nonce, key []byte) (string, error) {
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
    return string(plaintext), err
}
```

### Key Management
- Store encryption key in **GCP Secret Manager**
- Mount as env var in Cloud Run Job
- Rotate key annually (will require re-encryption migration)

---

## Rate Limiting Strategy

### Strava Limits
- 100 requests / 15 minutes
- 1,000 requests / day

### Our Strategy
```
Per athlete:
  - 1 request: refresh token
  - 1 request: fetch clubs (~5 clubs avg)
  - N requests: fetch events per club (~5 events per club)

Estimate: ~7 requests per athlete per sync
Safe capacity: ~140 athletes per sync run

If approaching limits:
  - Track usage in strava_api_logs
  - Sleep between athletes
  - Implement exponential backoff on 429 errors
```

---

## Event Comparison Logic

```sql
-- Fetch current Strava events for athlete's clubs
fetched_events = [Strava API response]

-- Fetch our stored events for this athlete
SELECT event_id, strava_event_id, strava_club_id
FROM strava_event_metadata
WHERE imported_by_athlete_id = ?
  AND event_id IN (
    SELECT id FROM events WHERE source = 'strava'
  )

-- Compare
FOR EACH stored_event:
  IF stored_event.strava_event_id NOT IN fetched_events:
    -- Event deleted on Strava
    DELETE FROM events WHERE id = stored_event.event_id
    -- CASCADE will delete from strava_event_metadata
    LOG "Deleted stale event: {event_id}"
  ELSE:
    -- Event still exists, refresh timestamp
    UPDATE strava_event_metadata
    SET last_refreshed_at = NOW(), refresh_count = refresh_count + 1
    WHERE event_id = stored_event.event_id
```

---

## Error Handling

### Scenarios

**1. Token Revoked (401 Unauthorized)**
```go
if resp.StatusCode == 401 {
    // User disconnected on Strava's side
    log.Warn("Token revoked for athlete", athleteID)

    // Option A: Delete connection and events (aggressive)
    db.Exec("DELETE FROM strava_connections WHERE athlete_id = ?", athleteID)

    // Option B: Mark as disabled, notify user (graceful)
    db.Exec("UPDATE strava_connections SET sync_enabled = 0 WHERE athlete_id = ?", athleteID)
    // TODO: Send email notification
}
```

**2. Rate Limited (429 Too Many Requests)**
```go
if resp.StatusCode == 429 {
    retryAfter := resp.Header.Get("Retry-After") // seconds
    log.Warn("Rate limited, backing off", retryAfter)
    time.Sleep(time.Duration(retryAfter) * time.Second)
    // Retry or skip to next athlete
}
```

**3. Network/API Errors (5xx)**
```go
if resp.StatusCode >= 500 {
    // Log error, continue to next athlete
    // Don't mark events as stale on API errors
    log.Error("Strava API error", athleteID, resp.Status)
    continue
}
```

---

## Database Queries

### Fetch connections to sync
```sql
SELECT athlete_id, refresh_token_encrypted, encryption_nonce, city_code
FROM strava_connections
WHERE sync_enabled = 1  -- If you add this column later
  AND (last_synced_at IS NULL OR last_synced_at < datetime('now', '-2 days'))
ORDER BY last_synced_at ASC NULLS FIRST
LIMIT 100;  -- Rate limit safety
```

### Delete stale events
```sql
-- Delete from events (cascade to strava_event_metadata)
DELETE FROM events
WHERE id IN (
    SELECT event_id FROM strava_event_metadata
    WHERE strava_event_id IN (?)  -- List of stale IDs
);
```

### Update refresh timestamp
```sql
UPDATE strava_event_metadata
SET last_refreshed_at = STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'),
    refresh_count = refresh_count + 1
WHERE event_id = ?;
```

---

## GCP Setup

### 1. Create Cloud Run Job
```bash
gcloud run jobs create strava-sync \
  --region us-west1 \
  --image gcr.io/cyclescene/strava-sync:latest \
  --max-retries 2 \
  --task-timeout 30m \
  --set-env-vars="TURSO_URL=...,STRAVA_CLIENT_ID=...,STRAVA_CLIENT_SECRET=..." \
  --set-secrets="STRAVA_TOKEN_ENCRYPTION_KEY=strava-encryption-key:latest"
```

### 2. Create Cloud Scheduler Job
```bash
gcloud scheduler jobs create http strava-sync-trigger \
  --location us-west1 \
  --schedule="0 2 */2 * *" \  # Every 2 days at 2am
  --uri="https://us-west1-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/cyclescene/jobs/strava-sync:run" \
  --http-method POST \
  --oauth-service-account-email=strava-sync@cyclescene.iam.gserviceaccount.com
```

### 3. Store Encryption Key
```bash
# Generate key
openssl rand -base64 32

# Store in Secret Manager
gcloud secrets create strava-encryption-key \
  --data-file=- \
  --replication-policy="automatic"

# Grant access to Cloud Run
gcloud secrets add-iam-policy-binding strava-encryption-key \
  --member="serviceAccount:strava-sync@cyclescene.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

---

## Code Structure

```
backend/
  cmd/
    strava-sync/
      main.go              # Entry point
  internal/
    strava/
      sync.go              # Sync orchestration
      encryption.go        # Token encryption/decryption
      client.go            # Reuse existing Strava API client
      service.go           # Reuse existing service logic
```

---

## Monitoring

### Structured Logging
```go
log.Info("sync_started", map[string]interface{}{
    "total_connections": connectionCount,
})

log.Info("sync_completed", map[string]interface{}{
    "total_synced": successCount,
    "total_failed": errorCount,
    "events_refreshed": refreshCount,
    "events_deleted": deleteCount,
    "duration_seconds": duration,
})
```

### Metrics to Track
- Total connections synced
- Events refreshed per sync
- Events deleted (stale)
- API rate limit usage
- Errors per connection
- Sync duration

### Alerts
- **Critical**: Sync job failure (exit code != 0)
- **Warning**: >10% connection failures
- **Warning**: Approaching rate limits (>800 requests used)

---

## Testing Strategy

### Local Testing
```bash
# Set env vars
export TURSO_URL=...
export STRAVA_CLIENT_ID=...
export STRAVA_CLIENT_SECRET=...
export STRAVA_TOKEN_ENCRYPTION_KEY=...

# Run sync
go run cmd/strava-sync/main.go
```

### Staging Environment
- Test with 1-2 real Strava connections
- Verify events refresh correctly
- Verify stale events delete
- Check logs in Cloud Logging

### Production Rollout
1. Deploy job (don't schedule yet)
2. Manually trigger once: `gcloud run jobs execute strava-sync`
3. Verify results
4. Enable scheduler

---

## Compliance Checklist

- ✅ Refresh data within 7 days (we do 2 days)
- ✅ Delete events within 48 hours if removed from Strava (immediate)
- ✅ Store tokens encrypted at rest (AES-256-GCM)
- ✅ Never log access/refresh tokens
- ✅ Respect rate limits
- ✅ Handle token revocation (401)

---

## Future Enhancements

1. **Webhook Support** (Phase 3)
   - Subscribe to Strava webhooks for real-time updates
   - Reduces polling frequency
   - Faster deletion of stale events

2. **Per-Athlete Scheduling**
   - Track last_synced_at per athlete
   - Sync different athletes on different schedules
   - Distribute load more evenly

3. **Event Update Detection**
   - Currently: only refresh timestamp
   - Future: detect title/time/location changes, update events

4. **Admin Dashboard**
   - Show sync status per connection
   - Manual sync trigger
   - Connection health monitoring
