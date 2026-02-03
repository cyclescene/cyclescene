# Milestone 3: Backend - HTTP Handlers & WebSocket

**Goal:** Create API endpoints for OAuth flow and WebSocket for progress tracking

**Status:** COMPLETED (2026-02-02)

---

## Design Decisions

**Route Fetching Strategy:**
- Use app-level `STRAVA_ACCESS_TOKEN` from environment (not user OAuth tokens)
- Routes are public data, don't require user authentication
- Create per-request RouteFetcher in WebSocket handler with app token

**Geocoding Strategy:**
- Strava provides accurate lat/lng coordinates
- Add new `SubmitRideWithCoordinates(submission, lat, lng)` to ride.Service
- Skip geocoding for Strava imports (use provided coordinates)

**Source Tracking:**
- Add `Source` and `SourceID` fields to Submission struct
- Update `CreateRide()` to handle source tracking
- Enable deduplication via UNIQUE constraint on (source, source_id)

**Email Strategy:**
- Send ONE summary email per import session (not per event)
- Email contains all imported events with edit links
- Collect edit tokens during import, send summary at end

**WebSocket Keep-Alive:**
- Send heartbeat messages every 30 seconds during import
- Prevents Cloud Run/load balancer from closing idle connections

**Concurrent Imports:**
- Add backend lock using `sync.Map` (map[athleteID]bool)
- Prevent multiple simultaneous imports per user

---

## Files Summary (Milestone 3)

**New Files Created:**
- `backend/internal/api/strava/handler.go` - HTTP handlers for OAuth
- `backend/internal/api/strava/websocket.go` - WebSocket for import progress

**Files Modified (Prerequisites):**
- `backend/internal/api/ride/models.go` - Added Source/SourceID to Submission
- `backend/internal/api/ride/repo.go` - Updated CreateRide INSERT with source fields
- `backend/internal/api/ride/service.go` - Added SubmitRideWithCoordinates() method
- `backend/internal/api/magiclink/service.go` - Added SendImportSummaryEmail() method
- `backend/internal/strava/service.go` - Added ProcessRouteWithToken() method

**Files Modified (M3 Core):**
- `backend/cmd/api/handler.go` - Registered Strava routes, wired dependencies
- `backend/.env.example` - Added STRAVA_CLIENT_ID, STRAVA_CLIENT_SECRET, STRAVA_CALLBACK_URL

**Test Tool Updated:**
- `backend/cmd/strava-test/main.go` - Added M3 import test endpoints

**Go Dependencies:**
- `github.com/coder/websocket` v1.8.12 - Already present in go.mod

---

## Tasks

### 3.0 - Prerequisites (MUST COMPLETE FIRST)

**3.0.1 - Add Source Tracking to ride.Service**
- [x] Add `Source string` and `SourceID string` fields to Submission struct
- [x] Update CreateRide INSERT statement to include `source, source_id` columns

**3.0.2 - Add SubmitRideWithCoordinates to ride.Service**
- [x] Add new method `SubmitRideWithCoordinates(submission *Submission, lat, lng float64) (*SubmissionResponse, error)`
- [x] Copy logic from `SubmitRide()` but skip geocoding step
- [x] Skip magic link email sending (will be handled by summary email)
- [x] Return SubmissionResponse with eventID and editToken

**3.0.3 - Add SendImportSummaryEmail to magiclink.Service**
- [x] Add new method `SendImportSummaryEmail(ctx context.Context, email string, events []ImportedEvent) error`
- [x] Create email template with list of all imported events
- [x] Include edit link for each event

**3.0.4 - Fix ProcessRoute to Use App Token**
- [x] Added `ProcessRouteWithToken()` method that accepts token directly
- [x] Routes are public data, don't need user authentication
- [x] Original `ProcessRoute()` now delegates to `ProcessRouteWithToken()`

### 3.1 - Create HTTP Handlers
- [x] Create `backend/internal/api/strava/handler.go`
- [x] `GET /v1/strava/auth/initiate` - Start OAuth flow
- [x] `GET /v1/strava/auth/callback` - OAuth callback handler
- [x] `GET /v1/strava/admin-clubs` - Get clubs where user is admin
- [x] `GET /v1/strava/clubs/{clubId}/events` - Get events for a club
- [x] `POST /v1/strava/logout` - Clear session

**Session Cookie Configuration:**
```go
http.SetCookie(w, &http.Cookie{
    Name:     "strava_session_id",
    Value:    sessionID,
    Path:     "/",
    MaxAge:   3600, // 1 hour
    HttpOnly: true,
    Secure:   os.Getenv("APP_ENV") != "dev", // Secure in production
    SameSite: http.SameSiteLaxMode,
})
```

