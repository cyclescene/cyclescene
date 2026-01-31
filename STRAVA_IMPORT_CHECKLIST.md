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

### Tasks

#### 1.1 - Create Strava Client Package
- [ ] Create `backend/internal/strava/client.go`
- [ ] Implement OAuth token exchange
- [ ] Implement token refresh logic
- [ ] Add rate limit tracking and headers parsing
- [ ] Add debug logging controlled by `STRAVA_DEBUG` env var

**Files to Create/Modify:**
- `backend/internal/strava/client.go` (new)

**Debug Points:**
- Log OAuth token exchange requests/responses
- Log rate limit usage from headers
- Log all API requests when `STRAVA_DEBUG=true`

#### 1.2 - Implement Club Methods
- [ ] `GetAthleteClubs()` - Fetch all clubs user belongs to
- [ ] `GetClubDetails(clubID)` - Get detailed club info with admin flags
- [ ] `FilterAdminClubs(clubs)` - Helper to filter clubs where user is admin/owner

**Files to Modify:**
- `backend/internal/strava/client.go`

**Debug Points:**
- Log club fetching with athlete ID
- Log admin/owner status for each club
- Log filtered admin clubs count

#### 1.3 - Implement Event Methods
- [ ] `GetClubEvents(clubID)` - Fetch group events for a club
- [ ] `GetRoute(routeID)` - Fetch route details (if event has route)

**Files to Modify:**
- `backend/internal/strava/client.go`

**Debug Points:**
- Log event fetching per club
- Log route fetching when route_id exists
- Log event counts returned

#### 1.4 - Add Error Handling
- [ ] Handle 401 (invalid/expired token)
- [ ] Handle 429 (rate limit exceeded)
- [ ] Handle 404 (club/event not found)
- [ ] Create custom error types

**Files to Modify:**
- `backend/internal/strava/client.go`
- `backend/internal/strava/errors.go` (new)

**Debug Points:**
- Log all API errors with status codes
- Log retry attempts for rate limits

#### 1.5 - Write Unit Tests
- [ ] Test OAuth token exchange
- [ ] Test admin club filtering
- [ ] Test error handling
- [ ] Mock HTTP responses

**Files to Create:**
- `backend/internal/strava/client_test.go` (new)

**Validation:**
```bash
cd backend
go test ./internal/strava/... -v
go build ./cmd/api
```

---

## Milestone 2: Backend - Service Layer

**Goal:** Build business logic for OAuth sessions and event conversion

### Tasks

#### 2.1 - Create Strava Service
- [ ] Create `backend/internal/strava/service.go`
- [ ] Integrate with SessionStore
- [ ] Integrate with Strava Client
- [ ] Add method to initiate OAuth flow
- [ ] Add method to handle OAuth callback
- [ ] Add method to get admin clubs for session

**Files to Create/Modify:**
- `backend/internal/strava/service.go` (new)

**Debug Points:**
- Log OAuth flow initiation with state token
- Log callback processing with athlete ID
- Log session creation/retrieval

#### 2.2 - Create Event Converter
- [ ] Create `backend/internal/strava/converter.go`
- [ ] Map Strava GroupEvent → CycleScene Submission
- [ ] Handle timezone conversion (UTC → local)
- [ ] Map city strings to city codes (Portland → pdx, Salt Lake City → slc)
- [ ] Set defaults for missing fields (duration, recurrence)

**Files to Create:**
- `backend/internal/strava/converter.go` (new)

**Debug Points:**
- Log each field mapping
- Log timezone conversions
- Log city code mapping
- Warn for missing/optional fields

#### 2.3 - Integrate Route Fetching
- [ ] Use existing `routes.RouteFetcher` for Strava routes
- [ ] Convert Strava route to GeoJSON
- [ ] Store in routes table with deduplication

**Files to Modify:**
- `backend/internal/strava/service.go`

**Debug Points:**
- Log route fetching initiation
- Log route conversion success/failure
- Log route deduplication (existing vs new)

#### 2.4 - Integrate Geocoding (Fallback)
- [ ] Use Strava's provided lat/lng as primary
- [ ] Fallback to geocoding if coordinates are 0,0
- [ ] Use existing `scraper.GeocodeQuery()` function

**Files to Modify:**
- `backend/internal/strava/converter.go`

**Debug Points:**
- Log when using Strava coordinates vs geocoding
- Log geocoding queries and results

