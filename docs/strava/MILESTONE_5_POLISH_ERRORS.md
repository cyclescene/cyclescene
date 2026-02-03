# Milestone 5: Frontend - Polish & Error Handling

**Goal:** Ensure robust error handling, polished UX, and accessibility compliance

**Prerequisites:** Milestones 1-4 complete

---

## Quick Reference Summary

### What M5 Adds

**Error Handling Improvements:**
- Proactive rate limit detection at 75% usage (25 request buffer for events/routes)
- Exact countdown timer showing minutes until rate limit reset (:00, :15, :30, :45)
- Automatic token refresh using Strava refresh_token (transparent to user)
- Manual reconnect fallback if automatic refresh fails
- Stop Import button during active imports
- Unlimited manual retry after WebSocket failures
- Monitoring DB logging for all errors and retries

**Loading & UX Polish:**
- Skeleton loaders (shadcn Skeleton) for clubs and events
- 60-second heartbeat timeout detection
- "Taking longer than expected..." warnings
- Button disable states to prevent double-clicks

**Accessibility (Comprehensive):**
- Live regions announcing progress updates for screen readers
- Focus management between steps (auto-focus as user progresses)
- Loading state announcements (aria-busy, role="status")
- Descriptive ARIA labels on all interactive elements
- Keyboard navigation verification and testing checklist
- Screen reader testing requirements (VoiceOver/NVDA)

**Mobile:**
- Verify current polling fallback works on iOS Safari & Android Chrome
- Add redirect flow if polling fails

**Pre-Milestone Blocker:**
- Fix 13 pre-existing TypeScript errors in non-Strava files (blocks build)

### Key Design Decisions

| Decision | Approach | Rationale |
|----------|----------|-----------|
| Rate limit check | Use GET /athlete/clubs as feeler | No extra API call, early detection |
| Rate limit threshold | 75% (75/100 requests) | Leaves 25 requests for events/routes |
| Token refresh | Automatic backend refresh + fallback | Transparent to user, uses OAuth refresh_token |
| WebSocket retry | Unlimited manual retries | User controls, log to monitoring DB |
| Mobile OAuth | Test polling fallback first | Current implementation should work |
| Skeleton loaders | Use shadcn Skeleton component | No custom code needed |
| Accessibility | Full WCAG AA compliance | Live regions, focus mgmt, screen reader support |

### Files Modified (10 frontend + 2 backend)

**Frontend:**
- `src/lib/api/strava.ts` - Rate limit & session error detection
- `src/lib/utils/strava-websocket.ts` - Manual retry, stop, timeout
- `src/lib/components/strava/StravaImport.svelte` - Focus management, step announcements
- `src/lib/components/strava/StravaImportButton.svelte` - ARIA labels, error messages
- `src/lib/components/strava/ImportProgress.svelte` - Live regions, retry/stop buttons
- `src/lib/components/strava/ClubList.svelte` - Skeleton loaders, loading announcements
- `src/lib/components/strava/EventCard.svelte` - ARIA labels, checkbox associations
- `src/lib/components/strava/EmailInput.svelte` - Auto-focus
- `src/lib/components/strava/ImportResults.svelte` - role="alert", announcements
- `src/lib/types/strava.ts` - Error classes (RateLimitError, SessionExpiredError)

**Backend:**
- `backend/internal/strava/client.go` - Rate limit parsing, token refresh
- `backend/internal/strava/session.go` - Automatic token refresh logic

### Implementation Order (Prioritized)

1. **Task 5.0** - Fix pre-existing TypeScript errors (BLOCKER)
2. **5.1.1** - Rate limit detection (high user impact)
3. **5.1.2** - Token refresh (prevents session expiry)
4. **5.1.4** - WebSocket retry/stop (critical for reliability)
5. **5.2.1** - Skeleton loaders (quick win, uses shadcn)
6. **5.4.1-5.4.7** - Accessibility (systematic implementation)
7. **5.3.1** - Mobile testing (verification task)
8. **5.1.3** - OAuth denial handling (polish)

### Success Criteria Checklist