### 3.2 - Add Routes to API Server
- [x] Register Strava routes in `backend/cmd/api/handler.go`
- [x] Pass routes.Repository to Handler constructor
- [x] Pass magiclink.Service to Handler constructor (via NewHandlerWithImport)
- [x] Pass ride.Service to Handler constructor (via NewHandlerWithImport)
- [x] CORS already configured for all origins

### 3.3 - Create WebSocket Handler for Import Progress
- [x] Create `backend/internal/api/strava/websocket.go`
- [x] `WS /v1/strava/import` - WebSocket endpoint for event import
- [x] Accept list of event IDs to import with optional overrides
- [x] Implement concurrent import detection (activeImports sync.Map)
- [x] Send heartbeat messages every 30 seconds
- [x] Use app-level access token for route fetching
- [x] Call `SubmitRideWithCoordinates()` instead of `SubmitRide()`
- [x] Collect all edit tokens during import
- [x] Send ONE summary email at end with all event links
- [x] Send progress updates per event (4 steps: fetching, coordinates, route, database)

**Message Format:**

```typescript
// Client → Server (initiate import)
{
  "session_id": "abc123",
  "organizer_email": "user@example.com",
  "events": [
    {
      "strava_event_id": 789,
      "club_id": 123,
      "overrides": {
        "audience": "All",
        "image_url": "https://..."
      }
    }
  ]
}

// Server → Client (heartbeat - every 30s)
{
  "type": "heartbeat",
  "message": "Import still in progress..."
}

// Server → Client (progress updates)
{
  "type": "progress",
  "event_index": 0,
  "total_events": 3,
  "strava_event_id": 789,
  "event_title": "Tuesday Night Ride",
  "step": "fetching" | "coordinates" | "route" | "database",
  "status": "in_progress" | "success" | "error",
  "message": "Fetching event data..."
}

// Server → Client (completion per event)
{
  "type": "complete",
  "event_index": 0,
  "strava_event_id": 789,
  "cyclescene_event_id": 42,
  "edit_token": "xyz123",
  "edit_url": "https://form.cyclescene.cc?token=xyz123",
  "success": true,
  "error": null
}

// Server → Client (final summary)
{
  "type": "done",
  "total_imported": 3,
  "total_failed": 0,
  "summary_email_sent": true,
  "results": [...]
}
```

### 3.4 - Add Environment Variables
- [x] Add `STRAVA_CLIENT_ID` to backend/.env.example
- [x] Add `STRAVA_CLIENT_SECRET` to backend/.env.example
- [x] Add `STRAVA_CALLBACK_URL` to backend/.env.example
- [x] Document `STRAVA_ACCESS_TOKEN` (app-level token for route fetching)

### 3.5 - Integration Testing

**Automated Tests (19 tests in `internal/api/strava/*_test.go`):**
- [x] OAuth initiate/callback handler tests
- [x] Admin clubs endpoint tests
- [x] Club events endpoint tests
- [x] Session/logout tests
- [x] Route registration tests
- [x] WebSocket message serialization tests
- [x] Concurrent import detection tests
- [x] Duplicate error detection tests
- [x] Heartbeat message format tests

**Manual Testing (requires real Strava credentials):**
- [x] Test OAuth flow end-to-end with real Strava
- [x] Test WebSocket import with single event
- [x] Test duplicate detection (re-import same event)
- [x] Test summary email delivery

**Test Tool Available:** Run `APP_ENV=dev go run ./cmd/strava-test` for interactive testing:
- `/test-import-select` - Select events for import testing
- `/test-convert` - Test event conversion to CycleScene submission
- `/test-import-preview` - Preview what would be imported
- `/test-websocket` - WebSocket import test tool (connects to main API on port 8081)

---

## M3 Design Decisions Summary

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| 1 | Session management | Cookie-based | Standard web pattern, works with OAuth redirects |
| 2 | WebSocket auth | Session ID in first message | Secure, clean protocol |
| 3 | Multi-event import | Sequential with continue-on-failure | Simple, rate-limit safe |
| 4 | Progress steps | 4 steps: fetching → coordinates → route → database | User-facing, debuggable |
| 5 | Error recovery | Continue-on-failure (except rate limits) | Partial success better than total failure |
| 6 | BFF token bypass | SubmitRideWithCoordinates() method | Strava session already validated |
| 7 | Magic links | Summary email after import | One email per session |
| 8 | Route processing | During import (step 3), optional | Event saved even if route fails |
| 9 | Duplicate handling | Database UNIQUE constraint | Automatic, fast |
| 10 | Rate limits | Log but don't preemptively stop | Relies on actual 429 responses |

---

## Edge Cases & Handling

| Edge Case | Handling | User Message |
|-----------|----------|--------------|
| Session expires during WebSocket | Import proceeds (cached session) | N/A - completes normally |
| WebSocket disconnects mid-import | Partial success saved | "Connection lost" |
| Same event imported twice | Skip with UNIQUE constraint | "This event has already been imported" |
| Event deleted from Strava | Skip event, continue | "Event no longer exists in Strava" |
| Geocoding fails | Skip event, continue | "Unable to find location" |
| Route fetch fails | Continue without route | "Route unavailable" |
| Rate limit exceeded (429) | Stop import | "Strava API limit reached. Wait 15 minutes." |
| Concurrent imports (same user) | Block second import | "Import already in progress" |

