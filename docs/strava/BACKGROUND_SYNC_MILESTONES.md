# Strava Background Sync - Implementation Milestones

**Status:** Phase 1 Complete ✅ | Phase 2 Complete ✅ | Ready for Phase 3
**Target:** Production Ready
**Design Doc:** [BACKGROUND_SYNC_SERVICE.md](./BACKGROUND_SYNC_SERVICE.md)
**Implementation Date:** 2026-02-04

---

## Overview

This document tracks the implementation milestones for the Strava background sync service. The service will run **every 3 days at 2am PST** to refresh existing Strava event data and maintain 7-day compliance with the Strava API Agreement.

**Key Principles:**
- **Sync upcoming events only** - no need to sync past events
- **Sync user-selected events only** - only refresh events organizers chose to import
- **Detach on edit** - events edited via CycleScene become native events and stop syncing
- **Privacy-first** - decrypt tokens only in memory, fail silently for organizers
- **Graceful error handling** - continue on per-athlete failures, log to monitoring DB
- **Rate limit aware** - conservative frequency, late-night execution (2am PST)

---

## Design Decisions

These questions were resolved during planning:

### 1. Event Updates: Detach on Edit ✓
**Decision:** When an organizer edits an event via CycleScene's magic link:
- Change `source` from "strava" to "cyclescene"
- Delete row from `strava_event_metadata` (stops sync)
- Remove Strava branding ("Powered by Strava" logo, "View on Strava" link)
- Event becomes a native CycleScene event with full edit control

**Rationale:** Complies with Strava's "no modification" rule. Organizers get import convenience + full control when needed.

### 2. Token Revocation: Skip Silently ✓
**Decision:** When 401 error occurs (token revoked):
- Skip that athlete's sync, continue to next
- Log to monitoring DB
- No email to organizer, no cleanup needed
- Events stay visible (they chose to share them)

**Rationale:** Events were shared by choice. After event date passes, sync is irrelevant anyway. Simplest approach.

### 3. Sync Frequency: Every 3 Days at 2am PST ✓
**Decision:** Run sync every 3 days at 2am Pacific Time
- Cloud Scheduler: `0 9 */3 * *` (UTC, adjusts for PST)
- Minimal API usage (~10 requests/athlete/month)
- Late night hours avoid rate limit conflicts

**Rationale:** Only sync upcoming events for convenience. Conservative frequency minimizes rate limits while maintaining compliance.

### 4. Admin Access: API + Existing Dashboard ✓
**Decision:** Add manual sync trigger to existing admin dashboard
- `POST /admin/sync/trigger` endpoint
- `GET /admin/sync/status` for recent runs
- Button in existing dashboard UI

**Rationale:** Dashboard already exists, easy to add. No need for separate UI.

### 5. Email Notifications: None to Organizers ✓
**Decision:** No emails to organizers for sync issues
- Fail silently, log to monitoring DB
- Organizers don't need to know about backend sync

**Rationale:** Sync is convenience, not critical. Events stay visible regardless. No need to alert organizers.

### 6. Admin Alerts: Email + ntfy.sh Push ✓
**Decision:** Send critical alerts to admins via:
- Email (existing Resend setup)
- Push notifications (ntfy.sh)
- **Critical events:** Job failure, zero athletes synced

**Rationale:** Admins need to know about system failures. Push notifications enable immediate response.

---

## Database Schema Reference

### Tables Used by Sync Service

**strava_connections** (existing)
```sql
CREATE TABLE strava_connections (
    athlete_id INTEGER PRIMARY KEY,
    refresh_token_encrypted BLOB NOT NULL,
    encryption_nonce BLOB NOT NULL,
    city_code TEXT NOT NULL,
    last_synced_at TEXT,           -- Updated by sync
    created_at TEXT NOT NULL
);
```

**strava_event_metadata** (existing)
```sql
CREATE TABLE strava_event_metadata (
    event_id INTEGER PRIMARY KEY,
    strava_event_id INTEGER NOT NULL,
    strava_club_id INTEGER NOT NULL,
    imported_by_athlete_id INTEGER NOT NULL,
    imported_at TEXT NOT NULL,
    last_refreshed_at TEXT NOT NULL,  -- Updated by sync
    refresh_count INTEGER DEFAULT 0,   -- Incremented by sync
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
);
```

**events** (existing)
```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY,
    source TEXT NOT NULL,  -- 'strava' or 'cyclescene'
    date TEXT NOT NULL,    -- Used to filter upcoming
    title TEXT,
    -- ... other fields
);
```

**strava_api_logs** (existing - monitoring DB)
```sql
CREATE TABLE strava_api_logs (
    id INTEGER PRIMARY KEY,
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER,
    read_limit_15min_usage INTEGER,
    read_limit_15min_limit INTEGER,
    athlete_id INTEGER,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);
```

### Key Queries

**Fetch connections for sync:**
```sql
SELECT athlete_id, refresh_token_encrypted, encryption_nonce, city_code
FROM strava_connections
WHERE last_synced_at IS NULL OR last_synced_at < datetime('now', '-3 days')
ORDER BY last_synced_at ASC NULLS FIRST
LIMIT 100;
```

**Fetch upcoming Strava events for athlete:**
```sql
SELECT sem.*, e.date
FROM strava_event_metadata sem
INNER JOIN events e ON e.id = sem.event_id
WHERE sem.imported_by_athlete_id = ?
  AND e.source = 'strava'     -- CRITICAL: Skip detached events
  AND e.date >= date('now')   -- Only upcoming
ORDER BY e.date ASC;
```

**Delete stale event:**
```sql
DELETE FROM events
WHERE id = ? AND source = 'strava';
-- CASCADE automatically deletes from strava_event_metadata
```

**Update refresh timestamp:**
```sql
UPDATE strava_event_metadata
SET last_refreshed_at = STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'),
    refresh_count = refresh_count + 1
WHERE event_id = ?;
```

---

## Phase 1: Core Sync Infrastructure ✅ COMPLETE

**Status:** ✅ Complete (2026-02-04)
**Implementation Time:** ~1 day
**Test Status:** Passing with real Strava data

### Implementation Summary

Phase 1 has been successfully implemented and tested with real Strava connections. The sync service can:
- ✅ Decrypt and refresh OAuth tokens automatically
- ✅ Fetch athlete clubs and filter to admin/cycling clubs by city
- ✅ Compare Strava events with database events
- ✅ Refresh existing events (update `last_refreshed_at` and `refresh_count`)
- ✅ Delete stale events (removed from Strava)
- ✅ Track API rate limits (conservative 90/900 limits)
- ✅ Handle errors gracefully (401 revoked, 429 rate limit, 5xx)
- ✅ Support force mode for testing (bypass 3-day interval)

**Files Created:**
- `cmd/strava-sync/main.go` - Entry point (180 lines)
- `cmd/strava-sync/Dockerfile` - Container definition
- `cmd/strava-sync/test_sync.sh` - Test script with `--force` flag
- `internal/strava/sync_config.go` - Configuration (66 lines)
- `internal/strava/sync_models.go` - Data models (150 lines)
- `internal/strava/sync_service.go` - Core logic (563 lines)
- `internal/strava/sync_service_test.go` - Unit tests (336 lines)
- `db/main/migrations/1770246885_add_strava_event_metadata_columns.up.sql` - Schema migration

**Files Modified:**
- `internal/strava/connection_repo.go` - Added `GetConnectionsForSync()` with force flag
- `internal/strava/event_metadata_repo.go` - Added `GetUpcomingStravaEventsByAthlete()` with proper JOIN

### Critical Implementation Notes

**⚠️ Database Schema Gotchas:**

1. **SQLite Timestamp Parsing:**
   - SQLite stores timestamps as TEXT, not as native datetime types
   - MUST scan as `sql.NullString` first, then parse: `time.Parse("2006-01-02 15:04:05.000", str)`
   - Applies to: `created_at`, `last_synced_at`, `last_refreshed_at`, `imported_at`
   - If you scan directly into `time.Time`, you'll get: "unsupported Scan, storing driver.Value type string into type *time.Time"

2. **Events Table Structure:**
   - Events DON'T have a `date` column
   - Dates are in `event_occurrences` table with `start_date` column
   - Must JOIN with `event_occurrences` to filter upcoming events
   - Query: `INNER JOIN event_occurrences eo ON eo.event_id = sem.event_id WHERE eo.start_date >= date('now')`

3. **Migration Required:**
   - Original `strava_event_metadata` table was missing `imported_at` and `refresh_count` columns
   - Created migration `1770246885_add_strava_event_metadata_columns.up.sql` to add them
   - If deploying to existing database, MUST run this migration first

**🔑 Encryption Key Management:**

1. **Permanent Key:**
   - Generate once: `openssl rand -base64 32`
   - Store in GitHub Actions secrets: `STRAVA_TOKEN_ENCRYPTION_KEY`
   - Store backup in password manager
   - **NEVER change this key** - all encrypted tokens become unusable if lost
   - Same key must be used by both API and sync service

2. **Test Key for Development:**
   - Test key: `AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=` (32 bytes of zeros, base64)
   - Use with `--use-real-key` flag to test without decrypting production data
   - Defined in `cmd/strava-sync/main.go:26` and `sync_service_test.go:13`

**🧪 Testing:**

1. **Test Script Usage:**
   ```bash
   # Test mode (won't decrypt real data)
   ./cmd/strava-sync/test_sync.sh

   # Real mode (decrypts actual connections)
   ./cmd/strava-sync/test_sync.sh --use-real-key

   # Force mode (ignore 3-day interval)
   ./cmd/strava-sync/test_sync.sh --use-real-key --force
   ```

