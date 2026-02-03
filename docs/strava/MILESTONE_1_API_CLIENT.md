# Milestone 1: Backend - Strava API Client

**Goal:** Create a robust Strava API client for OAuth and data fetching

## Files Summary (Milestone 1)

**New Files to Create:**
- `backend/internal/strava/client.go` - Main Strava API client
- `backend/internal/strava/errors.go` - Custom error types
- `backend/internal/strava/config.go` - Configuration loading
- `backend/internal/strava/monitoring.go` - Monitoring DB repository
- `backend/internal/strava/client_test.go` - Unit tests for client
- `backend/internal/strava/testutil.go` - Reusable test utilities

**Existing Files to Reference:**
- `backend/internal/strava/models.go` - Already has type definitions (Club, ClubDetail, GroupEvent, Route, TokenResponse, etc.)
- `backend/internal/strava/session.go` - Already has session management (SessionStore, GenerateState, CreateSession, etc.)
- `backend/.env.example` - Update with STRAVA_DEBUG

**Files to Update:**
- `backend/.env.example` - Add STRAVA environment variables
- `backend/cmd/api/main.go` - Initialize Strava client in init()

**Database Migration (MUST run before M1):**
- Run SQL migration against monitoring database to create `strava_api_logs` table

**No modifications to:**
- API handlers (not until Milestone 3)
- Frontend code (not until Milestone 4)
- Database schemas
- `go.mod` - All required dependencies already present (net/http, log/slog)

---

## User Flow & API Call Sequence

```
┌─────────────────────────────────────────────────────────────────────┐
│ PHASE 1: OAuth & Admin Detection (Automatic - happens once)        │
└─────────────────────────────────────────────────────────────────────┘
User clicks "Import from Strava"
  → OAuth popup → User authorizes
  → Backend receives OAuth code
  → API Call 1: ExchangeToken(code) → get access_token
  → API Call 2: GetAthleteClubs(token) → returns 10 clubs
  → API Calls 3-12: GetClubDetails(token, clubID) for each club
     - Club 1: admin=false (skip)
     - Club 2: admin=true ✓ (keep)
     - Club 3: admin=false (skip)
     - ...
     - Club 10: owner=true ✓ (keep)
  → Return 2 admin clubs to frontend

Frontend displays:
  [Portland Riders] [View Events →]
  [SLC Cycling Club] [View Events →]

┌─────────────────────────────────────────────────────────────────────┐
│ PHASE 2: Event Fetching (On-Demand - user triggered)               │
└─────────────────────────────────────────────────────────────────────┘
User clicks [View Events →] on "Portland Riders"
  → API Call 13: GetClubEvents(token, clubID=portland) → returns 5 events
  → Frontend displays event list with checkboxes

User clicks [View Events →] on "SLC Cycling Club"
  → API Call 14: GetClubEvents(token, clubID=slc) → returns 3 events
  → Frontend displays event list with checkboxes

┌─────────────────────────────────────────────────────────────────────┐
│ PHASE 3: Import (Milestone 3 - WebSocket)                          │
└─────────────────────────────────────────────────────────────────────┘
User selects 2 events and clicks "Import Selected"
  → WebSocket connection established
  → Import events with progress tracking
  → Routes fetched if needed (GetRoute API calls)
```

**Key Points:**
- **Events are NEVER fetched automatically** - only when user clicks "View Events"
- Admin detection happens once during OAuth (12 calls for 10 clubs example)
- Each club's events = 1 additional API call (lazy loaded)
- Total calls: ~12-20 for typical flow (well under 100/15min limit)

---

## BEFORE STARTING MILESTONE 1

**MANDATORY: Run these steps BEFORE writing any code:**

1. **Create monitoring database table:**
   ```bash
   # Connect to monitoring database
   # Run SQL from section #12 (Database Migration)
   ```

2. **Verify environment:**
   ```bash
   # Ensure you have access to:
   # - Monitoring DB connection
   # - Strava API credentials (for M3, but good to have ready)
   ```

3. **Review all 16 design decisions** in this document

---

## Tasks

### 1.1 - Create Strava Client Package
- [ ] Create `backend/internal/strava/client.go`
- [ ] Implement OAuth token exchange
- [ ] Implement token refresh logic
- [ ] Add rate limit tracking and headers parsing
- [ ] Add debug logging controlled by `STRAVA_DEBUG` env var

