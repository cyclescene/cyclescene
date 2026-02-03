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
  <div class="text-muted-foreground rounded-lg border p-8 text-center">
    <h3 class="mb-2 text-lg font-medium">No Admin Clubs Found</h3>
    <p>You're not an admin or owner of any cycling clubs on Strava.</p>
    <p class="mt-2 text-sm">
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

  <Accordion.Root type="multiple" bind:value={expandedClubs}>
    {#each clubs as club (club.id)}
      <Accordion.Item value={club.id.toString()} class="border-b">
        <Accordion.Trigger class="py-4 hover:no-underline">
          <div class="flex flex-1 items-center gap-3 text-left">
            {#if club.profile_medium}
              <img
                src={club.profile_medium}
                alt=""
                class="h-10 w-10 rounded-full object-cover"
              />
            {:else}
              <div class="flex h-10 w-10 items-center justify-center rounded-full bg-gray-200">
                <span class="text-lg">🚴</span>
              </div>
            {/if}
            <div class="min-w-0 flex-1">
              <h3 class="truncate font-medium">{club.name}</h3>
              <p class="text-muted-foreground text-sm">
                {#if hasLoadedEvents(club.id)}
                  {getEventCount(club.id)} upcoming event{getEventCount(club.id) !== 1 ? "s" : ""}
                {:else if loadingClubs.has(club.id)}
                  Loading events...
                {:else}
                  {club.member_count} members • Click to load events
                {/if}
              </p>
            </div>
          </div>
        </Accordion.Trigger>
        <Accordion.Content class="pb-4">
          {#if loadingClubs.has(club.id)}
            <div class="space-y-3 py-4" aria-busy="true" role="status">
              <span class="sr-only">Loading events</span>
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
          {:else if clubErrors.has(club.id)}
            <div class="rounded-md bg-red-50 p-4 text-center text-red-600" role="alert">
              <p>{clubErrors.get(club.id)}</p>
              <button
                class="mt-2 text-sm underline disabled:opacity-50 disabled:cursor-not-allowed"
                onclick={() => loadClubEvents(club.id)}
                disabled={loadingClubs.has(club.id)}
                aria-label="Retry loading events for {clubs.find(c => c.id === club.id)?.name || 'this club'}"
              >
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
        </Accordion.Content>
      </Accordion.Item>
    {/each}
  </Accordion.Root>
{/if}