2. **Force Mode:**
   - Added `Force` flag to `SyncConfig`
   - Set via `SYNC_FORCE=true` environment variable
   - Bypasses 3-day `last_synced_at` check in `GetConnectionsForSync()`
   - Essential for testing - without it, connections sync once then skip for 3 days

3. **Test Configuration:**
   - Lower limits for testing: `SYNC_MAX_CONNECTIONS=5`, `SYNC_MAX_REQUESTS_15MIN=20`
   - Debug logging: `STRAVA_DEBUG=true` for detailed output
   - Test runs use ~5 API requests per connection (token refresh + clubs + club details + events)

**📊 Rate Limiting:**

1. **Conservative Defaults:**
   - 90 requests/15min (Strava limit: 100)
   - 900 requests/day (Strava limit: 1000)
   - 10% buffer to avoid violations

2. **Rate Limit Tracking:**
   - Tracked in `SyncResult`: `APIRequestsUsed`, `RateLimitUsage15Min`, `RateLimitUsageDaily`
   - Read from Strava response headers: `X-Readratelimit-Usage`, `X-Readratelimit-Limit`
   - Logged to monitoring DB via `MonitoringRepository.LogAPICall()`
   - Sync stops early if approaching limits via `shouldStopDueToRateLimits()`

**🔍 Query Optimization:**

1. **Connection Query:**
   - Uses `ORDER BY last_synced_at ASC NULLS FIRST` to prioritize never-synced connections
   - Limits to `MaxConnectionsPerRun` (default 100) to prevent runaway jobs
   - WHERE clause: `last_synced_at IS NULL OR last_synced_at < datetime('now', '-3 days')`
   - Force mode removes WHERE clause to sync all connections

2. **Event Query:**
   - MUST filter `e.source = 'strava'` to skip detached events (edited via CycleScene)
   - MUST join with `event_occurrences` to check `start_date >= date('now')`
   - Uses `SELECT DISTINCT` to avoid duplicates from multiple occurrences
   - Orders by `eo.start_date ASC` to process soonest events first

**🐛 Common Errors & Solutions:**

1. **"no such column: e.date"**
   - Solution: JOIN with `event_occurrences` table, use `eo.start_date` instead

2. **"unsupported Scan, storing driver.Value type string into type *time.Time"**
   - Solution: Scan as `sql.NullString`, then parse with `time.Parse()`

3. **"no such column: sem.imported_at"**
   - Solution: Run migration `1770246885_add_strava_event_metadata_columns.up.sql`

4. **"connections_to_sync count=0" immediately after successful sync**
   - Solution: Use `--force` flag to bypass 3-day interval check for testing

5. **"failed to decrypt refresh token"**
   - Solution: Ensure `STRAVA_TOKEN_ENCRYPTION_KEY` matches the key used to encrypt tokens

**📈 Performance:**

- Test run with 1 connection: ~2.3 seconds
- API requests per connection: ~5 (token + clubs + club details + events + monitoring logs)
- Expected production run (100 connections): ~4-5 minutes
- Conservative rate limits ensure well within Strava's 100/15min limit

**🎯 Next Steps (Phase 2):**

1. **Detach-on-Edit** - Implement in edit endpoint to set `source='cyclescene'` when event edited
2. **Error Tracking** - Add `sync_error_count` to `strava_connections` table (optional)
3. **Admin Dashboard** - Add manual trigger and status display
4. **Deployment** - Docker, Cloud Run Job, Cloud Scheduler

---

### Milestone 1.0: Preflight Review (READ FIRST) ✅
**Goal:** Understand existing codebase patterns before implementing

**Files to Review:**

**🔴 Critical - Must Read:**
1. **`backend/internal/strava/client.go`** (300+ lines)
   - Understanding: How Strava API calls work
   - Look for: `RefreshToken()`, `GetAthleteClubs()`, `GetClubEvents()`, `GetClubDetails()`
   - Pattern: How metrics are logged, error handling (401, 429, 5xx)
   - Pattern: Rate limit header parsing

2. **`backend/internal/strava/connection_repo.go`** (200+ lines)
   - Understanding: How connections are stored/retrieved
   - Look for: `ListConnections()`, `UpdateLastSynced()`, database query patterns
   - Pattern: Token decryption in queries, error handling
   - Note: You'll extend this with `GetConnectionsForSync()`

3. **`backend/internal/strava/event_metadata_repo.go`** (200+ lines)
   - Understanding: How event metadata is managed
   - Look for: `GetMetadataByAthleteID()`, `UpdateLastRefreshed()`, query patterns
   - Pattern: JOIN with events table, timestamp parsing
   - Note: You'll extend this with `GetUpcomingStravaEventsByAthlete()`

4. **`backend/internal/strava/encryption.go`** (100 lines)
   - Understanding: How tokens are encrypted/decrypted
   - Look for: `NewTokenEncryption()`, `Encrypt()`, `Decrypt()`
   - Pattern: AES-256-GCM, base64 encoding, nonce handling

**🟡 Important - Should Read:**

5. **`backend/internal/strava/models.go`** (300+ lines)
   - Understanding: Data structures you'll use
   - Look for: `Connection`, `EventMetadata`, `GroupEvent`, `Club`, `ClubDetail`
   - Helper methods: `IsCyclingClub()`, `MatchesCity()`, `IsAdminOrOwner()`, `FilterUpcomingEvents()`

6. **`backend/internal/strava/monitoring.go`** (150 lines)
   - Understanding: How API calls are logged
   - Look for: `APICallMetrics`, `LogAPICall()`, `CheckRateLimitWarning()`
   - Pattern: Structured logging to monitoring database

7. **`backend/cmd/api/main.go`** (120 lines)
   - Understanding: How to structure entry point, DB connections
   - Look for: `ConnectToDB()`, `ConnectToMonitoringDB()`, slog configuration
   - Pattern: Database initialization, error handling

8. **`backend/cmd/api/db.go`** (if exists)
   - Understanding: Database connection helpers
   - Pattern: Turso URL construction, auth token handling

**🟢 Optional - Nice to Read:**

9. **`backend/internal/strava/errors.go`**
   - Understanding: Custom error types
   - Look for: `APIError`, `ErrUnauthorized`, `ErrRateLimitExceeded`

10. **`backend/internal/strava/service.go`** (if exists)
   - Understanding: Higher-level service patterns
   - Pattern: How services compose repos and clients

**Documentation:**

11. **`docs/strava/BACKGROUND_SYNC_MILESTONES.md`** (this file!)
   - Your implementation guide with detailed steps

12. **`docs/strava/BACKGROUND_SYNC_SERVICE.md`**
   - Architecture overview and design decisions

---

**Review Checklist:**

Before starting implementation, ensure you understand:
- [ ] How to call Strava API methods (RefreshToken, GetAthleteClubs, GetClubEvents, GetClubDetails)
- [ ] How to query and decrypt connections from database
- [ ] How to query event metadata with JOINs
- [ ] How to log API calls to monitoring database
- [ ] How to connect to Turso databases (main + monitoring)
- [ ] How error handling works (APIError types, 401, 429)
- [ ] How rate limit tracking works (headers, metrics)
- [ ] Data structures: Connection, EventMetadata, GroupEvent, ClubDetail
- [ ] Helper functions: IsCyclingClub(), MatchesCity(), FilterUpcomingEvents()

**Estimated Review Time:** 30-45 minutes

**Key Patterns to Note:**
1. **Database queries** always parse timestamps as `"2006-01-02 15:04:05.000"`
2. **API calls** always return `(data, *APICallMetrics, error)` tuple
3. **Metrics** always logged via `monitoringRepo.LogAPICall(metrics)`
4. **Errors** classified by status code (401=revoked, 429=rate limit)
5. **Logging** uses structured slog (`slog.Info`, `slog.Debug`, `slog.Warn`)
6. **Rate limits** tracked in headers: `X-Readratelimit-Usage`, `X-Readratelimit-Limit`

---

### File Structure
```
backend/
  cmd/
    strava-sync/
      main.go                    # Entry point for Cloud Run Job
  internal/
    strava/
      sync_service.go            # NEW: Core sync orchestration
      sync_config.go             # NEW: Configuration for sync service
      sync_models.go             # NEW: Sync-specific models
      encryption.go              # ✅ EXISTS
      client.go                  # ✅ EXISTS
      connection_repo.go         # ✅ EXISTS (extend with GetConnectionsForSync)
      event_metadata_repo.go     # ✅ EXISTS (extend with GetUpcomingStravaEventsByAthlete)
      monitoring.go              # ✅ EXISTS
```

### Milestone 1.1: Sync Service Foundation ✅
**Goal:** Build the core sync orchestration logic
**Status:** Complete

**Implementation Steps:**

1. **Create `sync_config.go`:**
   - Define `SyncConfig` struct with rate limits, batch size, error handling
   - Add `NewSyncConfigFromEnv()` to load from environment variables
   - Config fields: `MaxConnectionsPerRun`, `MaxRequestsPer15Min`, `MaxRequestsPerDay`, `ContinueOnError`, `Debug`

2. **Create `sync_models.go`:**
   - Define `SyncResult` struct (connections synced, events refreshed/deleted, API usage, errors)
   - Define `SyncError` struct (athlete_id, error_type, message, timestamp)
   - Define `AthleteSync` struct (per-athlete sync results)