#### 2.5 - Write Service Tests
- [ ] Test OAuth flow
- [ ] Test event conversion
- [ ] Test admin club filtering
- [ ] Mock Strava client responses

**Files to Create:**
- `backend/internal/strava/service_test.go` (new)
- `backend/internal/strava/converter_test.go` (new)

**Validation:**
```bash
cd backend
go test ./internal/strava/... -v -cover
go build ./cmd/api
```

---

## Milestone 3: Backend - HTTP Handlers & WebSocket

**Goal:** Create API endpoints for OAuth flow and WebSocket for progress tracking

### Tasks

#### 3.1 - Create HTTP Handlers
- [ ] Create `backend/internal/api/strava/handler.go`
- [ ] `GET /api/strava/auth/initiate` - Start OAuth flow
- [ ] `GET /api/strava/auth/callback` - OAuth callback handler
- [ ] `GET /api/strava/admin-clubs` - Get clubs where user is admin (requires session)
- [ ] `POST /api/strava/logout` - Clear session

**Files to Create:**
- `backend/internal/api/strava/handler.go` (new)

**Debug Points:**
- Log OAuth initiation with redirect URL
- Log callback with code and state validation
- Log admin clubs fetched per session
- Log session creation/deletion

#### 3.2 - Add Routes to API Server
- [ ] Register Strava routes in `backend/cmd/api/main.go`
- [ ] Add CORS configuration for form frontend
- [ ] Add session middleware

**Files to Modify:**
- `backend/cmd/api/main.go`
- `backend/cmd/api/handler.go` (possibly)

**Debug Points:**
- Log route registration
- Log CORS configuration

#### 3.3 - Create WebSocket Handler for Import Progress
- [ ] Create `backend/internal/api/strava/websocket.go`
- [ ] `WS /api/strava/import` - WebSocket endpoint for event import
- [ ] Accept list of event IDs to import
- [ ] Send progress updates: `{"current": 1, "total": 5, "step": "geocoding", "status": "success"}`
- [ ] Handle steps: image upload, geocoding, route building, database save
- [ ] Send final response with event IDs and edit tokens

**Files to Create:**
- `backend/internal/api/strava/websocket.go` (new)

**Message Format:**
```typescript
// Client → Server (initiate import)
{
  "session_id": "abc123",
  "events": [
    {
      "strava_event_id": 789,
      "club_id": 123,
      "customize": false,  // if true, user wants to edit before submit
      "overrides": {}      // optional field overrides
    }
  ]
}

// Server → Client (progress updates)
{
  "type": "progress",
  "event_index": 0,
  "total_events": 3,
  "strava_event_id": 789,
  "event_title": "Tuesday Night Ride",
  "step": "geocoding",
  "status": "in_progress" | "success" | "error",
  "message": "Geocoding starting location..."
}

// Server → Client (completion)
{
  "type": "complete",
  "event_index": 0,
  "strava_event_id": 789,
  "cyclescene_event_id": 42,
  "edit_token": "xyz123",
  "success": true
}

// Server → Client (final)
{
  "type": "done",
  "total_imported": 3,
  "total_failed": 0,
  "results": [...]
}
```

**Debug Points:**
- Log WebSocket connection establishment
- Log each progress step per event
- Log errors with event context
- Log final import results

#### 3.4 - Add Environment Variables
- [ ] Add `STRAVA_CLIENT_ID` to backend/.env.example
- [ ] Add `STRAVA_CLIENT_SECRET` to backend/.env.example
- [ ] Add `STRAVA_CALLBACK_URL` to backend/.env.example
- [ ] Add `STRAVA_DEBUG` to backend/.env.example

**Files to Modify:**
- `backend/.env.example`
- `backend/cmd/api/.env.example`

#### 3.5 - Integration Testing
- [ ] Test OAuth flow end-to-end
- [ ] Test admin club fetching
- [ ] Test WebSocket import with multiple events
- [ ] Test error handling (invalid session, rate limits)

**Validation:**
```bash
cd backend
go build ./cmd/api
# Manual testing with Strava OAuth
curl http://localhost:8080/api/strava/auth/initiate
# Complete OAuth flow
curl -H "Cookie: session_id=..." http://localhost:8080/api/strava/admin-clubs
```

---

## Milestone 4: Frontend - SvelteKit Integration

**Goal:** Add Strava import UI to the form frontend

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
