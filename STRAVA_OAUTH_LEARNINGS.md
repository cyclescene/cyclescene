# Strava OAuth Integration - Technical Learnings

## Overview
This document captures our learnings from exploring the Strava OAuth API for importing club group events into CycleScene. The implementation focuses on staying within Strava's Terms of Service by limiting event imports to club admins/owners only.

---

## OAuth Flow

### 1. Authorization URL
```
https://www.strava.com/oauth/authorize
```

**Required Parameters:**
- `client_id` - Your Strava application ID
- `redirect_uri` - OAuth callback URL (NO trailing slash - Strava appends this automatically)
- `response_type=code` - Standard OAuth code flow
- `scope=read,activity:read` - Minimum scopes for reading clubs and events
- `state` - CSRF protection token (we use 32-byte random token)

**Important Notes:**
- Strava does NOT allow `/` at the end of redirect_uri
- State token should expire after 10 minutes
- Redirect URI must match EXACTLY what's configured in Strava app settings

### 2. Token Exchange
After user authorizes, Strava redirects to your callback with `code` and `state`:

```
POST https://www.strava.com/oauth/token
Content-Type: application/x-www-form-urlencoded

client_id={client_id}
client_secret={client_secret}
code={authorization_code}
grant_type=authorization_code
```

**Response:**
```json
{
  "access_token": "...",
  "refresh_token": "...",
  "expires_at": 1234567890,
  "expires_in": 21600,
  "token_type": "Bearer",
  "athlete": {
    "id": 12345,
    "firstname": "John",
    "lastname": "Doe",
    ...
  }
}
```

**Key Points:**
- Access token expires in 6 hours (21600 seconds)
- Refresh token can be used to get new access tokens
- We store tokens ephemerally (in-memory only, no database)
- Sessions auto-expire after 1 hour

---

## Finding Admin Clubs

### The Challenge
We need to identify which clubs a user is an admin/owner of to enforce TOS compliance.

### Attempted Approaches

#### ❌ Approach 1: `/clubs/{id}/admins` endpoint
```
GET https://www.strava.com/api/v3/clubs/{id}/admins
```

**Result:** Returns admin athlete data but **without athlete IDs**, making it impossible to match against the authenticated user.

**Response Example:**
```json
[
  {
    "firstname": "John",
    "lastname": "Doe",
    "profile": "https://...",
    // No "id" field!
  }
]
```

#### ❌ Approach 2: `/clubs/{id}/members` endpoint
```
GET https://www.strava.com/api/v3/clubs/{id}/members
```

**Result:** Returns all members with `admin` boolean flag, but this is wasteful - we'd need to fetch potentially hundreds of members just to check one user's admin status.

#### ✅ Approach 3: Detailed club endpoint (SOLUTION)
```
GET https://www.strava.com/api/v3/clubs/{id}
```

**Result:** Returns detailed club info including **`admin`** and **`owner`** boolean flags for the authenticated user!

**Response Example:**
```json
{
  "id": 123456,
  "name": "Portland Bike Club",
  "admin": true,      // ← Current user is admin
  "owner": false,     // ← Current user is owner
  "membership": "member",
  "member_count": 342,
  ...
}
```

### Recommended Flow

1. Get all clubs user is a member of:
   ```
   GET /athlete/clubs
   ```

2. For each club, check admin status:
   ```
   GET /clubs/{club_id}
   ```
   Check `admin` or `owner` fields

3. Filter to only clubs where `admin === true || owner === true`

4. Only show events from these filtered clubs

---

## Fetching Group Events

### Endpoint Discovery
Through testing, we discovered an **undocumented endpoint** for fetching club group events:

```
GET https://www.strava.com/api/v3/clubs/{id}/group_events
```

