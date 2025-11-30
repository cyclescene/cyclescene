<script lang="ts">
  import { currentRoute } from "$lib/stores";
  import ElevationGraph from "../charts/ElevationGraph.svelte";

  const { ride } = $props();
</script>

<div class="space-y-6 p-4 sm:p-6">
  {#if $currentRoute}
    <div class="space-y-4">
      <!-- Route Distance Information -->
      <div class="grid grid-cols-2 gap-4">
        <div class="rounded-lg border border-border bg-card p-4">
          <p class="text-sm text-muted-foreground">Distance (km)</p>
          <p class="text-2xl font-semibold">
            {$currentRoute.properties.distance_km.toFixed(2)}
          </p>
        </div>
        <div class="rounded-lg border border-border bg-card p-4">
          <p class="text-sm text-muted-foreground">Distance (miles)</p>
          <p class="text-2xl font-semibold">
            {$currentRoute.properties.distance_mi.toFixed(2)}
          </p>
        </div>
      </div>

      <!-- Elevation Graph -->
      <div class="space-y-2">
        <h3 class="text-lg font-semibold">Elevation Profile</h3>
        <ElevationGraph coordinates={$currentRoute.geometry.coordinates} />
      </div>
    </div>
  {:else}
    <div class="rounded-lg border border-dashed border-border bg-muted/50 p-6 text-center">
      <p class="text-sm text-muted-foreground">
        No route information available for this ride
      </p>
    </div>
  {/if}
</div>

<style>
  :global(.elevation-graph) {
    width: 100%;
    height: 300px;
  }
</style>