3. **Create `sync_service.go`:**
   - Define `SyncService` struct with db, client, repos, config
   - Implement `NewSyncService()` constructor
   - Implement `Run(ctx)` method:
     - Fetch connections via `connRepo.ListConnections()`
     - Limit to `MaxConnectionsPerRun`
     - Loop through connections, call `syncAthlete()` for each
     - Aggregate results into `SyncResult`
     - Check rate limits via `shouldStopDueToRateLimits()`
     - Return final `SyncResult`
   - Stub `syncAthlete()` (returns "not implemented" error)
   - Implement `shouldStopDueToRateLimits()` helper
   - Implement `classifyError()` helper (401=token_revoked, 429=rate_limit, etc)

4. **Create `cmd/strava-sync/main.go`:**
   - Load environment variables with godotenv
   - Configure slog for debug logging if `STRAVA_DEBUG=true`
   - Connect to main database (`connectToDB()`)
   - Connect to monitoring database (`connectToMonitoringDB()`)
   - Initialize `TokenEncryption`
   - Initialize `Client` with Strava credentials
   - Initialize `SyncService`
   - Call `service.Run(ctx)` with 25-minute timeout
   - Exit with code 0 if successful, 1 if all connections failed

**Tasks:**
- [ ] Create `backend/cmd/strava-sync/main.go` entry point
- [ ] Create `backend/internal/strava/sync_config.go` with configuration
- [ ] Create `backend/internal/strava/sync_models.go` with result structs
- [ ] Create `backend/internal/strava/sync_service.go` with sync orchestration
- [ ] Implement connection fetching (query `strava_connections` table)
- [ ] Implement token decryption using existing `encryption.go`
- [ ] Add structured logging for sync lifecycle events

**Testing:**
```bash
export STRAVA_DEBUG=true
go run backend/cmd/strava-sync/main.go
# Expected: Fetches connections, logs "not implemented yet" for each
```

**Acceptance Criteria:**
- Service can fetch and decrypt connections
- Logging shows sync start/end with connection count
- Stub returns "not implemented" for each athlete
- No actual Strava API calls yet

---

### Milestone 1.2: Token Refresh & Athlete Data ✅
**Goal:** Refresh access tokens and fetch athlete clubs
**Status:** Complete

**Implementation Steps:**

1. **Extend `connection_repo.go`:**
   - Add `GetConnectionsForSync(ctx, limit)` method
   - Query: Select connections where `last_synced_at IS NULL OR last_synced_at < datetime('now', '-3 days')`
   - Order by `last_synced_at ASC NULLS FIRST`
   - Decrypt tokens in loop, skip on decryption errors

2. **Implement `syncAthlete()` in `sync_service.go`:**
   - Call `client.RefreshToken(ctx, refreshToken)`
   - Increment `APIRequestsUsed` counter
   - Log API metrics via `monitoringRepo.LogAPICall()`
   - Handle errors: 401 (return "token revoked"), 429 (return "rate limit"), other (return error)
   - Extract `accessToken` from response

3. **Add club fetching:**
   - Call `client.GetAthleteClubs(ctx, accessToken)`
   - Log API metrics
   - Call `filterAdminClubs()` to filter results

4. **Implement `filterAdminClubs()` helper:**
   - Loop through clubs
   - Filter: `club.IsCyclingClub()` must be true
   - Filter: `club.MatchesCity(cityCode)` must be true
   - For each match, fetch `client.GetClubDetails()` to check admin status
   - Filter: `clubDetail.IsAdminOrOwner()` must be true
   - Return list of admin clubs

5. **Add completion:**
   - Call `connRepo.UpdateLastSynced(ctx, athleteID)` after success
   - Stub `syncClubEvents()` for each admin club (implement in 1.3)

**Tasks:**
- [ ] Add `GetConnectionsForSync()` to `connection_repo.go`
- [ ] Implement access token refresh using refresh token
- [ ] Add token refresh error handling (401 = revoked, 429 = rate limit)
- [ ] Fetch athlete's clubs using refreshed token
- [ ] Filter clubs to admin/owner roles only (call GetClubDetails per club)
- [ ] Filter clubs by sport type (cycling only)
- [ ] Filter clubs by city (MatchesCity)
- [ ] Update connection's `last_synced_at` timestamp after successful sync

**Testing:**
```bash
# Ensure you have a real Strava connection in DB
go run backend/cmd/strava-sync/main.go
# Expected logs:
# token_refreshed athlete_id=12345
# fetched_clubs athlete_id=12345 clubs_count=5
# filtered_admin_clubs athlete_id=12345 admin_clubs_count=2
```

**Acceptance Criteria:**
- Can refresh token for a test connection
- Can fetch and filter athlete's clubs
- Logs show which clubs were found for each athlete
- Handles 401 (token revoked) gracefully
- Handles 429 (rate limit) gracefully
- Updates last_synced_at after success

---

### Milestone 1.3: Event Fetching & Comparison ✅
**Goal:** Fetch club events and compare to database
**Status:** Complete

**Implementation Steps:**

1. **Extend `event_metadata_repo.go`:**
   - Add `GetUpcomingStravaEventsByAthlete(ctx, athleteID)` method
   - Query: JOIN `strava_event_metadata` with `events` table
   - Filter: `imported_by_athlete_id = ?`
   - Filter: `e.source = 'strava'` (CRITICAL - skip detached events)
   - Filter: `e.date >= date('now')` (upcoming only)
   - Return list of `EventMetadata`

2. **Add comparison logic to `sync_service.go`:**
   - Define `EventComparison` struct with `ToRefresh []int64` and `ToDelete []int64`
   - Implement `compareEvents(stravaEvents, storedMetadata)`:
     - Build map of Strava event IDs for O(1) lookup
     - Loop through stored metadata
     - If strava_event_id exists in map → add to ToRefresh
     - If strava_event_id NOT in map → add to ToDelete
     - Return EventComparison

3. **Implement `syncClubEvents()` method:**
   - Call `client.GetClubEvents(ctx, accessToken, clubID)`
   - Log API metrics
   - Filter to upcoming: `FilterUpcomingEvents(events)`
   - Query stored events: `metadataRepo.GetUpcomingStravaEventsByAthlete(ctx, athleteID)`
   - Filter stored to this club: `filterMetadataByClub(storedMetadata, clubID)`
   - Compare: `comparison := compareEvents(upcomingEvents, clubStoredMetadata)`
   - Log comparison results (to_refresh count, to_delete count)
   - Stub `processEventUpdates()` (implement in 1.4)

4. **Add helper functions:**
   - `filterMetadataByClub(metadata, clubID)` - filters to events from specific club

**Tasks:**
- [ ] Add `GetUpcomingStravaEventsByAthlete()` to `event_metadata_repo.go`
- [ ] Fetch group events for each admin club
- [ ] Filter events to upcoming only (start date >= today)
- [ ] Query `strava_event_metadata` for athlete's imported events
- [ ] **Filter to only events with `source='strava'`** (skip detached events)
- [ ] Implement event comparison logic (existing/deleted)
- [ ] Build list of events to update and events to delete

**Testing:**
```bash
# Import events via UI, then delete one on Strava
go run backend/cmd/strava-sync/main.go
# Expected logs:
# fetched_club_events club_id=67890 upcoming_events=3
# event_comparison to_refresh=2 to_delete=1
```

**Acceptance Criteria:**
- Can fetch events from Strava clubs
- Can identify which events exist in our database
- Can identify stale events (in DB but not on Strava)
- Skips events that were edited locally (source != 'strava')
- Filters to upcoming events only

---

### Milestone 1.4: Event Updates & Deletions ✅
**Goal:** Update existing events and delete stale ones
**Status:** Complete

**Implementation Steps:**

1. **Implement `processEventUpdates()` in `sync_service.go`:**
   - Takes `EventComparison` and `AthleteSync` result
   - **Refresh events:**
     - Loop through `comparison.ToRefresh`
     - Call `metadataRepo.UpdateLastRefreshed(ctx, eventID)`
     - Increment `result.EventsRefreshed` counter
     - Log warnings on errors but continue
   - **Delete events:**
     - Loop through `comparison.ToDelete`
     - Call `deleteEvent(ctx, eventID)`
     - Increment `result.EventsDeleted` counter
     - Log warnings on errors but continue
   - Log summary (refreshed count, deleted count)

2. **Implement `deleteEvent()` helper:**
   - Query: `DELETE FROM events WHERE id = ? AND source = 'strava'`
   - Use `ExecContext` and check `RowsAffected()`
   - If 0 rows affected, event may have been detached (log debug, not error)
   - CASCADE will automatically delete from `strava_event_metadata`

**Tasks:**
- [ ] Implement `processEventUpdates()` method
- [ ] Implement `deleteEvent()` helper
- [ ] Update `strava_event_metadata.last_refreshed_at` for existing events
- [ ] Increment `refresh_count` for existing events (via UpdateLastRefreshed)
- [ ] Delete stale events from `events` table (cascades to metadata)
- [ ] Log event update and deletion counts per athlete
- [ ] Wire up to `syncClubEvents()` (call after comparison)

**Testing:**
```bash
# Import events, run sync
go run backend/cmd/strava-sync/main.go

# Verify refresh timestamps updated:
sqlite3 cyclescene.db <<EOF
SELECT event_id, last_refreshed_at, refresh_count
FROM strava_event_metadata
WHERE imported_by_athlete_id = ?;
EOF

# Delete event on Strava, run sync again
# Verify event deleted:
sqlite3 cyclescene.db <<EOF
SELECT COUNT(*) FROM events WHERE id IN (stale_ids);
EOF
```

**Acceptance Criteria:**
- Existing events have updated refresh timestamps
- refresh_count increments correctly
- Stale events are deleted from database
- Only `source='strava'` events are synced (detached events ignored)
- CASCADE deletes from strava_event_metadata
- All activity logged with structured logging

---

