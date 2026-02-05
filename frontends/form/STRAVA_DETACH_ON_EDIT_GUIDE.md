# Strava Detach-on-Edit Frontend Implementation Guide

**Phase 2, Milestone 2.3** - Detach-on-Edit Feature
**Status:** Backend Complete ✅ | Frontend Warning Added ✅ | Full Edit Dialog Pending

---

## Current Implementation (Feb 2026)

### ✅ What's Implemented

1. **Backend Detach Logic** (`backend/internal/api/ride/repo.go:148-250`)
   - Automatically detaches Strava events when edited
   - Sets `source=NULL` and deletes `strava_event_metadata`
   - Uses transactions for atomicity

2. **Frontend Warning Banner** (`frontends/form/src/routes/rides/edit/+page.svelte:160-174`)
   - Shows amber warning for Strava events (`source='strava'`)
   - Informs users the event is synced from Strava
   - Explains sync behavior

3. **API Returns Source Field**
   - `GET /api/v1/rides/edit/{token}` returns `source` field
   - Frontend can detect Strava events

### ⏸️ What's Pending

The current edit page (`/rides/edit`) only allows **occurrence editing** (time, cancellation), not full event editing. Full event editing would trigger the backend detachment logic when implemented.

---

## How to Implement Full Edit with Confirmation Dialog

When you add full event editing to the frontend, follow these steps:

### Step 1: Detect Strava Events

The `RideData` interface already includes the `source` field:

```typescript
interface RideData {
  event: {
    // ... other fields
    source?: string; // 'strava' | null
  };
}
```

Check if event is from Strava:

```typescript
const isStravaEvent = rideData?.event.source === 'strava';
```

### Step 2: Add Confirmation Dialog Before Submission

When the user submits a full event edit form, show a confirmation dialog for Strava events:

**Option A: Browser Confirm (Simple)**

```typescript
const handleSubmit = async (e: Event) => {
  e.preventDefault();

  // Show warning for Strava events
  if (isStravaEvent) {
    const confirmed = confirm(
      "⚠️ This event was imported from Strava and is automatically updated in the background.\n\n" +
      "Making changes will disconnect it from Strava and stop automatic updates.\n\n" +
      "Continue with edit?"
    );

    if (!confirmed) {
      return; // User cancelled
    }
  }

  // Proceed with submission
  await submitForm();
};
```

**Option B: Custom Modal (Recommended)**

Create a reusable confirmation dialog component:

```svelte
<!-- lib/components/StravaDetachDialog.svelte -->
<script lang="ts">
  import * as AlertDialog from "$lib/components/ui/alert-dialog";

  export let open = $state(false);
  export let onConfirm: () => void;
  export let onCancel: () => void;
</script>

<AlertDialog.Root bind:open>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Disconnect from Strava?</AlertDialog.Title>
      <AlertDialog.Description>
        This event was imported from Strava and is automatically synchronized every 3 days.

        Making changes to event details will:
        - Disconnect this event from Strava
        - Stop automatic updates
        - Convert it to a native CycleScene event
        - Give you full control over all fields

        This action cannot be undone. The event will remain on CycleScene but won't sync with Strava anymore.
      </AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <AlertDialog.Cancel onclick={onCancel}>Cancel</AlertDialog.Cancel>
      <AlertDialog.Action onclick={onConfirm}>Continue with Edit</AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
```

Usage in edit form:

```svelte
<script lang="ts">
  import StravaDetachDialog from "$lib/components/StravaDetachDialog.svelte";

  let showStravaDialog = $state(false);
  let pendingSubmit: (() => void) | null = $state(null);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();

    if (isStravaEvent && !userAlreadyConfirmed) {
      // Show dialog and wait for confirmation
      pendingSubmit = () => submitForm();
      showStravaDialog = true;
      return;
    }

    await submitForm();
  };

  const confirmStravaDetach = () => {
    showStravaDialog = false;
    if (pendingSubmit) {
      pendingSubmit();
      pendingSubmit = null;
    }
  };

  const cancelStravaDetach = () => {
    showStravaDialog = false;
    pendingSubmit = null;
  };
</script>

<form onsubmit={handleSubmit}>
  <!-- Form fields -->
  <Button type="submit">Save Changes</Button>
</form>

<StravaDetachDialog
  bind:open={showStravaDialog}
  onConfirm={confirmStravaDetach}
  onCancel={cancelStravaDetach}
/>
```

### Step 3: Backend Automatically Handles Detachment

When the form is submitted to `PUT /api/v1/rides/edit/{token}`, the backend:

1. Checks if `source='strava'` ✅
2. Calls `detachStravaEvent(tx, eventID)` ✅
3. Sets `source=NULL` and deletes from `strava_event_metadata` ✅
4. Updates the event fields ✅
5. Commits the transaction ✅

**No additional backend work needed!** The detachment is automatic.

### Step 4: Update UI After Detachment

After successful submission, the event will have `source=null` in future fetches:

```typescript
// After successful update
const updatedRide = await fetchRideData(token);
// updatedRide.event.source will now be null
// Warning banner won't show anymore
```

---

## Testing the Implementation

### Manual Test Steps

1. **Import a Strava Event**
   - Use the Strava import flow
   - Verify event has `source='strava'` (check dev tools or API response)

2. **Open Edit Page**
   - Navigate to `/rides/edit?token={edit_token}`
   - Verify amber warning banner appears

3. **Attempt Full Edit** (once full edit form is added)
   - Modify event title, description, or other core fields
   - Click "Save Changes"
   - **Expected:** Confirmation dialog appears
   - Verify dialog text explains detachment

4. **Cancel Edit**
   - Click "Cancel" in dialog
   - **Expected:** Form not submitted, event still has `source='strava'`

5. **Confirm Edit**
   - Try editing again
   - Click "Continue with Edit" in dialog
   - **Expected:** Form submits successfully
   - Verify success message shows

6. **Verify Detachment**
   - Refresh the page or fetch event data again
   - **Expected:** `source` field is now `null` or empty
   - **Expected:** Warning banner no longer appears
   - **Expected:** Event is still visible and fully editable

7. **Verify Sync Skips Detached Event**
   - Run sync service: `./cmd/strava-sync/test_sync.sh --use-real-key --force`
   - **Expected:** Event not mentioned in sync logs (filtered by `source='strava'`)

### Automated Tests

Add Playwright/Vitest tests:

```typescript
// tests/strava-detach.test.ts
import { test, expect } from '@playwright/test';

test('shows Strava warning banner for Strava events', async ({ page }) => {
  await page.goto('/rides/edit?token=strava_event_token');

  // Check for warning banner
  const warning = page.locator('text=Imported from Strava');
  await expect(warning).toBeVisible();
});

test('shows confirmation dialog when editing Strava event', async ({ page }) => {
  await page.goto('/rides/edit?token=strava_event_token');

  // Modify event
  await page.fill('[name="title"]', 'Updated Title');
  await page.click('button[type="submit"]');

  // Check for confirmation dialog
  const dialog = page.locator('text=Disconnect from Strava?');
  await expect(dialog).toBeVisible();
});

test('detaches event when user confirms', async ({ page }) => {
  // ... implementation
});
```

---

## Backend Integration Reference

The backend detachment logic is in `backend/internal/api/ride/repo.go`:

```go
// Lines 156-169: Check source and detach
var source sql.NullString
err = tx.QueryRow(`SELECT id, source FROM events WHERE edit_token = ?`, token).Scan(&eventID, &source)

if source.Valid && source.String == "strava" {
    if err := r.detachStravaEvent(tx, eventID); err != nil {
        return fmt.Errorf("failed to detach strava event: %w", err)
    }
    slog.Info("detached_strava_event_on_edit", "event_id", eventID)
}

// Lines 233-250: Detachment implementation
func (r *Repository) detachStravaEvent(tx *sql.Tx, eventID int64) error {
    // Set source=NULL, source_id=NULL
    // Delete from strava_event_metadata
    // Returns error if fails (transaction will rollback)
}
```

**Important:** The backend uses `sql.NullString` to handle NULL values correctly. The frontend just checks if `source === 'strava'`.

---

## Design Decisions

From `docs/strava/BACKGROUND_SYNC.md`:

**Decision:** When an organizer edits an event via CycleScene's magic link:
- Change `source` from "strava" to NULL (detach)
- Delete row from `strava_event_metadata` (stops sync)
- Event becomes a native CycleScene event with full edit control

**Rationale:** Complies with Strava's "no modification" rule. Organizers get import convenience + full control when needed.

---

## Summary Checklist

- [x] Backend detachment logic implemented (Phase 1)
- [x] API returns `source` field (Phase 2)
- [x] Frontend warning banner added (Phase 2)
- [ ] Full event edit form (future work)
- [ ] Confirmation dialog for full edits (future work)
- [ ] End-to-end tests (future work)

---

## Questions?

See:
- Backend implementation: `backend/internal/api/ride/repo.go`
- Current frontend warning: `frontends/form/src/routes/rides/edit/+page.svelte`
- Design documentation: `docs/strava/BACKGROUND_SYNC.md`
