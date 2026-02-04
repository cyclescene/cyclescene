<script lang="ts">
  import { SvelteSet, SvelteMap } from "svelte/reactivity";
  import * as Accordion from "$lib/components/ui/accordion";
  import { Skeleton } from "$lib/components/ui/skeleton";
  import EventList from "./EventList.svelte";
  import { fetchClubEvents } from "$lib/api/strava";
  import { PUBLIC_STRAVA_DEBUG } from "$env/static/public";
  import type {
    StravaClub,
    StravaGroupEvent,
    EventImportConfig,
  } from "$lib/types/strava";

  interface Props {
    clubs: StravaClub[];
    clubEvents: Map<number, StravaGroupEvent[]>;
    selectedEvents: Map<string, EventImportConfig>;
    cityCode: string;
    onEventsLoaded: (clubId: number, events: StravaGroupEvent[]) => void;
    onEventToggle: (eventId: string, selected: boolean) => void;
    onEventConfigChange: (eventId: string, config: EventImportConfig) => void;
  }

  let {
    clubs,
    clubEvents,
    selectedEvents,
    cityCode,
    onEventsLoaded,
    onEventToggle,
    onEventConfigChange,
  }: Props = $props();

  // Track which clubs are currently loading
  let loadingClubs = new SvelteSet<number>();
  let clubErrors = new SvelteMap<number, string>();

  // Track expanded clubs for accordion
  let expandedClubs = $state<string[]>([]);

  // Debug logging
  function debugLog(message: string, data?: unknown) {
    if (PUBLIC_STRAVA_DEBUG === "true") {
      console.log(`[Strava ClubList] ${message}`, data ?? "");
    }
  }

  // Load events for a club when expanded
  async function loadClubEvents(clubId: number) {
    // Skip if already loaded or loading
    if (clubEvents.has(clubId) || loadingClubs.has(clubId)) {
      return;
    }

    debugLog("Loading events for club", { clubId });
    loadingClubs.add(clubId);

    try {
      const events = await fetchClubEvents(clubId);
      debugLog("Events loaded", { clubId, count: events.length });
      onEventsLoaded(clubId, events);

      // Clear any previous error
      clubErrors.delete(clubId);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to load events";
      debugLog("Error loading events", { clubId, error: message });
      clubErrors.set(clubId, message);
    } finally {
      loadingClubs.delete(clubId);
    }
  }

  // Watch for changes to expanded clubs and load events
  $effect(() => {
    // Load events for newly expanded clubs
    for (const clubIdStr of expandedClubs) {
      const clubId = parseInt(clubIdStr, 10);
      if (!clubEvents.has(clubId) && !loadingClubs.has(clubId)) {
        loadClubEvents(clubId);
      }
    }
  });

  // Get event count for a club
  function getEventCount(clubId: number): number {
    const events = clubEvents.get(clubId);
    return events?.length ?? 0;
  }

  // Check if events have been loaded for a club
  function hasLoadedEvents(clubId: number): boolean {
    return clubEvents.has(clubId);
  }
</script>