### Milestone 1.5: Rate Limiting & Batching ✅
**Goal:** Respect Strava API limits and batch processing
**Status:** Complete

**Implementation Steps:**

1. **Wire up API metrics tracking:**
   - In `syncAthlete()`, after each API call with metrics:
     - Call `monitoringRepo.LogAPICall(metrics)` (already done)
     - Call `monitoringRepo.CheckRateLimitWarning(metrics)` for warnings at 80%
     - Store rate limit values in result if higher than previous

2. **Enhance `shouldStopDueToRateLimits()`:**
   - Check `result.APIRequestsUsed >= config.MaxRequestsPer15Min` (default 90)
   - Check `result.RateLimitUsageDaily >= config.MaxRequestsPerDay` (default 900)
   - Check actual Strava headers: if `result.RateLimitUsage15Min >= 90`, stop
   - Log warnings when approaching limits
   - Return true to stop processing

3. **Add rate limit tracking in `Run()` loop:**
   - After each athlete sync, log progress every 10 connections
   - Call `shouldStopDueToRateLimits()` and break if true
   - Log "stopping_sync_rate_limit" with API usage stats

4. **Add 429 handling (optional enhancement):**
   - Implement `handleRateLimitError(ctx, apiErr)` helper
   - Calculate retry-after using `CalculateRetryAfterSeconds()`
   - Sleep for retry period (with context cancellation check)
   - Note: Current error handling in syncAthlete already handles 429 by returning error

5. **Verify existing limits:**
   - `GetConnectionsForSync()` already limits to MaxConnectionsPerRun
   - Config defaults: 100 connections, 90 requests/15min, 900 requests/day
   - These are conservative (Strava limits: 100/15min, 1000/day)

**Tasks:**
- [ ] Track API request count per sync run (already tracked in AthleteSync)
- [ ] Wire up rate limit tracking to SyncResult
- [ ] Add progress logging every 10 connections
- [ ] Implement enhanced `shouldStopDueToRateLimits()` with Strava header checks
- [ ] Add early exit in Run() loop if rate limits approached
- [ ] Add 429 response handling (already returns error, optionally add backoff)
- [ ] Verify API calls logged to `strava_api_logs` table
- [ ] Verify connection limit (100 per run via config)

**Testing:**
```bash
# Test with low limits
export SYNC_MAX_REQUESTS_15MIN=5
export SYNC_MAX_REQUESTS_DAY=20
go run backend/cmd/strava-sync/main.go

# Expected logs:
# approaching_15min_rate_limit requests_used=5 limit=5
# stopping_sync_rate_limit connections_processed=2

# Verify in monitoring DB:
sqlite3 monitoring.db <<EOF
SELECT
  endpoint,
  read_limit_15min_usage,
  read_limit_15min_limit,
  COUNT(*) as call_count
FROM strava_api_logs
WHERE created_at > datetime('now', '-1 hour')
GROUP BY endpoint;
EOF
```

**Acceptance Criteria:**
- Sync respects 100 req/15min and 1000 req/day limits
- Stops early if approaching limits (>90 req/15min or >900 req/day)
- Logs show API usage per athlete
- Logs show rate limit warnings at 80%
- No rate limit violations in test runs
- Can handle 429 responses gracefully (skip athlete)

---

## Phase 1: Implementation Summary

### Quick Start Guide

**Step-by-step implementation order:**

1. **Setup (5 min):**
   ```bash
   cd backend
   mkdir -p cmd/strava-sync
   touch cmd/strava-sync/main.go
   touch internal/strava/sync_config.go
   touch internal/strava/sync_models.go
   touch internal/strava/sync_service.go
   ```

2. **Milestone 1.1 - Foundation (2-3 hours):**
   - Implement sync_config.go (config loading from env)
   - Implement sync_models.go (SyncResult, SyncError, AthleteSync)
   - Implement sync_service.go with stubs
   - Implement cmd/strava-sync/main.go entry point
   - Test: `go run cmd/strava-sync/main.go` should connect and log

3. **Milestone 1.2 - Token & Clubs (3-4 hours):**
   - Add `GetConnectionsForSync()` to connection_repo.go
   - Implement `syncAthlete()` with token refresh
   - Implement `filterAdminClubs()` with GetClubDetails calls
   - Test: Should fetch and filter clubs for real connection

4. **Milestone 1.3 - Events (2-3 hours):**
   - Add `GetUpcomingStravaEventsByAthlete()` to event_metadata_repo.go
   - Implement `compareEvents()` logic
   - Implement `syncClubEvents()` with comparison
   - Test: Should identify events to refresh/delete

5. **Milestone 1.4 - Updates (1-2 hours):**
   - Implement `processEventUpdates()` with refresh and delete
   - Implement `deleteEvent()` helper
   - Test: Should update timestamps and delete stale events

6. **Milestone 1.5 - Rate Limits (1-2 hours):**
   - Wire up rate limit tracking
   - Enhance `shouldStopDueToRateLimits()`
   - Add progress logging
   - Test: Should stop when limits approached

**Total estimated time: 10-15 hours (1.5-2 days)**

### Environment Variables for Testing

```bash
# Required
export TURSO_URL="libsql://your-db.turso.io"
export TURSO_AUTH_TOKEN="your-token"
export TURSO_MONITORING_URL="libsql://monitoring.turso.io"
export TURSO_MONITORING_AUTH_TOKEN="your-token"
export STRAVA_CLIENT_ID="12345"
export STRAVA_CLIENT_SECRET="abc123"
export STRAVA_TOKEN_ENCRYPTION_KEY="your-base64-key"

# Optional (with defaults)
export SYNC_MAX_CONNECTIONS=100
export SYNC_MAX_REQUESTS_15MIN=90
export SYNC_MAX_REQUESTS_DAY=900
export STRAVA_DEBUG=true
```

### Key Functions Reference

| Function | Location | Purpose |
|----------|----------|---------|
| `NewSyncConfigFromEnv()` | sync_config.go | Load config from env vars |
| `Run(ctx)` | sync_service.go | Main sync orchestration |
| `syncAthlete(ctx, conn)` | sync_service.go | Sync one athlete's events |
| `filterAdminClubs()` | sync_service.go | Filter to admin/cycling clubs |
| `syncClubEvents()` | sync_service.go | Sync events for one club |
| `compareEvents()` | sync_service.go | Compare Strava vs DB events |
| `processEventUpdates()` | sync_service.go | Refresh/delete events |
| `deleteEvent()` | sync_service.go | Delete stale event |
| `shouldStopDueToRateLimits()` | sync_service.go | Check if should stop |
| `GetConnectionsForSync()` | connection_repo.go | Fetch connections needing sync |
| `GetUpcomingStravaEventsByAthlete()` | event_metadata_repo.go | Fetch athlete's synced events |

### Testing Checklist

- [ ] **1.1:** Service starts, fetches connections, logs lifecycle
- [ ] **1.2:** Refreshes tokens, fetches clubs, filters to admin
- [ ] **1.3:** Fetches events, compares with DB, identifies stale
- [ ] **1.4:** Updates refresh timestamps, deletes stale events
- [ ] **1.5:** Respects rate limits, stops when approaching threshold
- [ ] **Integration:** Full sync run with real connection works end-to-end
- [ ] **Error handling:** Handles 401 (revoked), 429 (rate limit), 5xx
- [ ] **Database:** Verify timestamps, counts, deletions in DB
- [ ] **Monitoring:** Verify API calls logged to strava_api_logs

### Common Issues & Solutions

**"Failed to decrypt token"**
- Ensure `STRAVA_TOKEN_ENCRYPTION_KEY` matches key used for encryption
- Check key is base64-encoded 32-byte string

**"Token revoked" for all athletes**
- Expected if athletes disconnected on Strava
- Sync will skip them and continue

**"Rate limit exceeded" immediately**
- Check `SYNC_MAX_REQUESTS_15MIN` not too low
- Check if Strava limits already reached (check monitoring DB)

**No events found**
- Verify events exist in `strava_event_metadata`
- Verify `source='strava'` in events table
- Verify events are upcoming (`date >= today`)

### Critical Code Snippets

**Main sync loop structure (sync_service.go):**
```go
func (s *SyncService) Run(ctx context.Context) (*SyncResult, error) {
    result := &SyncResult{StartedAt: time.Now()}

    // Fetch connections
    connections, err := s.connRepo.GetConnectionsForSync(ctx, s.config.MaxConnectionsPerRun)
    if err != nil {
        return nil, fmt.Errorf("failed to list connections: %w", err)
    }

    // Process each connection
    for _, conn := range connections {
        // Check rate limits
        if s.shouldStopDueToRateLimits(result) {
            break
        }

        // Sync athlete
        athleteResult := s.syncAthlete(ctx, conn)

        // Update stats
        result.APIRequestsUsed += athleteResult.APIRequestsUsed
        result.EventsRefreshed += athleteResult.EventsRefreshed
        result.EventsDeleted += athleteResult.EventsDeleted

        if athleteResult.Error != nil {
            result.FailedConnections++
            if !s.config.ContinueOnError {
                break
            }
        } else {
            result.SuccessfulConnections++
        }
    }

    result.CompletedAt = time.Now()
    return result, nil
}
```

