# Milestone 2: Backend - Service Layer

**Goal:** Build business logic for OAuth sessions and event conversion

**Status:** Completed 2026-02-02

---

## Design Decisions (Finalized 2026-02-02)

All design considerations have been resolved through collaborative planning. These decisions guide the M2 implementation:

### Decision #1: Event Duration Default
- **Value:** 120 minutes (2 hours)
- **Rationale:** Sensible default for social rides, users can edit later via magic link
- **Implementation:** Set `EventDurationMinutes: 120` in converter

### Decision #2: Ride Length from Route Distance
- **Implementation:** Calculate from `event.Route.Distance` when available
- **Format:** "25.3 miles" (converted from meters using `distance * 0.000621371`)
- **Fallback:** Empty string if no route data
- **Rationale:** `ride_length` is freeform text field, distance is most useful info for users

### Decision #3: Route Deduplication
- **Method:** Use existing `routes.Repository.CreateRoute()`
- **Mechanism:** UNIQUE constraint on `(source, source_id)` already exists
- **Impact:** No M2 work needed - already handled by existing infrastructure

### Decision #4: City-First Filtering (CRITICAL - 83% API Call Reduction)
- **Source:** City code from form URL (`?city=pdx`) via PWA redirect
- **Flow:** Form → OAuth (stores cityCode with state) → Session (includes cityCode) → Filter clubs & events
- **Filter Strategy - Three Stages:**
  1. **City Match:** Club city contains "portland" (pdx) or "salt lake" (slc) - case-insensitive
  2. **Sport Type:** Club activity_types includes cycling types (Ride, EBikeRide, VirtualRide, Handcycle, Velomobile)
  3. **Admin Status:** User is admin OR owner of club
- **Critical:** Stages 1 & 2 happen BEFORE admin check (reduces API calls)
- **API Call Impact:**
  - Without filtering: 1 + 30 clubs = 31 GetClubDetails calls
  - With filtering: 1 + 5 Portland cycling clubs = 6 GetClubDetails calls
  - **83% reduction in API calls!**

### Decision #5: Unsupported Cities
- **Handling:** None needed - strict city filtering automatically excludes them
- **Result:** Seattle/Bend/Eugene events never shown to users

### Decision #6: Organizer Email (Required)
- **Source:** User input before import (one email applied to all events)
- **Rationale:** Edit links require email (magic link system)
- **Implementation:** Frontend collects email, backend applies to all imported events
- **Additional Contact Fields:**
  - `OrganizerName`: Empty (optional)
  - `OrganizerPhone`: Empty (optional)
  - `WebURL`: Link to Strava event (`https://www.strava.com/clubs/{club_id}/group_events/{event_id}`)
  - `WebName`: "View on Strava"

### Decision #7: Authentication for Multi-Event Import
- **Problem:** BFF tokens are single-use (marked `used=1` after first submission)
- **Solution:** Separate authentication paths
  - **Strava imports:** Validate OAuth session via `SessionStore.GetSession(sessionID)`
  - **Manual submissions:** Continue using BFF tokens (no change)

---

## Gap Resolutions (All 8 Gaps Addressed)

**Gap #1: City Code Validation**
- Validate in both Service (InitiateOAuth) and Converter
- Return 400 Bad Request if invalid/unsupported city code

**Gap #2: Session Expiry During Import**
- Non-issue - JSON fetching and submission is very fast (seconds, not minutes)

**Gap #3: Multiple Events Import - Partial Failures**
- Continue on failure strategy: Process all events sequentially
- On failure: Add to retry channel, continue to next event
- After first pass: Retry failures with exponential backoff

**Gap #4: Duplicate Event Detection**
- Add source tracking to events table via migration
- `source="strava"`, `source_id="{strava_event_id}"`

**Gap #5: Timezone & DST Edge Cases**
- Use Strava's `zone` field directly
- Trust Go's `time.LoadLocation()` - handles DST automatically

**Gap #6: Event with RouteID but No Route Data**
- Use existing `routes.RouteFetcher.FetchAndConvert()`
- Routes are optional (nice-to-have)

**Gap #7: Geocoding Failure Strategy**
- Three-tier location handling:
  1. **Best case:** Use `start_latlng` array directly
  2. **Fallback:** Geocode `address` field
  3. **Fail:** Return error if neither available

**Gap #8: TinyTitle Field**
- Leave NULL/empty - legacy field from shift2bikes schema

---

## Files Summary (Milestone 2)

**New Files to Create:**
- `backend/internal/strava/service.go` - Business logic layer
- `backend/internal/strava/converter.go` - Event conversion
- `backend/internal/strava/service_test.go` - Service layer tests
- `backend/internal/strava/converter_test.go` - Converter tests