{#if clubs.length === 0}
  <div class="rounded-xl border-2 border-dashed bg-muted/30 p-12 text-center">
    <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
      <svg class="h-8 w-8 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
      </svg>
    </div>
    <h3 class="mb-2 text-lg font-semibold">No Admin Clubs Found</h3>
    <p class="text-muted-foreground mb-1">You're not an admin or owner of any cycling clubs on Strava.</p>
    <p class="text-sm text-muted-foreground">
      Only club admins can import events to ensure proper authorization.
    </p>
  </div>
{:else}
  <!-- Screen reader live region for loading events -->
  <div class="sr-only" aria-live="polite">
    {#each Array.from(loadingClubs) as clubId}
      {@const club = clubs.find(c => c.id === clubId)}
      {#if club}
        Loading events for {club.name}
      {/if}
    {/each}
  </div>

  <div class="space-y-3">
    {#each clubs as club, i (club.id)}
      {@const isExpanded = expandedClubs.includes(club.id.toString())}
      <div class="rounded-xl border-2 transition-all duration-200 {isExpanded ? 'border-[#FC5200] shadow-md' : 'hover:border-[#FC5200]/40 hover:shadow-sm'} animate-in fade-in slide-in-from-bottom-2" style="animation-delay: {i * 80}ms">
        <button
          type="button"
          class="w-full p-4 text-left transition-colors {isExpanded ? 'bg-[#FC5200]/5' : 'hover:bg-muted/50'}"
          onclick={() => {
            if (isExpanded) {
              expandedClubs = expandedClubs.filter(id => id !== club.id.toString());
            } else {
              expandedClubs = [...expandedClubs, club.id.toString()];
            }
          }}
        >
          <div class="flex items-center gap-4">
            {#if club.profile_medium}
              <img
                src={club.profile_medium}
                alt=""
                class="h-12 w-12 rounded-full object-cover ring-2 ring-[#FC5200]/20"
              />
            {:else}
              <div class="flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-br from-[#FC5200]/20 to-orange-300/20 ring-2 ring-[#FC5200]/20">
                <svg class="h-6 w-6 text-[#FC5200]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
              </div>
            {/if}
            <div class="min-w-0 flex-1">
              <h3 class="truncate font-semibold text-base mb-1 {isExpanded ? 'text-[#FC5200]' : ''}">{club.name}</h3>
              <div class="flex items-center gap-2 text-sm text-muted-foreground">
                {#if hasLoadedEvents(club.id)}
                  <span class="inline-flex items-center gap-1 font-medium text-[#FC5200]">
                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                    {getEventCount(club.id)} event{getEventCount(club.id) !== 1 ? "s" : ""}
                  </span>
                {:else if loadingClubs.has(club.id)}
                  <span class="inline-flex items-center gap-1 font-medium">
                    <svg class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24" aria-hidden="true">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                    </svg>
                    Loading events...
                  </span>
                {:else}
                  <span class="inline-flex items-center gap-1">
                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                    </svg>
                    {club.member_count} members
                  </span>
                  <span class="text-xs">• Click to load events</span>
                {/if}
              </div>
            </div>
            <svg
              class="h-5 w-5 flex-shrink-0 text-muted-foreground transition-transform duration-200 {isExpanded ? 'rotate-180 text-[#FC5200]' : ''}"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </div>
        </button>
        {#if isExpanded}
          <div class="border-t p-4 bg-muted/20 animate-in fade-in slide-in-from-top-2 duration-300">
            {#if loadingClubs.has(club.id)}
              <div class="space-y-3 py-2" aria-busy="true" role="status">
                <span class="sr-only">Loading events</span>
                {#each Array(3) as _, index}
                  <div class="flex gap-3 rounded-lg border-2 bg-background p-3 animate-in fade-in duration-300" style="animation-delay: {index * 100}ms">
                    <Skeleton class="h-5 w-5 rounded" />
                    <div class="flex-1 space-y-2">
                      <Skeleton class="h-4 w-full" />
                      <Skeleton class="h-3 w-3/4" />
                      <Skeleton class="h-3 w-1/2" />
                    </div>
                  </div>
                {/each}
              </div>
            {:else if clubErrors.has(club.id)}
              <div class="rounded-lg border-2 border-red-300 bg-red-50 p-6 text-center" role="alert">
                <svg class="mx-auto mb-3 h-10 w-10 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <p class="mb-3 font-medium text-red-700">{clubErrors.get(club.id)}</p>
                <button
                  class="inline-flex items-center gap-2 rounded-lg border-2 border-red-300 bg-white px-4 py-2 text-sm font-semibold text-red-700 transition-all hover:bg-red-50 hover:border-red-400 disabled:opacity-50 disabled:cursor-not-allowed"
                  onclick={() => loadClubEvents(club.id)}
                  disabled={loadingClubs.has(club.id)}
                  aria-label="Retry loading events for {clubs.find(c => c.id === club.id)?.name || 'this club'}"
                >
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Try again
                </button>
              </div>
            {:else if hasLoadedEvents(club.id)}
              <EventList
                events={clubEvents.get(club.id) ?? []}
                {selectedEvents}
                {cityCode}
                onToggle={onEventToggle}
                onConfigChange={onEventConfigChange}
              />
            {/if}
          </div>
        {/if}
      </div>
    {/each}
  </div>
{/if}