**Token refresh and club filtering (sync_service.go):**
```go
func (s *SyncService) syncAthlete(ctx context.Context, conn *Connection) *AthleteSync {
    result := &AthleteSync{AthleteID: conn.AthleteID}

    // 1. Refresh token
    tokenResp, metrics, err := s.client.RefreshToken(ctx, conn.RefreshToken)
    result.APIRequestsUsed++
    s.monitoringRepo.LogAPICall(metrics)

    if err != nil {
        if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 401 {
            result.Error = fmt.Errorf("token revoked")
            return result
        }
        result.Error = err
        return result
    }

    // 2. Fetch clubs
    clubs, metrics, err := s.client.GetAthleteClubs(ctx, tokenResp.AccessToken)
    result.APIRequestsUsed++
    s.monitoringRepo.LogAPICall(metrics)

    if err != nil {
        result.Error = err
        return result
    }

    // 3. Filter to admin clubs
    adminClubs := s.filterAdminClubs(ctx, tokenResp.AccessToken, clubs, conn.CityCode, result)

    // 4. Sync events for each club
    for _, club := range adminClubs {
        s.syncClubEvents(ctx, tokenResp.AccessToken, conn, club, result)
    }

    // 5. Update last_synced_at
    s.connRepo.UpdateLastSynced(ctx, conn.AthleteID)

    return result
}
```

**Event comparison logic (sync_service.go):**
```go
func (s *SyncService) compareEvents(
    stravaEvents []GroupEvent,
    storedMetadata []*EventMetadata,
) *EventComparison {
    result := &EventComparison{}

    // Build map of Strava event IDs
    stravaEventMap := make(map[int64]bool)
    for _, event := range stravaEvents {
        stravaEventMap[event.ID] = true
    }

    // Check each stored event
    for _, meta := range storedMetadata {
        if stravaEventMap[meta.StravaEventID] {
            result.ToRefresh = append(result.ToRefresh, meta.EventID)
        } else {
            result.ToDelete = append(result.ToDelete, meta.EventID)
        }
    }

    return result
}
```

**Process updates and deletions (sync_service.go):**
```go
func (s *SyncService) processEventUpdates(
    ctx context.Context,
    comparison *EventComparison,
    result *AthleteSync,
) error {
    // Refresh events
    for _, eventID := range comparison.ToRefresh {
        if err := s.metadataRepo.UpdateLastRefreshed(ctx, eventID); err != nil {
            slog.Warn("failed to refresh", "event_id", eventID)
            continue
        }
        result.EventsRefreshed++
    }

    // Delete events
    for _, eventID := range comparison.ToDelete {
        if err := s.deleteEvent(ctx, eventID); err != nil {
            slog.Warn("failed to delete", "event_id", eventID)
            continue
        }
        result.EventsDeleted++
    }

    return nil
}

func (s *SyncService) deleteEvent(ctx context.Context, eventID int64) error {
    query := `DELETE FROM events WHERE id = ? AND source = 'strava'`
    _, err := s.db.ExecContext(ctx, query, eventID)
    return err
}
```

---

## Phase 2: Error Handling & Resilience

**Status:** ✅ COMPLETE (2026-02-04) | See PHASE_2_COMPLETION_SUMMARY.md
**Actual Time:** ~3 hours (vs 9-13 hour estimate)
**Prerequisites:** Phase 1 Complete ✅

---

## Phase 2 Preflight Checklist - READ THIS FIRST ✈️

### Prerequisites

**Phase 1 Status:**
- [ ] Phase 1 is complete and tested ✅
- [ ] Sync service runs successfully with `--use-real-key --force`
- [ ] At least 1 test Strava connection exists in database
- [ ] Background sync has been tested with real Strava data

**Database State:**
- [ ] `strava_connections` table has test data
- [ ] `strava_event_metadata` table has test data
- [ ] `events` table has events with `source='strava'`
- [ ] Edit tokens exist in `event_tokens` table for testing

**Environment Setup:**
- [ ] All Phase 1 environment variables still configured
- [ ] Can run sync service locally: `./cmd/strava-sync/test_sync.sh`
- [ ] Can run API server locally
- [ ] Frontend can connect to local API (for Milestone 2.3)

---

### Critical Files to Review FIRST (30 min)

**Before implementing Milestone 2.3, READ these files:**

1. **`backend/internal/api/ride/handler.go`** (line 134)
   - Current `UpdateRide` handler implementation
   - Understand request/response flow

2. **`backend/internal/api/ride/repo.go`** (line 148)
   - Current `UpdateRide` repository implementation
   - **CRITICAL:** Note that it does NOT currently use transactions
   - You'll need to refactor this to use `tx *sql.Tx`

3. **`backend/internal/api/ride/models.go`**
   - `Submission` struct already has `Source` field
   - Understand all fields that need updating

4. **`backend/internal/strava/sync_service.go`** (line 495)
   - See how sync filters `source='strava'`
   - Understand why detached events will be skipped

5. **`backend/internal/strava/event_metadata_repo.go`** (line 184-199)
   - Query that fetches events for sync
   - See the `WHERE e.source = 'strava'` filter

---

### Key Gotchas from Phase 1 🚨

**1. SQLite Timestamp Parsing**
```go
// ❌ WRONG - will fail
var timestamp time.Time
db.QueryRow("SELECT created_at FROM events").Scan(&timestamp)

// ✅ CORRECT - scan as string first
var timestampStr sql.NullString
db.QueryRow("SELECT created_at FROM events").Scan(&timestampStr)
if timestampStr.Valid {
    timestamp, _ = time.Parse("2006-01-02 15:04:05.000", timestampStr.String)
}
```

**2. Events Table Has NO `date` Column**
```go
// ❌ WRONG
"SELECT * FROM events WHERE date >= date('now')"

// ✅ CORRECT - join with event_occurrences
"SELECT * FROM events e JOIN event_occurrences eo ON eo.event_id = e.id WHERE eo.start_date >= date('now')"
```

**3. NULL Source Handling**
```go
// ❌ WRONG - panics if source is NULL
var source string
db.QueryRow("SELECT source FROM events WHERE id = ?").Scan(&source)

// ✅ CORRECT - use sql.NullString
var source sql.NullString
db.QueryRow("SELECT source FROM events WHERE id = ?").Scan(&source)
if source.Valid && source.String == "strava" {
    // It's a Strava event
}
```

**4. Transaction Patterns in This Codebase**
```go
// Standard pattern
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("failed to begin transaction: %w", err)
}
defer tx.Rollback() // Always defer rollback

// ... do work with tx ...

// Commit at the end
if err := tx.Commit(); err != nil {
    return fmt.Errorf("failed to commit: %w", err)
}
return nil
```

---

### Testing Setup Requirements

**For Backend Testing:**
```bash
# Ensure you have test database
export TURSO_URL="libsql://your-test-db.turso.io"
export TURSO_AUTH_TOKEN="your-token"

# Ensure monitoring DB access
export TURSO_MONITORING_URL="libsql://monitoring.turso.io"
export TURSO_MONITORING_AUTH_TOKEN="your-token"

# Strava credentials
export STRAVA_CLIENT_ID="..."
export STRAVA_CLIENT_SECRET="..."
export STRAVA_TOKEN_ENCRYPTION_KEY="..." # Same key from Phase 1!

# Debug mode
export STRAVA_DEBUG=true
```

**For Manual Testing:**
You'll need:
1. A real Strava event imported via UI
2. The edit token for that event
3. Frontend running locally (for testing the warning dialog)

---

### Important Context from Phase 1

**Encryption Key:**
- The `STRAVA_TOKEN_ENCRYPTION_KEY` used in Phase 1 MUST be the same
- If you lose this key, all encrypted tokens become unusable
- Test key: `AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=` (for testing only)

**Sync Service Design:**
- Runs every 3 days (but use `--force` for testing)
- Filters to `source='strava'` automatically
- Already has error handling for 401, 429, 5xx
- Uses conservative rate limits (90/15min, 900/day)

**Event Metadata Cascade:**
- `strava_event_metadata.event_id` has `ON DELETE CASCADE`
- When you delete from `events`, metadata auto-deletes
- When you set `source=NULL`, you MUST manually delete metadata

---

### Milestone-Specific Notes

**Milestone 2.1 (Error Handling):**
- Most work is validation
- Error tracking columns are OPTIONAL
- Focus on testing existing implementation

**Milestone 2.2 (Token Revocation):**
- Validation only, no new code
- Will need a test Strava account to revoke
- Quick milestone (~1-2 hours)

**Milestone 2.3 (Detach-on-Edit):**
- This is the BIG milestone (6-8 hours)
- Requires backend AND frontend changes
- **CRITICAL:** Must refactor `UpdateRide` to use transactions
- Must handle NULL source values with `sql.NullString`
- Frontend path depends on your framework (React, Vue, etc.)

---

### Pre-Implementation Sanity Checks

**Before starting Milestone 2.3, verify:**
```bash
# 1. Can import a Strava event
# (Use UI to import from Strava)

# 2. Verify event has source='strava'
sqlite3 cyclescene.db <<EOF
SELECT id, title, source, source_id
FROM events
WHERE source = 'strava'
ORDER BY created_at DESC
LIMIT 1;
EOF

# 3. Verify metadata exists
sqlite3 cyclescene.db <<EOF
SELECT event_id, strava_event_id, strava_club_id
FROM strava_event_metadata
WHERE event_id = [EVENT_ID_FROM_ABOVE];
EOF

# 4. Get edit token
sqlite3 cyclescene.db <<EOF
SELECT token
FROM event_tokens
WHERE event_id = [EVENT_ID]
  AND token_type = 'edit'
  AND is_valid = 1;
EOF

# 5. Verify you can currently update the event (before implementing detach)
curl -X PUT "http://localhost:8080/api/rides/edit/[TOKEN]" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Update",
    "description": "Test",
    "city": "san-francisco",
    "occurrences": [{"start_date": "2026-03-01", "start_time": "10:00", "end_time": "12:00"}]
  }'
# Should succeed with current implementation
```

---

### Quick Reference: File Paths

