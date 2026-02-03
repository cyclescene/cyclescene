# Milestone 4: Frontend - SvelteKit Integration

**Goal:** Add Strava import UI to the form frontend

**Status:** COMPLETED (2026-02-03)

---

## Design Decisions (Finalized 2026-02-03)

### Decision #1: UI Entry Point
- **Approach:** Same form page, toggle between manual form and Strava import mode
- **Rationale:** User stays in familiar context, no extra navigation
- **Implementation:** `stravaMode` state variable toggles between form and import UI

### Decision #2: OAuth Flow
- **Approach:** Popup window (not full-page redirect)
- **Rationale:** Preserves parent URL with token + city params, smoother UX
- **Implementation:** `window.open()` for OAuth, `postMessage` or poll for completion

### Decision #3: OAuth Return
- **Approach:** Popup closes, parent window detects and switches to import mode
- **Implementation:** Callback page sends `postMessage` to parent, then closes

### Decision #4: City Handling
- **Source:** Always present from URL params (`?city=pdx`)
- **Rationale:** City is required for form access (enforced by PWA redirect)
- **Implementation:** Read from `$page.url.searchParams`, pass to OAuth initiate

### Decision #5: Customization
- **Approach:** "Customize" button per event, expands inline options
- **Fields:** Audience (dropdown), Duration (dropdown), Image (upload)
- **Defaults:** Audience: `G` (All Ages), Duration: from route or null, Image: none
- **Rationale:** Quick import for most cases, customization available when needed

### Decision #6: Post-Import UX
- **Footer note:** "You can update event details after import using the edit links in your confirmation email."
- **Results:** Show success/error per event with edit links (magic links)

### Decision #7: Duplicate Events
- **Approach:** Let users attempt re-import, backend handles via DB constraint
- **Backend behavior:** source_id includes date (`{event_id}_{YYYY-MM-DD}`), so future occurrences of recurring events can be imported
- **True duplicates:** Same event + same date = blocked with "Already imported" message

---

## Files Summary (Milestone 4)

**New Files to Create:**
```
frontends/form/src/
├── routes/
│   └── strava/
│       └── callback/
│           └── +page.svelte          # OAuth callback (closes popup)
├── lib/
│   ├── components/
│   │   └── strava/
│   │       ├── StravaImportButton.svelte   # "Import from Strava" button
│   │       ├── StravaImport.svelte         # Main import container
│   │       ├── ClubList.svelte             # Admin clubs accordion
│   │       ├── EventList.svelte            # Events within a club
│   │       ├── EventCard.svelte            # Single event with checkbox + customize
│   │       ├── EmailInput.svelte           # Organizer email collection
│   │       ├── ImportProgress.svelte       # WebSocket progress display
│   │       └── ImportResults.svelte        # Final results with edit links
│   ├── api/
│   │   └── strava.ts                 # Strava API client functions
│   ├── types/
│   │   └── strava.ts                 # TypeScript interfaces
│   └── utils/
│       └── strava-websocket.ts       # WebSocket client helper
```

**Files to Modify:**
- `frontends/form/src/routes/+page.svelte` - Add mode toggle and Strava import
- `frontends/form/.env.example` - Add `PUBLIC_STRAVA_ENABLED`, `PUBLIC_STRAVA_DEBUG`

**Existing Files to Reference:**
- `frontends/form/src/lib/components/ImageUploader.svelte` - For image upload pattern
- `frontends/form/src/lib/components/ui/*` - shadcn/ui components (Button, Card, etc.)
- `frontends/form/src/lib/api/client.ts` - For API call patterns

**Backend Endpoints (from M3):**
- `GET /v1/strava/auth/initiate?city={code}` - Start OAuth, returns auth URL
- `GET /v1/strava/auth/callback` - OAuth callback (sets session cookie)
- `GET /v1/strava/admin-clubs` - Get clubs where user is admin
- `GET /v1/strava/clubs/{clubId}/events` - Get events for a club (lazy load)
- `POST /v1/strava/logout` - Clear session
- `WS /v1/strava/import` - WebSocket for import with progress

