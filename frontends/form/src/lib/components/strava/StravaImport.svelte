<script lang="ts">
  import { onMount } from "svelte";
  import { SvelteMap } from "svelte/reactivity";
  import * as Card from "$lib/components/ui/card";
  import { Button } from "$lib/components/ui/button";
  import { PUBLIC_STRAVA_DEBUG } from "$env/static/public";
  import EmailInput from "./EmailInput.svelte";
  import ClubList from "./ClubList.svelte";
  import ImportProgress from "./ImportProgress.svelte";
  import ImportResults from "./ImportResults.svelte";
  import { fetchAdminClubs, logout } from "$lib/api/strava";
  import type {
    ImportStep,
    StravaClub,
    StravaGroupEvent,
    EventImportConfig,
    ImportResult,
  } from "$lib/types/strava";

  interface Props {
    city: string;
    onClose: () => void;
  }

  let { city, onClose }: Props = $props();

  // Current step in the import flow
  let step = $state<ImportStep>("email");

  // Data state
  let organizerEmail = $state("");
  let groupCode = $state("");
  let adminClubs = $state<StravaClub[]>([]);
  let clubEvents = new SvelteMap<number, StravaGroupEvent[]>();
  let selectedEvents = new SvelteMap<string, EventImportConfig>();
  let importResults = $state<ImportResult[]>([]);

  // Loading/error states
  let isLoadingClubs = $state(false);
  let error = $state<string | null>(null);

  // Debug logging
  function debugLog(message: string, data?: unknown) {
    if (PUBLIC_STRAVA_DEBUG === "true") {
      console.log(`[Strava Import] ${message}`, data ?? "");
    }
  }

  // Load admin clubs on mount
  onMount(async () => {
    debugLog("Component mounted, loading clubs");
    await loadClubs();
  });

  // Load admin clubs
  async function loadClubs() {
    isLoadingClubs = true;
    error = null;

    try {
      adminClubs = await fetchAdminClubs();
      debugLog("Clubs loaded", { count: adminClubs.length });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to load clubs";
      debugLog("Error loading clubs", message);
      error = message;
    } finally {
      isLoadingClubs = false;
    }
  }

  // Handle email submission (now includes optional group code)
  function handleEmailSubmit(email: string, group: string) {
    debugLog("Email submitted", { email, groupCode: group });
    organizerEmail = email;
    groupCode = group;
    step = "select";
  }

  // Handle back from email step
  function handleEmailBack() {
    onClose();
  }

  // Handle events loaded for a club
  function handleEventsLoaded(clubId: number, events: StravaGroupEvent[]) {
    debugLog("Events loaded", { clubId, count: events.length });
    clubEvents.set(clubId, events);
  }

  // Handle event toggle
  function handleEventToggle(eventId: string, selected: boolean) {
    debugLog("Event toggled", { eventId, selected });

    if (selected) {
      // Find the event to get club_id
      for (const [, events] of clubEvents) {
        const event = events.find((e) => e.id === eventId);
        if (event) {
          selectedEvents.set(eventId, {
            strava_event_id: eventId,
            club_id: event.club_id,
          });
          break;
        }
      }
    } else {
      selectedEvents.delete(eventId);
    }
  }

  // Handle event config change
  function handleEventConfigChange(eventId: string, config: EventImportConfig) {
    debugLog("Event config changed", { eventId, config });
    selectedEvents.set(eventId, config);
  }

  // Start import
  function handleStartImport() {
    if (selectedEvents.size === 0) {
      error = "Please select at least one event to import";
      return;
    }

    debugLog("Starting import", { count: selectedEvents.size });
    error = null;
    step = "importing";
  }

  // Handle import completion
  function handleImportComplete(results: ImportResult[]) {
    debugLog("Import complete", { results });
    importResults = results;
    step = "complete";
  }

  // Handle import error
  function handleImportError(errorMessage: string) {
    debugLog("Import error", errorMessage);
    error = errorMessage;
    step = "select"; // Go back to selection
  }

  // Handle "Import More"
  function handleImportMore() {
    // Reset selection state but keep clubs/events loaded
    selectedEvents.clear(); // Clear the SvelteMap instead of replacing
    importResults = [];
    error = null;
    step = "select";
  }

  // Handle "Done" or close
  async function handleDone() {
    // Logout from Strava session
    try {
      await logout();
    } catch {
      // Ignore logout errors
    }
    onClose();
  }

  // Get event titles map for progress display
  function getEventTitles(): Map<string, string> {
    const titles = new Map<string, string>();
    for (const [, events] of clubEvents) {
      for (const event of events) {
        titles.set(event.id, event.title);
      }
    }
    return titles;
  }

  // Get selected events as array
  function getSelectedEventsArray(): EventImportConfig[] {
    return Array.from(selectedEvents.values());
  }

  // Count selected events
  let selectedCount = $derived(selectedEvents.size);
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h2 class="text-2xl font-semibold">Import from Strava</h2>
      <p class="text-muted-foreground text-sm">
        Import group events from clubs you manage
      </p>
    </div>
    {#if step !== "importing"}
      <Button variant="outline" onclick={handleDone}>
        ← Back to Form
      </Button>
    {/if}
  </div>

  <!-- Error display -->
  {#if error}
    <div class="rounded-lg border border-red-200 bg-red-50 p-4 text-red-700">
      <p>{error}</p>
    </div>
  {/if}

  <!-- Loading clubs state -->
  {#if isLoadingClubs}
    <Card.Root class="p-8">
      <div class="flex items-center justify-center">
        <svg
          class="text-muted-foreground h-8 w-8 animate-spin"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          />
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
        <span class="text-muted-foreground ml-3">Loading your clubs...</span>
      </div>
    </Card.Root>
  {:else if step === "email"}
    <!-- Email Input Step -->
    <EmailInput onSubmit={handleEmailSubmit} onBack={handleEmailBack} />
  {:else if step === "select"}
    <!-- Event Selection Step -->
    <div class="space-y-6">
      <ClubList
        clubs={adminClubs}
        {clubEvents}
        {selectedEvents}
        cityCode={city}
        onEventsLoaded={handleEventsLoaded}
        onEventToggle={handleEventToggle}
        onEventConfigChange={handleEventConfigChange}
      />

      <!-- Import button footer -->
      <div class="sticky bottom-0 border-t bg-white py-4">
        <div class="flex items-center justify-between">
          <p class="text-muted-foreground text-sm">
            {selectedCount} event{selectedCount !== 1 ? "s" : ""} selected
          </p>
          <Button
            onclick={handleStartImport}
            disabled={selectedCount === 0}
          >
            Import {selectedCount} Event{selectedCount !== 1 ? "s" : ""}
          </Button>
        </div>
      </div>
    </div>
  {:else if step === "importing"}
    <!-- Import Progress Step -->
    <ImportProgress
      {organizerEmail}
      {groupCode}
      events={getSelectedEventsArray()}
      eventTitles={getEventTitles()}
      onComplete={handleImportComplete}
      onError={handleImportError}
    />
  {:else if step === "complete"}
    <!-- Results Step -->
    <ImportResults
      results={importResults}
      {organizerEmail}
      onImportMore={handleImportMore}
      onDone={handleDone}
    />
  {/if}

  <!-- Strava Attribution (per Strava Brand Guidelines) -->
  <p class="text-muted-foreground mt-6 text-center text-xs">
    Powered by Strava
  </p>
</div>