**Files to Create/Modify:**
- `backend/internal/strava/client.go` (new)

**Implementation Details:**

**Client Structure:**
```go
type Client struct {
    httpClient   *http.Client
    clientID     string
    clientSecret string
    baseURL      string  // "https://www.strava.com/api/v3"
    debug        bool    // Set from STRAVA_DEBUG env var
}

func NewClient(clientID, clientSecret string) *Client
```

**Methods to Implement:**
1. `ExchangeToken(code string) (*TokenResponse, error)`
   - POST to `https://www.strava.com/oauth/token`
   - Content-Type: `application/x-www-form-urlencoded`
   - Body: `client_id`, `client_secret`, `code`, `grant_type=authorization_code`
   - Returns `TokenResponse` (defined in models.go)
   - Parse rate limit headers: `X-Ratelimit-Limit`, `X-Ratelimit-Usage`, `X-Readratelimit-Limit`

2. `RefreshToken(refreshToken string) (*TokenResponse, error)`
   - POST to `https://www.strava.com/oauth/token`
   - Body: `client_id`, `client_secret`, `refresh_token`, `grant_type=refresh_token`
   - Returns new access token

3. Helper method: `doRequest(req *http.Request) (*http.Response, error)`
   - Add `Authorization: Bearer {token}` header
   - Parse rate limit headers from response
   - Log request/response if debug=true
   - Handle common HTTP errors (401, 429, 404, 500)

**Rate Limit Tracking:**
- Parse headers: `X-Ratelimit-Limit: 200,2000` (15min, daily) - **Note: Actual limits are higher than documented!**
- Parse headers: `X-Ratelimit-Usage: 1,1` (current usage)
- Store in Client struct or return in response
- Log warnings when approaching limits (>80% usage)

**Debug Points:**
- Log OAuth token exchange requests/responses
- Log rate limit usage from headers
- Log all API requests when `STRAVA_DEBUG=true`

**Dependencies:**
- Uses existing `TokenResponse` from `models.go`
- Uses standard library `net/http`
- Uses `log/slog` for structured logging (see pattern in ride/service.go:8)
- Uses `os.Getenv("STRAVA_DEBUG")` to check debug flag

### 1.2 - Implement Club Methods
- [ ] `GetAthleteClubs(accessToken string)` - Fetch all clubs user belongs to
- [ ] `GetClubDetails(accessToken string, clubID int64)` - Get detailed club info with admin flags
- [ ] ~~`FilterAdminClubs(clubs []ClubDetail)`~~ - **REMOVED** - Moved to Service layer (M2)

**Files to Modify:**
- `backend/internal/strava/client.go`

**Implementation Details:**

**Methods to Add to Client:**

1. `GetAthleteClubs(accessToken string) ([]Club, error)`
   - GET `{baseURL}/athlete/clubs`
   - Header: `Authorization: Bearer {accessToken}`
   - Returns array of `Club` (basic info, defined in models.go)
   - Handle pagination if needed (check `per_page`, `page` params)

2. `GetClubDetails(accessToken string, clubID int64) (*ClubDetail, error)`
   - GET `{baseURL}/clubs/{clubID}`
   - Header: `Authorization: Bearer {accessToken}`
   - Returns `ClubDetail` with `admin` and `owner` boolean flags (defined in models.go)
   - This is the KEY method for admin detection!

**Usage Flow (OAuth callback → admin clubs):**
```go
// After OAuth callback, in the service layer:

// 1. Get all clubs user belongs to
clubs, err := client.GetAthleteClubs(session.AccessToken)
// Example: User is member of 10 clubs

// 2. For each club, get detailed info to check admin status
var adminClubs []ClubDetail
for _, club := range clubs {
    detail, err := client.GetClubDetails(session.AccessToken, club.ID)
    if err == nil && (detail.Admin || detail.Owner) {
        adminClubs = append(adminClubs, *detail)
    }
}
// Result: User is admin/owner of 2 clubs (filtered down from 10)

// 3. Return only admin clubs to frontend
// Frontend displays these 2 clubs with "View Events" button
// Events are NOT fetched yet (user hasn't clicked button)
```

**Important Notes:**
- **Events are fetched LATER** when user clicks "View Events" button on a specific club
- This minimizes API calls and respects rate limits
- Admin check happens once during OAuth flow
- Events fetched on-demand per club (lazy loading)