**No modifications to:**
- Backend code (complete in M1-M3)
- Other frontend apps (dashboard, pwa, directory)

---

## User Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ PWA (cyclescene.cc/pdx)                                                     │
│   → User clicks "Submit a ride"                                             │
│   → PWA generates BFF token                                                 │
│   → Redirects to /form?city=pdx&token=abc123                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│ Form Page (/form?city=pdx&token=abc123)                                     │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  [Submit a Ride]                    [Import from Strava] ← NEW      │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│   IF stravaMode = false:                                                    │
│     → Show existing manual form                                             │
│                                                                             │
│   IF stravaMode = true:                                                     │
│     → Show Strava import UI                                                 │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│ OAuth Popup Flow                                                            │
│                                                                             │
│   1. User clicks "Import from Strava"                                       │
│   2. Popup opens → /v1/strava/auth/initiate?city=pdx                       │
│   3. Redirects to Strava OAuth                                              │
│   4. User authorizes                                                        │
│   5. Strava → /v1/strava/auth/callback (sets cookie)                       │
│   6. Callback → /strava/callback (frontend)                                │
│   7. Callback page sends postMessage to parent, closes                      │
│   8. Parent detects auth complete, sets stravaMode = true                   │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│ Strava Import UI                                                            │
│                                                                             │
│   Step 1: Enter Email                                                       │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Email for edit links: [_______________________]                    │   │
│   │                                              [Continue]             │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│   Step 2: Select Events                                                     │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Portland Cycling Club (3 events)                          [▼]     │   │
│   │  ┌───────────────────────────────────────────────────────────────┐ │   │
│   │  │ ☑ Tuesday Night Ride                                          │ │   │
│   │  │   📍 Portland, OR • 📅 Feb 15, 6:00 PM • 🚴 25 mi            │ │   │
│   │  │                                              [Customize]      │ │   │
│   │  ├───────────────────────────────────────────────────────────────┤ │   │
│   │  │ ☐ Thursday Social Ride                                        │ │   │
│   │  │   📍 Portland, OR • 📅 Feb 17, 5:30 PM • 🚴 15 mi            │ │   │
│   │  │                                              [Customize]      │ │   │
│   │  └───────────────────────────────────────────────────────────────┘ │   │
│   │                                                                     │   │
│   │  SLC Bike Collective (2 events)                            [▶]     │   │
│   │  (Click to load events)                                            │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  [Back to Manual Form]              [Import 2 Selected Events]     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│   Footer: "You can update event details after import using the edit        │
│            links in your confirmation email."                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│ Event Card - Expanded (after clicking Customize)                            │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │ ☑ Tuesday Night Ride                                                │   │
│   │   📍 Portland, OR • 📅 Feb 15, 6:00 PM • 🚴 25 mi                  │   │
│   │                                                                     │   │
│   │   Audience:  [G - All Ages        ▼]                               │   │
│   │   Duration:  [2 hours             ▼]                               │   │
│   │   Image:     [📷 Upload Image]                                     │   │
│   │                                                        [Collapse]  │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│ Import Progress (WebSocket)                                                 │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Importing 2 Events                                                 │   │
│   │                                                                     │   │
│   │  ✓ Tuesday Night Ride                                              │   │
│   │    ✓ Fetching  ✓ Location  ✓ Route  ✓ Saved                       │   │
│   │    Event created! [Edit →]                                         │   │
│   │                                                                     │   │
│   │  ⟳ Thursday Social Ride                                            │   │
│   │    ✓ Fetching  ✓ Location  ⟳ Route...                             │   │
│   │                                                                     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│ Import Complete                                                             │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  ✓ Import Complete!                                                 │   │
│   │                                                                     │   │
│   │  2 events imported successfully                                     │   │
│   │  A confirmation email has been sent to user@example.com            │   │
│   │                                                                     │   │
│   │  ✓ Tuesday Night Ride          [Edit →]                            │   │
│   │  ✓ Thursday Social Ride        [Edit →]                            │   │
│   │                                                                     │   │
│   │  [Import More Events]              [Done]                          │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## TypeScript Types

