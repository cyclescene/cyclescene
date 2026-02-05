# CycleScene — Strava Compliance Technical Spec

## Overview

This spec covers all developer-facing changes required to bring CycleScene's Strava integration into full compliance with the Strava API Agreement (effective Oct 9, 2025) and Brand Guidelines (revised Sep 29, 2025) before submitting the Developer Program form.

---

## 1. OAuth Button Swap

### What to change
Replace the current "Import from Strava" button with the official "Connect with Strava" button asset.

### Assets
- Download: https://developers.strava.com/downloads/1.1-Connect-with-Strava-Buttons.zip
- Two color options: **orange** (recommended for light backgrounds) and **white** (for dark backgrounds)
- Provided in EPS, SVG, and PNG formats
- Button height: **48px @1x**, **96px @2x**

### Requirements
- The button **must** link to `https://www.strava.com/oauth/authorize` or `https://www.strava.com/oauth/mobile/authorize`
- Do not modify, stretch, recolor, or animate the button
- Do not make the Strava button more prominent than CycleScene's own branding

### Implementation
- Replace the current button element on the "Host a Ride" form page
- Use the SVG or @2x PNG for crisp rendering on retina displays
- The button should trigger your existing OAuth flow — only the visual asset changes

---

## 2. OAuth Token Storage & Refresh

### Context
Strava access tokens expire after ~6 hours. You need refresh tokens to run background syncs without requiring the admin to re-authenticate.

### Token flow
```
Initial auth → access_token (short-lived, ~6hrs)
             → refresh_token (long-lived, survives until revoked)
             → expires_at (unix timestamp)
```

### What to store per connected admin
| Field | Description |
|-------|-------------|
| `strava_athlete_id` | Unique Strava user ID |
| `access_token` | Current access token (encrypted at rest) |
| `refresh_token` | Current refresh token (encrypted at rest) |
| `expires_at` | Unix timestamp when access_token expires |
| `scopes` | Granted OAuth scopes |
| `connected_at` | Timestamp of initial authorization |
| `last_synced_at` | Timestamp of last successful data refresh |

### Refresh logic (before any API call)
```
if current_time >= expires_at:
    POST https://www.strava.com/oauth/token
    {
        client_id: 186625,
        client_secret: YOUR_SECRET,
        grant_type: "refresh_token",
        refresh_token: stored_refresh_token
    }
    → save new access_token, refresh_token, expires_at
```

### Security requirements
- Store tokens **encrypted at rest** (API Agreement Section 2.8)
- Never expose tokens client-side or in logs
- Transmit all Strava data over **HTTPS only**
- Notify Strava within **24 hours** if tokens are compromised (Section 1.2)

---

## 3. Background Sync Service (7-Day Cache Refresh)

### Requirement
Strava API Agreement Section 7.1: No Strava data shall remain in your cache longer than 7 days without being refreshed.

### Design

#### Scheduled job
- **Frequency:** Run every **24–48 hours** (gives comfortable margin under the 7-day limit)
- **Trigger:** Cron job, cloud scheduler, or queued background worker

#### Per connected admin, each run:
```
1. Check if admin's token is still valid (not revoked)
2. Refresh access_token if expired (see Section 2)
3. Fetch current club list → GET https://www.strava.com/api/v3/athlete/clubs
4. For each qualifying club (admin/owner role, matching sport + city):
     Fetch events → GET https://www.strava.com/api/v3/clubs/{club_id}/group_events
5. Compare fetched events to stored events:
     - New events → insert
     - Modified events → update
     - Missing events (existed in DB but not in API response) → mark for deletion
6. Delete stale events (see Section 4)
7. Update last_synced_at timestamp
```

#### Rate limit awareness
- Default limits: **200 requests / 15 min**, **2,000 requests / day**
- Read limits: **100 requests / 15 min**, **1,000 requests / day**
- Check response headers: `X-RateLimit-Usage` and `X-ReadRateLimit-Usage`
- If approaching limits, back off and retry later
- Implement exponential backoff on `429 Too Many Requests`