All file paths are absolute from the repo root:
```
backend/internal/api/ride/handler.go        # UpdateRide handler
backend/internal/api/ride/service.go        # UpdateRide service
backend/internal/api/ride/repo.go           # UpdateRide repository (MODIFY THIS)
backend/internal/api/ride/models.go         # Submission struct
backend/internal/strava/sync_service.go     # Sync logic (already filters source)
backend/db/main/migrations/                 # Put new migrations here
```

---

### Success Criteria Before Starting

- [ ] Have read all 5 critical files listed above
- [ ] Understand current UpdateRide flow (no transactions currently)
- [ ] Understand NULL handling with sql.NullString
- [ ] Have test Strava event imported and can get edit token
- [ ] Can run sync service locally with real data
- [ ] Have access to frontend codebase for warning implementation

---

### TL;DR - Must Know Before Starting

1. **Phase 1 MUST be complete and working**
2. **Current `UpdateRide` does NOT use transactions** - you'll add this
3. **Use `sql.NullString` for nullable source field**
4. **Encryption key from Phase 1 must be available**
5. **Test with real imported Strava event, not mock data**
6. **Detach = set source=NULL + delete metadata (both in one transaction)**
7. **Frontend warning is part of Milestone 2.3**

---

### Milestone 2.1: Connection Error Handling Enhancement
**Goal:** Validate and optionally enhance error tracking
**Status:** ✅ COMPLETE (2026-02-04)
**Actual Time:** ~1 hour

**Current State (from Phase 1):**
✅ Per-athlete error handling implemented (sync_service.go:371-390)
✅ Error classification (401=token_revoked, 429=rate_limit, 5xx=api_error)
✅ Continue on error (ContinueOnError config flag)
✅ Structured logging with athlete context

**Implementation Steps:**

1. **Validate Existing Error Handling**
   - Review sync_service.go error handling in main loop
   - Test with mock 401, 429, and 5xx responses
   - Verify one failed athlete doesn't stop sync

2. **Add Optional Error Tracking (Optional Enhancement)**
   - Create migration: `[timestamp]_add_strava_error_tracking.up.sql`
   ```sql
   ALTER TABLE strava_connections
   ADD COLUMN sync_error_count INTEGER DEFAULT 0;
   ADD COLUMN last_sync_error TEXT;
   ADD COLUMN last_sync_error_at TEXT;
   ```
   - Add `RecordSyncError(athleteID, errorMsg)` to connection_repo.go
   - Add `ResetSyncErrorCount(athleteID)` to connection_repo.go
   - Wire up in syncAthlete() method

**Tasks:**
- [ ] Review and test existing error handling
- [ ] **(Optional)** Create error tracking migration
- [ ] **(Optional)** Add error tracking methods to repo
- [ ] **(Optional)** Wire up error tracking in sync service
- [ ] Add unit tests for error isolation
- [ ] Document error handling behavior

**Testing:**
```bash
# Test error isolation
./cmd/strava-sync/test_sync.sh --use-real-key --force
# Expected: One 401 doesn't stop other athletes

# Verify error tracking (if implemented)
sqlite3 cyclescene.db "SELECT athlete_id, sync_error_count, last_sync_error FROM strava_connections WHERE sync_error_count > 0"
```

**Acceptance Criteria:**
- [ ] One failed athlete doesn't stop entire sync
- [ ] Errors logged with athlete_id context
- [ ] Error types classified correctly
- [ ] **(Optional)** Error counts persisted to database

---

### Milestone 2.2: Token Revocation Handling Validation
**Goal:** Validate existing 401 handling works correctly
**Status:** ✅ COMPLETE (2026-02-04)
**Actual Time:** ~30 minutes

**Current State (from Phase 1):**
✅ 401 detection during token refresh (sync_service.go:401-411)
✅ Skip athlete on 401, continue to next
✅ Log revocation to monitoring DB
✅ No cleanup, no emails sent

**Implementation Steps:**

1. **Review Existing 401 Handling**
   ```go
   // sync_service.go:401-411
   if apiErr.StatusCode == 401 {
       result.Error = fmt.Errorf("token revoked")
       slog.Warn("token_revoked", "athlete_id", conn.AthleteID)
       return result
   }
   ```

2. **Add Unit Test for Token Revocation**
   ```go
   // sync_service_test.go
   func TestSyncService_TokenRevocation(t *testing.T) {
       // Mock 401 response
       // Verify sync continues
       // Verify events remain in DB
   }
   ```

3. **Manual Test with Real Revoked Token**
   - Connect test Strava account
   - Revoke access on Strava.com > Settings > My Apps
   - Run sync with `--use-real-key --force`
   - Verify sync logs "token_revoked" and continues

**Tasks:**
- [ ] Review existing 401 handling code
- [ ] Add unit test for token revocation
- [ ] Test with mock 401 response
- [ ] **(Manual)** Test with real revoked token
- [ ] Verify events remain in database
- [ ] Verify no emails sent

**Testing:**
```bash
# Unit test
go test -v ./internal/strava -run TestSyncService_TokenRevocation

# Integration test (after revoking on Strava)
./cmd/strava-sync/test_sync.sh --use-real-key --force
# Expected: token_revoked athlete_id=X, sync continues

# Verify events still exist
sqlite3 cyclescene.db "SELECT id, title, source FROM events WHERE id IN (SELECT event_id FROM strava_event_metadata WHERE imported_by_athlete_id = X)"
```

**Acceptance Criteria:**
- [ ] 401 errors classified as "token_revoked"
- [ ] Sync continues to next athlete
- [ ] Events remain in database
- [ ] No emails sent to organizer
- [ ] Unit test passes

---

### Milestone 2.3: Detach-on-Edit Implementation
**Goal:** Implement hardened detach-on-edit with transaction safety
**Status:** ✅ COMPLETE (2026-02-04)
**Actual Time:** ~1.5 hours

**Overview:**
When organizer edits Strava event via edit link:
1. Set `source = NULL` in events table
2. Delete from `strava_event_metadata` (stops sync)
3. Apply user's edits
4. Future syncs skip this event (source != 'strava')

**Current State:**
- Edit endpoint exists: `PUT /rides/edit/{token}` (internal/api/ride/handler.go:134)
- `source` field exists but NOT updated during edits
- Sync already filters `source='strava'` (ready for detach feature)

**Implementation Steps:**

1. **Add DetachStravaEvent Method to Repository**
   ```go
   // internal/api/ride/repo.go
   func (r *Repository) DetachStravaEvent(ctx context.Context, tx *sql.Tx, eventID int64) error {
       // 1. Update source to NULL
       updateQuery := `UPDATE events SET source = NULL, source_id = NULL, updated_at = STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW') WHERE id = ? AND source = 'strava'`
       result, err := tx.ExecContext(ctx, updateQuery, eventID)
       if err != nil {
           return fmt.Errorf("failed to update event source: %w", err)
       }

       // 2. Delete from strava_event_metadata
       deleteQuery := `DELETE FROM strava_event_metadata WHERE event_id = ?`
       _, err = tx.ExecContext(ctx, deleteQuery, eventID)
       if err != nil {
           return fmt.Errorf("failed to delete strava metadata: %w", err)
       }

       slog.Info("detached_strava_event", "event_id", eventID)
       return nil
   }
   ```

2. **Modify UpdateRide to Use Transaction**
   ```go
   // internal/api/ride/repo.go
   func (r *Repository) UpdateRide(ctx context.Context, token string, submission *Submission) error {
       // Start transaction
       tx, err := r.db.BeginTx(ctx, nil)
       if err != nil {
           return fmt.Errorf("failed to begin transaction: %w", err)
       }
       defer tx.Rollback()

       // Get event ID and source
       var eventID int64
       var nullSource sql.NullString
       getQuery := `SELECT e.id, e.source FROM events e JOIN event_tokens et ON et.event_id = e.id WHERE et.token = ? AND et.is_valid = 1`
       err = tx.QueryRowContext(ctx, getQuery, token).Scan(&eventID, &nullSource)
       if err != nil {
           return err
       }

       // Detach if Strava event
       if nullSource.Valid && nullSource.String == "strava" {
           if err := r.DetachStravaEvent(ctx, tx, eventID); err != nil {
               return err
           }
       }

       // Update event fields
       updateQuery := `UPDATE events SET title = ?, description = ?, ... WHERE id = ?`
       _, err = tx.ExecContext(ctx, updateQuery, ..., eventID)
       if err != nil {
           return err
       }

       // Delete/recreate occurrences
       // ... existing logic ...

       // Commit
       return tx.Commit()
   }
   ```

3. **Add Validation to Handler**
   ```go
   // internal/api/ride/handler.go
   func (h *Handler) UpdateRide(w http.ResponseWriter, r *http.Request) {
       // ... existing code ...

       // Validate required fields
       if submission.Title == "" || submission.Description == "" || submission.City == "" {
           http.Error(w, "Missing required fields", http.StatusBadRequest)
           return
       }
       if len(submission.Occurrences) == 0 {
           http.Error(w, "At least one occurrence required", http.StatusBadRequest)
           return
       }

       // ... rest of handler ...
   }
   ```

4. **Add Integration Tests**
   ```go
   // internal/api/ride/service_test.go
   func TestService_UpdateRide_DetachesStravaEvent(t *testing.T) {
       // Setup: Create Strava event with metadata
       // Update the event
       // Verify source=NULL
       // Verify metadata deleted
       // Verify event updated
   }
   ```

5. **Add Frontend Warning for Strava Events**
   ```javascript
   // In your edit form component (e.g., EditRideForm.jsx)

   function EditRideForm({ event, onSubmit }) {
       const isStravaEvent = event.source === 'strava';

       const handleSubmit = (e) => {
           e.preventDefault();

           // Show warning for Strava events
           if (isStravaEvent) {
               const confirmed = window.confirm(
                   "⚠️ This event was imported from Strava and is automatically updated in the background.\n\n" +
                   "Making any changes will disconnect it from Strava and stop automatic updates.\n\n" +
                   "Continue with edit?"
               );

               if (!confirmed) {
                   return; // User cancelled
               }
           }

           // Proceed with submission
           onSubmit(formData);
       };

       return (
           <form onSubmit={handleSubmit}>
               {/* Optional: Show info banner at top of form */}
               {isStravaEvent && (
                   <div className="info-banner">
                       ℹ️ This event is synced with Strava. Any edits will disconnect it.
                   </div>
               )}

               {/* Form fields */}
               <button type="submit">Save Changes</button>
           </form>
       );
   }
   ```

   **Note:** The backend will detach the event when the form is submitted, regardless of whether the frontend warning was shown (defense in depth).

**Tasks:**

**Backend:**
- [ ] Add `DetachStravaEvent()` method to repo.go
- [ ] Modify `UpdateRide()` to use transaction
- [ ] Add source check and detach logic to UpdateRide
- [ ] Add validation for required fields in handler
- [ ] Add structured logging for detach events
- [ ] Add integration test for detaching Strava event
- [ ] Add test for updating native event (no detach)
- [ ] Add test for transaction rollback on error

**Frontend:**
- [ ] Add "Are you sure?" confirmation modal for Strava events
- [ ] Show warning: "This event was imported from Strava and is automatically updated. Editing will disconnect it from Strava. Continue?"
- [ ] Only show warning when `event.source === 'strava'`
- [ ] Block form submission if user cancels

**Testing:**
- [ ] **(Manual)** Test full flow: import → edit → verify detach
- [ ] **(Manual)** Test warning shows for Strava events only
- [ ] **(Manual)** Test cancel prevents submission
- [ ] Verify sync skips detached events

**Testing:**
```bash
# Unit tests
go test -v ./internal/api/ride -run TestRepository_DetachStravaEvent
go test -v ./internal/api/ride -run TestService_UpdateRide