- [ ] `npm run check` passes (0 TypeScript errors)
- [ ] `npm run build` succeeds
- [ ] Rate limit shows accurate countdown timer
- [ ] Token refresh happens transparently
- [ ] Stop Import button cancels active import
- [ ] Manual retry preserves completed events
- [ ] Skeleton loaders show during all loading states
- [ ] Screen reader announces progress updates
- [ ] Focus moves logically between steps
- [ ] Keyboard-only user can complete full flow
- [ ] Mobile OAuth works on iOS Safari & Android Chrome
- [ ] All errors logged to monitoring DB

---

## Pre-Milestone Blockers

**CRITICAL:** The following TypeScript errors exist in non-Strava files and MUST be fixed first to ensure the codebase builds cleanly.

### Task 5.0 - Fix Pre-existing TypeScript Errors

**Status:** Not Started
**Priority:** BLOCKER - Must complete before any M5 work

These 13 errors block `npm run check` and `npm run build`. Fix in a separate commit before M5 implementation.

| # | File | Line | Issue | Fix Approach |
|---|------|------|-------|--------------|
| 1 | `routes/group/+page.server.ts` | 113 | `edit_token`/`code` missing from ActionFailure type union | Add type guard to narrow union before accessing properties |
| 2 | `routes/group/edit/+page.server.ts` | 21 | `superValidate` API change - `data` option invalid | Check sveltekit-superforms docs for new API signature |
| 3 | `routes/rides/edit/+page.server.ts` | 40 | `image_srcset` not in type | Add `image_srcset` to the type definition or form schema |
| 4-7 | `components/group-form/CustomMarkerBuilder.svelte` | 143-449 | 4 null-check errors (`imagePreview`, `canvasRef`) | Add null guards before accessing `.length`, `.substring()`, `.toBlob()` |
| 8-11 | `components/ride-form/DateTimePicker.svelte` | 194-368 | 4 Calendar `bind:value` type mismatches | Check bits-ui Calendar component - may need `DateValue` vs `DateValue[]` |
| 12 | `components/ride-form/DateTimePicker.svelte` | 368 | `maxlength` expects number, got string | Change `maxlength="500"` to `maxlength={500}` |
| 13 | `routes/+page.svelte` | 370 | `$errors.occurrences` type mismatch | Check superforms error type, may need type assertion |

**Validation:**
```bash
cd frontends/form
npm run check   # Must show: "No errors found"
npm run build   # Must succeed
```

**Commit separately:** `fix(form): resolve pre-existing TypeScript errors`

---

## Files Summary (Milestone 5)

### Files to Modify

| File | Changes | Priority |
|------|---------|----------|
| `src/lib/api/strava.ts` | Rate limit detection, session expiry handling | High |
| `src/lib/utils/strava-websocket.ts` | Manual retry, connection timeout, reconnect UI state | High |
| `src/lib/components/strava/StravaImportButton.svelte` | Improve error messages, ARIA labels | High |
| `src/lib/components/strava/StravaImport.svelte` | Session expiry handling, ARIA live regions, focus mgmt | High |
| `src/lib/components/strava/ImportProgress.svelte` | Manual retry button, timeout handling, ARIA | High |
| `src/lib/components/strava/ClubList.svelte` | Rate limit error handling, keyboard nav | Medium |
| `src/lib/components/strava/EventCard.svelte` | Mobile optimization, focus management | Medium |
| `src/lib/components/strava/EmailInput.svelte` | Form accessibility, autofocus | Medium |
| `src/lib/components/strava/ImportResults.svelte` | Screen reader announcements, role="alert" | Medium |
| `src/lib/types/strava.ts` | Add error type constants | Low |

### New Files (Optional)

| File | Purpose | Priority |
|------|---------|----------|
| `src/lib/components/strava/SkeletonClubList.svelte` | Skeleton loader for clubs | Low |
| `src/lib/components/strava/ConnectionError.svelte` | Reusable connection error display | Low |

**No new backend files** - All backend work complete in M1-M3

---

## Current Implementation Status

### Already Implemented (M4)
- [x] Debug logging with `PUBLIC_STRAVA_DEBUG`
- [x] Basic error display (red alert boxes)
- [x] Loading spinners for clubs and events
- [x] WebSocket reconnection (3 attempts, exponential backoff)
- [x] Import results with success/partial/failure states
- [x] OAuth popup with postMessage + polling fallback
- [x] 5-minute OAuth timeout
- [x] Per-event error display in ImportProgress
- [x] Environment variables documented in `.env.example`