**Response:**
```json
[
  {
    "id": 789,
    "title": "Tuesday Night Social Ride",
    "description": "Easy-paced social ride...",
    "event_time": "2024-02-15T18:00:00Z",
    "address": "1234 Main St",
    "location_city": "Portland",
    "location_state": "OR",
    "location_country": "United States",
    "latitude": 45.5152,
    "longitude": -122.6784,
    "route_id": 12345,
    "route": {
      "id": 12345,
      "name": "Waterfront Loop",
      "distance": 25000,        // meters
      "elevation_gain": 150,    // meters
      "type": "ride",
      "sub_type": "road"
    },
    "skill_levels": "casual,recreational",
    "terrain": "mostly_flat",
    "visibility": "public",
    "women_only": false,
    "attending_count": 23,
    "interested_count": 45,
    "club_id": 123456,
    "organizer_id": 67890,
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-02-01T14:30:00Z"
  }
]
```

**Key Fields:**
- `event_time` - ISO 8601 timestamp (UTC)
- `latitude`/`longitude` - Starting location coordinates
- `route` - Embedded route object (if event has a route)
- `attending_count`/`interested_count` - RSVP counts
- All location fields provided (no need to geocode!)

---

## Rate Limiting

### Strava API Limits

**⚠️ IMPORTANT UPDATE (2026-01-31):** Live testing reveals **HIGHER limits** than documented!

**Documented limits:**
- 100 requests per 15 minutes
- 1000 requests per day

**Actual limits (verified from live API):**
- **200 requests per 15 minutes** (2x documented!)
- **2000 requests per day** (2x documented!)

**Rate limit headers returned in every response:**
```
X-Ratelimit-Limit: 200,2000         ← Actual limits
X-Ratelimit-Usage: 1,1              ← Current usage (15min, daily)
X-Readratelimit-Limit: 100,1000     ← Separate read-only limit
```

**⚠️ Case Sensitivity:**
- Headers use `X-Ratelimit-*` (lowercase 'l' in ratelimit)
- NOT `X-RateLimit-*` (capital R and L)
- Using wrong case returns empty string!

### Dual Rate Limit System

Strava enforces **TWO independent rate limit systems**:

1. **General Rate Limits** (applies to all API operations)
   - 200 requests per 15 minutes
   - 2000 requests per day
   - Headers: `X-Ratelimit-Limit` / `X-Ratelimit-Usage`

2. **Read-Only Rate Limits** (applies to GET requests specifically)
   - 100 requests per 15 minutes
   - 1000 requests per day
   - Headers: `X-Readratelimit-Limit` / `X-Readratelimit-Usage`

**⚠️ CRITICAL:** Since **ALL our operations are GET requests** (GetAthleteClubs, GetClubDetails, GetClubEvents), we're constrained by the **read-only limits (100/15min)**, not the general limits!

**Rate Limit Impact on User Flow:**
- User logs in → 1 call to exchange token
- Fetch clubs → 1 call to `/athlete/clubs`
- Check admin for 10 clubs → 10 calls to `/clubs/{id}`
- Fetch events for 5 clubs → 5 calls to `/clubs/{id}/group_events`
- **Total: ~17 calls per user session** (well within 100/15min limit)

**When we'd hit limits:**
- Power user with 50+ clubs → 1 + 50 admin checks = 51 calls (still safe)
- 5 concurrent users with 10 clubs each → 5 × 11 = 55 calls (safe)
- 10 concurrent users with 10 clubs each → 10 × 11 = 110 calls (**exceeds 100/15min!**)

### Recommendations
- **MUST track BOTH rate limit systems** in monitoring DB
- Alert when read limit exceeds 80% (80 of 100 requests in 15min window)
- Cache club admin status (store in session during OAuth)
- Batch event fetches per club
- Show rate limit warnings in UI if approaching limits
- Consider implementing exponential backoff for retries
- Monitor `read_limit_15min_usage` specifically (not just general limits!)

---

## Security & TOS Compliance

### Why Admin-Only?

1. **Respects Club Governance**: Ensures club leadership approves of event sharing
2. **TOS Compliance**: Strava expects responsible data usage
3. **Prevents Spam**: Random members can't import potentially inappropriate events
4. **Quality Control**: Admins curate what gets shared publicly

### Implementation Safeguards

1. **Hard Requirements:**
   - Always check `admin || owner` before showing events
   - Never allow event import if user is not admin/owner
   - Display clear error messages if user tries to import from non-admin club

2. **Session Security:**
   - Store tokens in-memory only (no database persistence)
   - Auto-expire sessions after 1 hour
   - Use CSRF state tokens (10-minute expiry)
   - Clear tokens on logout/window close