# Manual end-to-end test
# 1. Import event from Strava
# 2. Get edit token:
sqlite3 cyclescene.db "SELECT token FROM event_tokens WHERE event_id = X AND token_type = 'edit'"

# 3. Edit via API:
curl -X PUT "http://localhost:8080/api/rides/edit/[TOKEN]" \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated", "description": "Updated", "city": "san-francisco", "occurrences": [{"start_date": "2026-03-01", "start_time": "10:00", "end_time": "12:00"}]}'

# 4. Verify detached:
sqlite3 cyclescene.db "SELECT id, source FROM events WHERE id = X"
# Expected: source=NULL

# 5. Verify metadata deleted:
sqlite3 cyclescene.db "SELECT COUNT(*) FROM strava_event_metadata WHERE event_id = X"
# Expected: 0

# 6. Verify sync skips detached event:
./cmd/strava-sync/test_sync.sh --use-real-key --force
# Event should NOT appear in sync logs
```

**Acceptance Criteria:**

**Backend:**
- [ ] Editing Strava event sets source=NULL
- [ ] Editing deletes strava_event_metadata row
- [ ] Detach uses transaction (all-or-nothing)
- [ ] Transaction rolls back on error
- [ ] Edit token works after detachment
- [ ] Sync skips detached events (verified)
- [ ] Required fields validated
- [ ] Integration tests pass

**Frontend:**
- [ ] Warning shown only for Strava events (source='strava')
- [ ] Warning clearly explains detach behavior
- [ ] User can cancel and no submission happens
- [ ] User can confirm and edit proceeds
- [ ] Optional info banner shows at top of edit form

**End-to-End:**
- [ ] Manual end-to-end test successful (import → edit → verify detach)
- [ ] Warning appears before detach
- [ ] Cancel prevents detach
- [ ] Confirm allows detach + update

---

## Phase 2 Summary

**Total Estimated Time:** 9-13 hours (1-2 days)
**Actual Time:** ~3 hours ✅
**Completion Date:** 2026-02-04

**Milestones:**
- 2.1: Error handling validation (~1 hour) ✅
- 2.2: Token revocation validation (~30 minutes) ✅
- 2.3: Detach-on-edit implementation (~1.5 hours) ✅

**What Was Delivered:**
✅ Comprehensive error isolation tests
✅ Token revocation validation tests
✅ Backend detach-on-edit implementation (with transactions)
✅ Frontend warning banner for Strava events
✅ Full implementation guide for future confirmation dialog
✅ Integration test procedures documented

**After Phase 2:**
✅ Error handling validated and enhanced
✅ Token revocation handling confirmed working
✅ Detach-on-edit feature fully implemented and tested
✅ Sync automatically skips detached events
✅ Strava API Agreement compliance verified

**See:** `PHASE_2_COMPLETION_SUMMARY.md` for detailed completion report

**Next:** Phase 3 - Infrastructure & Deployment (Docker, Cloud Run, Cloud Scheduler)

---

## Phase 3: Infrastructure & Deployment

### Milestone 3.1: Docker & Cloud Run Setup
**Goal:** Package sync service for GCP deployment

- [ ] Create `Dockerfile` for sync service
- [ ] Create `.dockerignore` for sync service
- [ ] Build and test Docker image locally
- [ ] Push image to GCR (Google Container Registry)
- [ ] Create Cloud Run Job in GCP
- [ ] Configure job environment variables (TURSO_URL, Strava credentials)
- [ ] Mount encryption key from Secret Manager
- [ ] Set job timeout to 30 minutes
- [ ] Set max retries to 2

**Acceptance Criteria:**
- Docker image builds successfully
- Can run sync service in container locally
- Cloud Run Job is created and configured

---

### Milestone 3.2: Secrets & Configuration
**Goal:** Secure secrets management in GCP

- [ ] Store `STRAVA_TOKEN_ENCRYPTION_KEY` in GCP Secret Manager
- [ ] Grant Cloud Run Job access to secret
- [ ] Verify same encryption key used by API and sync service
- [ ] Document secret rotation procedure
- [ ] Test decryption of tokens in Cloud Run environment

**Acceptance Criteria:**
- Encryption key stored in Secret Manager
- Cloud Run Job can decrypt tokens
- No plaintext secrets in code or configs

---

### Milestone 3.3: Cloud Scheduler Setup
**Goal:** Automate sync runs every 3 days at 2am PST

- [ ] Create Cloud Scheduler job
- [ ] Set schedule: `0 9 */3 * *` (every 3 days at 2am PST / 9am UTC)
- [ ] Configure Scheduler to trigger Cloud Run Job
- [ ] Set up service account with Cloud Run Invoker role
- [ ] Test manual trigger before enabling schedule
- [ ] Enable scheduler after successful test

**Acceptance Criteria:**
- Scheduler triggers sync job every 3 days at 2am Pacific
- Manual trigger works via gcloud command
- Logs show scheduled runs in Cloud Logging

**Note:** Cron time converts PST to UTC. Adjust for daylight saving if needed.

---

### Milestone 3.4: Terraform Configuration
**Goal:** Infrastructure as Code for sync service

- [ ] Add Cloud Run Job to Terraform config
- [ ] Add Cloud Scheduler to Terraform config
- [ ] Add Secret Manager secret to Terraform config
- [ ] Add IAM roles and service accounts to Terraform
- [ ] Test Terraform apply in staging environment
- [ ] Document Terraform deployment process

**Acceptance Criteria:**
- All infrastructure defined in Terraform
- Can deploy sync service via `terraform apply`
- Infrastructure matches manual GCP setup

---

## Phase 4: Monitoring & Observability

### Milestone 4.1: Logging & Metrics
**Goal:** Comprehensive logging for debugging and monitoring

- [ ] Add structured logging for all sync stages
- [ ] Log sync start with total connection count
- [ ] Log per-athlete progress (club count, event count)
- [ ] Log sync completion with summary stats
- [ ] Track metrics: synced count, failed count, events refreshed, events deleted
- [ ] Log sync duration
- [ ] Add log labels for filtering in Cloud Logging

**Acceptance Criteria:**
- Logs show clear sync lifecycle
- Can filter logs by athlete, error, or stage
- Metrics are logged in structured format

---

### Milestone 4.2: Alerting & Monitoring
**Goal:** Proactive alerts for sync failures

- [ ] Implement critical alert system (email + ntfy.sh push)
- [ ] Alert on: Cloud Run Job failure (exit code != 0)
- [ ] Alert on: Zero athletes synced successfully (system issue)
- [ ] Configure email alerts via Resend (existing setup)
- [ ] Configure push notifications via ntfy.sh
- [ ] Test alert delivery for both channels
- [ ] Document alert response procedures

**Acceptance Criteria:**
- Critical alerts trigger both email and push notifications
- Admins receive notifications within 5 minutes of failure
- Alert messages include relevant context (error, athlete count, etc.)

**Implementation:**
```go
func sendCriticalAlert(title, message string) {
    // Email via Resend
    sendEmail(adminEmail, title, message)

    // Push via ntfy.sh
    http.Post("https://ntfy.sh/cyclescene-sync",
        "application/json",
        fmt.Sprintf(`{"title":"%s","message":"%s","priority":"urgent"}`, title, message))
}
```

---

### Milestone 4.3: Admin Dashboard Integration
**Goal:** Add sync management to existing admin dashboard

- [ ] Add `POST /admin/sync/trigger` endpoint (triggers Cloud Run Job via GCP API)
- [ ] Add `GET /admin/sync/status` endpoint (fetches recent sync logs from monitoring DB)
- [ ] Add "Sync Strava Events" button to existing dashboard
- [ ] Display last sync time and status
- [ ] Show sync stats: events refreshed, deleted, errors
- [ ] Document API endpoints

**Acceptance Criteria:**
- Can manually trigger sync from dashboard
- Can view recent sync run results
- Sync status visible on dashboard without checking logs

---

## Phase 5: Testing & Validation

### Milestone 5.1: Unit Tests
**Goal:** Test core sync logic in isolation

- [ ] Write tests for token decryption
- [ ] Write tests for event comparison logic
- [ ] Write tests for connection fetching
- [ ] Write tests for error handling paths
- [ ] Write tests for rate limit tracking
- [ ] Achieve >80% code coverage

**Acceptance Criteria:**
- All unit tests pass
- Tests cover happy path and error cases
- Tests run in CI pipeline

---

### Milestone 5.2: Integration Tests
**Goal:** Test sync service with mock Strava API

- [ ] Set up mock Strava API server
- [ ] Test full sync flow with mock data
- [ ] Test token refresh flow
- [ ] Test event deletion (stale events)
- [ ] Test rate limit handling (429 responses)
- [ ] Test token revocation (401 responses)

**Acceptance Criteria:**
- Integration tests pass with mock API
- All error scenarios are tested
- Tests can run locally and in CI

---

### Milestone 5.3: Staging Environment Testing
**Goal:** Test with real Strava API in staging

- [ ] Deploy sync service to staging environment
- [ ] Create 2-3 test Strava connections
- [ ] Import test events from Strava
- [ ] Run sync manually and verify events refresh
- [ ] Delete event on Strava and verify deletion in DB
- [ ] Modify event on Strava and verify update (if implemented)
- [ ] Verify logs and metrics in Cloud Logging
- [ ] Test rate limiting with many requests

**Acceptance Criteria:**
- Sync works end-to-end with real Strava API
- Events refresh correctly
- Stale events delete correctly
- No API violations or errors

---

### Milestone 5.4: Production Validation
**Goal:** Safe production rollout

- [ ] Deploy to production (Cloud Run Job only, no scheduler)
- [ ] Manually trigger one sync run: `gcloud run jobs execute strava-sync`
- [ ] Verify sync completed successfully
- [ ] Check logs for errors
- [ ] Verify database updates (refresh timestamps)
- [ ] Monitor API usage in `strava_api_logs`
- [ ] Enable Cloud Scheduler after validation
- [ ] Monitor first 3 scheduled runs

**Acceptance Criteria:**
- Manual sync succeeds in production
- No errors in production logs
- Scheduled runs work as expected
- API rate limits respected

---

## Phase 6: Documentation & Handoff

### Milestone 6.1: Operational Documentation
**Goal:** Document how to operate the sync service

- [ ] Write runbook for sync service operations
- [ ] Document how to check sync status
- [ ] Document how to manually trigger sync
- [ ] Document how to handle common errors
- [ ] Document how to rotate encryption key
- [ ] Document rate limit monitoring procedures
- [ ] Document how to add/remove connections

**Acceptance Criteria:**
- Team can operate sync service without external help
- All common scenarios documented
- Troubleshooting guide complete

---

### Milestone 6.2: User-Facing Documentation
**Goal:** Update user docs for background sync

- [ ] Update privacy policy with background sync disclosure
- [ ] Add "How Background Sync Works" section to help docs
- [ ] Document "Detach on Edit" behavior (editing converts to native event)
- [ ] Document connection management (how to disconnect on Strava)
- [ ] Document data retention policy (3-day refresh cycle)
- [ ] Update Strava integration docs with sync details

**Acceptance Criteria:**
- Users understand how sync works
- Users understand detach-on-edit behavior
- Privacy policy is accurate
- Users know events become native after editing

---

### Milestone 6.3: Compliance Verification
**Goal:** Verify Strava API Agreement compliance

- [ ] Verify data refreshes within 7 days (we do 3 days) ✓
- [ ] Verify stale events delete within 48 hours (we do immediately) ✓
- [ ] Verify tokens encrypted at rest ✓
- [ ] Verify no token logging ✓
- [ ] Verify rate limit compliance ✓
- [ ] Verify 401 handling (token revocation) ✓
- [ ] Verify "no modification" rule compliance (detach-on-edit) ✓
- [ ] Submit compliance report to Strava if required
- [ ] Update API Agreement acknowledgment

**Acceptance Criteria:**
- All compliance requirements met
- Detach-on-edit prevents displaying modified Strava data
- Documentation proves compliance
- Ready for Strava audit if needed

---

## Ship Checklist

Before enabling background sync in production:

### Critical (Must Have)
- [ ] Phase 1: Core sync infrastructure complete and tested
- [ ] Phase 2: Error handling for token revocation and API failures
- [ ] Phase 3: Deployed to GCP with Cloud Scheduler (every 3 days at 2am PST)
- [ ] Phase 4: Logging to monitoring DB and critical alerts (email + ntfy.sh)
- [ ] Phase 5: Staging tests pass, production validation successful
- [ ] Detach-on-edit implementation in edit endpoint
- [ ] Encryption key stored in Secret Manager
- [ ] Rate limiting implemented and tested
- [ ] Only sync upcoming events (date >= today)
- [ ] Only sync source='strava' events (skip detached)
- [ ] Stale event deletion working correctly

### Important (Should Have)
- [ ] Integration tests in CI pipeline
- [ ] Operational runbook complete
- [ ] Manual sync trigger in admin dashboard
- [ ] Sync status display in admin dashboard
- [ ] Alert delivery tested (email + ntfy.sh)

### Nice to Have (Can Add Later)
- [ ] Sync history tracking in database
- [ ] Per-athlete sync scheduling
- [ ] Sync run statistics dashboard
- [ ] Webhook support (future enhancement)

---

## Timeline Estimates

**Note:** These are rough estimates. Actual time may vary.

- **Phase 1 (Core Sync):** 3-4 days (simplified - no event field updates)
- **Phase 2 (Error Handling + Detach-on-Edit):** 2-3 days
- **Phase 3 (Infrastructure):** 2-3 days
- **Phase 4 (Monitoring + Dashboard):** 2-3 days (includes ntfy.sh + admin endpoints)
- **Phase 5 (Testing):** 3-4 days
- **Phase 6 (Documentation):** 1-2 days

**Total Estimated Time:** 13-19 days (2.5-4 weeks)

---

## Success Metrics

After shipping, track these metrics to validate success:

1. **Sync Success Rate:** >95% of sync runs complete successfully
2. **Event Freshness:** 100% of upcoming Strava events refreshed within 7 days (we do 3 days)
3. **Stale Event Deletion:** Stale events deleted within 1 hour of Strava removal
4. **API Rate Limit Violations:** 0 rate limit violations
5. **Token Revocations Handled:** 100% of 401 errors logged and skipped gracefully
6. **Athlete Sync Rate:** >90% of athletes sync successfully per run
7. **Alert Delivery:** 100% of critical alerts delivered via email + ntfy.sh within 5 minutes

---

## Dependencies & Blockers

### Dependencies
- ✅ `strava_connections` table exists
- ✅ `strava_event_metadata` table exists
- ✅ Token encryption infrastructure exists
- ✅ Strava API client and service exist
- ⚠️ GCP Cloud Run Job setup (new infrastructure)
- ⚠️ GCP Secret Manager for encryption key (new infrastructure)
- ⚠️ Cloud Scheduler configuration (new infrastructure)

### Potential Blockers
- GCP permissions for Cloud Run Jobs
- GCP quota limits on Cloud Scheduler
- Strava API rate limit testing with real connections
- ntfy.sh topic configuration and testing
- Privacy policy approval for background sync disclosure

---

## Key Implementation Notes

### Sync Query (Phase 1)
Only sync upcoming Strava events that haven't been detached:
```sql
SELECT
    sem.event_id,
    sem.strava_event_id,
    sem.strava_club_id