### Gaps to Address (M5)
- [ ] Rate limit (429) specific error handling
- [ ] Session expiry detection and re-auth flow
- [ ] Manual WebSocket retry button after max reconnects
- [ ] Connection timeout with user feedback
- [ ] Comprehensive ARIA labels
- [ ] Focus management between steps
- [ ] Screen reader live regions for progress
- [ ] Mobile-responsive event cards

---

## Tasks

### 5.1 - Error Handling Improvements

#### 5.1.1 - Rate Limit Detection & Prevention
**Files:** `src/lib/api/strava.ts`, `src/lib/types/strava.ts`, `backend/internal/strava/client.go`

**Strategy:** Use `GET /athlete/clubs` as both rate limit check AND data fetch (no extra API call).

**Strava Rate Limit Facts** (from official docs):
- "Non-upload" endpoints (all our GET requests): **100 requests per 15 minutes**, 1,000 daily
- Limits reset at **fixed 15-minute intervals**: :00, :15, :30, :45 past the hour (NOT rolling)
- Headers: `X-ReadRateLimit-Usage` and `X-ReadRateLimit-Limit` (format: "85,100")
- 429 responses still count toward daily limit
- Best practice: Monitor headers and throttle when approaching limits

**Frontend (`strava.ts`):**
- [ ] Create `RateLimitError` class extending `APIError` with `retry_after_seconds`
- [ ] Detect HTTP 429 responses in `fetchStravaAPI`
- [ ] Parse `retry_after_seconds` from response body (sent by backend)
- [ ] Show countdown message: "CycleScene is experiencing high Strava import usage. Try again in X minutes."
  - Example: "Try again in 8 minutes" (not always 15 - depends on when limit hit)
- [ ] Optional: Show countdown timer that updates every minute

**Backend (`client.go`):**
- [ ] Parse `X-ReadRateLimit-Usage` header from every Strava API response (format: "85,100")
- [ ] On `GetAthleteClubs()`, check if read limit usage > 75/100 (75% threshold)
- [ ] Threshold set conservatively to allow headroom for:
  - Event fetching when user expands clubs (1 call per club)
  - Route data fetching during import (1 call per event with route)
  - Example: 75 used + 10 clubs + 10 routes = 95 total (stays under 100)
- [ ] If usage > 75/100, calculate time until next reset:
  ```go
  // Limits reset at :00, :15, :30, :45 past the hour
  now := time.Now()
  minutesPastHour := now.Minute()
  nextReset := (15 - (minutesPastHour % 15)) // minutes until next interval
  ```
- [ ] Return 429 with `retry_after_seconds` in response body
- [ ] Log high-usage events to monitoring DB:
  ```json
  {
    "event": "strava_rate_limit_warning",
    "read_usage": 75,
    "read_limit": 100,
    "remaining": 25,
    "next_reset_minutes": 8,
    "timestamp": "2024-01-31T10:00:00Z"
  }
  ```
- [ ] If 429 received from Strava, parse `Retry-After` header OR calculate next reset time

**Flow:**
```
User completes OAuth
  ↓
Frontend calls /v1/strava/admin-clubs
  ↓
Backend calls GET /athlete/clubs (Strava)
  ↓
Backend checks X-Ratelimit-Usage header
  ↓
IF usage > 75/100:
  Return 429 to frontend
  Log to monitoring DB
  Frontend shows: "CycleScene is experiencing high Strava import usage.
                   Please try again in 15 minutes."
ELSE:
  Return clubs data
  Proceed to selection UI
  (Remaining budget for event/route fetching)
```

**Benefits:**
- Early detection (right after OAuth, before user invests time selecting events)
- No extra API calls (reuse required call)
- Clean UX (fail fast vs fail mid-import)

#### 5.1.2 - Session & Token Refresh (Hybrid Approach)
**Files:** `backend/internal/strava/session.go`, `backend/internal/strava/client.go`, `src/lib/api/strava.ts`

**Strategy:** Automatic backend token refresh with manual re-auth fallback