**Debug Points:**
- Log club fetching with athlete ID
- Log admin/owner status for each club
- Log filtered admin clubs count

**API Endpoints Reference:**
- GET `/athlete/clubs` - Returns `[]Club`
- GET `/clubs/{id}` - Returns `ClubDetail` with admin/owner flags

### 1.3 - Implement Event Methods
- [ ] `GetClubEvents(accessToken string, clubID int64)` - Fetch group events for a club
- [ ] `GetRoute(accessToken string, routeID int64)` - Fetch route details (if event has route)

**Files to Modify:**
- `backend/internal/strava/client.go`

**Implementation Details:**

**Methods to Add to Client:**

1. `GetClubEvents(accessToken string, clubID int64) ([]GroupEvent, error)`
   - GET `{baseURL}/clubs/{clubID}/group_events`
   - Header: `Authorization: Bearer {accessToken}`
   - Returns array of `GroupEvent` (defined in models.go)
   - NOTE: This is an UNDOCUMENTED endpoint (discovered during testing)
   - Returns all fields: title, description, event_time, address, lat/lng, route, etc.
   - Events are returned in chronological order (upcoming first)

2. `GetRoute(accessToken string, routeID int64) (*Route, error)`
   - GET `{baseURL}/routes/{routeID}`
   - Header: `Authorization: Bearer {accessToken}`
   - Returns `Route` (defined in models.go)
   - Includes: name, description, distance, elevation_gain, type, sub_type
   - NOTE: This returns metadata only, NOT the GPX track
   - For GPX export, use `/routes/{routeID}/export_gpx` (handled in routes/fetcher.go)

**Error Handling:**
- 404 Not Found: Club has no events OR club doesn't exist
- Return empty array `[]GroupEvent{}` for no events (not an error)
- Log warning if club not found (possible permissions issue)

**Debug Points:**
- Log event fetching per club
- Log route fetching when route_id exists
- Log event counts returned

**API Endpoints Reference:**
- GET `/clubs/{id}/group_events` - Returns `[]GroupEvent` (undocumented!)
- GET `/routes/{id}` - Returns `Route` metadata

### 1.4 - Add Error Handling
- [ ] Handle 401 (invalid/expired token)
- [ ] Handle 429 (rate limit exceeded)
- [ ] Handle 404 (club/event not found)
- [ ] Create custom error types

**Files to Create/Modify:**
- `backend/internal/strava/client.go` (add error handling to doRequest)
- `backend/internal/strava/errors.go` (new - custom error types)

**Implementation Details:**

**Create `errors.go` with custom error types:**
```go
package strava

import "errors"

var (
    ErrUnauthorized      = errors.New("strava: unauthorized or expired token")
    ErrRateLimitExceeded = errors.New("strava: rate limit exceeded")
    ErrNotFound          = errors.New("strava: resource not found")
    ErrForbidden         = errors.New("strava: access forbidden")
    ErrServerError       = errors.New("strava: server error")
)

// APIError wraps Strava API errors with details
type APIError struct {
    StatusCode int
    Message    string
    Err        error
}

func (e *APIError) Error() string {
    return fmt.Sprintf("strava api error %d: %s", e.StatusCode, e.Message)
}

func (e *APIError) Unwrap() error {
    return e.Err
}
```

**Error Handling in `doRequest` method:**
- 401 Unauthorized → Return `ErrUnauthorized` (token invalid/expired)
  - Action: Delete session, redirect user to re-authenticate
- 429 Rate Limit → Return `ErrRateLimitExceeded`
  - Parse `X-RateLimit-*` headers to show time until reset
  - Consider implementing exponential backoff (optional for M1)
- 403 Forbidden → Return `ErrForbidden`
  - User doesn't have permission (e.g., not admin of club)
- 404 Not Found → Return `ErrNotFound`
  - Resource doesn't exist (club, event, route)
- 500/502/503 Server Error → Return `ErrServerError`
  - Transient error, safe to retry

**Debug Points:**
- Log all API errors with status codes
- Log retry attempts for rate limits
- Log error response body when STRAVA_DEBUG=true

**Pattern Reference:**
See `backend/internal/api/common/apierror/apierror.go` for existing error patterns

### 1.5 - Write Unit Tests
- [ ] Test OAuth token exchange
- [ ] Test admin club filtering
- [ ] Test error handling
- [ ] Mock HTTP responses

**Files to Create:**
- `backend/internal/strava/client_test.go` (new)

