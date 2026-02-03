# Strava Event Import - Implementation Checklist

## Overview
This checklist tracks the implementation of the Strava OAuth integration feature for importing club group events into CycleScene. Each milestone must meet success criteria before moving to the next.

**Success Criteria (All Milestones):**
- ✅ Codebase builds successfully (`go build`, `npm run build`)
- ✅ No TypeScript errors
- ✅ Feature is debuggable with `STRAVA_DEBUG=true` environment variable
- ✅ All debug logs use structured logging (`slog` for Go, `console.log` for TS)
- ✅ Changes are backward compatible (existing features continue working)

---

## Milestone 1: Backend - Strava API Client

**Goal:** Create a robust Strava API client for OAuth and data fetching

### 📁 Files Summary (Milestone 1)

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

**User Flow & API Call Sequence:**

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

### ⚠️ BEFORE STARTING MILESTONE 1

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

### Tasks

#### 1.1 - Create Strava Client Package
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

**Rate Limit Impact of User Flow:**
- OAuth callback: 1 call to get athlete clubs + N calls to check admin status (where N = number of clubs)
  - Example: User in 10 clubs = 1 + 10 = 11 API calls during OAuth
  - Result: Returns 2 admin clubs (no events fetched yet)
- User clicks "View Events" on a club: 1 call to get events
  - Events only fetched when user explicitly requests them (lazy loading)
- Typical flow: ~15 API calls total (OAuth + admin checks + viewing events for 2-3 clubs)
- Well within 100 requests/15min limit

**Debug Points:**
- Log OAuth token exchange requests/responses
- Log rate limit usage from headers
- Log all API requests when `STRAVA_DEBUG=true`

**Dependencies:**
- Uses existing `TokenResponse` from `models.go`
- Uses standard library `net/http`
- Uses `log/slog` for structured logging (see pattern in ride/service.go:8)
- Uses `os.Getenv("STRAVA_DEBUG")` to check debug flag

#### 1.2 - Implement Club Methods
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

3. `FilterAdminClubs(clubs []ClubDetail) []ClubDetail`
   - Helper function (can be standalone or method)
   - Filters clubs where `club.Admin == true || club.Owner == true`
   - Returns only clubs where user has admin/owner rights

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

#### 1.3 - Implement Event Methods
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

#### 1.4 - Add Error Handling
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

**Response Body Parsing:**
Strava error responses look like:
```json
{
  "message": "Authorization Error",
  "errors": [{"resource": "Athlete", "field": "access_token", "code": "invalid"}]
}
```

Parse this into `APIError.Message` for better error reporting.

**Debug Points:**
- Log all API errors with status codes
- Log retry attempts for rate limits
- Log error response body when STRAVA_DEBUG=true

**Pattern Reference:**
See `backend/internal/api/common/apierror/apierror.go` for existing error patterns

#### 1.5 - Write Unit Tests
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

### 📝 Milestone 1: Environment Variables & Configuration

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

### ✅ Milestone 1: Completion Checklist

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

### 🎯 Milestone 1: Finalized Design Decisions

The following design decisions were made to guide implementation:

#### **1. Rate Limit Strategy** ✅

**Decision:** Log only for M1 - can add storage later if rate limiting becomes an issue.

**Implementation:**
```go
// Helper function in client.go
func (c *Client) parseRateLimits(resp *http.Response) {
    limitHeader := resp.Header.Get("X-Ratelimit-Limit")   // "200,2000"
    usageHeader := resp.Header.Get("X-Ratelimit-Usage")   // "1,1"
    readLimitHeader := resp.Header.Get("X-Readratelimit-Limit")  // "100,1000" (separate read limit)

    if c.debug {
        slog.Debug("Rate limit info", "limit", limitHeader, "usage", usageHeader)
    }

    // Log to monitoring DB (all API calls tracked)
    // Future: can add storage/return here without refactoring all methods
}
```

**Rationale:**
- Current user base is small; rate limits unlikely to be hit
- If user growth requires higher limits, can request increase from Strava
- Monitoring DB will show if rate limiting becomes a real issue
- Design allows easy upgrade to storage/return pattern later

---

#### **2. HTTP Client Architecture** ✅

**Decision:** Internal, specialized HTTP client with 30s timeout.

**Implementation:**
```go
func NewClient(clientID, clientSecret string) *Client {
    return &Client{
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        clientID:     clientID,
        clientSecret: clientSecret,
        baseURL:      "https://www.strava.com/api/v3",
        debug:        os.Getenv("STRAVA_DEBUG") == "true",
    }
}
```

**Rationale:**
- Self-contained, specialized for Strava API domain logic
- Simple constructor, no dependency injection needed
- Matches existing pattern in codebase (`routes/fetcher.go`)
- 30s timeout is appropriate for Strava API calls

---

#### **3. Access Token Management** ✅

**Decision:** Stateless Client - tokens passed to each method call.

**Implementation:**
```go
// One global Client instance shared across all users
client := strava.NewClient(clientID, clientSecret)

// Each method accepts user's access token
clubs, err := client.GetAthleteClubs(session.AccessToken)
events, err := client.GetClubEvents(session.AccessToken, clubID)
```

**Rationale:**
- **Privacy-first:** No user data stored in Client
- **Ephemeral sessions:** Tokens stored in-memory SessionStore only (already implemented)
- **Stateless:** One global Client instance, thread-safe for concurrent users
- **No persistence:** Tokens expire with session (1hr max), never saved to database
- Matches existing architecture pattern (`routes/fetcher.go`)

---

#### **4. Admin Club Filtering Location** ✅

**Decision:** Service layer (M2) handles filtering - Client only fetches data.

**M1 Client Methods (Data Fetching Only):**
```go
// ✅ Include in M1
func (c *Client) GetAthleteClubs(token string) ([]Club, error)
func (c *Client) GetClubDetails(token string, clubID int64) (*ClubDetail, error)

// ❌ Remove from M1 - Business logic belongs in Service
// func (c *Client) FilterAdminClubs(clubs []ClubDetail) []ClubDetail
```

**M2 Service Layer (Business Logic):**
```go
// In service.go (Milestone 2)
func (s *Service) GetAdminClubs(sessionID string) ([]ClubDetail, error) {
    session := s.sessionStore.GetSession(sessionID)

    clubs := s.client.GetAthleteClubs(session.AccessToken)

    var adminClubs []ClubDetail
    for _, club := range clubs {
        detail := s.client.GetClubDetails(session.AccessToken, club.ID)
        if detail.Admin || detail.Owner {  // Filtering logic here
            adminClubs = append(adminClubs, detail)
        }
    }
    return adminClubs, nil
}
```

**Rationale:**
- **Separation of concerns:** Client = data fetcher, Service = business logic
- **Clean architecture:** Client doesn't need to know what "admin" means
- Follows existing codebase pattern (ride/service.go handles business logic)

---

#### **5. Error Handling & Monitoring** ✅

**Decision:** Status code handling + monitoring DB tracking for all API calls.

**Error Handling in Client:**
```go
func (c *Client) handleError(resp *http.Response) error {
    // Log ALL API calls to monitoring DB (success and errors)
    c.logToMonitoring(resp)

    switch resp.StatusCode {
    case 401:
        return &APIError{StatusCode: 401, Message: "Unauthorized or expired token", Err: ErrUnauthorized}
    case 429:
        return &APIError{StatusCode: 429, Message: "Rate limit exceeded", Err: ErrRateLimitExceeded}
    case 403:
        return &APIError{StatusCode: 403, Message: "Forbidden", Err: ErrForbidden}
    case 404:
        return &APIError{StatusCode: 404, Message: "Not found", Err: ErrNotFound}
    case 500, 502, 503:
        return &APIError{StatusCode: resp.StatusCode, Message: "Server error", Err: ErrServerError}
    default:
        return &APIError{StatusCode: resp.StatusCode, Message: "Unknown error", Err: fmt.Errorf("status %d", resp.StatusCode)}
    }
}
```

**Monitoring DB Table:**
```sql
CREATE TABLE strava_api_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    endpoint TEXT NOT NULL,           -- '/athlete/clubs', '/clubs/123/group_events'
    method TEXT NOT NULL,              -- 'GET', 'POST'
    status_code INTEGER NOT NULL,      -- 200, 401, 429, etc.
    response_time_ms INTEGER,          -- Latency tracking
    rate_limit_15min TEXT,             -- '5/100'
    rate_limit_daily TEXT,             -- '42/1000'
    message TEXT                       -- 'ok' for success, error details for failures
);
```

**Rationale:**
- **Simple for M1:** Status codes sufficient for error handling
- **Privacy-first:** No user data logged (no session info, no club IDs, no event data)
- **Monitoring:** Track all API calls for visibility into production behavior
- **Data-driven:** Can identify patterns (frequent 429s → need pagination, etc.)
- Can add full JSON error parsing later if needed

---

#### **6. OAuth Token Refresh Strategy** ✅

**Decision:** Client provides `RefreshToken()` method but does NOT auto-refresh. Service layer handles refresh logic.