```typescript
// frontends/form/src/lib/types/strava.ts

// ============================================================================
// API Response Types (from backend)
// ============================================================================

export interface StravaClub {
  id: number;
  name: string;
  profile_medium: string;
  city: string;
  state: string;
  country: string;
  member_count: number;
  is_admin: boolean;
  is_owner: boolean;
}

export interface StravaGroupEvent {
  id: number;  // Note: Display only - send as string to WebSocket!
  title: string;
  description: string;
  activity_type: string;
  upcoming_occurrences: string[];  // ISO 8601 timestamps
  zone: string;  // IANA timezone
  address: string;
  start_latlng: [number, number];  // [lat, lng]
  route_id: number | null;
  route: StravaRoute | null;
  skill_levels: string | null;
  terrain: string | null;
  women_only: boolean;
  private: boolean;
  club_id: number;
}

export interface StravaRoute {
  id: number;
  name: string;
  distance: number;  // meters
  elevation_gain: number;  // meters
}

// ============================================================================
// WebSocket Message Types (match backend exactly)
// ============================================================================

// Client → Server
export interface ImportRequest {
  session_id?: string;  // Optional - backend reads from HttpOnly cookie if not provided
  organizer_email: string;
  events: EventImportConfig[];
}

export interface EventImportConfig {
  strava_event_id: string;  // MUST be string (int64 precision issue)
  club_id: number;
  overrides?: EventOverrides;
}

export interface EventOverrides {
  audience?: string;  // "G", "F", "A", "E"
  image_url?: string;
  event_duration_minutes?: number;
}

// Server → Client
export interface ProgressMessage {
  type: 'heartbeat' | 'progress' | 'complete' | 'done' | 'error';
  event_index?: number;
  total_events?: number;
  strava_event_id?: string;
  event_title?: string;
  step?: 'fetching' | 'coordinates' | 'route' | 'database';
  status?: 'in_progress' | 'success' | 'error';
  message?: string;
  // Complete message fields
  cyclescene_event_id?: number;
  edit_token?: string;
  edit_url?: string;
  success?: boolean;
  error?: string | null;
  // Done message fields
  total_imported?: number;
  total_failed?: number;
  summary_email_sent?: boolean;
  results?: ImportResult[];
}

export interface ImportResult {
  strava_event_id: string;
  title: string;
  success: boolean;
  cyclescene_event_id?: number;
  edit_token?: string;
  edit_url?: string;
  error?: string;
}

// ============================================================================
// UI State Types
// ============================================================================

export type ImportStep = 'connect' | 'email' | 'select' | 'importing' | 'complete';

export interface StravaState {
  authenticated: boolean;
  step: ImportStep;
  organizerEmail: string;
  adminClubs: StravaClub[];
  clubEvents: Map<number, StravaGroupEvent[]>;  // clubId → events
  selectedEvents: Map<string, EventImportConfig>;  // eventId → config
  expandedClubs: Set<number>;
  expandedEvents: Set<string>;  // Events with customize panel open
  importProgress: Map<string, ProgressMessage>;  // eventId → latest progress
  importResults: ImportResult[];
  error: string | null;
}

export interface ClubWithEvents extends StravaClub {
  events?: StravaGroupEvent[];
  loading?: boolean;
  error?: string;
}
```

---

## Tasks

### 4.1 - Create TypeScript Types
- [x] Create `frontends/form/src/lib/types/strava.ts`
- [x] Define all types as specified above
- [x] Ensure WebSocket message types match backend exactly

**Validation:**
```bash
cd frontends/form && pnpm run check
```