**Implementation Details:**

**NOTE:** The codebase does not currently have any test files in `backend/internal/`. This task will establish the testing pattern for future milestones.

**Test Structure:**
```go
package strava

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

// Use httptest.NewServer to mock Strava API responses
func TestExchangeToken(t *testing.T) {
    // Mock server returning TokenResponse
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request method, headers, body
        // Return mock TokenResponse JSON
    }))
    defer server.Close()

    client := NewClient("test_id", "test_secret")
    client.baseURL = server.URL  // Point to mock server

    token, err := client.ExchangeToken("test_code")
    // Assert no error, validate token fields
}

func TestGetClubDetails_AdminDetection(t *testing.T) {
    // Test that admin=true and owner=true are correctly detected
}

func TestFilterAdminClubs(t *testing.T) {
    // Test filtering logic (admin=true || owner=true)
}

func TestErrorHandling_401(t *testing.T) {
    // Mock server returns 401
    // Assert ErrUnauthorized is returned
}

func TestErrorHandling_429(t *testing.T) {
    // Mock server returns 429 with rate limit headers
    // Assert ErrRateLimitExceeded is returned
}
```

**Test Cases to Cover:**
1. OAuth token exchange (success)
2. OAuth token refresh (success)
3. Get athlete clubs (returns Club array)
4. Get club details (admin=true, owner=false)
5. Get club details (admin=false, owner=true)
6. Filter admin clubs (returns only admin/owner clubs)
7. Get club events (returns GroupEvent array)
8. Get club events (empty array for no events)
9. Error 401 (unauthorized)
10. Error 429 (rate limit)
11. Error 404 (not found)

**Validation:**
```bash
cd backend
go test ./internal/strava/... -v
go build ./cmd/api
```

**Dependencies:**
- Standard library `testing`
- Standard library `net/http/httptest` for mocking HTTP responses
- No external testing frameworks needed (keep it simple)

---

## Environment Variables & Configuration

**File to Update:** `backend/.env.example`

Add the following section (after the SCRAPER CONFIGURATION section):

```bash
# ============================================================================
# STRAVA INTEGRATION (OPTIONAL - for event import)
# ============================================================================

# Enable verbose debug logging for Strava OAuth and API operations
# Set to 'true' during development to see detailed logs
# MUST be 'false' or unset in production (to avoid logging tokens)
STRAVA_DEBUG=false

# OAuth credentials (will be added in Milestone 3)
# Get these from: https://www.strava.com/settings/api
# STRAVA_CLIENT_ID=your_strava_client_id
# STRAVA_CLIENT_SECRET=your_strava_client_secret
# STRAVA_CALLBACK_URL=http://localhost:8080/v1/strava/auth/callback
```

**Note:** Only `STRAVA_DEBUG` is needed for Milestone 1 (client testing). The OAuth credentials will be added in Milestone 3 when we create the HTTP handlers.

---

## Completion Checklist

Before marking Milestone 1 as complete, verify ALL of the following:

**Code Quality:**
- [ ] All six files created:
  - [ ] `client.go` - API client
  - [ ] `errors.go` - Error types
  - [ ] `config.go` - Configuration loading
  - [ ] `monitoring.go` - Monitoring repository
  - [ ] `client_test.go` - Unit tests
  - [ ] `testutil.go` - Test utilities
