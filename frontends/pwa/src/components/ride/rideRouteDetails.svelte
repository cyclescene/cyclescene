<script lang="ts">
  import { currentRoute } from "$lib/stores";
  import ElevationGraph from "../charts/ElevationGraph.svelte";

  const { ride } = $props();
</script>

{#if $currentRoute}
  <div class="space-y-6">
    <!-- Route Distance Information -->
    <div class="grid grid-cols-2 gap-4">
      <div class="rounded-lg border border-border bg-muted/50 p-4">
        <p class="text-sm text-muted-foreground">Distance (km)</p>
        <p class="text-2xl font-semibold">
          {$currentRoute.geojson.properties.distance_km.toFixed(2)}
        </p>
      </div>
      <div class="rounded-lg border border-border bg-muted/50 p-4">
        <p class="text-sm text-muted-foreground">Distance (miles)</p>
        <p class="text-2xl font-semibold">
          {$currentRoute.geojson.properties.distance_mi.toFixed(2)}
        </p>
      </div>
    </div>

    <!-- Elevation Graph -->
    <div class="space-y-2">
      <h3 class="text-lg font-semibold">Elevation Profile</h3>
      <ElevationGraph coordinates={$currentRoute.geojson.geometry.coordinates} />
    </div>
  </div>
{/if}