**Backend - Automatic Token Refresh:**
- [ ] Before any Strava API call, check if access token expires soon (< 5 min buffer)
- [ ] If expiring, use `refresh_token` to get new access token (standard OAuth flow):
  ```go
  POST https://www.strava.com/oauth/token
  client_id={client_id}
  client_secret={client_secret}
  grant_type=refresh_token
  refresh_token={refresh_token}
  ```
- [ ] Update session with new access_token and expires_at
- [ ] This happens transparently - user never notices

**Frontend - Pre-Import Check:**
- [ ] Before starting WebSocket import, call `/v1/strava/check-session`
- [ ] Backend checks token age and refreshes if needed
- [ ] If refresh succeeds, proceed with import
- [ ] If refresh fails (token revoked, expired refresh_token), return 401
- [ ] Show brief "Preparing import..." during check

**Fallback - Manual Re-auth:**
- [ ] Create `SessionExpiredError` class for when automatic refresh fails
- [ ] Detect HTTP 401 responses (refresh_token no longer valid)
- [ ] Show message: "Your session expired. Please reconnect to Strava."
- [ ] Add "Reconnect" button that restarts OAuth flow
- [ ] Clear session and reset to initial state

**Monitoring:**
- [ ] Log successful automatic refreshes:
  ```json
  {
    "event": "strava_token_auto_refresh",
    "session_id": "...",
    "old_expires_at": "2024-01-31T10:00:00Z",
    "new_expires_at": "2024-01-31T16:00:00Z"
  }
  ```
- [ ] Log refresh failures (requires manual re-auth):
  ```json
  {
    "event": "strava_token_refresh_failed",
    "error": "invalid_grant",
    "fallback": "manual_reauth"
  }
  ```

**User Experience:**
- **Normal case** (99%): Token refresh happens automatically, user never sees anything
- **Edge case**: If automatic refresh fails, user gets clear "Reconnect" button
- No localStorage complexity needed

#### 5.1.3 - OAuth Denial Handling
**Files:** `src/lib/api/strava.ts`, `src/lib/components/strava/StravaImportButton.svelte`

- [ ] Detect `error=access_denied` in OAuth callback
- [ ] Show friendly message: "Authorization cancelled. You can try again anytime."
- [ ] Clear loading state on denial
- [ ] Don't show error if user just closes popup (ambiguous)

#### 5.1.4 - WebSocket Retry & Stop Controls
**Files:** `src/lib/utils/strava-websocket.ts`, `src/lib/components/strava/ImportProgress.svelte`

**Current Gap:** After 3 auto-reconnects fail, user has no way to retry without page refresh. Also no way to stop an in-progress import.

- [ ] Add "Stop Import" button during active import (cancels WebSocket, returns to selection)
- [ ] After 3 failed reconnects, show "Connection lost" UI state
- [ ] Add "Retry Import" button (unlimited retries allowed)
- [ ] Display reconnection attempts: "Reconnecting (2/3)..." during auto-reconnect
- [ ] Add `manualRetry()` method to WebSocket class
- [ ] Add `stop()` method to WebSocket class
- [ ] Preserve already-completed results on retry
- [ ] Log retry events to monitoring DB with context:
  ```typescript
  {
    event: 'strava_import_retry',
    attempt_number: number,
    total_events: number,
    completed_before_failure: number,
    error_reason: string,
    athlete_id: string
  }
  ```
- [ ] Log manual stops to monitoring DB:
  ```typescript
  {
    event: 'strava_import_stopped',
    total_events: number,
    completed: number,
    stopped_by_user: true
  }
  ```

**Retry Behavior:**
- Manual retry resets auto-reconnect counter (3 fresh attempts)
- Unlimited manual retries (no cap)
- Already-completed events not re-imported
- Failed events retried in same order

#### 5.1.5 - Partial Import Handling
**Files:** `src/lib/components/strava/ImportProgress.svelte`

- [ ] If WebSocket disconnects mid-import, show partial results
- [ ] Show "X of Y events imported before connection lost"
- [ ] Option to retry only failed events (future enhancement)

**Error Message Standards:**