### 4.2 - Create Strava API Client
- [x] Create `frontends/form/src/lib/api/strava.ts`
- [x] `initiateAuth(city: string)` - Opens OAuth popup
- [x] `checkSession()` - Verifies session cookie is valid
- [x] `fetchAdminClubs()` - GET /v1/strava/admin-clubs
- [x] `fetchClubEvents(clubId: number)` - GET /v1/strava/clubs/{id}/events
- [x] `logout()` - POST /v1/strava/logout

**Debug Points:**
```typescript
if (import.meta.env.PUBLIC_STRAVA_DEBUG === 'true') {
  console.log('[Strava] Fetching admin clubs...');
}
```

### 4.3 - Create WebSocket Client Helper
- [x] Create `frontends/form/src/lib/utils/strava-websocket.ts`
- [x] Handle connection lifecycle (connect, reconnect, close)
- [x] Parse incoming messages to typed `ProgressMessage`
- [x] Handle heartbeat messages (keep connection alive)
- [x] Emit events for progress updates

**Key considerations:**
- WebSocket URL: `wss://{API_HOST}/v1/strava/import`
- Must send `ImportRequest` as first message after connect
- Event IDs must be strings (int64 precision)

### 4.4 - Create OAuth Callback Page
- [x] Create `frontends/form/src/routes/strava/callback/+page.svelte`
- [x] Send `postMessage` to parent window on load
- [x] Auto-close popup after short delay
- [x] Show fallback "Close this window" message

```svelte
<script>
  import { onMount } from 'svelte';

  onMount(() => {
    if (window.opener) {
      window.opener.postMessage({ type: 'strava-auth-complete' }, '*');
      setTimeout(() => window.close(), 1000);
    }
  });
</script>

<div class="text-center p-8">
  <p>Authentication successful!</p>
  <p class="text-sm text-muted-foreground">This window should close automatically.</p>
  <button onclick={() => window.close()}>Close</button>
</div>
```

### 4.5 - Install Additional shadcn Components

Before creating components, install these shadcn components:

```bash
cd frontends/form
pnpm dlx shadcn-svelte@latest add accordion
pnpm dlx shadcn-svelte@latest add collapsible
pnpm dlx shadcn-svelte@latest add progress
pnpm dlx shadcn-svelte@latest add badge
pnpm dlx shadcn-svelte@latest add alert
```

**shadcn Component Mapping:**

| M4 Component | shadcn Components Used |
|--------------|----------------------|
| StravaImportButton | `Button` |
| StravaImport | `Card`, `Button` |
| EmailInput | `Card`, `Input`, `Label`, `Button` |
| ClubList | `Accordion` |
| EventList | (container only) |
| EventCard | `Card`, `Checkbox`, `Collapsible`, `Select`, `Button` |
| ImportProgress | `Card`, `Progress`, `Badge` |
| ImportResults | `Card`, `Button`, `Badge`, `Alert` |

### 4.6 - Create Strava Import Components