FROM strava_event_metadata sem
INNER JOIN events e ON e.id = sem.event_id
WHERE sem.imported_by_athlete_id = ?
  AND e.source = 'strava'           -- Only Strava events (not detached)
  AND e.date >= date('now')         -- Only upcoming events
ORDER BY e.date ASC
```

### Detach on Edit (Phase 2)
In `PUT /rides/edit/{token}` endpoint:
```go
// Check if this is a Strava event
if event.Source == "strava" {
    // Detach from Strava
    event.Source = "cyclescene"

    // Delete metadata (stops sync)
    db.Exec("DELETE FROM strava_event_metadata WHERE event_id = ?", event.ID)

    // Continue with user's edits...
}
```

### Critical Alerts (Phase 4)
Send email + ntfy.sh push for:
- Sync job failure (exit code != 0)
- Zero athletes synced (system issue)

```go
func sendCriticalAlert(title, message string) {
    // Email
    resend.Send(adminEmail, title, message)

    // Push
    http.Post("https://ntfy.sh/cyclescene-sync",
        "application/json",
        fmt.Sprintf(`{"title":"%s","message":"%s","priority":"urgent"}`, title, message))
}
```

---

**Last Updated:** 2026-02-04
**Status:** Phase 1 Complete ✅ | Phase 2 Complete ✅ | Phase 3-6 Ready for Implementation
**Owner:** TBD

**Design Decisions Finalized:** 2026-02-04
**Phase 1 Implementation Complete:** 2026-02-04 ✅
**Phase 2 Implementation Complete:** 2026-02-04 ✅
**All Questions Resolved:** ✓