---

## Completion Checklist

- [x] All files created: `handler.go`, `websocket.go`, `handler_test.go`, `websocket_test.go`
- [x] HTTP handlers: InitiateOAuth, HandleOAuthCallback, Logout, GetAdminClubs, GetClubEvents
- [x] WebSocket: Session validation, email validation, progress tracking, heartbeats
- [x] Error handling: Continue-on-failure, rate limit detection, duplicate detection
- [x] Build succeeds: `cd backend && go build ./cmd/api`
- [x] Routes registered in `cmd/api/handler.go`
- [x] Environment variables documented in `.env.example`
- [x] Automated tests pass: `go test ./internal/api/strava/...` (19 tests)
- [x] Manual OAuth flow test passes
- [x] Manual WebSocket import test passes
- [x] Summary email sent with all event links

---

## Implementation Notes (2026-02-02)

### Files Created/Modified

1. **`backend/internal/api/ride/models.go`**
   - Added `Source` and `SourceID` fields to `Submission` struct

2. **`backend/internal/api/ride/repo.go`**
   - Updated `CreateRide` INSERT to include `source, source_id` columns

3. **`backend/internal/api/ride/service.go`**
   - Added `SubmitRideWithCoordinates()` method

4. **`backend/internal/api/magiclink/service.go`**
   - Added `ImportedEvent` struct
   - Added `SendImportSummaryEmail()` method with HTML template

5. **`backend/internal/strava/service.go`**
   - Refactored `ProcessRoute()` to use new `ProcessRouteWithToken()`
   - Added `ProcessRouteWithToken()` for app-level token support

6. **`backend/internal/api/strava/handler.go`** (NEW)
   - `Handler` struct with `NewHandler()` and `NewHandlerWithImport()`
   - OAuth endpoints: InitiateOAuth, HandleOAuthCallback, Logout
   - Data endpoints: GetAdminClubs, GetClubEvents
   - Cookie-based session management

7. **`backend/internal/api/strava/websocket.go`** (NEW)
   - `ImportHandler` struct with WebSocket handling
   - Progress tracking with 4 steps
   - Heartbeat goroutine (30s interval)
   - Concurrent import detection via `sync.Map`
   - Summary email after successful imports

8. **`backend/cmd/api/handler.go`**
   - Added Strava handler initialization
   - Wired dependencies (stravaService, rideService, magicLinkService)
   - Registered routes under `/v1/strava/*`

9. **`backend/.env.example`**
   - Added STRAVA_CLIENT_ID, STRAVA_CLIENT_SECRET
   - Added STRAVA_CALLBACK_URL
   - Documented STRAVA_ACCESS_TOKEN

10. **`backend/cmd/strava-test/main.go`**
    - Added `/test-import-select` endpoint
    - Added `/test-convert` endpoint
    - Added `/test-import-preview` endpoint

11. **`backend/internal/api/strava/handler_test.go`** (NEW)
    - Tests for OAuth initiate/callback handlers
    - Tests for logout with/without session
    - Tests for admin clubs/events endpoints
    - Tests for route registration
    - Mock Strava server for isolated testing

12. **`backend/internal/api/strava/websocket_test.go`** (NEW)
    - Tests for concurrent import detection
    - Tests for message serialization (ProgressMessage, ImportRequest, ImportResult)
    - Tests for duplicate error detection
    - WebSocket integration test with mock server
    - Tests for heartbeat message format

### Fixes During Manual Testing (2026-02-03)

1. **JavaScript int64 Precision Issue**
   - Strava event IDs exceed JavaScript's MAX_SAFE_INTEGER (9007199254740991)
   - Fixed by using `json:",string"` tag in `EventImportConfig` struct
   - WebSocket clients must send event IDs as strings: `"strava_event_id": "3452409311648303182"`

2. **Resend Email Sender Domain**
   - Updated sender from `magic@cyclescene.cc` to `CycleScene <noreply@email.cyclescene.cc>`
   - Matches verified domain in Resend

3. **WebSocket Test Tool**
   - Added `/test-websocket` endpoint to strava-test for manual WebSocket testing
   - Accessible at `http://localhost:3000/test-websocket`

### Build Verification

All builds pass:
- `go build ./cmd/api` ✓
- `go build ./cmd/image-optimizer` ✓
- `go build ./cmd/scraperv2` ✓
- `go build ./cmd/strava-test` ✓
- `go vet ./...` ✓
- `go test ./internal/strava/...` ✓ (19 tests pass)
- `go test ./internal/api/strava/...` ✓ (19 tests pass)