#### 4.6.1 - StravaImportButton.svelte
**shadcn:** `Button`
- [x] "Import from Strava" button with Strava brand color (#FC4C02)
- [x] Initiates OAuth popup on click
- [x] Listens for `postMessage` from popup
- [x] Calls `onAuthComplete` callback when done

```svelte
<Button variant="outline" class="border-[#FC4C02] text-[#FC4C02] hover:bg-[#FC4C02] hover:text-white">
  <StravaIcon class="mr-2 h-4 w-4" />
  Import from Strava
</Button>
```

#### 4.6.2 - StravaImport.svelte (Main Container)
**shadcn:** `Card`, `Button`
- [x] Manages overall import state/flow
- [x] Renders current step component (email → select → importing → complete)
- [x] "Back to Manual Form" button at top

#### 4.6.3 - EmailInput.svelte
**shadcn:** `Card`, `Input`, `Label`, `Button`
- [x] Email input with validation (reuse existing email validation pattern)
- [x] "Continue" button to proceed to event selection
- [x] Explanation text about why email is needed (for edit links)

#### 4.6.4 - ClubList.svelte
**shadcn:** `Accordion`
- [x] `Accordion.Root` with `multiple` for multi-expand
- [x] `Accordion.Item` per club
- [x] `Accordion.Trigger` shows club name + event count
- [x] `Accordion.Content` renders EventList
- [x] Lazy-loads events when club first expanded
- [x] Shows loading spinner in content while fetching

```svelte
<Accordion.Root type="multiple" bind:value={expandedClubs}>
  {#each clubs as club}
    <Accordion.Item value={club.id.toString()}>
      <Accordion.Trigger>
        {club.name} ({club.event_count} events)
      </Accordion.Trigger>
      <Accordion.Content>
        <EventList events={clubEvents.get(club.id)} />
      </Accordion.Content>
    </Accordion.Item>
  {/each}
</Accordion.Root>
```

#### 4.6.5 - EventList.svelte
- [x] Container for EventCard components
- [x] Renders EventCard for each event
- [x] Shows "No upcoming events" message if empty

#### 4.6.6 - EventCard.svelte
**shadcn:** `Card`, `Checkbox`, `Collapsible`, `Select`, `Button`
- [x] `Checkbox` for selection (left side)
- [x] Event details (title, date, location, distance)
- [x] `Collapsible` for customize panel
- [x] `Select` for Audience (G/F/A/E options)
- [x] `Select` for Duration (preset options)
- [x] Reuse existing `ImageUploader` component

```svelte
<Card.Root class="p-4">
  <div class="flex items-start gap-3">
    <Checkbox bind:checked={selected} />
    <div class="flex-1">
      <h4 class="font-medium">{event.title}</h4>
      <p class="text-sm text-muted-foreground">
        📍 {event.address} • 📅 {formatDate(event)} • 🚴 {formatDistance(event)}
      </p>

      <Collapsible.Root bind:open={customizeOpen}>
        <Collapsible.Trigger asChild let:builder>
          <Button builders={[builder]} variant="ghost" size="sm">
            {customizeOpen ? 'Collapse' : 'Customize'}
          </Button>
        </Collapsible.Trigger>
        <Collapsible.Content class="mt-3 space-y-3">
          <div>
            <Label>Audience</Label>
            <Select.Root bind:value={audience}>
              <Select.Trigger><Select.Value /></Select.Trigger>
              <Select.Content>
                <Select.Item value="G">G - All Ages</Select.Item>
                <Select.Item value="F">F - Family Friendly</Select.Item>
                <Select.Item value="A">A - Adult (21+)</Select.Item>
              </Select.Content>
            </Select.Root>
          </div>
          <div>
            <Label>Duration</Label>
            <Select.Root bind:value={duration}>
              <!-- duration options -->
            </Select.Root>
          </div>
          <div>
            <Label>Image (optional)</Label>
            <ImageUploader onUploadComplete={handleImageUpload} />
          </div>
        </Collapsible.Content>
      </Collapsible.Root>
    </div>
  </div>
</Card.Root>
```

#### 4.6.7 - ImportProgress.svelte
**shadcn:** `Card`, `Progress`, `Badge`
- [x] WebSocket connection management
- [x] `Card` per event being imported
- [x] `Progress` bar showing overall progress
- [x] `Badge` for step status (in_progress, success, error)
- [x] Handles heartbeat messages (keep-alive)

```svelte
<Card.Root class="p-4">
  <h3 class="font-medium mb-4">Importing {events.length} Events</h3>

  {#each events as event, i}
    <div class="mb-4">
      <div class="flex items-center gap-2 mb-2">
        {#if progress[i]?.status === 'success'}
          <Badge variant="success">✓</Badge>
        {:else if progress[i]?.status === 'error'}
          <Badge variant="destructive">✗</Badge>
        {:else}
          <Badge variant="secondary">⟳</Badge>
        {/if}
        <span>{event.title}</span>
      </div>
      <div class="flex gap-2 text-xs text-muted-foreground">
        <span class={stepClass('fetching', i)}>Fetching</span>
        <span class={stepClass('coordinates', i)}>Location</span>
        <span class={stepClass('route', i)}>Route</span>
        <span class={stepClass('database', i)}>Saved</span>
      </div>
    </div>
  {/each}
</Card.Root>
```

#### 4.6.8 - ImportResults.svelte
**shadcn:** `Card`, `Button`, `Badge`, `Alert`
- [x] `Alert` for success summary message
- [x] `Card` with list of imported events
- [x] `Badge` for success/error per event
- [x] `Button` for edit links (opens in new tab)
- [x] "Import More Events" and "Done" buttons

```svelte
<Alert.Root variant="success">
  <Alert.Title>Import Complete!</Alert.Title>
  <Alert.Description>
    {successCount} events imported. A confirmation email has been sent.
  </Alert.Description>
</Alert.Root>

<Card.Root class="mt-4 p-4">
  {#each results as result}
    <div class="flex items-center justify-between py-2">
      <div class="flex items-center gap-2">
        <Badge variant={result.success ? 'success' : 'destructive'}>
          {result.success ? '✓' : '✗'}
        </Badge>
        <span>{result.title}</span>
      </div>
      {#if result.edit_url}
        <Button variant="outline" size="sm" href={result.edit_url} target="_blank">
          Edit →
        </Button>
      {/if}
    </div>
  {/each}
</Card.Root>

<div class="flex justify-end gap-2 mt-4">
  <Button variant="outline" onclick={onImportMore}>Import More Events</Button>
  <Button onclick={onDone}>Done</Button>
</div>
```

### 4.7 - Integrate into Form Page
- [x] Modify `frontends/form/src/routes/+page.svelte`
- [x] Add `stravaMode` state variable
- [x] Conditionally render form OR Strava import UI
- [x] Add StravaImportButton near page header
- [ ] Handle OAuth callback via `postMessage` listener

```svelte
<script>
  let stravaMode = $state(false);

  // Listen for OAuth completion from popup
  function handleMessage(event: MessageEvent) {
    if (event.data?.type === 'strava-auth-complete') {
      stravaMode = true;
    }
  }

  onMount(() => {
    window.addEventListener('message', handleMessage);
    return () => window.removeEventListener('message', handleMessage);
  });
</script>

{#if stravaMode}
  <StravaImport
    city={data.city}
    onClose={() => stravaMode = false}
  />
{:else}
  <!-- Existing form -->
  <div class="flex justify-between items-center mb-6">
    <h1>Submit a Ride</h1>
    <StravaImportButton onAuthComplete={() => stravaMode = true} />
  </div>
  <!-- ... rest of form ... -->
{/if}
```

### 4.8 - Add Environment Variables
- [ ] Update `frontends/form/.env.example`

```bash
# Strava Integration
PUBLIC_STRAVA_ENABLED=true
PUBLIC_STRAVA_DEBUG=false
PUBLIC_API_URL=https://api.cyclescene.cc
```

### 4.9 - Testing & Validation
- [ ] `pnpm run check` - Zero TypeScript errors
- [ ] `pnpm run build` - Production build succeeds
- [ ] `pnpm run lint` - No lint errors
- [ ] Manual testing with real Strava account
- [ ] Test OAuth popup flow
- [ ] Test event selection and customization
- [ ] Test WebSocket import with progress
- [ ] Test error states (no clubs, no events, import failure)

---

## Component Props & Events

### StravaImportButton
```typescript
interface Props {
  city: string;
  onAuthComplete: () => void;
}
```

### StravaImport
```typescript
interface Props {
  city: string;
  onClose: () => void;  // Return to manual form
}
```

### EventCard
```typescript
interface Props {
  event: StravaGroupEvent;
  selected: boolean;
  config: EventImportConfig | null;
  onToggle: (selected: boolean) => void;
  onConfigChange: (config: EventImportConfig) => void;
}
```

### ImportProgress
```typescript
interface Props {
  organizerEmail: string;
  events: EventImportConfig[];
  onComplete: (results: ImportResult[]) => void;
  onError: (error: string) => void;
}
```

---

## Debug Logging

All components should respect `PUBLIC_STRAVA_DEBUG`:

```typescript
function debugLog(message: string, data?: any) {
  if (import.meta.env.PUBLIC_STRAVA_DEBUG === 'true') {
    console.log(`[Strava] ${message}`, data ?? '');
  }
}

// Usage
debugLog('OAuth popup opened', { city });
debugLog('Admin clubs fetched', { count: clubs.length });
debugLog('WebSocket message received', message);
```

---

## Error Handling

| Scenario | User Message | Action |
|----------|--------------|--------|
| OAuth popup blocked | "Please allow popups to connect with Strava" | Show instructions |
| OAuth denied | "Strava authorization was cancelled" | Return to form |
| No admin clubs | "You're not an admin of any cycling clubs on Strava" | Show explanation |
| No events in club | "No upcoming events in this club" | Disable club row |
| WebSocket disconnect | "Connection lost. Retrying..." | Auto-reconnect |
| Import failed (event) | "Failed: {error message}" | Show per-event, continue others |
| Rate limit exceeded | "Strava API limit reached. Please try again in 15 minutes." | Stop import |

---

## Completion Checklist

**Prerequisites (BEFORE starting build):**
```bash
cd frontends/form

# Install required shadcn components
pnpm dlx shadcn-svelte@latest add accordion
pnpm dlx shadcn-svelte@latest add collapsible
pnpm dlx shadcn-svelte@latest add progress
pnpm dlx shadcn-svelte@latest add badge
pnpm dlx shadcn-svelte@latest add alert

# Verify installation
ls src/lib/components/ui/
# Should now include: accordion, alert, badge, collapsible, progress
```
- [ ] `accordion` component installed
- [ ] `collapsible` component installed
- [ ] `progress` component installed
- [ ] `badge` component installed
- [ ] `alert` component installed

**Code Quality:**
- [ ] All components created with TypeScript
- [ ] Types match backend WebSocket protocol exactly
- [ ] Event IDs sent as strings (int64 precision issue)
- [ ] Debug logging respects PUBLIC_STRAVA_DEBUG
- [ ] Error handling for all failure scenarios

**Build & Validation:**
- [ ] `pnpm run check` - Zero TypeScript errors
- [ ] `pnpm run build` - Production build succeeds
- [ ] `pnpm run lint` - No lint errors
- [ ] `pnpm run dev` - Works locally

**Functionality:**
- [ ] OAuth popup flow works
- [ ] Admin clubs load correctly
- [ ] Events lazy-load per club
- [ ] Event selection with checkboxes
- [ ] Customize panel (audience, duration, image)
- [ ] WebSocket import with real-time progress
- [ ] Results with edit links
- [ ] "Back to Manual Form" works
- [ ] Footer note about editing displayed

**Documentation:**
- [ ] Environment variables in `.env.example`
- [ ] No TODO comments without issues
- [ ] Debug logging documented

---

## Notes

### JavaScript int64 Precision Issue
Strava event IDs exceed `Number.MAX_SAFE_INTEGER`. Always:
- Display event IDs for debugging only
- Send to WebSocket as **strings**: `"strava_event_id": "3453605542995245000"`
- Backend expects string format in `EventImportConfig`

### Session Cookie
- Backend sets `strava_session_id` cookie on OAuth callback
- Cookie is HttpOnly, so frontend can't read it directly
- Session validation happens server-side on API calls
- Frontend just needs to include `credentials: 'include'` in fetch requests
- WebSocket handler reads session from cookie automatically (no need to send session_id in message)

### Audience Codes
- `G` = All Ages (General)
- `F` = Family Friendly
- `A` = Adult (21+)
- `E` = Explicit/Adult Content

Default to `G` for Strava imports.
