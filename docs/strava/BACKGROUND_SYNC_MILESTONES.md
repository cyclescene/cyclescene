# Strava Background Sync - Implementation Milestones

**Status:** Planning
**Target:** Production Ready
**Design Doc:** [BACKGROUND_SYNC_SERVICE.md](./BACKGROUND_SYNC_SERVICE.md)

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

## Phase 1: Core Sync Infrastructure

### Milestone 1.1: Sync Service Foundation
**Goal:** Build the core sync orchestration logic

- [ ] Create `backend/cmd/strava-sync/main.go` entry point
- [ ] Create `backend/internal/strava/sync_service.go` with sync orchestration
- [ ] Implement connection fetching (query `strava_connections` table)
- [ ] Implement token decryption using existing `encryption.go`
- [ ] Add structured logging for sync lifecycle events
- [ ] Add configuration for sync settings (batch size, timeouts)

**Acceptance Criteria:**
- Service can fetch and decrypt connections
- Logging shows sync start/end with connection count
- No actual Strava API calls yet

---

### Milestone 1.2: Token Refresh & Athlete Data
**Goal:** Refresh access tokens and fetch athlete clubs

- [ ] Implement access token refresh using refresh token
- [ ] Add token refresh error handling (401 = revoked, 429 = rate limit)
- [ ] Fetch athlete's clubs using refreshed token
- [ ] Filter clubs to admin/owner roles only
- [ ] Filter clubs by sport type (cycling only)
- [ ] Update connection's `last_synced_at` timestamp after successful sync

**Acceptance Criteria:**
- Can refresh token for a test connection
- Can fetch and filter athlete's clubs
- Logs show which clubs were found for each athlete

---

### Milestone 1.3: Event Fetching & Comparison
**Goal:** Fetch club events and compare to database

- [ ] Fetch group events for each admin club
- [ ] Filter events to upcoming only (start date >= today)
- [ ] Query `strava_event_metadata` for athlete's imported events
- [ ] **Filter to only events with `source='strava'`** (skip detached events)
- [ ] Implement event comparison logic (new/existing/deleted)
- [ ] Build list of events to update and events to delete

**Acceptance Criteria:**
- Can fetch events from Strava clubs
- Can identify which events exist in our database
- Can identify stale events (in DB but not on Strava)
- Skips events that were edited locally (source != 'strava')

---

### Milestone 1.4: Event Updates & Deletions
**Goal:** Update existing events and delete stale ones

- [ ] Update `strava_event_metadata.last_refreshed_at` for existing events (timestamp only, no field updates)
- [ ] Delete stale events from `events` table (cascades to metadata)
- [ ] Log event update and deletion counts per athlete
- [ ] Log to monitoring DB: sync stats, errors, event counts

**Acceptance Criteria:**
- Existing events have updated refresh timestamps
- Stale events are deleted from database
- Only `source='strava'` events are synced (detached events ignored)
- All activity logged to monitoring database

---

### Milestone 1.5: Rate Limiting & Batching
**Goal:** Respect Strava API limits and batch processing

- [ ] Track API request count per sync run
- [ ] Implement request throttling (sleep between athletes if needed)
- [ ] Add 429 (rate limit) response handling with backoff
- [ ] Log API usage to `strava_api_logs` table (reuse existing monitoring)
- [ ] Limit sync to 100 connections per run (safety margin)
- [ ] Add early exit if approaching rate limits (>900 requests/day)

**Acceptance Criteria:**
- Sync respects 100 req/15min and 1000 req/day limits
- Logs show API usage per athlete
- No rate limit violations in test runs

---

## Phase 2: Error Handling & Resilience

### Milestone 2.1: Connection Error Handling
**Goal:** Handle per-athlete errors gracefully

- [ ] Add error handling for token refresh failures
- [ ] Add error handling for API timeouts
- [ ] Add error handling for network errors (5xx)
- [ ] Continue sync on per-athlete errors (don't fail entire job)
- [ ] Track error count and log at end of sync
- [ ] Add `sync_error_count` and `last_sync_error` to `strava_connections` table (optional)

**Acceptance Criteria:**
- One failed athlete doesn't stop the entire sync
- Errors are logged with athlete context
- Error counts are tracked and reported

---

### Milestone 2.2: Token Revocation Handling
**Goal:** Handle revoked tokens (401 unauthorized)

- [ ] Detect 401 responses during token refresh
- [ ] Log revocation events to monitoring DB with athlete_id
- [ ] Skip that athlete, continue to next (no cleanup, no emails)
- [ ] Events remain visible (they were shared by choice)

**Acceptance Criteria:**
- 401 errors logged but don't fail entire sync
- Sync continues to next athlete
- No emails sent, events remain intact

---

### Milestone 2.3: Detach-on-Edit Implementation
**Goal:** Implement logic to detach events when edited via CycleScene

**Note:** This is implemented in the **edit endpoint**, not the sync service. Listed here for completeness.

- [ ] Modify edit endpoint (`PUT /rides/edit/{token}`)
- [ ] Check if event has `source='strava'` before applying edits
- [ ] If yes, update `source` to 'cyclescene' in `events` table
- [ ] Delete row from `strava_event_metadata` table
- [ ] Apply user's edits to the now-native event
- [ ] Update UI to remove Strava branding on detached events

**Acceptance Criteria:**
- Editing a Strava event converts it to native event
- Sync service ignores detached events (source != 'strava')
- Strava branding removed from UI for detached events
- Edit token continues to work after detachment

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
**Status:** Ready for Implementation
**Owner:** TBD

**Design Decisions Finalized:** 2026-02-04
**All Questions Resolved:** ✓