3. **Logging:**
   - Log all OAuth flows with athlete IDs
   - Log admin checks and denials
   - Use `STRAVA_DEBUG=true` for verbose logging in development

---

## Data Mapping: Strava → CycleScene

### Direct Mappings
| Strava Field | CycleScene Field | Notes |
|--------------|------------------|-------|
| `title` | `title` | Event name |
| `description` | `details` | Event description |
| `event_time` | `startTime` | Parse ISO 8601 to local time |
| `latitude` | `latitude` | Already provided! |
| `longitude` | `longitude` | Already provided! |
| `address` | `address` | Full address string |
| `location_city` | `city` | Need to map to CycleScene city codes (pdx, slc, etc.) |

### Route Handling
- If `route_id` exists, fetch full route data from Strava
- Convert to GPX → GeoJSON (use existing route converter)
- Store in `routes` table with source="strava"

### Fields Requiring Transformation
- **Event Time:** Strava uses UTC ISO 8601, convert to local timezone
- **Duration:** Not provided by Strava, use default or prompt user
- **Recurrence:** Strava events are single-instance, map to `recurrence: "once"`
- **Image:** Strava doesn't provide event images, leave optional

### Missing/Optional Fields
- `organizer_email` - Not from Strava, use OAuth user's email
- `venue_name` - May need to geocode address or use club name
- `image_url` - Not available, allow manual upload
- `event_duration_minutes` - Default to 120 or make customizable

---

## Error Handling

### Common Errors

1. **Invalid/Expired Token (401)**
   ```json
   {
     "message": "Authorization Error",
     "errors": [{"resource": "Athlete", "field": "access_token", "code": "invalid"}]
   }
   ```
   **Solution:** Delete session, redirect to re-authenticate

2. **Rate Limit Exceeded (429)**
   ```json
   {
     "message": "Rate Limit Exceeded",
     "errors": [{"resource": "Application", "field": "rate limit", "code": "exceeded"}]
   }
   ```
   **Solution:** Show user-friendly message, retry after 15 minutes

3. **Club Not Found (404)**
   ```json
   {
     "message": "Record Not Found",
     "errors": [{"resource": "Club", "field": "id", "code": "not found"}]
   }
   ```
   **Solution:** Filter out club from list

4. **No Events (Empty Array)**
   **Solution:** Show "No upcoming events" message

---

## Testing & Debugging

### Environment Variables
```bash
# Enable verbose Strava OAuth logging
STRAVA_DEBUG=true

# OAuth credentials
STRAVA_CLIENT_ID=your_client_id
STRAVA_CLIENT_SECRET=your_client_secret

# For local testing
STRAVA_CALLBACK_URL=http://localhost:8080/api/strava/callback
```

### Debug Logging Points
1. OAuth initiation (state generation)
2. OAuth callback (code exchange)
3. Token storage in session
4. Club fetching
5. Admin status checks
6. Event fetching
7. Event conversion and submission

### Test Tool
We created `backend/cmd/strava-test/main.go` - a standalone OAuth testing tool with web UI for exploring:
- Full OAuth flow
- Club listing
- Admin detection
- Event fetching

**Run with:**
```bash
cd backend/cmd/strava-test
go run main.go
# Visit http://localhost:3000
```

---

## Implementation Checklist

- [x] OAuth models and types
- [x] Session management (in-memory)
- [x] CSRF protection (state tokens)
- [ ] Strava API client
- [ ] Admin detection service
- [ ] Event fetching service
- [ ] Event converter (Strava → CycleScene)
- [ ] HTTP handlers (OAuth flow)
- [ ] WebSocket progress tracking
- [ ] Frontend OAuth button
- [ ] Frontend event selection UI
- [ ] Multi-event import with progress

---

## Resources

- **Strava API Docs:** https://developers.strava.com/docs/reference/
- **OAuth Guide:** https://developers.strava.com/docs/authentication/
- **Rate Limits:** https://developers.strava.com/docs/rate-limits/
- **Test Tool Location:** `backend/cmd/strava-test/main.go`

---

**Last Updated:** 2026-01-31
**Status:** Foundation complete, ready for full implementation
