<script>
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
  import { cn, formatDate, formatTime } from "$lib/utils";
  import RideLabels from "./ride/rideLabels.svelte";

  export let ride;

  function onCardClick(ride) {
    navigateTo(VIEW_RIDE_DETAILS);
    currentRideStore.setRide(ride);
  }
</script>

{#if ride}
  <Card.Root
    tabindex="0"
    role="button"
    class={cn(
      "mb-2 last:mb-0",
      "cursor-pointer transition-colors hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-2, focus-visible:ring-ring focus-visible:ring-offset-2 gap-1",
    )}
    onclick={() => onCardClick(ride)}
    on:keydown={(e) => onCardClick(ride)}
  >
    <Card.Header class="">
      <Card.Title class={`${ride.cancelled ? "line-through" : ""} text-xl`}
        >{ride.title}</Card.Title
      >
    </Card.Header>

    <Card.Content>
      <p>{ride.newsflash.String}</p>
      <p>{ride.venue.String}</p>
      <div class="flex flex-row gap-1">
        <p>{formatTime(ride.starttime)}</p>
        {#if $activeView === SUB_VIEW_ADULT_ONLY_RIDES || $activeView === SUB_VIEW_FAMILY_FRIENDLY_RIDES || $activeView === SUB_VIEW_COVID_SAFETY_RIDES}
          <p>
            <strong>
              {formatDate(ride.date)}
            </strong>
          </p>
        {/if}
      </div>
      <br />
      <RideLabels {ride} />
    </Card.Content>
  </Card.Root>
{/if}