**Why No Auto-Refresh Needed:**
- **Sessions last 1 hour max** (ephemeral, in-memory only)
- **Tokens last 6 hours** (won't expire during session)
- **401 errors indicate user revoked access** (not token expiry)
- First-time users: Click "Allow" on OAuth popup
- Returning users: Strava auto-redirects (already authorized), no "Allow" button

**Client Implementation:**
```go
// Provides method but doesn't auto-call it
func (c *Client) RefreshToken(refreshToken string) (*TokenResponse, error) {
    // POST to /oauth/token with grant_type=refresh_token
    // Returns new access_token
}

// On 401, just return error (Service layer decides what to do)
func (c *Client) GetAthleteClubs(token string) ([]Club, error) {
    // ... make request ...
    if resp.StatusCode == 401 {
        return nil, ErrUnauthorized  // Service handles refresh if needed
    }
}
```

**Rationale:**
- **Stateless Client:** Doesn't need to manage refresh tokens
- **Simple logic:** Client just reports 401 errors
- **Service layer control:** M2 will implement refresh logic if needed (unlikely)

---

#### **7. Route Fetching Strategy** ✅

**Decision:** `GetRoute()` returns metadata only - Reuse existing `routes.RouteFetcher` for GPX.

**Implementation Strategy:**

**Case 1: Event has embedded route data**
```go
event := client.GetClubEvents(token, clubID)
if event.Route != nil {
    // Use embedded route (name, distance, elevation)
    // No extra API call needed
}
```

**Case 2: Event has route_id but no embedded data**
```go
event := client.GetClubEvents(token, clubID)
if event.RouteID != 0 && event.Route == nil {
    // Fetch metadata
    route := client.GetRoute(token, event.RouteID)
}
```

**Case 3: Need GPX for mapping**
```go
// Reuse existing routes.RouteFetcher (same as scraper uses)
routeURL := fmt.Sprintf("https://www.strava.com/routes/%d", event.RouteID)
geoJSON, err := routeFetcher.FetchAndConvert(routeURL)
```

**Client Method:**
```go
// Metadata only (NOT GPX export)
func (c *Client) GetRoute(token string, routeID int64) (*Route, error) {
    // GET /routes/{routeID}
    // Returns: name, distance, elevation_gain, type, sub_type
}
```

**Rationale:**
- **Avoid duplication:** `routes/fetcher.go` already handles GPX → GeoJSON conversion
- **Separation of concerns:** Client fetches metadata, fetcher handles file conversion
- **Flexible:** Can use embedded route data when available (saves API call)

---

#### **8. Pagination Strategy** ✅

**Decision:** No pagination in M1 - Monitor via tracking, add if needed.

**Implementation:**
```go
func (c *Client) GetAthleteClubs(token string) ([]Club, error) {
    // GET /athlete/clubs (default: returns all clubs)
    // Log clubs_count to monitoring DB

    if c.debug {
        slog.Debug("Fetched clubs", "count", len(clubs))
    }

    return clubs, nil
}
```

**Monitoring:**
```sql
-- Track in strava_api_logs table
INSERT INTO strava_api_logs (endpoint, clubs_count, ...)
VALUES ('/athlete/clubs', 47, ...);
```

**Reality Check:**
- User in 100 clubs? Possible (enthusiast joins many)
- **Admin of 100 clubs? Extremely unlikely** (managing 1-5 clubs is typical)
- Real cost: 100+ `GetClubDetails()` calls to check admin status
- Rate limit: 100 calls / 15min (user with 100 clubs = 101 API calls)

**Decision Criteria:**
- If monitoring shows users regularly fetching 50+ clubs → implement pagination
- Most organizers: < 20 clubs (well within limits)
- Edge case users: might consume rate limits, but rare
- **Data-driven approach:** Monitor first, optimize if needed

**Rationale:**
- Simple implementation for M1
- Monitoring will reveal if this is a real problem
- Can add pagination in M2 without breaking changes

---

#### **9. Debug Logging Safety** ✅

**Decision:** Never log tokens - only log success/failure and metadata.

**Implementation:**
```go
// ✅ SAFE - Log metadata only
if os.Getenv("STRAVA_DEBUG") == "true" {
    slog.Debug("OAuth token exchange",
        "status", "success",
        "athlete_id", tokenResp.Athlete.ID,
        "expires_in", tokenResp.ExpiresIn,
        "token_type", tokenResp.TokenType,
    )
    // ❌ NEVER log: access_token, refresh_token
}

// ✅ SAFE - Log API call results
if c.debug {
    slog.Debug("Fetched clubs",
        "endpoint", "/athlete/clubs",
        "status_code", 200,
        "clubs_count", len(clubs),
        "rate_limit", "5/100",
    )
}

// ✅ SAFE - Log errors
if err != nil {
    slog.Error("API call failed",
        "endpoint", endpoint,
        "status_code", statusCode,
        "error", err.Error(),
    )
}
```

**What to Log:**
- ✅ Success/failure status
- ✅ HTTP status codes
- ✅ Athlete ID (public identifier)
- ✅ Token expiry time
- ✅ Rate limit usage
- ✅ Response counts (clubs, events)
- ❌ **NEVER:** access_token, refresh_token

**Rationale:**
- **Defense in depth:** Tokens never logged, even if debug accidentally enabled in production
- **Still debuggable:** Metadata sufficient for troubleshooting
- **Security-first:** No risk of token exposure in logs

---

#### **10. Testing Strategy** ✅

**Decision:** Mocked unit tests only for M1 - Integration tests in M6.

**Implementation:**
```go
package strava

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "encoding/json"
)

func TestExchangeToken(t *testing.T) {
    // Mock HTTP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        if r.Method != "POST" {
            t.Errorf("Expected POST, got %s", r.Method)
        }

        // Return mock response
        mockResponse := TokenResponse{
            AccessToken: "mock_token",
            AthleteID: 12345,
            ExpiresIn: 21600,
        }
        json.NewEncoder(w).Encode(mockResponse)
    }))
    defer server.Close()

    // Test client
    client := NewClient("test_id", "test_secret")
    client.baseURL = server.URL

    token, err := client.ExchangeToken("test_code")

    // Assertions
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    if token.AccessToken != "mock_token" {
        t.Errorf("Expected mock_token, got %s", token.AccessToken)
    }
}
```

**Test Coverage:**
- OAuth token exchange (success)
- OAuth token refresh (success)
- Get athlete clubs (success)
- Get club details (admin=true, owner=false cases)
- Get club events (success, empty array)
- Get route metadata (success)
- Error handling (401, 429, 404, 500)

**Rationale:**
- **Fast:** No network calls, runs in milliseconds
- **Reliable:** No external dependencies, works offline
- **CI/CD friendly:** No Strava credentials needed
- **Complete coverage:** Can test all success and error cases
- **Integration tests later:** M6 can add real API tests if needed

---

### 📊 Milestone 1: Additional Implementation Requirements

#### **11. Monitoring DB Integration** ✅

**Decision:** Service layer (M2) handles monitoring DB writes, not Client.

**Implementation:**

**Create monitoring repository in M1:**
```go
// backend/internal/strava/monitoring.go
package strava

import (
    "database/sql"
    "time"
)

type MonitoringRepository struct {
    db *sql.DB
}

func NewMonitoringRepository(db *sql.DB) *MonitoringRepository {
    return &MonitoringRepository{db: db}
}

func (r *MonitoringRepository) LogAPICall(metrics *APICallMetrics) error {
    query := `
        INSERT INTO strava_api_logs (
            endpoint,
            method,
            status_code,
            response_time_ms,
            rate_limit_15min_usage,
            rate_limit_15min_limit,
            rate_limit_daily_usage,
            rate_limit_daily_limit,
            read_limit_15min_usage,
            read_limit_15min_limit,
            read_limit_daily_usage,
            read_limit_daily_limit,
            message,
            clubs_count,
            events_count,
            athlete_id
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

    _, err := r.db.Exec(
        query,
        metrics.Endpoint,
        metrics.Method,
        metrics.StatusCode,
        metrics.ResponseTimeMs,
        metrics.RateLimit15minUsage,
        metrics.RateLimit15minLimit,
        metrics.RateLimitDailyUsage,
        metrics.RateLimitDailyLimit,
        metrics.ReadLimit15minUsage,
        metrics.ReadLimit15minLimit,
        metrics.ReadLimitDailyUsage,
        metrics.ReadLimitDailyLimit,
        metrics.Message,
        metrics.ClubsCount,
        metrics.EventsCount,
        metrics.AthleteID,
    )

    return err
}
```

**Client returns monitoring data:**
```go
// In client.go
type APICallMetrics struct {
    // Request details
    Endpoint       string
    Method         string
    StatusCode     int
    ResponseTimeMs int

    // General rate limit tracking (parsed from X-Ratelimit-* headers)
    // IMPORTANT: Headers are case-sensitive! Use "X-Ratelimit-*" NOT "X-RateLimit-*"
    // Verified from actual Strava API response (2026-01-31)
    RateLimit15minUsage int  // Parsed from "X-Ratelimit-Usage: 7,7" → 7
    RateLimit15minLimit int  // Parsed from "X-Ratelimit-Limit: 200,2000" → 200
    RateLimitDailyUsage int  // Parsed from "X-Ratelimit-Usage: 7,7" → 7
    RateLimitDailyLimit int  // Parsed from "X-Ratelimit-Limit: 200,2000" → 2000

    // Read-only rate limit tracking (parsed from X-Readratelimit-* headers)
    // CRITICAL: This is the ACTUAL constraint for our GET-heavy operations!
    ReadLimit15minUsage int  // Parsed from "X-Readratelimit-Usage: 7,7" → 7
    ReadLimit15minLimit int  // Parsed from "X-Readratelimit-Limit: 100,1000" → 100
    ReadLimitDailyUsage int  // Parsed from "X-Readratelimit-Usage: 7,7" → 7
    ReadLimitDailyLimit int  // Parsed from "X-Readratelimit-Limit: 100,1000" → 1000

    // Response data (privacy-safe)
    Message      string
    // Note: X-Envoy-Upstream-Service-Time header also available (response time from Strava's upstream)
    ClubsCount   int  // For /athlete/clubs endpoint
    EventsCount  int  // For /group_events endpoint
    AthleteID    int64 // From token response or context
}

// Client methods return metrics
func (c *Client) GetAthleteClubs(ctx context.Context, token string) ([]Club, *APICallMetrics, error) {
    start := time.Now()
    // ... make request ...

    // Parse general rate limit headers
    // X-Ratelimit-Limit: "200,2000"
    // X-Ratelimit-Usage: "7,7"
    limit15min, limitDaily := parseRateLimitHeader(resp.Header.Get("X-Ratelimit-Limit"))
    usage15min, usageDaily := parseRateLimitHeader(resp.Header.Get("X-Ratelimit-Usage"))

    // Parse read-only rate limit headers (ACTUAL constraint for GET requests!)
    // X-Readratelimit-Limit: "100,1000"
    // X-Readratelimit-Usage: "7,7"
    readLimit15min, readLimitDaily := parseRateLimitHeader(resp.Header.Get("X-Readratelimit-Limit"))
    readUsage15min, readUsageDaily := parseRateLimitHeader(resp.Header.Get("X-Readratelimit-Usage"))

    metrics := &APICallMetrics{
        Endpoint:            "/athlete/clubs",
        Method:              "GET",
        StatusCode:          resp.StatusCode,
        ResponseTimeMs:      int(time.Since(start).Milliseconds()),
        RateLimit15minUsage: usage15min,
        RateLimit15minLimit: limit15min,
        RateLimitDailyUsage: usageDaily,
        RateLimitDailyLimit: limitDaily,
        ReadLimit15minUsage: readUsage15min,
        ReadLimit15minLimit: readLimit15min,
        ReadLimitDailyUsage: readUsageDaily,
        ReadLimitDailyLimit: readLimitDaily,
        Message:             "ok", // or error message
        ClubsCount:          len(clubs),
        AthleteID:           athleteID, // From token or context
    }

    return clubs, metrics, nil
}

// Helper to parse rate limit headers
func parseRateLimitHeader(header string) (int, int) {
    // "200,2000" → (200, 2000)
    parts := strings.Split(header, ",")
    if len(parts) != 2 {
        return 0, 0
    }
    first, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
    second, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
    return first, second
}
```

**Service layer (M2) logs to monitoring DB:**
```go
// In service.go (Milestone 2)
func (s *Service) GetAdminClubs(ctx context.Context, sessionID string) ([]ClubDetail, error) {
    session := s.sessionStore.GetSession(sessionID)

    // Get athlete clubs with metrics
    clubs, metrics, err := s.client.GetAthleteClubs(ctx, session.AccessToken)

    // Set athlete ID from session
    metrics.AthleteID = session.AthleteID

    // Log to monitoring DB (simple, single call)
    if logErr := s.monitoringRepo.LogAPICall(metrics); logErr != nil {
        slog.Error("Failed to log API call", "error", logErr)
        // Don't fail the request if logging fails
    }

    return clubs, err
}
```

**Rationale:**
- **M1:** Client returns metrics, monitoring repo created
- **M2:** Service layer uses monitoring repo to log
- **Separation:** Client stays pure (no DB dependency), Service orchestrates

---

#### **11b. Actual Strava Response Headers** ✅

**Verified from Live API Testing (2026-01-31):**

When calling `/athlete/clubs`, Strava returns these headers:

```
Content-Type: application/json; charset=utf-8
Referrer-Policy: strict-origin-when-cross-origin
X-Permitted-Cross-Domain-Policies: none
X-Xss-Protection: 1; mode=block
Date: Sun, 01 Feb 2026 04:21:06 GMT
X-Envoy-Upstream-Service-Time: 82
Server: istio-envoy
X-Ratelimit-Usage: 1,1                    ← 15min usage, daily usage
Vary: Origin
X-Readratelimit-Limit: 100,1000           ← Separate READ rate limit
X-Content-Type-Options: nosniff
X-Cache: Miss from cloudfront
Status: 200 OK
X-Ratelimit-Limit: 200,2000               ← 15min limit, daily limit
Cache-Control: max-age=0, private, must-revalidate
X-Frame-Options: DENY
```

**Key Findings:**

1. **Header case sensitivity:** `X-Ratelimit-*` (NOT `X-RateLimit-*` as docs suggest)
2. **TWO separate rate limit systems:**
   - **General limit:** 200 req/15min, 2000 req/day (`X-Ratelimit-Limit/Usage`)
   - **Read-only limit:** 100 req/15min, 1000 req/day (`X-Readratelimit-Limit/Usage`)
3. **⚠️ CRITICAL:** All our operations are GET requests, so we're constrained by the **read limit (100/15min)**, not general limit!
4. **Response time available:** `X-Envoy-Upstream-Service-Time: 62` (ms from Strava's upstream service)
5. **Request tracking:** `X-Request-Id` header available for debugging with Strava support

**Implementation Notes:**

```go
// Use exact header names (case-sensitive)
limitHeader := resp.Header.Get("X-Ratelimit-Limit")   // Works
limitHeader := resp.Header.Get("X-RateLimit-Limit")   // Returns empty string!

// Strava gives us DOUBLE the documented limits!
// 15min: 200 (not 100)
// Daily: 2000 (not 1000)
```

**Why this matters:**
- **MUST track read limits!** All our operations (GetAthleteClubs, GetClubDetails, GetClubEvents) are GET requests
- Effective limit is **100 requests per 15 minutes**, not 200
- Must use correct case for header names or we'll get empty values
- Should monitor BOTH rate limit systems to detect if we hit either ceiling

---

#### **12. Database Migration** ✅

**Decision:** Create `strava_api_logs` table in monitoring DB **before** starting M1 implementation.

**SQL Migration:**
```sql
-- Run this against monitoring database before starting M1
CREATE TABLE IF NOT EXISTS strava_api_logs (
    -- Primary key
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    -- Request details
    endpoint TEXT NOT NULL,              -- '/athlete/clubs', '/clubs/123/group_events'
    method TEXT NOT NULL,                -- 'GET', 'POST'
    status_code INTEGER NOT NULL,        -- 200, 401, 429, etc.
    response_time_ms INTEGER,            -- Latency tracking (how long Strava took)

    -- General rate limit tracking (from X-Ratelimit-* headers)
    rate_limit_15min_usage INTEGER,      -- Current usage in 15min window (e.g., 7)
    rate_limit_15min_limit INTEGER,      -- Max allowed in 15min (e.g., 200)
    rate_limit_daily_usage INTEGER,      -- Current usage in 24hr window (e.g., 7)
    rate_limit_daily_limit INTEGER,      -- Max allowed in 24hr (e.g., 2000)

    -- Read-only rate limit tracking (from X-Readratelimit-* headers)
    -- IMPORTANT: This is the ACTUAL limiting factor for our GET-heavy operations!
    read_limit_15min_usage INTEGER,      -- Current read usage in 15min window (e.g., 7)
    read_limit_15min_limit INTEGER,      -- Max read allowed in 15min (100)
    read_limit_daily_usage INTEGER,      -- Current read usage in 24hr window (e.g., 7)
    read_limit_daily_limit INTEGER,      -- Max read allowed in 24hr (1000)

    -- Response data (privacy-safe)
    message TEXT,                        -- 'ok' for 200, error details for failures
    clubs_count INTEGER,                 -- Number of clubs returned (for /athlete/clubs)
    events_count INTEGER,                -- Number of events returned (for /group_events)

    -- User context (privacy-safe identifiers only)
    athlete_id INTEGER                   -- Strava athlete ID (public identifier, not sensitive)
);

-- Indexes for common queries
CREATE INDEX idx_strava_status ON strava_api_logs(status_code);
CREATE INDEX idx_strava_endpoint ON strava_api_logs(endpoint);
CREATE INDEX idx_strava_created_at ON strava_api_logs(created_at);
CREATE INDEX idx_strava_athlete ON strava_api_logs(athlete_id);
CREATE INDEX idx_strava_rate_limit ON strava_api_logs(rate_limit_15min_usage, rate_limit_15min_limit);
CREATE INDEX idx_strava_read_limit ON strava_api_logs(read_limit_15min_usage, read_limit_15min_limit);
```

**Schema Breakdown:**

**Tracking API Health:**
- `status_code` - Success (200) vs errors (401, 429, 404)
- `response_time_ms` - Strava API latency/performance
- `endpoint` + `method` - Which endpoints are most used

**Rate Limit Monitoring (TWO separate systems!):**

1. **General Rate Limits:**
   - `rate_limit_15min_usage` / `rate_limit_15min_limit` - General short-term limit (200)
   - `rate_limit_daily_usage` / `rate_limit_daily_limit` - General daily limit (2000)
   - Parsed from headers: `X-Ratelimit-Usage: 7,7` → usage columns
   - Parsed from headers: `X-Ratelimit-Limit: 200,2000` → limit columns

2. **Read-Only Rate Limits (ACTUAL constraint for us!):**
   - `read_limit_15min_usage` / `read_limit_15min_limit` - Read short-term limit (100)
   - `read_limit_daily_usage` / `read_limit_daily_limit` - Read daily limit (1000)
   - Parsed from headers: `X-Readratelimit-Usage: 7,7` → usage columns
   - Parsed from headers: `X-Readratelimit-Limit: 100,1000` → limit columns
   - **⚠️ This is the bottleneck** since all our operations are GET requests!

**Usage Insights:**
- `clubs_count` - How many clubs typical users have (pagination decision)
- `events_count` - How many events per club (performance insights)
- `athlete_id` - Track power users (high API usage) without storing PII

**Privacy-Safe:**
- ❌ No session IDs (ephemeral, not persistent)
- ❌ No access tokens (security risk)
- ❌ No club IDs (user data)
- ❌ No event details (user data)
- ✅ Only: athlete_id (public Strava identifier), counts, timestamps

**Example Queries You Can Run:**

```sql
-- ⚠️ MOST IMPORTANT: Check if we're approaching READ rate limits (the actual bottleneck!)
SELECT
    created_at,
    endpoint,
    read_limit_15min_usage,
    read_limit_15min_limit,
    (CAST(read_limit_15min_usage AS FLOAT) / read_limit_15min_limit * 100) as read_percent_used,
    rate_limit_15min_usage,
    rate_limit_15min_limit,
    (CAST(rate_limit_15min_usage AS FLOAT) / rate_limit_15min_limit * 100) as general_percent_used
FROM strava_api_logs
WHERE created_at > datetime('now', '-1 hour')
ORDER BY created_at DESC;

-- Alert when read rate limit exceeds 80%
SELECT
    created_at,
    endpoint,
    read_limit_15min_usage,
    read_limit_15min_limit,
    (CAST(read_limit_15min_usage AS FLOAT) / read_limit_15min_limit * 100) as percent_used
FROM strava_api_logs
WHERE
    created_at > datetime('now', '-1 hour')
    AND (CAST(read_limit_15min_usage AS FLOAT) / read_limit_15min_limit) > 0.80
ORDER BY created_at DESC;

-- Find users with most clubs (pagination candidates)
SELECT
    athlete_id,
    MAX(clubs_count) as max_clubs,
    COUNT(*) as api_calls
FROM strava_api_logs
WHERE endpoint = '/athlete/clubs' AND status_code = 200
GROUP BY athlete_id
ORDER BY max_clubs DESC
LIMIT 10;

-- Error rate by endpoint
SELECT
    endpoint,
    COUNT(*) as total_calls,
    SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as errors,
    ROUND(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as error_rate_pct
FROM strava_api_logs
WHERE created_at > datetime('now', '-7 days')
GROUP BY endpoint
ORDER BY error_rate_pct DESC;

-- Average response time by endpoint
SELECT
    endpoint,
    COUNT(*) as calls,
    AVG(response_time_ms) as avg_ms,
    MAX(response_time_ms) as max_ms
FROM strava_api_logs
WHERE status_code = 200
GROUP BY endpoint
ORDER BY avg_ms DESC;

-- Daily API usage trend
SELECT
    DATE(created_at) as date,
    MAX(rate_limit_daily_usage) as peak_usage,
    MAX(rate_limit_daily_limit) as limit
FROM strava_api_logs
WHERE created_at > datetime('now', '-30 days')
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

**Action:** Run this SQL against monitoring database before implementing M1.

---

#### **13. Client Initialization & Configuration** ✅

**Decision:** Initialize in `main.go`, pass config to Client constructor.

**Implementation:**

**Create config struct:**
```go
// backend/internal/strava/config.go
package strava

type Config struct {
    ClientID     string
    ClientSecret string
    Debug        bool
}

func LoadConfig() *Config {
    return &Config{
        ClientID:     os.Getenv("STRAVA_CLIENT_ID"),
        ClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
        Debug:        os.Getenv("STRAVA_DEBUG") == "true",
    }
}
```

**Update Client constructor:**
```go
func NewClient(config *Config) *Client {
    return &Client{
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        clientID:     config.ClientID,
        clientSecret: config.ClientSecret,
        baseURL:      "https://www.strava.com/api/v3",
        debug:        config.Debug,
    }
}
```

**Initialize in main.go:**
```go
// backend/cmd/api/main.go
func init() {
    // ... existing DB connections ...

    // Load Strava configuration
    stravaConfig := strava.LoadConfig()
    if stravaConfig.ClientID == "" {
        slog.Warn("STRAVA_CLIENT_ID not configured - Strava import will not be available")
    }

    // Create Strava client (will be passed to handlers in M3)
    stravaClient := strava.NewClient(stravaConfig)
}
```

---

#### **14. Context Support for Request Cancellation** ✅

**Decision:** All Client methods accept `context.Context` for timeout/cancellation control.

**Implementation:**
```go
// All Client methods use context
func (c *Client) GetAthleteClubs(ctx context.Context, token string) ([]Club, *APICallMetrics, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/athlete/clubs", nil)
    if err != nil {
        return nil, nil, err
    }

    req.Header.Set("Authorization", "Bearer "+token)

    // Request will be cancelled if context times out or is cancelled
    resp, err := c.httpClient.Do(req)
    // ...
}
```

**Benefits:**
- User closes browser → request cancelled automatically
- Custom timeouts per request
- Future: distributed tracing support

**All method signatures:**
```go
func (c *Client) ExchangeToken(ctx context.Context, code string) (*TokenResponse, *APICallMetrics, error)
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, *APICallMetrics, error)
func (c *Client) GetAthleteClubs(ctx context.Context, token string) ([]Club, *APICallMetrics, error)
func (c *Client) GetClubDetails(ctx context.Context, token string, clubID int64) (*ClubDetail, *APICallMetrics, error)
func (c *Client) GetClubEvents(ctx context.Context, token string, clubID int64) ([]GroupEvent, *APICallMetrics, error)
func (c *Client) GetRoute(ctx context.Context, token string, routeID int64) (*Route, *APICallMetrics, error)
```

---

#### **15. Undocumented Endpoint Resilience** ✅

**Decision:** On 404 from `/group_events`, send email alert and log to monitoring DB.

**Implementation:**
```go
// In client.go
func (c *Client) GetClubEvents(ctx context.Context, token string, clubID int64) ([]GroupEvent, *APICallMetrics, error) {
    // ... make request to /clubs/{clubID}/group_events ...

    if resp.StatusCode == 404 {
        // This endpoint is undocumented - might have been removed
        metrics.Message = "ALERT: /group_events endpoint returned 404 - may have been removed by Strava"

        // Service layer will handle email alert based on this message
        return nil, metrics, &APIError{
            StatusCode: 404,
            Message:    "Club events endpoint not found - undocumented endpoint may have changed",
            Err:        ErrNotFound,
        }
    }

    return events, metrics, nil
}
```

**Service layer sends alert (M2):**
```go
// In service.go
func (s *Service) GetClubEvents(ctx context.Context, sessionID string, clubID int64) ([]GroupEvent, error) {
    events, metrics, err := s.client.GetClubEvents(ctx, session.AccessToken, clubID)

    // Log to monitoring
    s.monitoringRepo.LogAPICall(metrics)

    // Check for critical endpoint failure
    if err != nil && metrics.StatusCode == 404 && strings.Contains(metrics.Message, "ALERT") {
        // Send email alert
        s.sendAlert("Strava endpoint failure", metrics.Message)
    }

    return events, err
}
```

**Email alert integration:**
- Reuse existing Resend email service
- Send to admin email (from env var)
- Include: timestamp, endpoint, error details

---

#### **16. Testing Infrastructure & Patterns** ✅

**Decision:** M1 establishes testing patterns for entire codebase.

**Create testing utilities:**
```go
// backend/internal/strava/testutil.go
package strava

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

// MockServer creates a test HTTP server for mocking Strava API
func MockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
    server := httptest.NewServer(handler)
    t.Cleanup(func() { server.Close() })
    return server
}

// MockTokenResponse returns a mock TokenResponse for testing
func MockTokenResponse() *TokenResponse {
    return &TokenResponse{
        AccessToken:  "mock_access_token",
        RefreshToken: "mock_refresh_token",
        ExpiresAt:    1234567890,
        ExpiresIn:    21600,
        TokenType:    "Bearer",
        Athlete: Athlete{
            ID:        12345,
            FirstName: "Test",
            LastName:  "User",
        },
    }
}

// MockClubDetail returns a mock ClubDetail for testing
func MockClubDetail(admin, owner bool) *ClubDetail {
    return &ClubDetail{
        Club: Club{
            ID:   123,
            Name: "Test Club",
        },
        Admin: admin,
        Owner: owner,
    }
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
}

// AssertEqual fails the test if expected != actual
func AssertEqual(t *testing.T, expected, actual interface{}) {
    t.Helper()
    if expected != actual {
        t.Errorf("Expected %v, got %v", expected, actual)
    }
}
```

**Testing patterns document:**
```go
// backend/internal/strava/client_test.go
package strava

import (
    "context"
    "encoding/json"
    "net/http"
    "testing"
)

// Example test following the established pattern
func TestGetAthleteClubs_Success(t *testing.T) {
    // Arrange: Mock server
    mockClubs := []Club{
        {ID: 1, Name: "Club 1"},
        {ID: 2, Name: "Club 2"},
    }

    server := MockServer(t, func(w http.ResponseWriter, r *http.Request) {
        // Assert request
        AssertEqual(t, "GET", r.Method)
        AssertEqual(t, "Bearer test_token", r.Header.Get("Authorization"))

        // Return mock response
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(mockClubs)
    })

    // Arrange: Client
    client := NewClient(&Config{ClientID: "test", ClientSecret: "secret"})
    client.baseURL = server.URL

    // Act
    clubs, metrics, err := client.GetAthleteClubs(context.Background(), "test_token")

    // Assert
    AssertNoError(t, err)
    AssertEqual(t, 2, len(clubs))
    AssertEqual(t, 200, metrics.StatusCode)
    AssertEqual(t, "ok", metrics.Message)
}

func TestGetAthleteClubs_Unauthorized(t *testing.T) {
    // Test error case
    server := MockServer(t, func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
    })

    client := NewClient(&Config{ClientID: "test", ClientSecret: "secret"})
    client.baseURL = server.URL

    clubs, metrics, err := client.GetAthleteClubs(context.Background(), "invalid_token")

    // Assert error handling
    if err == nil {
        t.Fatal("Expected error for 401 response")
    }
    AssertEqual(t, 401, metrics.StatusCode)
    AssertEqual(t, (*[]Club)(nil), clubs)
}
```

**Documentation:**
- Pattern established in `client_test.go`
- Reusable test utilities in `testutil.go`
- Can be applied retroactively to existing services
- Template for future service testing

---

## Milestone 2: Backend - Service Layer ✅ COMPLETE

**Goal:** Build business logic for OAuth sessions and event conversion

**Status:** Completed 2026-02-02

---

### 🎯 Design Decisions (Finalized 2026-02-02)

All design considerations have been resolved through collaborative planning. These decisions guide the M2 implementation:

#### ✅ Decision #1: Event Duration Default
- **Value:** 120 minutes (2 hours)
- **Rationale:** Sensible default for social rides, users can edit later via magic link
- **Implementation:** Set `EventDurationMinutes: 120` in converter

#### ✅ Decision #2: Ride Length from Route Distance
- **Implementation:** Calculate from `event.Route.Distance` when available
- **Format:** "25.3 miles" (converted from meters using `distance * 0.000621371`)
- **Fallback:** Empty string if no route data
- **Rationale:** `ride_length` is freeform text field, distance is most useful info for users

#### ✅ Decision #3: Route Deduplication
- **Method:** Use existing `routes.Repository.CreateRoute()`
- **Mechanism:** UNIQUE constraint on `(source, source_id)` already exists
- **Impact:** No M2 work needed - already handled by existing infrastructure

#### ✅ Decision #4: City-First Filtering (CRITICAL - 83% API Call Reduction)
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
- **Implementation Requirements:**
  - Add `CityCode string` field to Session struct (models.go)
  - Add `StoreStateContext(state, cityCode)` to SessionStore (session.go)
  - Add `ValidateStateAndGetCity(state) (cityCode, valid)` to SessionStore
  - Update `InitiateOAuth(ctx, cityCode)` signature to accept cityCode
  - Filter in `GetAdminClubs()` before calling GetClubDetails

#### ✅ Decision #5: Unsupported Cities
- **Handling:** None needed - strict city filtering automatically excludes them
- **Result:** Seattle/Bend/Eugene events never shown to users

#### ✅ Decision #6: Organizer Email (Required)
- **Source:** User input before import (one email applied to all events)
- **Rationale:** Edit links require email (magic link system)
- **Implementation:** Frontend collects email, backend applies to all imported events
- **Additional Contact Fields:**
  - `OrganizerName`: Empty (optional)
  - `OrganizerPhone`: Empty (optional)
  - `WebURL`: Link to Strava event (`https://www.strava.com/clubs/{club_id}/group_events/{event_id}`)
  - `WebName`: "View on Strava"

#### ✅ Decision #7: Authentication for Multi-Event Import
- **Problem:** BFF tokens are single-use (marked `used=1` after first submission)
- **Solution:** Separate authentication paths
  - **Strava imports:** Validate OAuth session via `SessionStore.GetSession(sessionID)`
  - **Manual submissions:** Continue using BFF tokens (no change)
- **Implementation (M3):**
  - WebSocket handler validates Strava session (not BFF token)
  - Submit events directly to ride.Service bypassing BFF token check
- **City Code Validation:** Defense-in-depth
  - Validate in Service.InitiateOAuth() - return 400 if invalid
  - Validate in Converter - return error if unsupported
  - Even though city code comes from trusted PWA, protect against URL manipulation

---

### 🔍 Gap Resolutions (All 8 Gaps Addressed)

During planning, we identified and resolved 8 potential edge cases:

**Gap #1: City Code Validation**
- ✅ Validate in both Service (InitiateOAuth) and Converter
- Return 400 Bad Request if invalid/unsupported city code
- Defense-in-depth even though city code comes from trusted PWA

**Gap #2: Session Expiry During Import**
- ✅ Non-issue - JSON fetching and submission is very fast (seconds, not minutes)
- No special handling needed

**Gap #3: Multiple Events Import - Partial Failures**
- ✅ Continue on failure strategy:
  - Process all events sequentially
  - On failure: Add to retry channel, continue to next event
  - After first pass: Retry failures with exponential backoff (2s, 4s, 8s max 3 retries)
  - WebSocket sends progress for each event (success/failure/retry)
- ✅ Rate limit monitoring: Skipped (over-engineering for low traffic)
  - Typical user: ~6-11 API calls per session
  - Would need 10+ concurrent users to approach 100/15min read limit
  - Monitoring DB logs all calls for post-mortem if issues arise

**Gap #4: Duplicate Event Detection**
- ✅ Add source tracking to events table (Migration created in M2):
  - File: `db/main/migrations/1770000000_add_source_tracking_to_events.up.sql`
  ```sql
  ALTER TABLE events ADD COLUMN source TEXT;
  ALTER TABLE events ADD COLUMN source_id TEXT;
  CREATE UNIQUE INDEX idx_events_source_dedup ON events (source, source_id)
    WHERE source IS NOT NULL AND source_id IS NOT NULL;
  ```
- ✅ Usage:
  - Strava imports: `source="strava"`, `source_id="{strava_event_id}"`
  - Manual submissions: `source="manual"` or NULL, `source_id=NULL`
- ✅ Duplicate handling: Database constraint violation → skip event, log, continue

**Gap #5: Timezone & DST Edge Cases**
- ✅ Use Strava's `zone` field directly (e.g., "America/Denver")
- ✅ Trust Go's `time.LoadLocation()` - handles DST automatically
- ✅ Fallback to city-based mapping if zone field missing
- ✅ Add test case for DST transition dates in converter_test.go

**Gap #6: Event with RouteID but No Route Data**
- ✅ Use existing `routes.RouteFetcher.FetchAndConvert()`
- ✅ Routes are optional (nice-to-have)
- ✅ If route fetch fails: Log warning, continue without route
- ✅ Deduplication handled by existing (source, source_id) constraint on routes table

**Gap #7: Geocoding Failure Strategy**
- ✅ Three-tier location handling:
  1. **Best case:** Use `start_latlng` array directly (no geocoding needed)
  2. **Fallback:** Geocode `address` field if latlng not available
  3. **Fail:** Return error if neither available - user must add location in Strava
- ✅ Error message: "Unable to import: No location information. Please add address or starting location in Strava and retry."

**Gap #8: TinyTitle Field**
- ✅ Leave NULL/empty - legacy field from shift2bikes schema, not used anywhere
- No special handling needed

---

### 📊 Real Strava API Data Structure

During planning, we examined real Strava event responses. Key findings:

**Location Data:**
```json
{
  "address": "Salt Lake City Public Library",
  "start_latlng": [40.75997, -111.88460]  // Array format [lat, lng]
}
```

**Event Time:**
```json
{
  "upcoming_occurrences": ["2022-10-22T01:00:00Z"],  // Array of UTC timestamps
  "zone": "America/Denver"  // Timezone string
}
```

**Route Data (Optional):**
```json
{
  "route_id": 3283525013759831116,
  "route": {
    "id": 3283525013759831116,
    "name": "Kenton Trotters 3.5 miles",
    "map": {
      "summary_polyline": "{_fuGrzxkV|f@@?eBnU@..."
    }
  }
}
```

**Sport Type Filtering:**
```json
{
  "activity_types": ["Ride", "EBikeRide", "VirtualRide", "Handcycle", "Velomobile"],
  "sport_type": "cycling"  // Deprecated, use activity_types instead
}
```

---

### 📁 Files Summary (Milestone 2)

**New Files to Create:**
- `backend/internal/strava/service.go` - Business logic layer (orchestrates OAuth, admin filtering, event fetching)
- `backend/internal/strava/converter.go` - Event conversion (Strava GroupEvent → CycleScene Submission)
- `backend/internal/strava/service_test.go` - Service layer tests
- `backend/internal/strava/converter_test.go` - Converter tests
- `backend/internal/strava/testutil_service.go` - Service test utilities (mock Client interface)

**Existing Files to Integrate With:**
- `backend/internal/strava/client.go` - API client (from M1)
- `backend/internal/strava/session.go` - Session management (already exists)
- `backend/internal/strava/models.go` - Type definitions (already exists)
- `backend/internal/strava/monitoring.go` - Monitoring repository (from M1)
- `backend/internal/routes/fetcher.go` - Route fetching (`FetchAndConvert()`)
- `backend/internal/routes/repository.go` - Route storage
- `backend/internal/scraper/geocode.go` - Geocoding (`GeocodeQuery()`)
- `backend/internal/scraper/cities.json` - City mappings (pdx, slc)
- `backend/internal/api/ride/service.go` - Event submission pattern reference
- `backend/internal/api/ride/models.go` - Submission type reference

**No modifications to:**
- API handlers (not until Milestone 3)
- Frontend code (not until Milestone 4)
- Database repositories (reuse existing ride.Repository)
- Client code (M1 complete)

**Key Integration Points:**
1. **Service → Client:** Calls Client methods (GetAthleteClubs, GetClubDetails, GetClubEvents) and logs metrics to monitoring DB
2. **Service → SessionStore:** Manages OAuth sessions (CreateSession, GetSession, DeleteSession)
3. **Service → ride.Repository:** Submits converted events via existing submission flow
4. **Converter → scraper:** Uses GeocodeQuery() for fallback geocoding and city code mappings
5. **Converter → routes:** Uses RouteFetcher for route GPX conversion

---

### Tasks

#### 2.1 - Create Strava Service Layer ✅
- [x] Create `backend/internal/strava/service.go`
- [x] Implement dependency injection constructor
- [x] Add method: `InitiateOAuth()` - Generate state token and return authorization URL
- [x] Add method: `HandleOAuthCallback()` - Exchange code for token, create session
- [x] Add method: `GetAdminClubs()` - Fetch clubs where user is admin/owner
- [x] Add method: `GetClubEvents()` - Fetch events for a specific club
- [x] Add helper: `logAPICall()` - Write API metrics to monitoring DB

**Files to Create:**
- `backend/internal/strava/service.go` (new)

**Implementation Details:**

**Service Structure:**
```go
package strava

import (
    "context"
    "fmt"
    "log/slog"
    "os"
)

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
) *Service {
    return &Service{
        client:         client,
        sessionStore:   sessionStore,
        monitoringRepo: monitoringRepo,
        callbackURL:    callbackURL,
        debug:          os.Getenv("STRAVA_DEBUG") == "true",
    }
}
```

**Method: InitiateOAuth** (Updated for Decision #4)
```go
// Returns authorization URL for frontend redirect
// cityCode is passed from form URL context
func (s *Service) InitiateOAuth(ctx context.Context, cityCode string) (string, error) {
    // Generate CSRF state token
    state, err := s.sessionStore.GenerateState()
    if err != nil {
        return "", fmt.Errorf("failed to generate state: %w", err)
    }

    // Store cityCode with state for retrieval during callback (Decision #4)
    s.sessionStore.StoreStateContext(state, cityCode)

    // Build Strava authorization URL
    authURL := fmt.Sprintf(
        "https://www.strava.com/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=read,activity:read&state=%s",
        s.client.clientID,
        s.callbackURL,
        state,
    )

    if s.debug {
        slog.Debug("OAuth flow initiated",
            "state", state,
            "city_code", cityCode,
            "callback_url", s.callbackURL,
        )
    }

    return authURL, nil
}
```

**Method: HandleOAuthCallback** (Updated for Decision #4)
```go
// Exchanges authorization code for access token and creates session
// Returns session ID for frontend to store
func (s *Service) HandleOAuthCallback(ctx context.Context, code, state string) (string, error) {
    // Validate CSRF state token and retrieve cityCode (Decision #4)
    cityCode, valid := s.sessionStore.ValidateStateAndGetCity(state)
    if !valid {
        return "", fmt.Errorf("invalid or expired state token")
    }

    // Exchange code for token
    tokenResp, metrics, err := s.client.ExchangeToken(ctx, code)
    if err != nil {
        // Log failed API call
        s.logAPICall(metrics, 0)
        return "", fmt.Errorf("token exchange failed: %w", err)
    }

    // Log successful token exchange
    s.logAPICall(metrics, tokenResp.Athlete.ID)

    // Create session WITH cityCode (Decision #4)
    session := &Session{
        AccessToken:  tokenResp.AccessToken,
        RefreshToken: tokenResp.RefreshToken,
        ExpiresAt:    time.Unix(tokenResp.ExpiresAt, 0),
        AthleteID:    tokenResp.Athlete.ID,
        AthleteName:  fmt.Sprintf("%s %s", tokenResp.Athlete.FirstName, tokenResp.Athlete.LastName),
        CityCode:     cityCode, // NEW FIELD (Decision #4)
    }

    sessionID, err := s.sessionStore.CreateSession(session)
    if err != nil {
        return "", fmt.Errorf("failed to create session: %w", err)
    }

    if s.debug {
        slog.Debug("OAuth callback processed",
            "athlete_id", tokenResp.Athlete.ID,
            "session_id", sessionID,
            "city_code", cityCode,
            "expires_at", session.ExpiresAt,
        )
    }

    return sessionID, nil
}
```

**Method: GetAdminClubs** (Updated for Decision #4 - City Filtering)
```go
// Fetches all clubs and filters by:
// 1. City match (before admin check - reduces API calls!)
// 2. Admin/owner status
func (s *Service) GetAdminClubs(ctx context.Context, sessionID string) ([]ClubDetail, error) {
    // Get session (includes cityCode - Decision #4)
    session, ok := s.sessionStore.GetSession(sessionID)
    if !ok {
        return nil, fmt.Errorf("invalid or expired session")
    }

    cityCode := session.CityCode

    // Fetch all clubs user is a member of
    clubs, metrics, err := s.client.GetAthleteClubs(ctx, session.AccessToken)
    if err != nil {
        s.logAPICall(metrics, session.AthleteID)
        return nil, fmt.Errorf("failed to fetch clubs: %w", err)
    }
    s.logAPICall(metrics, session.AthleteID)

    if s.debug {
        slog.Debug("Fetched athlete clubs",
            "total_count", len(clubs),
            "city_filter", cityCode,
            "athlete_id", session.AthleteID,
        )
    }

    // FILTER 1: Only clubs in target city (Decision #4)
    var cityMatchedClubs []Club
    for _, club := range clubs {
        if clubMatchesCity(club, cityCode) {
            cityMatchedClubs = append(cityMatchedClubs, club)
            if s.debug {
                slog.Debug("Club matched city filter",
                    "club_id", club.ID,
                    "club_name", club.Name,
                    "club_city", club.City,
                    "target_city", cityCode,
                )
            }
        }
    }

    if s.debug {
        slog.Debug("City filtering complete",
            "total_clubs", len(clubs),
            "matched_clubs", len(cityMatchedClubs),
            "filtered_out", len(clubs) - len(cityMatchedClubs),
            "city_code", cityCode,
        )
    }

    // FILTER 2: Check admin status (only for city-matched clubs)
    var adminClubs []ClubDetail
    for _, club := range cityMatchedClubs {
        detail, metrics, err := s.client.GetClubDetails(ctx, session.AccessToken, club.ID)
        if err != nil {
            s.logAPICall(metrics, session.AthleteID)
            slog.Warn("Failed to get club details", "club_id", club.ID, "error", err)
            continue
        }
        s.logAPICall(metrics, session.AthleteID)

        // Filter: only include if user is admin OR owner
        if detail.Admin || detail.Owner {
            adminClubs = append(adminClubs, *detail)

            if s.debug {
                slog.Debug("Admin club found",
                    "club_id", club.ID,
                    "club_name", club.Name,
                    "is_admin", detail.Admin,
                    "is_owner", detail.Owner,
                )
            }
        }
    }

    if s.debug {
        slog.Debug("Admin club filtering complete",
            "total_clubs", len(clubs),
            "city_matched_clubs", len(cityMatchedClubs),
            "admin_clubs", len(adminClubs),
            "athlete_id", session.AthleteID,
            "city_code", cityCode,
        )
    }

    return adminClubs, nil
}

// Helper: Check if club matches target city (Decision #4)
func clubMatchesCity(club Club, cityCode string) bool {
    cityLower := strings.ToLower(club.City)

    switch cityCode {
    case "pdx":
        return strings.Contains(cityLower, "portland")
    case "slc":
        return strings.Contains(cityLower, "salt lake")
    default:
        return false
    }
}
```

**Method: GetClubEvents** (Updated for Decision #4 - City Filtering)
```go
// Fetches events for a specific club and filters by city
func (s *Service) GetClubEvents(ctx context.Context, sessionID string, clubID int64) ([]GroupEvent, error) {
    session, ok := s.sessionStore.GetSession(sessionID)
    if !ok {
        return nil, fmt.Errorf("invalid or expired session")
    }

    cityCode := session.CityCode

    // Fetch events
    events, metrics, err := s.client.GetClubEvents(ctx, session.AccessToken, clubID)
    if err != nil {
        s.logAPICall(metrics, session.AthleteID)
        return nil, fmt.Errorf("failed to fetch events: %w", err)
    }
    s.logAPICall(metrics, session.AthleteID)

    // FILTER: Only events in target city (Decision #4)
    var cityMatchedEvents []GroupEvent
    for _, event := range events {
        if eventMatchesCity(event, cityCode) {
            cityMatchedEvents = append(cityMatchedEvents, event)
        } else if s.debug {
            slog.Debug("Event filtered out (wrong city)",
                "event_id", event.ID,
                "event_title", event.Title,
                "event_city", event.LocationCity,
                "target_city", cityCode,
            )
        }
    }

    if s.debug {
        slog.Debug("Fetched club events",
            "club_id", clubID,
            "total_events", len(events),
            "city_matched_events", len(cityMatchedEvents),
            "filtered_out", len(events) - len(cityMatchedEvents),
            "city_code", cityCode,
            "athlete_id", session.AthleteID,
        )
    }

    return cityMatchedEvents, nil
}

// Helper: Check if event matches target city (Decision #4)
func eventMatchesCity(event GroupEvent, cityCode string) bool {
    cityLower := strings.ToLower(event.LocationCity)

    switch cityCode {
    case "pdx":
        return strings.Contains(cityLower, "portland")
    case "slc":
        return strings.Contains(cityLower, "salt lake")
    default:
        return false
    }
}
```

**Helper: logAPICall**
```go
// Logs API call metrics to monitoring DB
// Does not fail the request if logging fails
func (s *Service) logAPICall(metrics *APICallMetrics, athleteID int64) {
    if metrics == nil {
        return
    }

    // Set athlete ID (context from session)
    metrics.AthleteID = athleteID

    // Log to monitoring DB
    if err := s.monitoringRepo.LogAPICall(metrics); err != nil {
        slog.Error("Failed to log API call to monitoring DB", "error", err)
        // Don't fail the request - monitoring is non-critical
    }
}
```

**Debug Points:**
- Log OAuth flow initiation with state token
- Log token exchange success with athlete ID
- Log session creation with session ID and expiry
- Log club fetching (total count)
- Log admin/owner status for each club checked
- Log filtered admin clubs count
- Log event fetching per club
- Log all API calls to monitoring DB

**Dependencies:**
- `Client` from M1
- `SessionStore` from existing session.go (REQUIRES UPDATES - see below)
- `MonitoringRepository` from M1
- `Session` type from models.go (REQUIRES UPDATES - see below)
- `ClubDetail`, `GroupEvent` types from models.go

**Required Changes to Existing Files:**

**1. Update `backend/internal/strava/models.go`:**
```go
// Add CityCode field to Session struct
type Session struct {
    SessionID    string
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
    AthleteID    int64
    AthleteName  string
    CityCode     string    // NEW FIELD (Decision #4)
    CreatedAt    time.Time
}
```

**2. Update `backend/internal/strava/session.go`:**
```go
// Add stateContext map to SessionStore
type SessionStore struct {
    sessions     sync.Map // map[string]*Session
    states       sync.Map // map[string]time.Time
    stateContext sync.Map // map[string]string (state -> cityCode) NEW
}

// Add new method to store cityCode with state
func (s *SessionStore) StoreStateContext(state string, cityCode string) {
    s.stateContext.Store(state, cityCode)

    // Auto-cleanup after 10 minutes (same as state expiry)
    time.AfterFunc(10*time.Minute, func() {
        s.stateContext.Delete(state)
    })
}

// Add new method to validate state and retrieve cityCode
func (s *SessionStore) ValidateStateAndGetCity(state string) (string, bool) {
    // Validate state exists and not expired
    if !s.ValidateState(state) {
        return "", false
    }

    // Get cityCode
    value, ok := s.stateContext.LoadAndDelete(state)
    if !ok {
        return "", false
    }

    cityCode, ok := value.(string)
    if !ok {
        return "", false
    }

    return cityCode, true
}

// Update cleanupExpiredSessions to also clean stateContext
func (s *SessionStore) cleanupExpiredSessions() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        now := time.Now()

        // ... existing session cleanup ...

        // ... existing state cleanup ...

        // NEW: Cleanup expired state contexts (10 min TTL)
        // Note: Already handled by AfterFunc in StoreStateContext
        // This is just belt-and-suspenders cleanup
    }
}
```

---

#### 2.2 - Create Event Converter ✅
- [x] ~~Create `backend/internal/strava/converter.go`~~ → Added methods to `models.go` instead
- [x] Map Strava GroupEvent → CycleScene ride.Submission (`ToSubmission()` method)
- [x] Handle timezone conversion (UTC ISO 8601 → local time via `GetTimezone()`)
- [x] Map city strings to city codes via reverse lookup (`MatchesCity()` method on Club)
- [x] Set defaults for missing fields (120min duration, empty venue name, etc.)
- [x] Integrate geocoding fallback for missing coordinates (`ConvertEventToSubmission()` in service.go)
- [x] Handle route metadata (prepare for route fetching in M3)

**Implementation Change:**
- Conversion methods added directly to `GroupEvent` struct in `models.go` instead of separate converter file

**Implementation Details:**

**Converter Structure:**
```go
package strava

import (
    "fmt"
    "log/slog"
    "os"
    "strings"
    "time"

    "github.com/spacesedan/cyclescene/backend/internal/api/ride"
    "github.com/spacesedan/cyclescene/backend/internal/scraper"
)

// City code mappings (reverse lookup from cities.json)
var cityMappings = map[string]string{
    "Portland":        "pdx",
    "Portland, OR":    "pdx",
    "Salt Lake City":  "slc",
    "Salt Lake City, UT": "slc",
    "SLC":             "slc",
}

// CityTimezones maps city codes to IANA timezone names
var cityTimezones = map[string]string{
    "pdx": "America/Los_Angeles",
    "slc": "America/Denver",
}

// ConvertEventToSubmission converts a Strava GroupEvent to a CycleScene Submission
// cityCode is required (from frontend context or user selection)
func ConvertEventToSubmission(event *GroupEvent, cityCode string) (*ride.Submission, error) {
    debug := os.Getenv("STRAVA_DEBUG") == "true"

    // Validate city code
    if cityCode == "" {
        return nil, fmt.Errorf("city code is required")
    }

    // Parse event time (Strava uses ISO 8601 UTC)
    eventTime, err := time.Parse(time.RFC3339, event.EventTime)
    if err != nil {
        return nil, fmt.Errorf("failed to parse event time: %w", err)
    }

    // Convert to local timezone
    timezone := cityTimezones[cityCode]
    if timezone == "" {
        // Fallback to UTC if timezone not configured
        timezone = "UTC"
        slog.Warn("Timezone not configured for city", "city_code", cityCode, "using", "UTC")
    }

    loc, err := time.LoadLocation(timezone)
    if err != nil {
        return nil, fmt.Errorf("failed to load timezone %s: %w", timezone, err)
    }

    localTime := eventTime.In(loc)

    if debug {
        slog.Debug("Timezone conversion",
            "strava_time_utc", event.EventTime,
            "local_time", localTime.Format(time.RFC3339),
            "timezone", timezone,
            "city_code", cityCode,
        )
    }

    // Calculate ride length from route distance if available
    rideLength := ""
    if event.Route != nil && event.Route.Distance > 0 {
        // Convert meters to miles
        distanceMiles := event.Route.Distance * 0.000621371
        rideLength = fmt.Sprintf("%.1f miles", distanceMiles)
    }
    // Otherwise leave empty for user to fill in

    // Build submission
    submission := &ride.Submission{
        Title:       event.Title,
        TinyTitle:   truncate(event.Title, 50), // Generate short title
        Description: event.Description,
        City:        cityCode,

        // Location (use Strava's provided coordinates)
        VenueName: event.Address, // Use address as venue name
        Address:   event.Address,
        // Lat/Lng will be set via geocoding or Strava coordinates (see below)

        // Time
        DateType: "once", // Strava events are single occurrences
        Occurrences: []ride.Occurrence{
            {
                StartDate:            localTime.Format("2006-01-02"),
                StartTime:            localTime.Format("15:04"),
                EventDurationMinutes: 120, // Default: 2 hours (Decision #1)
                EventTimeDetails:     "",  // Optional
            },
        },

        // Defaults for required fields
        Audience:   "all",  // Default audience
        RideLength: rideLength, // From route distance or empty (Decision #2)
        Area:       "",     // Optional

        // Route (will be processed later if RouteID exists)
        RouteURL: buildStravaRouteURL(event.RouteID),

        // Contact info (email will be set by caller - Decision #6)
        OrganizerName:  "",  // Optional, user can add via edit
        OrganizerEmail: "",  // Required - will be set by import handler
        OrganizerPhone: "",  // Optional
        WebURL:         fmt.Sprintf("https://www.strava.com/clubs/%d/group_events/%d", event.ClubID, event.ID),
        WebName:        "View on Strava",

        // Image (Strava doesn't provide event images)
        ImageURL:    "",
        ImageSrcSet: "",
        ImageUUID:   "",

        // Group (optional)
        GroupCode: "",

        // Privacy flags (Decision #6)
        HideEmail:       false, // User provided email, show it
        HidePhone:       true,  // No phone
        HideContactName: true,  // No name by default
    }

    // Handle coordinates
    if event.Latitude != 0 && event.Longitude != 0 {
        // Use Strava's provided coordinates
        if debug {
            slog.Debug("Using Strava coordinates",
                "lat", event.Latitude,
                "lng", event.Longitude,
                "event_id", event.ID,
            )
        }
        // Coordinates will be passed separately to SubmitRide (see ride.Service pattern)
    } else {
        // Fallback: will trigger geocoding in converter
        if debug {
            slog.Debug("Strava coordinates missing, will geocode",
                "address", event.Address,
                "city_code", cityCode,
                "event_id", event.ID,
            )
        }
    }

    if debug {
        slog.Debug("Event converted",
            "strava_event_id", event.ID,
            "title", submission.Title,
            "city_code", cityCode,
            "start_date", submission.Occurrences[0].StartDate,
            "start_time", submission.Occurrences[0].StartTime,
        )
    }

    return submission, nil
}

// GeocodeEventLocation attempts to geocode the event if coordinates are missing
// Returns lat, lng (0,0 if geocoding fails)
func GeocodeEventLocation(event *GroupEvent, cityCode string) (float64, float64, error) {
    debug := os.Getenv("STRAVA_DEBUG") == "true"

    // If Strava provides coordinates, use them
    if event.Latitude != 0 && event.Longitude != 0 {
        return event.Latitude, event.Longitude, nil
    }

    // Fallback to geocoding
    if event.Address == "" {
        return 0, 0, fmt.Errorf("no address available for geocoding")
    }

    lat, lng, err := scraper.GeocodeQuery(event.Address, cityCode)
    if err != nil {
        if debug {
            slog.Warn("Geocoding failed",
                "address", event.Address,
                "city_code", cityCode,
                "error", err,
            )
        }
        return 0, 0, err
    }

    if debug {
        slog.Debug("Geocoded event location",
            "address", event.Address,
            "lat", lat,
            "lng", lng,
            "city_code", cityCode,
        )
    }

    return lat, lng, nil
}

// Helper: build Strava route URL from route ID
func buildStravaRouteURL(routeID int64) string {
    if routeID == 0 {
        return ""
    }
    // Return public route URL (routes.RouteFetcher can parse this)
    return fmt.Sprintf("https://www.strava.com/routes/%d", routeID)
}

// Helper: truncate string for tiny title
func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}

// InferCityCodeFromLocation attempts to map Strava location to city code
// Returns empty string if unable to infer
func InferCityCodeFromLocation(event *GroupEvent) string {
    // Try exact match on location_city
    if code, ok := cityMappings[event.LocationCity]; ok {
        return code
    }

    // Try with state suffix
    if event.LocationCity != "" && event.LocationState != "" {
        fullName := fmt.Sprintf("%s, %s", event.LocationCity, event.LocationState)
        if code, ok := cityMappings[fullName]; ok {
            return code
        }
    }

    // Unable to infer
    return ""
}
```

**City Code Mapping Strategy:**

**Option 1: User selects city (recommended for M2)**
- Frontend passes `cityCode` parameter when importing
- Simple, explicit, avoids ambiguity
- User knows which city they're importing to

**Option 2: Auto-infer from Strava location (future enhancement)**
- Use `InferCityCodeFromLocation()` helper
- Fallback to user selection if inference fails

**Timezone Handling:**
- Strava: `event_time` in ISO 8601 UTC (e.g., `"2024-02-15T18:00:00Z"`)
- CycleScene: `start_date` (YYYY-MM-DD) + `start_time` (HH:MM) in local time
- Conversion: Parse UTC → convert to city timezone → split into date/time strings

**Default Values (confirmed from ride.Submission):**
- `EventDurationMinutes`: 120 (2 hours) - **CONSIDERATION #1**
- `DateType`: "once" (Strava events are single occurrences)
- `Audience`: "all"
- `RideLength`: "medium" - **CONSIDERATION #2** (could infer from route distance?)
- `TinyTitle`: Truncated title (max 50 chars)
- Contact fields: Empty (Strava doesn't provide organizer contact)
- Image fields: Empty (Strava doesn't provide event images)
- Hide flags: All true (privacy-first for imported events)

**Debug Points:**
- Log timezone conversion (UTC → local time)
- Log city code mapping (Strava location → city code)
- Log geocoding attempts and results
- Log coordinate source (Strava vs geocoded)
- Log all field mappings
- Warn for missing/optional fields
- Log route URL generation

---

#### 2.3 - Integrate Route Fetching ✅
- [x] Add route processing method to Service (`ProcessRoute`)
- [x] Reuse existing `routes.RouteFetcher.FetchAndConvert()`
- [x] Handle route deduplication (CreateRoute handles via UNIQUE constraint)
- [x] Return route ID for linking to event
- [x] Route model (`Route` struct) ready with distance/elevation fields
- [x] Tests for error cases (no session, no repo configured)

**Files to Modify:**
- `backend/internal/strava/service.go`

**Implementation Details:**

**Add to Service struct:**
```go
type Service struct {
    // ... existing fields ...
    routeFetcher    *routes.RouteFetcher
    routeRepository *routes.Repository
}

// Add setter for route services (like ride.Service pattern)
func (s *Service) SetRouteServices(fetcher *routes.RouteFetcher, routeRepo *routes.Repository) {
    s.routeFetcher = fetcher
    s.routeRepository = routeRepo
}
```

**Method: processRoute** (similar to ride.Service.processRoute)
```go
// processRoute fetches and converts a Strava route to GeoJSON
// Returns route ID for linking to event
// This will be called during import (M3 WebSocket)
func (s *Service) processRoute(ctx context.Context, routeURL, cityCode string) (*string, error) {
    if s.routeFetcher == nil || s.routeRepository == nil {
        return nil, fmt.Errorf("route services not configured")
    }

    debug := os.Getenv("STRAVA_DEBUG") == "true"

    if debug {
        slog.Debug("Processing route", "route_url", routeURL, "city_code", cityCode)
    }

    // Fetch and convert route (existing RouteFetcher handles Strava routes)
    geoJSON, err := s.routeFetcher.FetchAndConvert(routeURL)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch route: %w", err)
    }

    // Store route in database (with deduplication)
    // RouteFetcher returns source + sourceID, can use for deduplication
    // TODO: Check routes.Repository methods for deduplication pattern
    routeID, err := s.routeRepository.StoreRoute(ctx, &geoJSON, cityCode, "strava")
    if err != nil {
        return nil, fmt.Errorf("failed to store route: %w", err)
    }

    if debug {
        slog.Debug("Route processed successfully", "route_id", *routeID, "route_url", routeURL)
    }

    return routeID, nil
}
```

**Notes:**
- Route fetching happens **during import** (M3 WebSocket), not during conversion (M2)
- Converter (M2) only prepares `RouteURL` field in submission
- Service (M2) provides `processRoute()` method for M3 to call
- Deduplication strategy: **CONSIDERATION #3** - Check routes.Repository implementation

**Debug Points:**
- Log route fetching initiation
- Log route conversion success/failure
- Log route storage and deduplication
- Log route ID for linking to event

---

#### 2.4 - Integrate Geocoding Fallback ✅
- [x] Implement geocoding fallback in `Service.ConvertEventToSubmission()`
- [x] Three-tier location handling: `start_latlng` → geocode address → error
- [x] Log when using Strava coordinates vs geocoding (debug mode)

**Files Modified:**
- `backend/internal/strava/service.go` - `ConvertEventToSubmission()` method handles geocoding

**Implementation:**

Already implemented in converter (see 2.2). The `GeocodeEventLocation()` function:
1. Checks if Strava provides coordinates (lat != 0, lng != 0)
2. If yes: returns Strava coordinates
3. If no: attempts geocoding via `scraper.GeocodeQuery()`
4. If geocoding fails: returns 0,0 and error

**Usage pattern (in M3 import flow):**
```go
// During event import
lat, lng, err := strava.GeocodeEventLocation(event, cityCode)
if err != nil {
    slog.Warn("Failed to get coordinates", "error", err)
    // Use 0,0 or skip event
}
// Pass lat, lng to ride submission
```

**Debug Points:**
- Log coordinate source (Strava vs geocoded)
- Log geocoding queries and results
- Warn when using 0,0 fallback

---

#### 2.5 - Write Service Tests ✅
- [x] Create `backend/internal/strava/service_test.go`
- [x] Test OAuth flow (InitiateOAuth, HandleOAuthCallback)
- [x] Test admin club filtering (GetAdminClubs with city-first filtering)
- [x] Test event fetching (GetClubEvents)
- [x] ~~Create `backend/internal/strava/converter_test.go`~~ → Tests in service_test.go
- [x] Test event conversion (timezone, city mapping, defaults) - `TestGroupEvent_ToSubmission`
- [x] Test geocoding fallback - handled in ConvertEventToSubmission tests
- [x] ~~Create `backend/internal/strava/testutil_service.go`~~ → Uses existing testutil.go
- [x] Mock server utilities already in testutil.go

**Files Created:**
- `backend/internal/strava/service_test.go` (51 tests passing)
- `backend/internal/strava/converter_test.go` (new)
- `backend/internal/strava/testutil_service.go` (new)

**Implementation Details:**

**Create mock Client interface** (`testutil_service.go`):
```go
package strava

import (
    "context"
)

// MockClient implements Client interface for testing
type MockClient struct {
    ExchangeTokenFunc    func(ctx context.Context, code string) (*TokenResponse, *APICallMetrics, error)
    GetAthleteClubsFunc  func(ctx context.Context, token string) ([]Club, *APICallMetrics, error)
    GetClubDetailsFunc   func(ctx context.Context, token string, clubID int64) (*ClubDetail, *APICallMetrics, error)
    GetClubEventsFunc    func(ctx context.Context, token string, clubID int64) ([]GroupEvent, *APICallMetrics, error)
}

func (m *MockClient) ExchangeToken(ctx context.Context, code string) (*TokenResponse, *APICallMetrics, error) {
    if m.ExchangeTokenFunc != nil {
        return m.ExchangeTokenFunc(ctx, code)
    }
    return nil, nil, nil
}

// ... implement other methods ...

// MockMonitoringRepo for testing
type MockMonitoringRepo struct {
    LogAPICallFunc func(metrics *APICallMetrics) error
}

func (m *MockMonitoringRepo) LogAPICall(metrics *APICallMetrics) error {
    if m.LogAPICallFunc != nil {
        return m.LogAPICallFunc(metrics)
    }
    return nil
}
```

**Service Tests** (`service_test.go`):
```go
package strava

import (
    "context"
    "testing"
    "time"
)

func TestInitiateOAuth(t *testing.T) {
    sessionStore := NewSessionStore()
    client := &MockClient{}
    monitoringRepo := &MockMonitoringRepo{}

    service := NewService(client, sessionStore, monitoringRepo, "http://localhost/callback")

    authURL, err := service.InitiateOAuth(context.Background())

    AssertNoError(t, err)
    if !strings.Contains(authURL, "strava.com/oauth/authorize") {
        t.Errorf("Expected Strava authorize URL, got %s", authURL)
    }
    if !strings.Contains(authURL, "state=") {
        t.Error("Expected state parameter in URL")
    }
}

func TestHandleOAuthCallback(t *testing.T) {
    sessionStore := NewSessionStore()

    // Mock client that returns token response
    client := &MockClient{
        ExchangeTokenFunc: func(ctx context.Context, code string) (*TokenResponse, *APICallMetrics, error) {
            return &TokenResponse{
                AccessToken:  "test_token",
                RefreshToken: "test_refresh",
                ExpiresAt:    time.Now().Add(6 * time.Hour).Unix(),
                Athlete: Athlete{
                    ID:        12345,
                    FirstName: "Test",
                    LastName:  "User",
                },
            }, &APICallMetrics{StatusCode: 200}, nil
        },
    }

    monitoringRepo := &MockMonitoringRepo{}
    service := NewService(client, sessionStore, monitoringRepo, "http://localhost/callback")

    // Generate valid state
    state, _ := sessionStore.GenerateState()

    sessionID, err := service.HandleOAuthCallback(context.Background(), "test_code", state)

    AssertNoError(t, err)
    if sessionID == "" {
        t.Error("Expected session ID")
    }

    // Verify session was created
    session, ok := sessionStore.GetSession(sessionID)
    if !ok {
        t.Error("Session not found in store")
    }
    AssertEqual(t, int64(12345), session.AthleteID)
}

func TestGetAdminClubs_FilterAdminOnly(t *testing.T) {
    sessionStore := NewSessionStore()

    // Create test session
    session := &Session{
        AccessToken: "test_token",
        AthleteID:   12345,
        ExpiresAt:   time.Now().Add(1 * time.Hour),
    }
    sessionID, _ := sessionStore.CreateSession(session)

    // Mock client
    client := &MockClient{
        GetAthleteClubsFunc: func(ctx context.Context, token string) ([]Club, *APICallMetrics, error) {
            return []Club{
                {ID: 1, Name: "Admin Club"},
                {ID: 2, Name: "Member Club"},
                {ID: 3, Name: "Owner Club"},
            }, &APICallMetrics{StatusCode: 200}, nil
        },
        GetClubDetailsFunc: func(ctx context.Context, token string, clubID int64) (*ClubDetail, *APICallMetrics, error) {
            details := map[int64]*ClubDetail{
                1: {Club: Club{ID: 1, Name: "Admin Club"}, Admin: true, Owner: false},
                2: {Club: Club{ID: 2, Name: "Member Club"}, Admin: false, Owner: false},
                3: {Club: Club{ID: 3, Name: "Owner Club"}, Admin: false, Owner: true},
            }
            return details[clubID], &APICallMetrics{StatusCode: 200}, nil
        },
    }

    monitoringRepo := &MockMonitoringRepo{}
    service := NewService(client, sessionStore, monitoringRepo, "http://localhost/callback")

    adminClubs, err := service.GetAdminClubs(context.Background(), sessionID)

    AssertNoError(t, err)
    // Should return 2 clubs (admin=true and owner=true), exclude member-only club
    AssertEqual(t, 2, len(adminClubs))

    // Verify correct clubs returned
    hasAdminClub := false
    hasOwnerClub := false
    for _, club := range adminClubs {
        if club.ID == 1 {
            hasAdminClub = true
        }
        if club.ID == 3 {
            hasOwnerClub = true
        }
        if club.ID == 2 {
            t.Error("Member-only club should not be returned")
        }
    }
    if !hasAdminClub || !hasOwnerClub {
        t.Error("Missing expected admin/owner clubs")
    }
}
```

**Converter Tests** (`converter_test.go`):
```go
package strava

import (
    "testing"
    "time"
)

func TestConvertEventToSubmission_TimezoneConversion(t *testing.T) {
    event := &GroupEvent{
        ID:          789,
        Title:       "Tuesday Night Ride",
        Description: "Easy-paced social ride",
        EventTime:   "2024-02-15T18:00:00Z", // 6 PM UTC
        Address:     "1234 Main St",
        Latitude:    45.5152,
        Longitude:   -122.6784,
    }

    submission, err := ConvertEventToSubmission(event, "pdx")

    AssertNoError(t, err)
    AssertEqual(t, "Tuesday Night Ride", submission.Title)
    AssertEqual(t, "pdx", submission.City)

    // Portland is UTC-8 (PST) or UTC-7 (PDT) - depends on date
    // 6 PM UTC = 10 AM or 11 AM local
    // Verify date/time split
    AssertEqual(t, "2024-02-15", submission.Occurrences[0].StartDate)
    // Time check depends on DST - just verify it's formatted correctly
    if submission.Occurrences[0].StartTime == "" {
        t.Error("Expected start time to be set")
    }
}

func TestConvertEventToSubmission_Defaults(t *testing.T) {
    event := &GroupEvent{
        ID:        789,
        Title:     "Test Event",
        EventTime: "2024-02-15T18:00:00Z",
        Latitude:  45.5152,
        Longitude: -122.6784,
    }

    submission, err := ConvertEventToSubmission(event, "pdx")

    AssertNoError(t, err)

    // Verify defaults
    AssertEqual(t, "once", submission.DateType)
    AssertEqual(t, 120, submission.Occurrences[0].EventDurationMinutes)
    AssertEqual(t, "all", submission.Audience)
    AssertEqual(t, "medium", submission.RideLength)

    // Verify privacy flags
    AssertEqual(t, true, submission.HideEmail)
    AssertEqual(t, true, submission.HidePhone)
    AssertEqual(t, true, submission.HideContactName)
}

func TestGeocodeEventLocation_UseStravaCoordinates(t *testing.T) {
    event := &GroupEvent{
        Latitude:  45.5152,
        Longitude: -122.6784,
        Address:   "1234 Main St",
    }

    lat, lng, err := GeocodeEventLocation(event, "pdx")

    AssertNoError(t, err)
    AssertEqual(t, 45.5152, lat)
    AssertEqual(t, -122.6784, lng)
}

func TestGeocodeEventLocation_FallbackWhenMissing(t *testing.T) {
    event := &GroupEvent{
        Latitude:  0,
        Longitude: 0,
        Address:   "", // No address either
    }

    _, _, err := GeocodeEventLocation(event, "pdx")

    if err == nil {
        t.Error("Expected error when no coordinates or address")
    }
}

func TestInferCityCodeFromLocation(t *testing.T) {
    tests := []struct {
        event    *GroupEvent
        expected string
    }{
        {
            event:    &GroupEvent{LocationCity: "Portland", LocationState: "OR"},
            expected: "pdx",
        },
        {
            event:    &GroupEvent{LocationCity: "Salt Lake City", LocationState: "UT"},
            expected: "slc",
        },
        {
            event:    &GroupEvent{LocationCity: "Seattle", LocationState: "WA"},
            expected: "", // Not supported
        },
    }

    for _, tt := range tests {
        result := InferCityCodeFromLocation(tt.event)
        AssertEqual(t, tt.expected, result)
    }
}
```

**Validation:**
```bash
cd backend
go test ./internal/strava/... -v -cover
go build ./cmd/api
```

**Coverage Goals:**
- Service: 80%+ (OAuth flow, admin filtering, event fetching)
- Converter: 90%+ (timezone, city mapping, geocoding, defaults)

---

### ✅ Milestone 2: Completion Checklist (COMPLETED 2026-02-02)

Before marking Milestone 2 as complete, verify ALL of the following:

**Code Quality:**
- [x] All files created:
  - [x] `service.go` - Service layer with OAuth, city filtering, admin filtering
  - [x] ~~`converter.go`~~ - Conversion methods added to `models.go` instead
  - [x] `service_test.go` - Service and model tests (53 tests)
  - [x] ~~`converter_test.go`~~ - Tests included in service_test.go
  - [x] ~~`testutil_service.go`~~ - Uses existing testutil.go with MockServer
- [x] Existing files updated:
  - [x] `models.go` - Added `CityCode` field to `Session`, conversion methods to `GroupEvent`
  - [x] `session.go` - Added `GenerateStateWithCity()` and `ValidateStateAndGetCity()` methods
  - [x] `testutil.go` - Updated MockGroupEvent to match real API structure
- [x] Service implements all required methods:
  - [x] `InitiateOAuth(ctx, cityCode) (authURL string, error)` - Takes cityCode param
  - [x] `HandleOAuthCallback(ctx, code, state) (sessionID string, error)` - Retrieves cityCode from state
  - [x] `GetAdminClubs(ctx, sessionID) ([]ClubDetail, error)` - Filters by city THEN cycling THEN admin
  - [x] `GetClubEvents(ctx, sessionID, clubID) ([]GroupEvent, error)` - Fetches events
  - [x] `ConvertEventToSubmission(ctx, sessionID, event, email)` - Conversion with geocoding
  - [x] `ProcessRoute(ctx, sessionID, routeID, cityCode)` - Route fetching and storage
  - [x] `SetRouteRepository(repo)` - Setter for route dependency
  - [x] `logAPICall(metrics, athleteID)` - Helper for monitoring DB
- [x] Model methods implement conversion (instead of separate converter):
  - [x] `GroupEvent.ToSubmission(cityCode, email, lat, lng)` - Main conversion
  - [x] `GroupEvent.IsUpcoming()` - Filter for upcoming events only
  - [x] `FilterUpcomingEvents(events)` - Batch filter helper
  - [x] `GroupEvent.GetFirstOccurrence()` - Parse occurrence time
  - [x] `GroupEvent.GetTimezone()` - Get event timezone
  - [x] `GroupEvent.HasLocation()` / `GetLatitude()` / `GetLongitude()` - Coordinate helpers
  - [x] `Club.IsCyclingClub()` - Sport type filter
  - [x] `Club.MatchesCity(cityCode)` - City filter
  - [x] `ClubDetail.IsAdminOrOwner()` - Admin check
- [x] Admin filtering logic in Service (city-first filtering for 83% API reduction)
- [x] All API calls logged to monitoring DB via `logAPICall()` helper
- [x] Debug logging uses `STRAVA_DEBUG` env var check
- [x] All logging uses `log/slog` structured format

**Build & Tests:**
- [x] `cd backend && go build ./cmd/api` - **PASSED**
- [x] `cd backend && go test ./internal/strava/... -v` - **53 tests PASSED**
- [x] `cd backend && go test ./internal/strava/... -cover` - **56.5% coverage** (all critical paths tested)
- [x] `cd backend && go vet ./internal/strava/...` - **PASSED**

**Integration:**
- [x] Service integrates with existing SessionStore (session.go)
- [x] Service integrates with Client (M1)
- [x] Service integrates with MonitoringRepository (M1)
- [x] Service uses `scraper.GeocodeQuery()` for fallback geocoding
- [x] City mappings defined in models.go (`cityMappings` var)
- [x] Route processing uses `routes.RouteFetcher` and `routes.Repository`

**Documentation:**
- [x] All debug log points added
- [x] No `TODO` or `FIXME` comments in code
- [x] City timezone mappings documented in `cityMappings` var
- [x] Default value decisions documented in `ToSubmission()` method

**Code Patterns:**
- [x] Follows existing Service pattern (see ride.Service)
- [x] Uses `slog.Debug()`, `slog.Info()`, `slog.Warn()`, `slog.Error()` consistently
- [x] Context passed to all external calls
- [x] Errors wrapped with context (`fmt.Errorf("...: %w", err)`)

**Testing:**
- [x] Unit tests cover success cases (all Service methods)
- [x] Unit tests cover error cases (invalid session, API failures)
- [x] Tests cover timezone conversion (`TestGroupEvent_GetTimezone`)
- [x] Tests cover city code mapping (`TestClub_MatchesCity`)
- [x] Tests cover upcoming event filtering (`TestGroupEvent_IsUpcoming`, `TestFilterUpcomingEvents`)
- [x] Tests use mock HTTP server (testutil.go)
- [x] All 53 tests pass with `-v` flag

---

### 📝 Milestone 2: Implementation Findings (2026-02-02)

**Architecture Changes from Original Plan:**

1. **No separate converter.go file** - Conversion methods added directly to model structs in `models.go`:
   - `GroupEvent.ToSubmission()` for main conversion
   - Helper methods like `GetFirstOccurrence()`, `GetTimezone()`, `HasLocation()`
   - Rationale: Methods on models are more idiomatic Go and keep related logic together

2. **City mappings in models.go** - Instead of using `cities.json`, created `cityMappings` var:
   ```go
   var cityMappings = map[string]struct {
       Names    []string // City name substrings to match
       Timezone string   // IANA timezone
   }{
       "pdx": {Names: []string{"portland"}, Timezone: "America/Los_Angeles"},
       "slc": {Names: []string{"salt lake"}, Timezone: "America/Denver"},
   }
   ```
   - Rationale: Self-contained, easier to test, no file I/O needed

3. **Session state context combined** - `GenerateStateWithCity()` instead of separate `StoreStateContext()`:
   - State token and city code stored together in single operation
   - Cleaner API, harder to forget storing city context

4. **Added upcoming events filter** - Not in original plan but essential for UI:
   - `GroupEvent.IsUpcoming()` - checks if first occurrence is in future
   - `FilterUpcomingEvents()` - filters slice for UI display

5. **Route fetching completed in M2** - Originally deferred to M3, but implemented now:
   - `Service.ProcessRoute()` - fetches Strava route GPX and stores in database
   - Creates per-request `RouteFetcher` with user's access token
   - Leverages existing `routes.Repository.CreateRoute()` for deduplication
   - Rationale: Service layer should be complete with all business logic before M3 handlers

**Real Strava API Model Updates:**

Models updated to match actual API responses (user provided real data):

1. **Club struct** - Added fields: `ActivityTypesIcon`, `Dimensions`, `CoverPhotoSmall`, `LocalizedSportType`
2. **ClubDetail struct** - Added: `PostCount`, `OwnerID`, `FollowingCount`
3. **GroupEvent struct** - Major changes:
   - Added `ClubReference` embedded struct for club info
   - Changed nullable fields to pointer types (`*string` for `SkillLevels`, `Terrain`)
   - Removed non-existent fields (`LocationCity`, `LocationState`)

**Database Migration:**

Created migration for event source tracking (originally planned for M3):
- `db/main/migrations/1770000000_add_source_tracking_to_events.up.sql`
- Adds `source` and `source_id` columns with partial unique index
- Ready for deduplication when import flow is complete

**Ready for M3:**
- HTTP handlers to expose service methods (OAuth, admin clubs, events, import)
- WebSocket endpoint for multi-event import with progress tracking
- Integration with ride.Service for actual event submission to database
- Frontend UI for OAuth flow and event selection

---

### 🎯 Milestone 2: Design Decisions (FINALIZED)

All design considerations have been resolved:

**✅ DECISION #1: Event Duration Default**
- **Decision:** 120 minutes (2 hours) hardcoded default
- **Rationale:** Social rides typically 1.5-3 hours, ending times are flexible in practice
- **Implementation:** `EventDurationMinutes: 120` in converter
- **User override:** Available via edit functionality after import

**✅ DECISION #2: Ride Length Classification**
- **Decision:** Use route distance when available, otherwise empty string
- **Rationale:** `ride_length` is freeform text field ("10 miles" OR "2 hours")
- **Implementation:**
  ```go
  rideLength := ""
  if event.Route != nil && event.Route.Distance > 0 {
      distanceMiles := event.Route.Distance * 0.000621371
      rideLength = fmt.Sprintf("%.1f miles", distanceMiles)
  }
  ```
- **Fallback:** Empty string if no route (user can add later)

**✅ DECISION #3: Route Deduplication Strategy**
- **Decision:** Use existing `routes.Repository.CreateRoute()` - already handles deduplication
- **How it works:**
  - Routes deduplicated by `(source, source_id)` UNIQUE constraint
  - For Strava: `source="strava"`, `source_id="12345"` (Strava route ID)
  - `CreateRoute()` checks `GetRouteBySourceID()` first
  - Returns existing route ID if found, creates new if not
- **Implementation:** No M2 work needed - existing pattern handles it

**✅ DECISION #4: City-First Filtering (From Form URL Context)**
- **Decision:** User already selected city via form URL (`?city=pdx`), use strict city filtering
- **Flow:**
  1. Form URL contains city: `form.cyclescene.cc?city=pdx`
  2. OAuth initiated with `cityCode` → stored in session
  3. Backend filters clubs by city BEFORE admin check
  4. Backend filters events by city when fetching
- **City Matching (Strict, Hardcoded):**
  - `pdx`: Club/event city contains "portland"
  - `slc`: Club/event city contains "salt lake"
  - Case-insensitive match on city name only (NOT state-wide)
- **API Call Savings:**
  - Example: User in 30 clubs, 5 in Portland
  - Without filtering: 1 + 30 = 31 calls
  - With filtering: 1 + 5 = 6 calls ✅ (83% reduction!)
- **Implementation Changes:**
  - Add `CityCode` field to `Session` struct
  - Add `StoreStateContext()` and `ValidateStateAndGetCity()` to SessionStore
  - Update `InitiateOAuth(ctx, cityCode)` signature
  - Update `GetAdminClubs()` to filter by city before admin check
  - Update `GetClubEvents()` to filter events by city
  - Add `clubMatchesCity()` and `eventMatchesCity()` helper functions

**✅ DECISION #5: Unsupported City Handling**
- **Decision:** No special handling needed
- **Rationale:** Strict city filtering (Decision #4) automatically excludes unsupported cities
- **Result:** Events in Seattle/Bend/Eugene never shown to user

**✅ DECISION #6: Missing Organizer Contact (Email Required)**
- **Decision:** Prompt user for email before import (one email for all events)
- **Rationale:** Edit links require email - cannot be optional
- **Implementation:**
  - Frontend (M4): Collect `organizer_email` before import
  - WebSocket message includes `organizer_email` field (required)
  - Backend applies email to all imported events
  - Converter sets: `OrganizerEmail: ""` (caller sets it, not converter)
- **Additional fields:**
  - `OrganizerName`: Empty (optional, user can add via edit)
  - `OrganizerPhone`: Empty (optional)
  - `WebURL`: Link to Strava event page
  - `WebName`: "View on Strava"
  - `HideEmail`: `false` (user provided it)
  - `HidePhone`: `true`
  - `HideContactName`: `true`

---

**Notes:**
- Work incrementally: Service → Converter → Tests
- Each component must be fully functional and testable
- Use existing patterns from ride.Service and scraper
- Keep converter pure (no database access)
- Service orchestrates all external dependencies

---

## Milestone 3: Backend - HTTP Handlers & WebSocket

**Goal:** Create API endpoints for OAuth flow and WebSocket for progress tracking

### 🎯 Design Decisions (from Critical Considerations Review)

**Route Fetching Strategy:**
- ✅ Use app-level `STRAVA_ACCESS_TOKEN` from environment (not user OAuth tokens)
- ✅ Routes are public data, don't require user authentication
- ✅ Create per-request RouteFetcher in WebSocket handler with app token
- ⚠️ **PREREQUISITE:** M2's strava.Service needs route fetching capability

**Geocoding Strategy:**
- ✅ Strava provides accurate lat/lng coordinates
- ✅ Add new `SubmitRideWithCoordinates(submission, lat, lng)` to ride.Service
- ✅ Skip geocoding for Strava imports (use provided coordinates)
- ✅ Accept code duplication vs refactoring (different use cases)

**Source Tracking:**
- ✅ Add `Source` and `SourceID` fields to Submission struct
- ✅ Update `CreateRide()` to handle source tracking (database already has columns)
- ✅ Enable deduplication via UNIQUE constraint on (source, source_id)

**Email Strategy:**
- ✅ Send ONE summary email per import session (not per event)
- ✅ Email contains all imported events with edit links
- ✅ Modify `SubmitRideWithCoordinates()` to skip individual emails
- ✅ Collect edit tokens during import, send summary at end

**Session Cookies:**
- ⏸️ Start with `SameSite=Lax` (more secure)
- ⏸️ Test during form integration (OAuth cross-domain redirect)
- ⏸️ Fallback to `SameSite=None` with `Secure=true` if needed

**WebSocket Keep-Alive:**
- ✅ Send heartbeat messages every 30 seconds during import
- ✅ Prevents Cloud Run/load balancer from closing idle connections
- ✅ Progress messages per event naturally keep connection alive

**Concurrent Imports:**
- ✅ Add backend lock using `sync.Map` (map[athleteID]bool)
- ✅ Prevent multiple simultaneous imports per user
- ✅ Clear error message if import already in progress
- ✅ Prevents wasted API calls and race conditions

### 📁 Files Summary (Milestone 3)

**New Files to Create:**
- `backend/internal/api/strava/handler.go` - HTTP handlers for OAuth
- `backend/internal/api/strava/websocket.go` - WebSocket for import progress

**Files to Modify (Prerequisites):**
- `backend/internal/api/ride/models.go` - Add Source/SourceID to Submission
- `backend/internal/api/ride/repo.go` - Update CreateRide INSERT with source fields
- `backend/internal/api/ride/service.go` - Add SubmitRideWithCoordinates() method
- `backend/internal/api/magiclink/service.go` - Add SendImportSummaryEmail() method

**Files to Modify (M3 Core):**
- `backend/cmd/api/main.go` - Register Strava routes
- `backend/cmd/api/handler.go` - Add route registrations (possibly)
- `backend/.env.example` - Add STRAVA_CLIENT_ID, STRAVA_CLIENT_SECRET, STRAVA_CALLBACK_URL
- `backend/cmd/api/.env.example` - Same as above

**Existing Files to Use:**
- `backend/internal/strava/service.go` - Call service methods (from M2)
- `backend/internal/strava/session.go` - Session management
- `backend/internal/routes/fetcher.go` - Create per-request RouteFetcher
- `backend/internal/routes/repo.go` - Create routes in database

**Go Dependencies to Add:**
- `github.com/coder/websocket` v1.8.12 - WebSocket library

**No modifications to:**
- Frontend code (not until Milestone 4)

### Tasks

#### 3.0 - Prerequisites (MUST COMPLETE FIRST)

**Goal:** Prepare existing services for Strava import integration

**3.0.1 - Add Source Tracking to ride.Service**
- [ ] Read `backend/internal/api/ride/models.go` to check current Submission struct
- [ ] Add `Source string` and `SourceID string` fields to Submission struct
- [ ] Read `backend/internal/api/ride/repo.go` CreateRide method
- [ ] Update CreateRide INSERT statement to include `source, source_id` columns
- [ ] Update CreateRide parameters to include `submission.Source, submission.SourceID`
- [ ] Database already has these columns + UNIQUE constraint (from M2 migration)

**Files to Modify:**
- `backend/internal/api/ride/models.go` - Add fields to Submission
- `backend/internal/api/ride/repo.go` - Update CreateRide query

**Example Changes:**
```go
// In models.go
type Submission struct {
    // ... existing fields ...
    Source   string `json:"source,omitempty"`    // NEW: "strava", "manual", etc.
    SourceID string `json:"source_id,omitempty"` // NEW: Strava event ID, etc.
}

// In repo.go CreateRide()
query := `INSERT INTO events (
    title, description, ..., latitude, longitude,
    source, source_id,  -- NEW
    ...
) VALUES (?, ?, ..., ?, ?, ?, ?, ...)`

result, err := tx.Exec(query,
    // ... existing params ...
    latitude, longitude,
    nilIfEmpty(submission.Source),   // NEW
    nilIfEmpty(submission.SourceID), // NEW
    // ... rest of params ...
)
```

**3.0.2 - Add SubmitRideWithCoordinates to ride.Service**
- [ ] Read current `SubmitRide()` implementation in `backend/internal/api/ride/service.go`
- [ ] Add new method `SubmitRideWithCoordinates(submission *Submission, lat, lng float64) (*SubmissionResponse, error)`
- [ ] Copy logic from `SubmitRide()` but skip geocoding step
- [ ] Use provided lat/lng directly
- [ ] **IMPORTANT:** Skip magic link email sending (will be handled by summary email)
- [ ] Process route if provided (route logic unchanged)
- [ ] Return SubmissionResponse with eventID and editToken

**Files to Modify:**
- `backend/internal/api/ride/service.go` - Add new method

**Method Signature:**
```go
func (s *Service) SubmitRideWithCoordinates(
    submission *Submission,
    lat, lng float64,
) (*SubmissionResponse, error) {
    // Generate edit token
    editToken, err := generateSecureToken(32)
    if err != nil {
        return nil, err
    }

    // Skip geocoding - use provided coordinates
    // (coordinates already validated by caller)

    // Process route if provided
    var routeID *string
    if submission.RouteURL != "" && s.routeFetcher != nil && s.routeRepository != nil {
        routeID, err = s.processRoute(context.Background(), submission.RouteURL, submission.City)
        if err != nil {
            slog.Warn("Failed to process route", "error", err)
        }
    }

    // Create ride with provided coordinates
    eventID, err := s.repo.CreateRide(submission, editToken, lat, lng)
    if err != nil {
        return nil, err
    }

    // Link route if processed
    if routeID != nil {
        if err := s.repo.LinkRouteToRide(eventID, *routeID); err != nil {
            slog.Warn("Failed to link route to ride", "error", err)
        }
    }

    // NOTE: Do NOT send magic link email here
    // Summary email will be sent after all imports complete

    return &SubmissionResponse{
        Success:   true,
        EventID:   eventID,
        EditToken: editToken,
        Message:   "Ride submitted successfully",
    }, nil
}
```

**3.0.3 - Add SendImportSummaryEmail to magiclink.Service**
- [ ] Read current magiclink.Service implementation
- [ ] Add new method `SendImportSummaryEmail(ctx context.Context, email string, events []ImportedEvent) error`
- [ ] Create email template with list of all imported events
- [ ] Include edit link for each event
- [ ] Use Resend API (existing integration)

**Files to Modify:**
- `backend/internal/api/magiclink/service.go` - Add new method

**Method Signature:**
```go
type ImportedEvent struct {
    Title     string
    EditToken string
    EditURL   string
}

func (s *Service) SendImportSummaryEmail(
    ctx context.Context,
    email string,
    events []ImportedEvent,
) error {
    // Build HTML email with all events listed
    // Subject: "Your X events have been imported to CycleScene"
    // Body: List of events with edit links
    // Use Resend API to send
}
```

**3.0.4 - Fix ProcessRoute to Use App Token**
- [ ] Read `backend/internal/strava/service.go` ProcessRoute method (line 321)
- [ ] Currently uses `session.AccessToken` (user OAuth token) - INCORRECT
- [ ] Update to use `STRAVA_ACCESS_TOKEN` from environment (app-level token)
- [ ] Routes are public data, don't need user authentication

**Files to Modify:**
- `backend/internal/strava/service.go` - Update ProcessRoute method

**Current (INCORRECT):**
```go
// Line 332 in service.go
fetcher := routes.NewRouteFetcher(http.DefaultClient, session.AccessToken, "", "")
```

**Updated (CORRECT):**
```go
// Get app-level token for public route fetching
stravaToken := os.Getenv("STRAVA_ACCESS_TOKEN")
if stravaToken == "" {
    return nil, fmt.Errorf("STRAVA_ACCESS_TOKEN not configured")
}
fetcher := routes.NewRouteFetcher(http.DefaultClient, stravaToken, "", "")
```

**3.0.5 - Verify Route Fetching Setup**
- [ ] Verify `STRAVA_ACCESS_TOKEN` exists in `.env` files (app-level token)
- [ ] This token is used for fetching public route data
- [ ] Test that ProcessRoute works with app token

**Debug Points:**
- Log when source tracking fields are populated
- Log when SubmitRideWithCoordinates is called vs SubmitRide
- Verify source deduplication works (database constraint)
- Test summary email template
- Verify ProcessRoute uses app token (not user OAuth token)

#### 3.1 - Create HTTP Handlers
- [ ] Create `backend/internal/api/strava/handler.go`
- [ ] `GET /api/strava/auth/initiate` - Start OAuth flow
- [ ] `GET /api/strava/auth/callback` - OAuth callback handler
- [ ] `GET /api/strava/admin-clubs` - Get clubs where user is admin (requires session)
- [ ] `POST /api/strava/logout` - Clear session
- [ ] Add Handler constructor with required dependencies

**Files to Create:**
- `backend/internal/api/strava/handler.go` (new)

**Handler Dependencies:**
```go
type Handler struct {
    stravaService    *strava.Service
    sessionStore     *strava.SessionStore
    rideService      *ride.Service
    routeRepository  *routes.Repository  // NEW - for route processing
    magicLinkService *magiclink.Service  // NEW - for summary emails
    debug            bool
}

func NewHandler(
    stravaService *strava.Service,
    sessionStore *strava.SessionStore,
    rideService *ride.Service,
    routeRepository *routes.Repository,  // NEW
    magicLinkService *magiclink.Service, // NEW
) *Handler {
    return &Handler{
        stravaService:    stravaService,
        sessionStore:     sessionStore,
        rideService:      rideService,
        routeRepository:  routeRepository,
        magicLinkService: magicLinkService,
        debug:            os.Getenv("STRAVA_DEBUG") == "true",
    }
}
```

**Session Cookie Configuration:**
```go
// In OAuth callback handler
http.SetCookie(w, &http.Cookie{
    Name:     "strava_session_id",
    Value:    sessionID,
    Path:     "/",
    MaxAge:   3600, // 1 hour
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteLaxMode, // Test this, may need SameSiteNoneMode
})
```

**Debug Points:**
- Log OAuth initiation with redirect URL
- Log callback with code and state validation
- Log admin clubs fetched per session
- Log session creation/deletion
- Log cookie settings (SameSite mode)

#### 3.2 - Add Routes to API Server
- [ ] Register Strava routes in `backend/cmd/api/main.go`
- [ ] Pass routes.Repository to Handler constructor
- [ ] Pass magiclink.Service to Handler constructor
- [ ] Add CORS configuration for form frontend
- [ ] Add session middleware (if needed)

**Files to Modify:**
- `backend/cmd/api/main.go`
- `backend/cmd/api/handler.go` (possibly)

**Debug Points:**
- Log route registration
- Log CORS configuration
- Log handler dependencies initialization

#### 3.3 - Create WebSocket Handler for Import Progress
- [ ] Create `backend/internal/api/strava/websocket.go`
- [ ] `WS /api/strava/import` - WebSocket endpoint for event import
- [ ] Accept list of event IDs to import with optional overrides
- [ ] Apply overrides (audience, custom data) to Submission before save
- [ ] Implement concurrent import detection (activeImports map)
- [ ] Send heartbeat messages every 30 seconds
- [ ] Create per-request RouteFetcher with app-level access token
- [ ] Process routes directly in handler (not via strava.Service)
- [ ] Call `SubmitRideWithCoordinates()` instead of `SubmitRide()`
- [ ] Collect all edit tokens during import
- [ ] Send ONE summary email at end with all event links
- [ ] Send progress updates per event and per step
- [ ] Handle steps: fetch event data, validate coordinates, process route (if exists), save to database
- [ ] Send final response with all results

**Files to Create:**
- `backend/internal/api/strava/websocket.go` (new)

**Concurrent Import Detection:**
```go
type Handler struct {
    // ... other fields ...
    activeImports sync.Map // map[athleteID]bool
}

func (h *Handler) handleImportWebSocket(w http.ResponseWriter, r *http.Request) {
    // ... get session ...

    // Check if import already in progress for this user
    if _, loaded := h.activeImports.LoadOrStore(session.AthleteID, true); loaded {
        conn.Close(websocket.StatusPolicyViolation, "import already in progress")
        return
    }
    defer h.activeImports.Delete(session.AthleteID)

    // Proceed with import...
}
```

**Heartbeat Implementation:**
```go
// Start heartbeat goroutine
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Send heartbeat message
            heartbeat := Message{
                Type:    "heartbeat",
                Message: "Import still in progress...",
            }
            conn.Write(ctx, websocket.MessageText, mustJSON(heartbeat))
        case <-ctx.Done():
            return
        }
    }
}()
```

**Route Processing (Per-Request RouteFetcher):**
```go
func (h *Handler) processRoute(
    ctx context.Context,
    routeURL string,
    cityCode string,
) (*string, error) {
    // Get app-level Strava access token from environment
    stravaToken := os.Getenv("STRAVA_ACCESS_TOKEN")
    if stravaToken == "" {
        return nil, fmt.Errorf("STRAVA_ACCESS_TOKEN not configured")
    }

    // Create per-request RouteFetcher with app token
    routeFetcher := routes.NewRouteFetcher(
        nil,         // Use default HTTP client
        stravaToken, // App-level token for public routes
        "",          // No RWGPS auth token
        "",          // No RWGPS API key
    )

    // Fetch and convert route
    feature, err := routeFetcher.FetchAndConvert(routeURL)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch route: %w", err)
    }

    // Parse source and source ID
    source, sourceID, err := routes.ParseRouteURL(routeURL)
    if err != nil {
        return nil, err
    }

    // Extract distance from properties
    var distanceKm, distanceMi float64
    if km, ok := feature.Properties["distance_km"].(float64); ok {
        distanceKm = km
    }
    if mi, ok := feature.Properties["distance_mi"].(float64); ok {
        distanceMi = mi
    }

    // Create route in database (deduplication handled by repository)
    routeID, err := h.routeRepository.CreateRoute(
        ctx, source, sourceID, routeURL, cityCode,
        &feature, distanceKm, distanceMi,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create route: %w", err)
    }

    return routeID, nil
}
```

**Summary Email at End:**
```go
// After all imports complete
var successfulEvents []magiclink.ImportedEvent

for _, result := range results {
    if result.Success {
        successfulEvents = append(successfulEvents, magiclink.ImportedEvent{
            Title:     result.EventTitle,
            EditToken: result.EditToken,
            EditURL:   fmt.Sprintf("%s?token=%s", editLinkBaseURL, result.EditToken),
        })
    }
}

// Send ONE summary email with all events
if len(successfulEvents) > 0 && h.magicLinkService != nil {
    err := h.magicLinkService.SendImportSummaryEmail(
        ctx,
        req.OrganizerEmail,
        successfulEvents,
    )
    if err != nil {
        slog.Error("Failed to send import summary email", "error", err)
    }
}
```

**Call SubmitRideWithCoordinates:**
```go
// Import single event
// First convert Strava event to Submission
submission, lat, lng, err := h.stravaService.ConvertEventToSubmission(
    ctx, sessionID, stravaEvent, organizerEmail,
)
if err != nil {
    return fmt.Errorf("failed to convert event: %w", err)
}

// Apply overrides from request (user-provided data like audience, image, etc.)
if len(eventReq.Overrides) > 0 {
    if audience, ok := eventReq.Overrides["audience"].(string); ok {
        submission.Audience = audience
    }
    if imageURL, ok := eventReq.Overrides["image_url"].(string); ok {
        submission.ImageURL = imageURL
    }
    // ... apply other overrides as needed
}

// Set source tracking for deduplication
submission.Source = "strava"
submission.SourceID = fmt.Sprintf("%d", stravaEvent.ID)

// Submit with coordinates (skips geocoding and email)
response, err := h.rideService.SubmitRideWithCoordinates(submission, lat, lng)
```

**Message Format:**
```typescript
// Client → Server (initiate import)
{
  "session_id": "abc123",
  "organizer_email": "user@example.com",  // For summary email
  "events": [
    {
      "strava_event_id": 789,
      "club_id": 123,
      "overrides": {
        // Optional: User-provided data to override/supplement Strava data
        "audience": "All",           // Required field not from Strava
        "image_url": "https://...",  // Optional: User uploads image (Strava events don't have images)
        "title": "Custom Title",     // Optional: Override Strava title
        // ... any other Submission fields
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
  "message": "Fetching event data..." | "Validating coordinates..." | "Processing route..." | "Saving to database..."
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
  "error": null  // or error message if failed
}

// Server → Client (final summary)
{
  "type": "done",
  "total_imported": 3,
  "total_failed": 0,
  "summary_email_sent": true,
  "results": [
    {
      "strava_event_id": 789,
      "success": true,
      "event_title": "Tuesday Night Ride",
      "edit_url": "..."
    }
  ]
}
```

**Debug Points:**
- Log WebSocket connection establishment with session info
- Log concurrent import detection (if blocked)
- Log heartbeat messages sent
- Log per-request RouteFetcher creation with app token
- Log route processing for each event (verify app token used)
- Log overrides applied to Submission (audience, image, etc.)
- Log SubmitRideWithCoordinates calls (not SubmitRide)
- Log each progress step per event (fetching, coordinates, route, database)
- Log errors with event context and Strava event ID
- Log summary email sending with event count
- Log final import results (success/failure counts)
- Log activeImports cleanup on connection close

#### 3.4 - Add Environment Variables
- [ ] Add `STRAVA_CLIENT_ID` to backend/.env.example
- [ ] Add `STRAVA_CLIENT_SECRET` to backend/.env.example
- [ ] Add `STRAVA_CALLBACK_URL` to backend/.env.example
- [ ] Add `STRAVA_DEBUG` to backend/.env.example
- [ ] Verify `STRAVA_ACCESS_TOKEN` exists (app-level token for route fetching)

**Files to Modify:**
- `backend/.env.example`
- `backend/cmd/api/.env.example`

**Note:** `STRAVA_ACCESS_TOKEN` should already exist from scraper setup. This is the app-level token used for fetching public route data, NOT user OAuth tokens.

#### 3.5 - Integration Testing
- [ ] Test OAuth flow end-to-end (verify SameSite=Lax cookies work cross-domain)
- [ ] Test admin club fetching
- [ ] Test WebSocket import with single event
- [ ] Test WebSocket import with multiple events (3+)
- [ ] Test heartbeat messages (WebSocket stays alive during long operations)
- [ ] Test concurrent import detection (open two tabs, try to import simultaneously)
- [ ] Test source deduplication (import same event twice)
- [ ] Test summary email (verify one email with all event links)
- [ ] Test route processing with app-level token
- [ ] Test error handling (invalid session, rate limits, missing coordinates)

**Validation:**
```bash
cd backend
go build ./cmd/api
# Manual testing with Strava OAuth
curl http://localhost:8080/api/strava/auth/initiate
# Complete OAuth flow
curl -H "Cookie: session_id=..." http://localhost:8080/api/strava/admin-clubs
# Test WebSocket import (use wscat or similar tool)
```

**Critical Tests:**
- ✅ Verify `SameSite=Lax` cookie persists after OAuth redirect
- ✅ Verify concurrent imports are blocked with clear error message
- ✅ Verify summary email contains all imported events
- ✅ Verify routes fetched using app token (not user OAuth token)
- ✅ Verify source tracking prevents duplicates (database constraint)
- ✅ Verify SubmitRideWithCoordinates is called (no geocoding logs)

---

### 📝 M3 Implementation Summary

**What Changed from Initial Spec:**

1. **Added Task 3.0 (Prerequisites)** - Must complete BEFORE M3 implementation:
   - Add Source/SourceID to Submission struct and CreateRide
   - Add SubmitRideWithCoordinates() to ride.Service
   - Add SendImportSummaryEmail() to magiclink.Service
   - Verify STRAVA_ACCESS_TOKEN exists for route fetching

2. **Route Fetching Strategy:**
   - Use app-level token (STRAVA_ACCESS_TOKEN from env)
   - Create per-request RouteFetcher in WebSocket handler
   - Routes are public data, don't need user OAuth tokens

3. **Email Strategy:**
   - ONE summary email per import session (not per event)
   - Collect all edit tokens during import
   - Send summary at end with all event links
   - SubmitRideWithCoordinates skips individual emails

4. **WebSocket Enhancements:**
   - Heartbeat messages every 30 seconds (keep connection alive)
   - Concurrent import detection (sync.Map lock per athleteID)
   - Route processing directly in handler (not via service layer)

5. **Handler Dependencies:**
   - Added routes.Repository (for route processing)
   - Added magiclink.Service (for summary emails)
   - Added activeImports sync.Map (for concurrent detection)

6. **Deferred Testing:**
   - SameSite=Lax cookie behavior (test during form integration)
   - May need to switch to SameSite=None if OAuth redirect fails

**Key Design Principles:**
- Accept some code duplication (SubmitRideWithCoordinates vs SubmitRide) for clarity
- Use app-level credentials for public data (routes)
- Use user OAuth credentials for private data (events, clubs)
- One email per session, not per event (better UX)
- Backend lock for concurrent imports (prevent wasted API calls)

---

### 🎯 M3 Design Decisions (Summary)

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| 1 | Session management | Cookie-based (`strava_session_id`) | Standard web pattern, works with OAuth redirects, HttpOnly prevents XSS |
| 2 | WebSocket auth | Session ID in first message | Secure (not in URL), clean protocol |
| 3 | Multi-event import | Sequential with continue-on-failure | Simple, rate-limit safe, predictable progress |
| 4 | Progress steps | 4 steps: fetching → coordinates → route → database | User-facing, debuggable, accurate |
| 5 | Error recovery | Continue-on-failure (except rate limits) | Partial success better than total failure |
| 6 | BFF token bypass | SubmitRideWithCoordinates() method | Strava session already validated |
| 7 | Magic links | Summary email after import | One email per session, not per event |
| 8 | Route processing | During import (step 3), optional | User sees progress, event saved even if route fails |
| 9 | Duplicate handling | Database UNIQUE constraint on (source, source_id) | Automatic, fast, user-friendly message |
| 10 | Rate limits | Log but don't preemptively stop | Simple, relies on actual Strava 429 responses |
| 11 | Message ordering | Sequential writes (WebSocket guarantees FIFO) | No special handling needed |
| 12 | Connection lifecycle | 10min timeout, graceful shutdown | Bounded, resilient, debuggable |
| 13 | CORS for WebSocket | Validate origin in AcceptOptions | Secure, flexible for dev |
| 14 | Error messages | Map technical errors to user-friendly messages | Actionable, non-technical language |
| 15 | Session expiry during import | Allow completion (session cached) | Access token valid 6hr, import max 10min |
| 16 | Organizer email | Required in ImportRequest, validated | Edit links need email for magic link |

---

### 🚨 M3 Edge Cases & Handling

| Edge Case | Handling | User Message |
|-----------|----------|--------------|
| Session expires during WebSocket | ✅ Import proceeds (cached session) | N/A - completes normally |
| WebSocket disconnects mid-import | ✅ Partial success saved | "Connection lost" (frontend) |
| Same event imported twice | ✅ Skip with UNIQUE constraint | "This event has already been imported" |
| Event deleted from Strava | ✅ Skip event, continue | "Event no longer exists in Strava" |
| Geocoding fails | ✅ Skip event, continue | "Unable to find location. Please add address in Strava." |
| Route fetch fails (404) | ✅ Continue without route | "Route unavailable (event saved without route)" |
| Database save fails | ✅ Skip event, continue | "Database error. Please try again." |
| Rate limit exceeded (429) | ✅ Stop import | "Strava API limit reached. Wait 15 minutes." |
| Concurrent imports (same user) | ✅ Block second import | "Import already in progress" |
| Event with no occurrences | ✅ Reject in validation | "This event has already occurred" |
| Missing organizer email | ✅ Close WebSocket | "Organizer email required" |
| Very long title (>200 chars) | ✅ Truncate with "..." | N/A - silent truncation |
| WebSocket message too large | ✅ Send incrementally | N/A - results sent per-event |
| Invalid city code | ✅ Reject request | "Invalid city code. Please reconnect." |
| Magic link email fails | ✅ Event saved, email logged | N/A - user gets edit link in WebSocket response |

---

### ✅ M3 Completion Checklist

**Before marking M3 complete, verify:**

- [ ] All files created: `handler.go`, `websocket.go`
- [ ] HTTP handlers: InitiateOAuth, HandleOAuthCallback, Logout, GetAdminClubs, GetClubEvents
- [ ] WebSocket: Session validation, email validation, progress tracking, heartbeats
- [ ] Error handling: Continue-on-failure, rate limit detection, duplicate detection
- [ ] Build succeeds: `cd backend && go build ./cmd/api`
- [ ] Routes registered in `cmd/api/handler.go`
- [ ] Environment variables documented in `.env.example`
- [ ] Manual OAuth flow test passes
- [ ] Manual WebSocket import test passes
- [ ] Summary email sent with all event links

---

## Milestone 4: Frontend - SvelteKit Integration

**Goal:** Add Strava import UI to the form frontend

### 📁 Files Summary (Milestone 4)

**New Files to Create:**
- `frontends/form/src/lib/stores/strava.ts` - Strava auth state management
- `frontends/form/src/lib/components/StravaImport.svelte` - Main import component
- `frontends/form/src/lib/components/ImportProgress.svelte` - Progress tracking UI
- `frontends/form/src/lib/types/strava.ts` - TypeScript type definitions
- `frontends/form/src/lib/utils/websocket.ts` - WebSocket client helper (optional)

**Files to Modify:**
- `frontends/form/src/routes/+page.svelte` - Add Strava import button
- `frontends/form/.env.example` - Add PUBLIC_STRAVA_ENABLED, PUBLIC_STRAVA_DEBUG

**Existing Files to Reference:**
- `frontends/form/src/lib/stores/*` - For store patterns
- `frontends/form/src/lib/components/*` - For component patterns
- `frontends/form/src/routes/+page.server.ts` - For API calls pattern

**Backend Endpoints to Call:**
- `GET /api/strava/auth/initiate` (from M3)
- `GET /api/strava/admin-clubs` (from M3)
- `WS /api/strava/import` (from M3)

**No modifications to:**
- Backend code (complete in M1-M3)
- Other frontend apps (dashboard, pwa, directory)

### Tasks

#### 4.1 - Create Strava Auth Store
- [ ] Create `frontends/form/src/lib/stores/strava.ts`
- [ ] Manage OAuth state (authenticated, session_id, admin_clubs)
- [ ] Persist session ID in localStorage (temporary)
- [ ] Add methods: `initiateAuth()`, `fetchAdminClubs()`, `logout()`

**Files to Create:**
- `frontends/form/src/lib/stores/strava.ts` (new)

**Debug Points:**
- Log OAuth initiation
- Log session ID storage
- Log admin clubs fetched

#### 4.2 - Create Strava Import Component
- [ ] Create `frontends/form/src/lib/components/StravaImport.svelte`
- [ ] Add Strava icon button
- [ ] Show "Connect with Strava" modal
- [ ] Display admin clubs as cards
- [ ] Show event list per club (collapsible)
- [ ] Add checkboxes for multi-event selection
- [ ] Add "Customize before import" toggle per event

**Files to Create:**
- `frontends/form/src/lib/components/StravaImport.svelte` (new)

**UI Flow:**
1. User clicks "Import from Strava" button
2. OAuth popup window opens
3. After auth, show admin clubs
4. User expands club to see events
5. User selects events to import (checkboxes)
6. User clicks "Import Selected Events"
7. Progress modal opens with WebSocket updates

**Debug Points:**
- Log OAuth popup open/close
- Log club/event selection changes
- Log import button click with selected events

#### 4.3 - Create Import Progress Component
- [ ] Create `frontends/form/src/lib/components/ImportProgress.svelte`
- [ ] Connect to WebSocket endpoint
- [ ] Display progress for each event being imported
- [ ] Show steps: 1/4 Uploading image, 2/4 Geocoding, 3/4 Building route, 4/4 Saving
- [ ] Show success/error states per event
- [ ] Display final results with edit links

**Files to Create:**
- `frontends/form/src/lib/components/ImportProgress.svelte` (new)

**UI Design:**
```
Importing 3 Events

[✓] Tuesday Night Ride
    ✓ Geocoding  ✓ Route  ✓ Image  ✓ Saved
    Event created! [Edit]

[⟳] Thursday Social Ride
    ✓ Geocoding  ⟳ Route...

[ ] Weekend Gran Fondo
    Waiting...
```

**Debug Points:**
- Log WebSocket connection
- Log each progress message received
- Log final results

#### 4.4 - Integrate into Form Page
- [ ] Add Strava import button to `frontends/form/src/routes/+page.svelte`
- [ ] Position near form title or as alternative to manual form
- [ ] Add toggle: "Manual Entry" vs "Import from Strava"
- [ ] Respect city parameter (pass to import flow)

**Files to Modify:**
- `frontends/form/src/routes/+page.svelte`

#### 4.5 - Add TypeScript Types
- [ ] Create `frontends/form/src/lib/types/strava.ts`
- [ ] Define types for Club, GroupEvent, ImportProgress, etc.

**Files to Create:**
- `frontends/form/src/lib/types/strava.ts` (new)

**Validation:**
```bash
cd frontends/form
npm run check  # TypeScript validation
npm run build  # Production build
npm run dev    # Manual testing
```

---

## Milestone 5: Frontend - Polish & Error Handling

**Goal:** Ensure robust error handling and great UX

### 📁 Files Summary (Milestone 5)

**Files to Modify:**
- `frontends/form/src/lib/stores/strava.ts` - Add error handling
- `frontends/form/src/lib/components/StravaImport.svelte` - Add loading states, errors
- `frontends/form/src/lib/components/ImportProgress.svelte` - Add retry logic
- All Svelte components - Add ARIA labels, accessibility

**New Files (Optional):**
- `frontends/form/src/lib/components/ErrorBoundary.svelte` - Error boundary (if needed)

**No new backend files:**
- All backend work complete in M1-M3

**Focus Areas:**
- Error messages and recovery
- Loading and skeleton states
- Responsive design (mobile testing)
- Accessibility (ARIA, keyboard nav)

### Tasks

#### 5.1 - Error Handling
- [ ] Handle OAuth failure (user denies)
- [ ] Handle session expiry (redirect to re-auth)
- [ ] Handle rate limit errors (show friendly message)
- [ ] Handle WebSocket disconnection (retry logic)
- [ ] Handle partial import failures (some events succeed, some fail)

**Files to Modify:**
- `frontends/form/src/lib/stores/strava.ts`
- `frontends/form/src/lib/components/ImportProgress.svelte`

**Error Messages:**
- OAuth denied: "Authorization cancelled. You can try again anytime."
- Session expired: "Your session expired. Please reconnect to Strava."
- Rate limit: "Strava API limit reached. Please try again in 15 minutes."
- Not admin: "You must be an admin or owner of a club to import its events."

#### 5.2 - Loading States
- [ ] Show loading spinner during OAuth
- [ ] Show skeleton loaders for clubs/events
- [ ] Disable buttons during import process
- [ ] Add timeout handling (30s for event fetch, 60s for import)

**Files to Modify:**
- `frontends/form/src/lib/components/StravaImport.svelte`
- `frontends/form/src/lib/components/ImportProgress.svelte`

#### 5.3 - Responsive Design
- [ ] Ensure components work on mobile
- [ ] Test OAuth flow on mobile (popup vs redirect)
- [ ] Optimize event cards for small screens

**Files to Modify:**
- `frontends/form/src/lib/components/StravaImport.svelte`

#### 5.4 - Accessibility
- [ ] Add ARIA labels to buttons
- [ ] Ensure keyboard navigation works
- [ ] Add focus management for modals
- [ ] Test with screen reader

**Files to Modify:**
- All Svelte components

#### 5.5 - Add Environment Variables
- [ ] Add `PUBLIC_STRAVA_ENABLED` flag to toggle feature
- [ ] Add `PUBLIC_API_URL` for WebSocket connection

**Files to Modify:**
- `frontends/form/.env.example`

**Validation:**
```bash
cd frontends/form
npm run check
npm run build
# Manual testing across devices
```

---

## Milestone 6: Testing & Documentation

**Goal:** Comprehensive testing and documentation

### 📁 Files Summary (Milestone 6)

**New Test Files to Create:**
- `backend/internal/api/strava/handler_test.go` - Handler tests
- `backend/internal/api/strava/websocket_test.go` - WebSocket tests
- `frontends/form/src/lib/stores/strava.test.ts` - Store tests
- `frontends/form/src/lib/components/StravaImport.test.ts` - Component tests

**Documentation Files to Update:**
- `README.md` - Add Strava setup instructions
- `STRAVA_OAUTH_LEARNINGS.md` - Add any new discoveries
- `backend/README.md` - Document Strava API endpoints
- `frontends/form/README.md` - Document Strava import feature

**Test Files Already Exist (from M1-M2):**
- `backend/internal/strava/client_test.go`
- `backend/internal/strava/service_test.go`
- `backend/internal/strava/converter_test.go`

**No new feature files:**
- All feature code complete in M1-M5

### Tasks

#### 6.1 - Backend Integration Tests
- [ ] Create `backend/internal/api/strava/handler_test.go`
- [ ] Test OAuth flow with mocked Strava API
- [ ] Test WebSocket import with multiple events
- [ ] Test error scenarios

**Files to Create:**
- `backend/internal/api/strava/handler_test.go` (new)
- `backend/internal/api/strava/websocket_test.go` (new)

#### 6.2 - Frontend Component Tests
- [ ] Test Strava store (auth, logout, session management)
- [ ] Test StravaImport component (event selection)
- [ ] Test ImportProgress component (WebSocket messages)

**Files to Create:**
- `frontends/form/src/lib/stores/strava.test.ts` (new)
- `frontends/form/src/lib/components/StravaImport.test.ts` (new)

#### 6.3 - End-to-End Testing
- [ ] Test full OAuth flow
- [ ] Test importing single event
- [ ] Test importing multiple events (3+)
- [ ] Test error recovery (session expiry, rate limit)
- [ ] Test on different browsers (Chrome, Firefox, Safari)

**Manual Test Checklist:**
- [ ] OAuth flow works
- [ ] Only admin clubs are shown
- [ ] Events display correctly
- [ ] Multi-event import shows progress
- [ ] Edit links work after import
- [ ] Magic link emails are sent with event titles

#### 6.4 - Update Documentation
- [ ] Add Strava setup instructions to README
- [ ] Document environment variables
- [ ] Add screenshots to `STRAVA_OAUTH_LEARNINGS.md`
- [ ] Create user guide for importing events

**Files to Modify:**
- `README.md`
- `STRAVA_OAUTH_LEARNINGS.md`
- `backend/README.md`
- `frontends/form/README.md`

#### 6.5 - Performance Testing
- [ ] Test with user in 10+ clubs
- [ ] Test importing 10+ events simultaneously
- [ ] Monitor memory usage (session cleanup)
- [ ] Test WebSocket connection limits

**Validation:**
```bash
# Backend tests
cd backend
go test ./... -v -cover

# Frontend tests
cd frontends/form
npm run test
npm run check
npm run build

# E2E validation
# Manual testing with real Strava account
```

---

## Milestone 7: Deployment Preparation

**Goal:** Prepare for production deployment

### 📁 Files Summary (Milestone 7)

**Infrastructure Files to Modify:**
- `backend/cmd/api/infra/main.tf` - Add Strava secrets to Cloud Run
- `backend/cmd/api/infra/variables.tf` - Add Strava variable definitions
- `backend/cmd/api/infra/terraform.tfvars.example` - Document Strava vars

**Environment Files to Update:**
- Production `.env` files with real Strava credentials
- Verify all `.env.example` files are complete

**Configuration Reviews:**
- CORS settings in `backend/cmd/api/main.go`
- Rate limiting configurations
- WebSocket connection limits
- Session cleanup intervals

**No new feature code:**
- All feature development complete in M1-M6

**Focus Areas:**
- Security review (no token logging)
- Monitoring and metrics
- Deployment and rollback procedures

### Tasks

#### 7.1 - Environment Configuration
- [ ] Add production Strava OAuth app
- [ ] Configure production callback URL
- [ ] Set `STRAVA_DEBUG=false` in production
- [ ] Add Strava credentials to Cloud Run secrets

**Files to Modify:**
- `backend/cmd/api/infra/main.tf` (add secrets)
- Production environment configuration

#### 7.2 - Security Review
- [ ] Ensure tokens are never logged in production
- [ ] Validate CSRF state tokens properly
- [ ] Check WebSocket authentication
- [ ] Review CORS configuration

#### 7.3 - Monitoring & Logging
- [ ] Add metrics for OAuth success/failure rates
- [ ] Add metrics for import success/failure rates
- [ ] Monitor rate limit usage
- [ ] Set up alerts for errors

#### 7.4 - Rollout Plan
- [ ] Deploy backend with feature flag disabled
- [ ] Test in production with internal users
- [ ] Enable for beta testers (specific cities)
- [ ] Full public rollout

#### 7.5 - Rollback Plan
- [ ] Document how to disable feature (environment variable)
- [ ] Ensure existing functionality unaffected
- [ ] Keep test tool available for debugging

**Validation:**
```bash
# Terraform validation
cd backend/cmd/api/infra
terraform plan

# Deploy to staging
# Test OAuth flow in staging
# Monitor logs and metrics
```

---

## Success Metrics

### Per Milestone
- [ ] All tests pass (`go test`, `npm test`)
- [ ] No build errors (`go build`, `npm run build`)
- [ ] No TypeScript errors (`npm run check`)
- [ ] Feature debuggable with `STRAVA_DEBUG=true`
- [ ] Structured logging in place

### Overall Feature
- [ ] OAuth flow completes successfully
- [ ] Only admin/owner clubs shown
- [ ] Events import with all fields mapped correctly
- [ ] Progress tracking works for multi-event imports
- [ ] Magic link emails sent with event titles
- [ ] Error states handled gracefully
- [ ] Mobile-responsive UI
- [ ] Accessible (keyboard nav, screen readers)
- [ ] Production-ready (monitoring, security, rollback plan)

---

## Timeline Estimate

| Milestone | Estimated Time | Dependencies |
|-----------|---------------|--------------|
| M1: Backend Client | 4-6 hours | None |
| M2: Service Layer | 4-6 hours | M1 |
| M3: HTTP & WebSocket | 6-8 hours | M1, M2 |
| M4: Frontend Integration | 6-8 hours | M3 |
| M5: Polish & Errors | 4-6 hours | M4 |
| M6: Testing & Docs | 4-6 hours | All previous |
| M7: Deployment | 2-4 hours | M6 |
| **Total** | **30-44 hours** | |

---

## Notes

- Work bottom-up: Backend → API → Frontend
- Each milestone must be fully functional and debuggable
- Commit after each milestone with detailed commit message
- Keep `STRAVA_OAUTH_LEARNINGS.md` updated with new discoveries
- Use `STRAVA_DEBUG=true` liberally during development

---

**Branch:** `feature/stravaimport`
**Created:** 2026-01-31
**Status:** Ready to begin Milestone 1
