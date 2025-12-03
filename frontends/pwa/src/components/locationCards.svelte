<script lang="ts">
  import { ScrollArea } from "$lib/components/ui/scroll-area/index";
  import { currentRide, mapStore } from "$lib/stores";
  import type { RideData } from "$lib/types";
  import Card from "./card.svelte";

  let ride = $state<RideData | null>();

  $effect(() => {
    ride = $currentRide;
  });
</script>

{#if $mapStore.showCurrentRide && $currentRide}
  <div class="location-card-container">
    <Card {ride} />
  </div>
{/if}

<style>
  .location-card-container {
    position: absolute;
    bottom: calc(var(--footer-height) + 35px + 15px);
    left: 0;
    width: 100%;
    background-color: transparent;
    padding: 0 1.25rem;
    z-index: 1000;
    max-height: 20vh;
  }
</style>
