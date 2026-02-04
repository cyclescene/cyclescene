<script lang="ts">
  import { onMount } from "svelte";
  import { SvelteMap } from "svelte/reactivity";
  import * as Card from "$lib/components/ui/card";
  import { Button } from "$lib/components/ui/button";
  import { Skeleton } from "$lib/components/ui/skeleton";
  import { PUBLIC_STRAVA_DEBUG } from "$env/static/public";
  import EmailInput from "./EmailInput.svelte";
  import ClubList from "./ClubList.svelte";
  import ImportProgress from "./ImportProgress.svelte";
  import ImportResults from "./ImportResults.svelte";
  import {
    fetchAdminClubs,
    logout,
    checkSessionForImport,
    formatRetryTime,
  } from "$lib/api/strava";
  import {
    RateLimitError,
    SessionExpiredError,
    type ImportStep,
    type StravaClub,
    type StravaGroupEvent,
    type EventImportConfig,
    type ImportResult,
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
  let rateLimitRetryAfter = $state<number | null>(null); // Seconds until rate limit resets
  let sessionExpired = $state(false);

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
    rateLimitRetryAfter = null;
    sessionExpired = false;

    try {
      adminClubs = await fetchAdminClubs();
      debugLog("Clubs loaded", { count: adminClubs.length });
    } catch (err) {
      debugLog("Error loading clubs", err);

      // Handle rate limit errors
      if (err instanceof RateLimitError) {
        rateLimitRetryAfter = err.retry_after_seconds;
        error = `CycleScene is experiencing high Strava import usage. Please try again in ${formatRetryTime(err.retry_after_seconds)}.`;
        // Log to monitoring (console for now)
        console.log(
          JSON.stringify({
            event: "strava_rate_limit_hit",
            retry_after_seconds: err.retry_after_seconds,
            timestamp: new Date().toISOString(),
          })
        );
      }
      // Handle session expired errors
      else if (err instanceof SessionExpiredError) {
        sessionExpired = true;
        error = err.message;
      }
      // Handle other errors
      else {
        const message = err instanceof Error ? err.message : "Failed to load clubs";
        error = message;
      }
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

  // Start import (with session check)
  async function handleStartImport() {
    if (selectedEvents.size === 0) {
      error = "Please select at least one event to import";
      return;
    }

    debugLog("Starting import", { count: selectedEvents.size });
    error = null;

    // Check session validity before starting import
    try {
      const valid = await checkSessionForImport();
      if (!valid) {
        sessionExpired = true;
        error = "Your session expired. Please reconnect to Strava.";
        return;
      }
    } catch (err) {
      if (err instanceof SessionExpiredError) {
        sessionExpired = true;
        error = err.message;
        return;
      }
      // Other errors, proceed anyway (will be caught by WebSocket)
      debugLog("Session check warning", err);
    }

    step = "importing";
  }

  // Handle reconnect for expired sessions
  function handleReconnect() {
    debugLog("Reconnecting after session expiry");
    sessionExpired = false;
    error = null;
    rateLimitRetryAfter = null;
    // Reload clubs to trigger new OAuth
    loadClubs();
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

  // Focus management: Move focus when step changes
  $effect(() => {
    // Small delay to ensure DOM is updated
    setTimeout(() => {
      if (step === "email") {
        document.getElementById("organizer-email")?.focus();
      } else if (step === "select") {
        // Focus first accordion item or instructions
        const firstAccordion = document.querySelector('[data-accordion-item]') as HTMLElement;
        if (firstAccordion) {
          firstAccordion.focus();
        }
      } else if (step === "importing") {
        // Focus progress heading
        document.getElementById("import-progress-heading")?.focus();
      } else if (step === "complete") {
        // Focus results heading
        document.getElementById("results-heading")?.focus();
      }
    }, 100);
  });
</script>

<div class="space-y-6 animate-in fade-in duration-500">
  <!-- Screen reader live region for step announcements -->
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

  <!-- Header -->
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 rounded-xl border-2 bg-gradient-to-br from-background to-muted/30 p-4 sm:p-6 shadow-sm">
    <div class="flex items-center gap-3 sm:gap-4">
      <div class="flex h-12 w-12 sm:h-14 sm:w-14 flex-shrink-0 items-center justify-center rounded-xl bg-[#FC5200] shadow-md">
        <svg class="h-6 w-6 sm:h-7 sm:w-7 text-white" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
          <path d="M15.387 17.944l-2.089-4.116h-3.065L15.387 24l5.15-10.172h-3.066m-7.008-5.599l2.836 5.598h4.172L10.463 0l-7 13.828h4.169" />
        </svg>
      </div>
      <div class="min-w-0">
        <h2 class="text-xl sm:text-2xl font-bold truncate">Import from Strava</h2>
        <p class="text-muted-foreground text-xs sm:text-sm font-medium">
          Import group events from clubs you manage
        </p>
      </div>
    </div>
    {#if step !== "importing"}
      <Button variant="outline" size="lg" onclick={handleDone} class="border-2 w-full sm:w-auto flex-shrink-0">
        <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
        </svg>
        Back to Form
      </Button>
    {/if}
  </div>

  <!-- Error display -->
  {#if error}
    <div class="rounded-xl border-2 border-red-300 bg-red-50 p-5 shadow-sm animate-in fade-in slide-in-from-top-2 duration-300" role="alert">
      <div class="flex items-start gap-3">
        <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-red-500">
          <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="flex-1">
          <p class="font-medium text-red-800 mb-3">{error}</p>
          {#if sessionExpired}
            <Button
              variant="outline"
              size="sm"
              class="border-2 border-red-300 text-red-700 hover:bg-red-100 hover:border-red-400"
              onclick={handleReconnect}
              aria-label="Reconnect to Strava to continue"
            >
              <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              Reconnect to Strava
            </Button>
          {:else if rateLimitRetryAfter}
            <Button
              variant="outline"
              size="sm"
              class="border-2 border-red-300 text-red-700 hover:bg-red-100 hover:border-red-400"
              onclick={loadClubs}
              aria-label="Try loading clubs again"
            >
              <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              Try Again
            </Button>
          {/if}
        </div>
      </div>
    </div>
  {/if}

  <!-- Loading clubs state -->
  {#if isLoadingClubs}
    <div class="space-y-4" aria-busy="true" role="status">
      <span class="sr-only">Loading your clubs</span>
      {#each Array(3) as _, i}
        <Card.Root class="border-2 p-4 animate-in fade-in duration-300" style="animation-delay: {i * 100}ms">
          <div class="flex items-center gap-4">
            <Skeleton class="h-12 w-12 rounded-full" />
            <div class="flex-1 space-y-2">
              <Skeleton class="h-5 w-[200px]" />
              <Skeleton class="h-4 w-[150px]" />
            </div>
            <Skeleton class="h-5 w-5 rounded" />
          </div>
        </Card.Root>
      {/each}
    </div>
  {:else if step === "email"}
    <!-- Email Input Step -->
    <EmailInput onSubmit={handleEmailSubmit} onBack={handleEmailBack} city={city} />
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
      <div class="sticky bottom-0 rounded-xl border-2 bg-gradient-to-br from-background to-muted/30 p-4 sm:p-5 shadow-lg backdrop-blur-sm">
        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 sm:gap-4">
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-muted-foreground" aria-live="polite">
              {selectedCount} event{selectedCount !== 1 ? "s" : ""} selected
            </p>
            {#if selectedCount > 0}
              <p class="text-xs text-muted-foreground mt-0.5">
                Ready to import to CycleScene
              </p>
            {/if}
          </div>
          <Button
            size="lg"
            onclick={handleStartImport}
            disabled={selectedCount === 0}
            aria-label="Start importing {selectedCount} selected event{selectedCount !== 1 ? 's' : ''}"
            class="bg-[#FC5200] hover:bg-[#E04A00] text-white shadow-sm hover:shadow-md transition-all duration-200 group disabled:opacity-50 w-full sm:w-auto flex-shrink-0"
          >
            <svg class="h-5 w-5 mr-2 group-hover:scale-110 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10" />
            </svg>
            <span class="truncate">Import {selectedCount} Event{selectedCount !== 1 ? "s" : ""}</span>
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
  <div class="flex items-center justify-center mt-8 -mx-6 -mb-6 py-6 bg-[#2D2D32] rounded-b-xl">
    <img
      src="/api_logo_pwrdBy_strava_horiz_white.png"
      alt="Powered by Strava"
      class="h-5"
    />
  </div>
</div>