#### Suggested DB fields for synced events
| Field | Description |
|-------|-------------|
| `id` | Internal primary key |
| `strava_event_id` | Strava's event ID (for dedup and deletion tracking) |
| `strava_club_id` | Source club ID |
| `source` | `"strava"` (to distinguish from manually created events) |
| `title` | Event name |
| `description` | Event description |
| `start_date` | Event start datetime |
| `city` | City (from your filtering logic) |
| `group_id` | CycleScene group ID this was imported into |
| `imported_by` | Strava athlete ID of the admin who imported it |
| `last_refreshed_at` | Timestamp of last successful refresh from API |
| `created_at` | When first imported |

---

## 4. Event Deletion & Staleness Handling

### Requirement
- API Agreement Section 2.14.6: Deletions must be reflected within **48 hours**
- API Agreement Section 7.1: If a resource is no longer available from Strava, remove from cache **immediately** regardless of refresh schedule

### Implementation

#### During sync (Section 3):
```
fetched_event_ids = [events from Strava API response]
stored_event_ids = [events in DB for this club with source="strava"]

stale_ids = stored_event_ids - fetched_event_ids
→ DELETE all events where strava_event_id IN stale_ids
```

#### Staleness safety net:
```
If last_refreshed_at is older than 7 days AND sync has been failing:
    → Soft-delete or hide event from public display
    → Log alert for investigation
```

#### Past events:
- Cleaning up past events is good hygiene but not the core requirement
- The critical path is: if Strava says it doesn't exist anymore → remove it

---

## 5. Deauthorization Flow

### Requirement
- Users must be able to disconnect their Strava account from CycleScene
- On disconnect, delete all Strava data associated with that admin
- Also support Strava-side revocation (user disconnects from strava.com/settings/apps)

### A. User-initiated disconnect (from CycleScene)

Add a "Disconnect Strava" option in the admin's account/settings area.

#### On click:
```
1. POST https://www.strava.com/oauth/deauthorize
   Headers: Authorization: Bearer {access_token}
   → This revokes access on Strava's side

2. Delete from your DB:
   - The admin's stored tokens (access_token, refresh_token)
   - All events imported by this admin where source="strava"
   - Any cached club data for this admin

3. Confirm disconnection in the UI
```

### B. Strava-initiated revocation

If a user revokes access from strava.com/settings/apps, your stored tokens will start returning `401 Unauthorized`.

#### Handle during sync:
```
if API returns 401 for a stored token:
    → Treat as revoked
    → Delete tokens and associated Strava data
    → Optionally notify the admin via email that their Strava connection was lost
```

### C. Webhook option (recommended but not required)
Strava webhooks can notify you of deauthorization events in real time:
- Event: `athlete` with `aspect_type: "update"` and `updates: {"authorized": "false"}`
- Docs: https://developers.strava.com/docs/webhooks/
- This is a nice-to-have for faster response but not strictly required if your sync handles 401s

---

## 6. "Powered by Strava" Logo Placement

### Assets
- Download: https://developers.strava.com/downloads/1.2-Strava-API-Logos.zip
- 3 color options: orange, white, black
- Horizontal and stacked versions
- EPS, SVG, PNG formats

### Where to display
- On any screen/component where **Strava-sourced events** are shown to users
- This includes:
  - Event cards/listings on the city PWA (e.g., pdx.cyclescene.cc) where `source="strava"`
  - Event detail pages for Strava-imported events
  - The import confirmation/preview screen (admin side)

### Rules
- Logo must be **completely separate** from CycleScene branding
- Must **not appear more prominently** than CycleScene's own name/logo
- Do not modify, alter, recolor, or animate the logo
- Do not use any part of the Strava logo as the CycleScene app icon

### Suggested implementation
- Add a small "Powered by Strava" badge/footer on event cards that have `source="strava"`
- Use the horizontal version at a reasonable size (not dominating the card)

---

## 7. "View on Strava" Links

### Requirement
API Brand Guidelines Section 3: Any link back to original Strava data must use the text **"View on Strava"**.

### Styling (must meet at least one):
- **Bold** weight
- **Underlined**
- **Orange** color `#FC5200`