| Error | User Message |
|-------|-------------|
| OAuth denied | "Authorization cancelled. You can try again anytime." |
| OAuth timeout | "Connection timed out. Please try again." |
| Session expired | "Your session expired. Please reconnect to Strava." |
| Rate limit | "CycleScene is experiencing high Strava import usage. Try again in X minutes." (X = calculated time until next reset at :00, :15, :30, or :45) |
| Not admin | "You must be an admin or owner of a club to import its events." |
| No events | "No upcoming events found for this club." |
| WebSocket failed | "Connection lost. Please check your internet and try again." |
| Import failed | "Failed to import event: [specific error from server]" |

---

### 5.2 - Loading States & Timeouts

#### 5.2.1 - Skeleton Loaders (Use shadcn Skeleton)
**Files:** `StravaImport.svelte`, `ClubList.svelte`

**Use shadcn's Skeleton component** instead of spinners for better perceived performance.

**StravaImport - Loading Clubs:**
- [ ] Replace spinner with 3-5 skeleton club cards while `isLoadingClubs`
- [ ] Each skeleton shows: circle (profile image) + 2 lines (name, members)

```svelte
<!-- StravaImport.svelte -->
{#if isLoadingClubs}
  <div class="space-y-4">
    {#each Array(3) as _}
      <Card.Root class="p-4">
        <div class="flex items-center gap-3">
          <Skeleton class="h-10 w-10 rounded-full" />
          <div class="flex-1 space-y-2">
            <Skeleton class="h-4 w-[200px]" />
            <Skeleton class="h-3 w-[150px]" />
          </div>
        </div>
      </Card.Root>
    {/each}
  </div>
{/if}
```

**ClubList - Loading Events:**
- [ ] When accordion expands, show 2-3 skeleton event cards
- [ ] Each skeleton: checkbox + 3 lines (title, date, location)

```svelte
<!-- ClubList.svelte -->
{#if loadingClubs.has(club.id)}
  <div class="space-y-3 py-4">
    {#each Array(3) as _}
      <div class="flex gap-3 rounded-lg border p-3">
        <Skeleton class="h-5 w-5 rounded" />
        <div class="flex-1 space-y-2">
          <Skeleton class="h-4 w-full" />
          <Skeleton class="h-3 w-3/4" />
          <Skeleton class="h-3 w-1/2" />
        </div>
      </div>
    {/each}
  </div>
{/if}
```

**Import shadcn Skeleton:**
```typescript
import { Skeleton } from "$lib/components/ui/skeleton";
```

**Benefits:**
- Shows content structure while loading
- Reduces perceived wait time
- No layout shift when content loads
- Uses existing shadcn component (no custom code)

---

#### 5.2.2 - Timeout Handling
**Files:** `src/lib/utils/strava-websocket.ts`

- [ ] Add 60-second heartbeat timeout (no heartbeat = stale connection)
- [ ] Show "Taking longer than expected..." after 15 seconds of no activity
- [ ] Trigger reconnect attempt on timeout

---

#### 5.2.3 - Button State Management
**Files:** All Strava components

- [ ] Verify buttons disabled during async ops (mostly done)
- [ ] Add brief disable on "Import More" to prevent double-click
- [ ] ClubList "Try again" disabled while retrying

---

### 5.3 - Responsive Design

#### 5.3.1 - Mobile OAuth Flow Verification
**Files:** `src/lib/api/strava.ts`

**Current implementation should work on mobile** via polling fallback:
- Desktop: `window.open()` → popup → postMessage works
- Mobile: `window.open()` → new tab → postMessage fails → polling detects tab close → `checkSession()` validates

**Testing required:**
- [ ] Test on iOS Safari (15+): OAuth in new tab, verify polling fallback works
- [ ] Test on Android Chrome: OAuth in new tab, verify success detection
- [ ] Test on iPad Safari: May use popup or tab depending on viewport
- [ ] Verify OAuth callback page works in mobile tab context
- [ ] Confirm `popup.closed` detects mobile tab closure

**If polling fallback fails on mobile:**
- [ ] Add mobile detection (user agent or `window.matchMedia('(max-width: 768px)')`)
- [ ] Use redirect flow for mobile instead of popup:
  ```typescript
  if (isMobile) {
    // Store return URL in sessionStorage
    sessionStorage.setItem('strava_return_url', window.location.href);
    // Redirect current window
    window.location.href = authUrl;
    // Callback redirects back with success flag
  }
  ```

