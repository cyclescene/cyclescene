<script lang="ts">
  import * as Card from "$lib/components/ui/card";
  import {
    navigateTo,
    currentRideStore,
    VIEW_RIDE_DETAILS,
    activeView,
    SUB_VIEW_ADULT_ONLY_RIDES,
    SUB_VIEW_FAMILY_FRIENDLY_RIDES,
    SUB_VIEW_COVID_SAFETY_RIDES,
  } from "$lib/stores";
  import type { RideData } from "$lib/types";
  import { cn, formatDate, formatTime } from "$lib/utils";
  import CardLabel from "./cardLabel.svelte";
  import RideLabels from "./ride/rideLabels.svelte";

  let { ride }: { ride: RideData } = $props();

  function onCardClick(ride: RideData) {
    navigateTo(VIEW_RIDE_DETAILS);
    currentRideStore.setRide(ride);
  }
</script>

{#if ride}
  <Card.Root
    tabindex="0"
    role="button"
    class={cn(
      "py-3",
      "cursor-pointer transition-colors hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-2, focus-visible:ring-ring focus-visible:ring-offset-2 gap-1",
    )}
    onclick={() => onCardClick(ride)}
    on:keydown={() => onCardClick(ride)}
  >
    <Card.Header class="">
      <Card.Title class={`${ride.cancelled ? "line-through" : ""} text-xl`}
        >{ride.title.length > 80
          ? ride.title.substring(0, 80) + "..."
          : ride.title}</Card.Title
      >
    </Card.Header>

    <Card.Content>
      {#if ride.newsflash}
        <CardLabel label="newsflash">
          <p class="text-lg">{ride.newsflash}</p>
        </CardLabel>
      {/if}
      <CardLabel label="venue">
        <p class="text-lg">
          {ride.venue.length > 30
            ? ride.venue.substring(0, 30) + "..."
            : ride.venue}
        </p>
      </CardLabel>
      {#if $activeView === SUB_VIEW_ADULT_ONLY_RIDES || $activeView === SUB_VIEW_FAMILY_FRIENDLY_RIDES || $activeView === SUB_VIEW_COVID_SAFETY_RIDES}
        <CardLabel label="startTime">
          <p class="text-lg">
            {formatTime(ride.starttime)} on {formatDate(ride.date)}
          </p>
        </CardLabel>
      {:else}
        <CardLabel label="startTime"
          ><p class="text-lg">{formatTime(ride.starttime)}</p></CardLabel
        >
      {/if}
      <RideLabels {ride} />
    </Card.Content>
  </Card.Root>
{/if}