### Where to add
- On Strava-sourced event cards or detail pages, link back to the event or club on Strava
- Link URL pattern: `https://www.strava.com/clubs/{club_id}/group_events/{event_id}` (verify against your API response data for exact URLs)

### Implementation
```html
<a href="https://www.strava.com/clubs/{club_id}/group_events/{event_id}"
   style="color: #FC5200; font-weight: bold;">
   View on Strava
</a>
```

### Rules
- Text must be **legible**
- Must be **identifiable as a link**
- Do not use other text like "See on Strava" or "Open in Strava" — must be exactly **"View on Strava"**

---

## 8. Privacy Policy Updates

### Requirement
API Agreement Section 5.2 requires your privacy policy to cover:
1. What type of Strava data is collected
2. How the data is collected
3. How a user can withdraw consent
4. How a user can request deletion

### Location
Update the existing privacy policy at `cyclescene.cc/privacy`.

### Content
See the drafted privacy policy sections in the form draft document (`strava-form-draft.md`). Key sections to add:
- Strava Integration (what data is accessed)
- How We Use Strava Data (display purposes only, no sharing/selling)
- Data Retention and Refresh (7-day refresh cycle)
- Strava Usage Data (Strava may collect API usage data — Section 2.12)
- Revoking Strava Access (how to disconnect)
- Requesting Data Deletion (contact email + process)

### Additional requirements
- Privacy policy must be accessible via a **reasonably prominent hyperlink**
- Must not conflict with or supersede the Strava Privacy Policy
- Must comply with GDPR/UK GDPR if serving users in those regions

---

## 9. Data Scope Restrictions

### What you CAN access
- Club details for clubs where the authenticated admin is an admin/owner
- Upcoming club events from those clubs
- The authenticated admin's own athlete profile (for identification)

### What you CANNOT do
- Display one user's Strava data to another user (UNLESS classified as a Community Application — which CycleScene qualifies for, so displaying club events publicly is permitted under Section 2.10)
- Store Strava data longer than 7 days without refreshing
- Use Strava data for advertising, AI/ML training, or selling to third parties
- Combine Strava data with other data for analytics or insights (Section 2.14.7)
- Modify or edit any Strava content, links, or metadata when displaying (Section 2.14.19)
- Cache or aggregate geographic location data from Strava (Section 2.14.14)
- Charge users for access to Strava data or functionality (Section 2.14.15)

---

## 10. Implementation Priority

### Phase 1 — Blockers (must complete before form submission)
1. **OAuth button swap** — Replace with official "Connect with Strava" asset
2. **Privacy policy update** — Add Strava-specific sections
3. **"Powered by Strava" logo** — Add to all Strava-sourced event displays
4. **"View on Strava" links** — Add to Strava-sourced events

### Phase 2 — Required (implement before or shortly after submission)
5. **Background sync service** — Cron/scheduler refreshing events within 7-day window
6. **Staleness deletion** — Remove events no longer on Strava within 48 hours
7. **Token refresh logic** — Handle expired access tokens silently via refresh tokens

### Phase 3 — Recommended
8. **Deauthorization endpoint** — "Disconnect Strava" in admin settings + Strava API deauth call
9. **401 handling in sync** — Detect revoked tokens and clean up data
10. **Webhook subscription** — Real-time deauth and event update notifications

---

## API Endpoints Reference

| Purpose | Method | Endpoint |
|---------|--------|----------|
| OAuth authorize | GET | `https://www.strava.com/oauth/authorize` |
| Token exchange | POST | `https://www.strava.com/oauth/token` |
| Token refresh | POST | `https://www.strava.com/oauth/token` (grant_type=refresh_token) |
| Deauthorize | POST | `https://www.strava.com/oauth/deauthorize` |
| List athlete's clubs | GET | `/api/v3/athlete/clubs` |
| Club details | GET | `/api/v3/clubs/{id}` |
| Club events | GET | `/api/v3/clubs/{id}/group_events` |
| Authenticated athlete | GET | `/api/v3/athlete` |
| Webhook subscription | POST | `/api/v3/push_subscriptions` |

Base URL for API endpoints: `https://www.strava.com`

Docs: https://developers.strava.com/docs/reference/