**Expected result:** Current implementation works without changes (polling handles mobile tabs)

#### 5.3.2 - Mobile Component Optimization
**Files:** `EventCard.svelte`, `ImportProgress.svelte`

- [ ] Stack checkbox and event details vertically on narrow screens
- [ ] Ensure customize panel usable on mobile
- [ ] Verify touch targets >= 44x44px
- [ ] Step badges wrap correctly (already uses flex-wrap)

**Breakpoint Testing:**
- [ ] 320px (iPhone SE)
- [ ] 375px (iPhone 12 mini)
- [ ] 414px (iPhone 12 Pro Max)
- [ ] 768px (iPad)

---

### 5.4 - Accessibility (Comprehensive)

**Goal:** Make Strava import fully accessible for keyboard users, screen reader users, and users with cognitive/motor disabilities.

**What shadcn/Radix already provides:** Component-level accessibility (keyboard nav, ARIA roles/states, focus management within components)

**What we need to add:** Application-level accessibility (live regions, focus between steps, loading states, context-specific labels)

---

#### 5.4.1 - Live Regions for Screen Readers
**Files:** `ImportProgress.svelte`, `ImportResults.svelte`, `ClubList.svelte`, `StravaImport.svelte`

**ImportProgress - Real-time Progress Announcements:**
- [ ] Add `aria-live="polite"` region that announces progress updates
- [ ] Announce connection state: "Connecting to import service"
- [ ] Announce at milestones: "Imported 5 of 20 events", "Imported 10 of 20 events"
- [ ] Announce completion: "Import complete, 20 events imported successfully"
- [ ] Announce errors: "Connection error, attempting to reconnect"
- [ ] Use `aria-atomic="true"` for full context in each announcement

```svelte
<!-- ImportProgress.svelte -->
<div class="sr-only" aria-live="polite" aria-atomic="true">
  {#if wsState === "connecting"}
    Connecting to import service
  {:else if wsState === "error"}
    Connection error, attempting to reconnect
  {:else if completedCount > 0 && completedCount < events.length}
    Imported {completedCount} of {events.length} events
  {:else if completedCount === events.length}
    Import complete, {completedCount} events imported successfully
  {/if}
</div>
```

**ImportResults - Alert Announcement:**
- [ ] Add `role="alert"` to summary alert (auto-announced by screen readers)
- [ ] Ensure alert has clear heading and description

```svelte
<!-- ImportResults.svelte -->
<Alert.Root role="alert" class="border-green-200 bg-green-50">
  <Alert.Title id="results-heading">Import Complete!</Alert.Title>
  <Alert.Description>
    {successCount} event{successCount !== 1 ? "s" : ""} imported successfully.
  </Alert.Description>
</Alert.Root>
```

**ClubList - Loading Announcements:**
- [ ] Add `aria-live="polite"` for club loading states
- [ ] Announce when events are loaded: "Loaded 5 events for Portland Bike Club"

```svelte
<!-- ClubList.svelte -->
<div class="sr-only" aria-live="polite">
  {#each Array.from(loadingClubs) as clubId}
    Loading events for {clubs.find(c => c.id === clubId)?.name}
  {/each}
</div>
```

**StravaImport - Step Announcements:**
- [ ] Announce step changes: "Step 2 of 4: Select events to import"

```svelte
<!-- StravaImport.svelte -->
<div class="sr-only" aria-live="polite">
  {#if step === "email"}
    Step 1 of 4: Enter your email address
  {:else if step === "select"}
    Step 2 of 4: Select events to import
  {:else if step === "importing"}
    Step 3 of 4: Importing events
  {:else if step === "complete"}
    Step 4 of 4: Import complete
  {/if}
</div>
```

---

#### 5.4.2 - Focus Management Between Steps
**Files:** `StravaImport.svelte`, `EmailInput.svelte`, `ClubList.svelte`

