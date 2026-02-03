# Milestone 4: Frontend - SvelteKit Integration

**Goal:** Add Strava import UI to the form frontend

---

## Files Summary (Milestone 4)

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

---

## Tasks

### 4.1 - Create Strava Auth Store
- [ ] Create `frontends/form/src/lib/stores/strava.ts`
- [ ] Manage OAuth state (authenticated, session_id, admin_clubs)
- [ ] Persist session ID in localStorage (temporary)
- [ ] Add methods: `initiateAuth()`, `fetchAdminClubs()`, `logout()`

**Debug Points:**
- Log OAuth initiation
- Log session ID storage
- Log admin clubs fetched

### 4.2 - Create Strava Import Component
- [ ] Create `frontends/form/src/lib/components/StravaImport.svelte`
- [ ] Add Strava icon button
- [ ] Show "Connect with Strava" modal
- [ ] Display admin clubs as cards
- [ ] Show event list per club (collapsible)
- [ ] Add checkboxes for multi-event selection
- [ ] Add "Customize before import" toggle per event

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

### 4.3 - Create Import Progress Component
- [ ] Create `frontends/form/src/lib/components/ImportProgress.svelte`
- [ ] Connect to WebSocket endpoint
- [ ] Display progress for each event being imported
- [ ] Show steps: 1/4 Uploading image, 2/4 Geocoding, 3/4 Building route, 4/4 Saving
- [ ] Show success/error states per event
- [ ] Display final results with edit links

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

### 4.4 - Integrate into Form Page
- [ ] Add Strava import button to `frontends/form/src/routes/+page.svelte`
- [ ] Position near form title or as alternative to manual form
- [ ] Add toggle: "Manual Entry" vs "Import from Strava"
- [ ] Respect city parameter (pass to import flow)

### 4.5 - Add TypeScript Types
- [ ] Create `frontends/form/src/lib/types/strava.ts`
- [ ] Define types for Club, GroupEvent, ImportProgress, etc.

**Validation:**
```bash
cd frontends/form
npm run check  # TypeScript validation
npm run build  # Production build
npm run dev    # Manual testing
```