**Existing Files to Integrate With:**
- `backend/internal/strava/client.go` - API client (from M1)
- `backend/internal/strava/session.go` - Session management
- `backend/internal/strava/models.go` - Type definitions
- `backend/internal/strava/monitoring.go` - Monitoring repository (from M1)
- `backend/internal/routes/fetcher.go` - Route fetching
- `backend/internal/scraper/geocode.go` - Geocoding

---

## Tasks

### 2.1 - Create Strava Service Layer
- [x] Create `backend/internal/strava/service.go`
- [x] Implement dependency injection constructor
- [x] Add method: `InitiateOAuth()` - Generate state token and return authorization URL
- [x] Add method: `HandleOAuthCallback()` - Exchange code for token, create session
- [x] Add method: `GetAdminClubs()` - Fetch clubs where user is admin/owner
- [x] Add method: `GetClubEvents()` - Fetch events for a specific club
- [x] Add helper: `logAPICall()` - Write API metrics to monitoring DB

**Service Structure:**
```go
type Service struct {
    client         *Client
    sessionStore   *SessionStore
    monitoringRepo *MonitoringRepository
    callbackURL    string
    debug          bool
}

func NewService(
    client *Client,
    sessionStore *SessionStore,
    monitoringRepo *MonitoringRepository,
    callbackURL string,
) *Service
```

### 2.2 - Create Event Converter
- [x] Map Strava GroupEvent → CycleScene ride.Submission (`ToSubmission()` method)
- [x] Handle timezone conversion (UTC ISO 8601 → local time via `GetTimezone()`)
- [x] Map city strings to city codes via reverse lookup
- [x] Set defaults for missing fields (120min duration, empty venue name, etc.)
- [x] Integrate geocoding fallback for missing coordinates

### 2.3 - Integrate Route Fetching
- [x] Add route processing method to Service (`ProcessRoute`)
- [x] Reuse existing `routes.RouteFetcher.FetchAndConvert()`
- [x] Handle route deduplication
- [x] Return route ID for linking to event

### 2.4 - Integrate Geocoding Fallback
- [x] Implement geocoding fallback in `Service.ConvertEventToSubmission()`
- [x] Three-tier location handling: `start_latlng` → geocode address → error
- [x] Log when using Strava coordinates vs geocoding

### 2.5 - Write Service Tests
- [x] Create `backend/internal/strava/service_test.go`
- [x] Test OAuth flow (InitiateOAuth, HandleOAuthCallback)
- [x] Test admin club filtering (GetAdminClubs with city-first filtering)
- [x] Test event fetching (GetClubEvents)
- [x] Test event conversion (timezone, city mapping, defaults)

---

## Completion Checklist (COMPLETED 2026-02-02)

**Code Quality:**
- [x] All files created:
  - [x] `service.go` - Service layer with OAuth, city filtering, admin filtering
  - [x] Conversion methods added to `models.go` instead of separate `converter.go`
  - [x] `service_test.go` - Service and model tests (53 tests)
- [x] Existing files updated:
  - [x] `models.go` - Added `CityCode` field to `Session`, conversion methods to `GroupEvent`
  - [x] `session.go` - Added `GenerateStateWithCity()` and `ValidateStateAndGetCity()` methods

**Build & Tests:**
- [x] `cd backend && go build ./cmd/api` - **PASSED**
- [x] `cd backend && go test ./internal/strava/... -v` - **53 tests PASSED**
- [x] `cd backend && go test ./internal/strava/... -cover` - **56.5% coverage**
- [x] `cd backend && go vet ./internal/strava/...` - **PASSED**

---

## Implementation Findings (2026-02-02)

**Architecture Changes from Original Plan:**

1. **No separate converter.go file** - Conversion methods added directly to model structs in `models.go`

2. **City mappings in models.go** - Instead of using `cities.json`, created `cityMappings` var

3. **Session state context combined** - `GenerateStateWithCity()` instead of separate `StoreStateContext()`

4. **Added upcoming events filter** - `GroupEvent.IsUpcoming()` and `FilterUpcomingEvents()`

5. **Route fetching completed in M2** - Originally deferred to M3, but implemented now

**Database Migration:**
- Created migration for event source tracking: `db/main/migrations/1770000000_add_source_tracking_to_events.up.sql`
- Adds `source` and `source_id` columns with partial unique index

**Ready for M3:**
- HTTP handlers to expose service methods
- WebSocket endpoint for multi-event import with progress tracking
- Integration with ride.Service for actual event submission
