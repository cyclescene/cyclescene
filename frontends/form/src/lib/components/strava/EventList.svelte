<script lang="ts">
  import EventCard from "./EventCard.svelte";
  import type {
    StravaGroupEvent,
    EventImportConfig,
  } from "$lib/types/strava";

  interface Props {
    events: StravaGroupEvent[];
    selectedEvents: Map<string, EventImportConfig>;
    cityCode: string;
    onToggle: (eventId: string, selected: boolean) => void;
    onConfigChange: (eventId: string, config: EventImportConfig) => void;
  }

  let {
    events,
    selectedEvents,
    cityCode,
    onToggle,
    onConfigChange,
  }: Props = $props();
</script>

{#if events.length === 0}
  <div class="text-muted-foreground py-8 text-center">
    <p>No upcoming events in this club.</p>
  </div>
{:else}
  <div class="space-y-3">
    {#each events as event (event.id)}
      <EventCard
        {event}
        selected={selectedEvents.has(event.id)}
        config={selectedEvents.get(event.id) ?? null}
        {cityCode}
        onToggle={(selected) => onToggle(event.id, selected)}
        onConfigChange={(config) => onConfigChange(event.id, config)}
      />
    {/each}
  </div>
{/if}
