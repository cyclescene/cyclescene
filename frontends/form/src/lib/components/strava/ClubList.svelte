<script lang="ts">
  import { SvelteSet, SvelteMap } from "svelte/reactivity";
  import * as Accordion from "$lib/components/ui/accordion";
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

  // Handle accordion value change
  function handleValueChange(value: string[]) {
    expandedClubs = value;

    // Load events for newly expanded clubs
    for (const clubIdStr of value) {
      const clubId = parseInt(clubIdStr, 10);
      if (!clubEvents.has(clubId) && !loadingClubs.has(clubId)) {
        loadClubEvents(clubId);
      }
    }
  }

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
  <Accordion.Root type="multiple" value={expandedClubs} onValueChange={handleValueChange}>
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
            <div class="flex items-center justify-center py-8">
              <svg
                class="text-muted-foreground h-6 w-6 animate-spin"
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
              <span class="text-muted-foreground ml-2">Loading events...</span>
            </div>
          {:else if clubErrors.has(club.id)}
            <div class="rounded-md bg-red-50 p-4 text-center text-red-600">
              <p>{clubErrors.get(club.id)}</p>
              <button
                class="mt-2 text-sm underline"
                onclick={() => loadClubEvents(club.id)}
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
