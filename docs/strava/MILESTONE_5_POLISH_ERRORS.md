# Milestone 5: Frontend - Polish & Error Handling

**Goal:** Ensure robust error handling and great UX

---

## Files Summary (Milestone 5)

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

---

## Tasks

### 5.1 - Error Handling
- [ ] Handle OAuth failure (user denies)
- [ ] Handle session expiry (redirect to re-auth)
- [ ] Handle rate limit errors (show friendly message)
- [ ] Handle WebSocket disconnection (retry logic)
- [ ] Handle partial import failures (some events succeed, some fail)

**Error Messages:**
- OAuth denied: "Authorization cancelled. You can try again anytime."
- Session expired: "Your session expired. Please reconnect to Strava."
- Rate limit: "Strava API limit reached. Please try again in 15 minutes."
- Not admin: "You must be an admin or owner of a club to import its events."

### 5.2 - Loading States
- [ ] Show loading spinner during OAuth
- [ ] Show skeleton loaders for clubs/events
- [ ] Disable buttons during import process
- [ ] Add timeout handling (30s for event fetch, 60s for import)

### 5.3 - Responsive Design
- [ ] Ensure components work on mobile
- [ ] Test OAuth flow on mobile (popup vs redirect)
- [ ] Optimize event cards for small screens

### 5.4 - Accessibility
- [ ] Add ARIA labels to buttons
- [ ] Ensure keyboard navigation works
- [ ] Add focus management for modals
- [ ] Test with screen reader

### 5.5 - Add Environment Variables
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