- [ ] Monitoring DB table created: `strava_api_logs` (SQL in section #12)
- [ ] Client implements all required methods (with context.Context and metrics):
  - [ ] `NewClient(config *Config) *Client`
  - [ ] `ExchangeToken(ctx, code) (*TokenResponse, *APICallMetrics, error)`
  - [ ] `RefreshToken(ctx, refreshToken) (*TokenResponse, *APICallMetrics, error)`
  - [ ] `GetAthleteClubs(ctx, token) ([]Club, *APICallMetrics, error)`
  - [ ] `GetClubDetails(ctx, token, clubID) (*ClubDetail, *APICallMetrics, error)`
  - [ ] ~~`FilterAdminClubs()`~~ - **REMOVED** (moved to Service layer M2)
  - [ ] `GetClubEvents(ctx, token, clubID) ([]GroupEvent, *APICallMetrics, error)`
  - [ ] `GetRoute(ctx, token, routeID) (*Route, *APICallMetrics, error)`
- [ ] MonitoringRepository implements `LogAPICall()` method
- [ ] Config implements `LoadConfig()` method
- [ ] Client initialized in `cmd/api/main.go`
- [ ] Error handling implemented for 401, 429, 404, 403, 500
- [ ] Rate limit headers parsed and logged
- [ ] **All API calls logged to monitoring DB** (success and errors)
- [ ] Debug logging uses `STRAVA_DEBUG` env var check
- [ ] **Tokens NEVER logged** (access_token, refresh_token redacted)
- [ ] All logging uses `log/slog` structured format

**Build & Tests:**
- [ ] `cd backend && go build ./cmd/api` - **MUST succeed**
- [ ] `cd backend && go test ./internal/strava/... -v` - **MUST pass**
- [ ] `cd backend && go vet ./internal/strava/...` - **MUST have no warnings**

**Documentation:**
- [ ] `STRAVA_DEBUG` added to `backend/.env.example`
- [ ] All debug log points added as documented in each task
- [ ] No `TODO` or `FIXME` comments in code

**Code Patterns:**
- [ ] Follows existing patterns (see `routes/fetcher.go` for HTTP client example)
- [ ] Uses `slog.Info()`, `slog.Warn()`, `slog.Error()`, `slog.Debug()` consistently
- [ ] No hardcoded URLs (use `baseURL` field)
- [ ] No secrets logged (even in debug mode)

**Testing:**
- [ ] Unit tests cover success cases (all 6 methods)
- [ ] Unit tests cover error cases (401, 429, 404, 500)
- [ ] Tests use `httptest.NewServer` for mocking
- [ ] Test utilities in `testutil.go` (MockServer, AssertNoError, etc.)
- [ ] All tests pass with `-v` flag
- [ ] **Testing pattern documented** for retroactive service testing

---

## Design Decisions

### 1. Rate Limit Strategy
**Decision:** Log only for M1 - can add storage later if rate limiting becomes an issue.

### 2. HTTP Client Architecture
**Decision:** Internal, specialized HTTP client with 30s timeout.

### 3. Access Token Management
**Decision:** Stateless Client - tokens passed to each method call.

### 4. Admin Club Filtering Location
**Decision:** Service layer (M2) handles filtering - Client only fetches data.

### 5. Error Handling & Monitoring
**Decision:** Status code handling + monitoring DB tracking for all API calls.

### 6. OAuth Token Refresh Strategy
**Decision:** Client provides `RefreshToken()` method but does NOT auto-refresh. Service layer handles refresh logic.

### 7. Route Fetching Strategy
**Decision:** `GetRoute()` returns metadata only - Reuse existing `routes.RouteFetcher` for GPX.

### 8. Pagination Strategy
**Decision:** No pagination in M1 - Monitor via tracking, add if needed.

### 9. Debug Logging Safety
**Decision:** Never log tokens - only log success/failure and metadata.

### 10. Testing Strategy
**Decision:** Mocked unit tests only for M1 - Integration tests in M6.

---

## Database Migration

**SQL Migration (run against monitoring database before starting M1):**

```sql
CREATE TABLE IF NOT EXISTS strava_api_logs (
    -- Primary key
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    -- Request details
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER,

    -- General rate limit tracking
    rate_limit_15min_usage INTEGER,
    rate_limit_15min_limit INTEGER,
    rate_limit_daily_usage INTEGER,
    rate_limit_daily_limit INTEGER,

    -- Read-only rate limit tracking
    read_limit_15min_usage INTEGER,
    read_limit_15min_limit INTEGER,
    read_limit_daily_usage INTEGER,
    read_limit_daily_limit INTEGER,

    -- Response data (privacy-safe)
    message TEXT,
    clubs_count INTEGER,
    events_count INTEGER,

    -- User context (privacy-safe identifiers only)
    athlete_id INTEGER
);

-- Indexes for common queries
CREATE INDEX idx_strava_status ON strava_api_logs(status_code);
CREATE INDEX idx_strava_endpoint ON strava_api_logs(endpoint);
CREATE INDEX idx_strava_created_at ON strava_api_logs(created_at);
CREATE INDEX idx_strava_athlete ON strava_api_logs(athlete_id);
CREATE INDEX idx_strava_rate_limit ON strava_api_logs(rate_limit_15min_usage, rate_limit_15min_limit);
CREATE INDEX idx_strava_read_limit ON strava_api_logs(read_limit_15min_usage, read_limit_15min_limit);
```