- [ ] **Email step:** Auto-focus email input on mount
- [ ] **Selection step:** Move focus to first club or instructions when entering step
- [ ] **Import step:** Move focus to progress heading (non-interactive, use tabindex="-1")
- [ ] **Complete step:** Move focus to results alert heading
- [ ] **After error:** Move focus to error message

```svelte
<!-- StravaImport.svelte -->
$effect(() => {
  // Small delay to ensure DOM is updated
  setTimeout(() => {
    if (step === "email") {
      document.getElementById("strava-email-input")?.focus();
    } else if (step === "select") {
      // Focus first interactive element or instructions
      const firstClub = document.querySelector('[data-accordion-item]');
      if (firstClub) {
        (firstClub as HTMLElement).focus();
      }
    } else if (step === "importing") {
      // Focus heading (make focusable with tabindex="-1")
      document.getElementById("import-progress-heading")?.focus();
    } else if (step === "complete") {
      // Focus results heading
      document.getElementById("results-heading")?.focus();
    }
  }, 100);
});
```

**Focus visible styles:**
- [ ] Verify all focusable elements have visible focus indicators (Tailwind: `focus-visible:ring-2`)
- [ ] Test with keyboard navigation (Tab, Shift+Tab)

---

#### 5.4.3 - Loading State Announcements
**Files:** `StravaImportButton.svelte`, `ClubList.svelte`, `StravaImport.svelte`

**StravaImportButton - OAuth Loading:**
- [ ] Add `aria-busy="true"` during OAuth
- [ ] Include sr-only status text

```svelte
<Button
  aria-busy={isLoading}
  aria-label="Import events from your Strava clubs"
>
  {#if isLoading}
    <span class="sr-only">Connecting to Strava</span>
  {/if}
  <!-- visual content -->
</Button>
```

**ClubList - Event Loading:**
- [ ] Add `aria-busy="true"` to accordion items while loading events
- [ ] Screen reader text: "Loading events..."

**StravaImport - Club Loading:**
- [ ] Card wrapper has `aria-busy="true"` while `isLoadingClubs`
- [ ] Spinner has `role="status"` with sr-only text

```svelte
<Card.Root aria-busy={isLoadingClubs}>
  <div role="status">
    <svg class="animate-spin" aria-hidden="true">...</svg>
    <span class="sr-only">Loading your clubs</span>
  </div>
</Card.Root>
```

---

#### 5.4.4 - Context-Specific ARIA Labels
**Files:** All Strava components

**Descriptive Button Labels:**
- [ ] StravaImportButton: `aria-label="Import cycling events from your Strava clubs"`
- [ ] Stop Import button: `aria-label="Stop current import and return to event selection"`
- [ ] Retry Import button: `aria-label="Retry failed import from where it stopped"`
- [ ] Import More button: `aria-label="Import additional events from Strava"`
- [ ] Done button: `aria-label="Close import and return to form"`
- [ ] Reconnect button: `aria-label="Reconnect to Strava to continue"`

**EventCard - Checkbox Association:**
- [ ] Add `aria-describedby` linking checkbox to event details

```svelte
<Checkbox
  id="event-{event.id}"
  aria-describedby="event-{event.id}-details"
  aria-label="Select {event.title} for import"
/>
<div id="event-{event.id}-details">
  <h4>{event.title}</h4>
  <p>{event.description}</p>
</div>
```

**Progress Indicators:**
- [ ] Add `aria-label` to step badges: `aria-label="Fetching event data: in progress"`
- [ ] Overall progress bar already has `aria-valuenow` (shadcn Progress component)

---

#### 5.4.5 - Semantic HTML & Landmarks
**Files:** All Strava components

- [ ] Ensure proper heading hierarchy (h1 → h2 → h3, no skipping)
- [ ] Use `<form>` elements for email input (already done via EmailInput)
- [ ] Add `role="region"` with `aria-label` to major sections:

```svelte
<!-- StravaImport.svelte -->
<section role="region" aria-label="Strava event import">
  <!-- All content -->
</section>
```

---

#### 5.4.6 - Keyboard Navigation Verification
**Files:** All interactive components

- [ ] **Tab order:** Verify logical tab sequence through entire flow
- [ ] **Enter/Space:** All buttons activate with both keys (native button behavior)
- [ ] **Space:** Checkboxes toggle (Radix handles this)
- [ ] **Arrow keys:** Accordion navigation works (Radix handles this)
- [ ] **Escape:** Close customize panel (Collapsible handles this)
- [ ] **No keyboard traps:** User can always navigate out of any section

---

#### 5.4.7 - Error Messages Accessibility
**Files:** `StravaImport.svelte`, `ClubList.svelte`, error displays

- [ ] Error messages have `role="alert"` for immediate announcement
- [ ] Errors are programmatically associated with their input (aria-describedby)
- [ ] Error color contrast meets WCAG AA (text on red background)

```svelte
{#if error}
  <div role="alert" class="rounded-lg border border-red-200 bg-red-50 p-4">
    <p class="text-red-700">{error}</p>
  </div>
{/if}
```

---

#### 5.4.8 - Testing Checklist

**Keyboard-only testing:**
- [ ] Complete entire OAuth flow using only keyboard
- [ ] Navigate through club list and event selection
- [ ] Start import and monitor progress
- [ ] Tab order is logical at every step
- [ ] Focus visible on all interactive elements

**Screen reader testing (VoiceOver/NVDA):**
- [ ] All interactive elements properly announced
- [ ] Progress updates announced as they happen
- [ ] Step changes announced
- [ ] Error messages announced immediately
- [ ] Button purposes are clear from labels
- [ ] Loading states announced

**Color contrast:**
- [ ] All text meets WCAG AA (4.5:1 for normal text, 3:1 for large)
- [ ] Focus indicators visible
- [ ] Error messages readable

---

**Success Criteria:**
- [ ] Screen reader user can complete full import flow
- [ ] Keyboard-only user can complete full import flow
- [ ] Progress updates are announced as they happen
- [ ] All interactive elements have descriptive labels
- [ ] No accessibility errors in automated testing (axe, Lighthouse)
- [ ] Focus always visible and logical

---

### 5.5 - Environment Variables

**Status:** ✅ Already complete in `.env.example`

| Variable | Purpose | Status |
|----------|---------|--------|
| `PUBLIC_STRAVA_ENABLED` | Feature flag | ✅ Done |
| `PUBLIC_STRAVA_DEBUG` | Debug logging | ✅ Done |
| `PUBLIC_API_URL` | Backend URL (also for WebSocket) | ✅ Done |

---

## Validation Checklist

### Build Verification
```bash
cd frontends/form
npm run check    # Zero TypeScript errors
npm run build    # Successful build
npm run lint     # No lint errors
```

### Manual Testing

#### Error Scenarios
- [ ] Deny OAuth → "cancelled" message
- [ ] Session expiry → re-auth prompt
- [ ] Rate limit (429) → friendly message with timer
- [ ] Kill network mid-import → retry option shown
- [ ] Close WebSocket tab → partial results displayed

#### Responsive
- [ ] Mobile (320-414px) → all content visible, touch-friendly
- [ ] Tablet (768px) → good space usage
- [ ] Desktop (1024px+) → centered, max-width respected

#### Accessibility
- [ ] Keyboard-only → all features usable
- [ ] Screen reader → progress announced
- [ ] Focus visible → clear indicators

---

## Implementation Order

1. **Fix pre-existing TypeScript errors** (required blocker)
2. **5.1.1** - Rate limit detection
3. **5.1.2** - Session expiry handling
4. **5.1.4** - WebSocket retry UI
5. **5.4.1** - ARIA labels (quick wins)
6. **5.4.2** - Live regions
7. **5.4.3** - Focus management
8. **5.3** - Mobile testing and fixes
9. **5.1.3** - OAuth denial handling
10. **5.2.3** - Skeleton loaders (if time)

---

## Success Criteria

- [ ] `npm run check` passes with zero errors
- [ ] `npm run build` succeeds
- [ ] All error scenarios show user-friendly messages
- [ ] WebSocket has manual retry after failed reconnects
- [ ] Screen reader users can track import progress
- [ ] Mobile users can complete full import flow
- [ ] Focus moves logically through steps

---

**Branch:** `feature/stravaimport`
**Created:** 2026-01-31
**Updated:** 2026-02-03
**Status:** Ready for implementation (after fixing pre-existing TS errors)
