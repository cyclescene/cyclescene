<script lang="ts">
  import { currentRoute } from "$lib/stores";
  import ElevationGraph from "../charts/ElevationGraph.svelte";
  import LinkIcon from "~icons/bx/link-external";

  const { ride } = $props();

  const sourceLabel = $derived.by(() => {
    switch ($currentRoute?.source) {
      case "strava":
        return "Strava";
      case "ridewithgps":
        return "Ride with GPS";
      default:
        return $currentRoute?.source ?? "Route source";
    }
  });
</script>

{#if $currentRoute}
  <div class="space-y-6">
    {#if $currentRoute.source_url}
      <a
        href={$currentRoute.source_url}
        target="_blank"
        rel="noopener noreferrer"
        class="flex items-center justify-between gap-3 rounded-lg border border-border bg-muted/50 p-4 transition-colors hover:bg-muted"
      >
        <span>
          <span class="block text-sm text-muted-foreground">Route Source</span>
          <span class="text-lg font-semibold">Open in {sourceLabel}</span>
        </span>
        <LinkIcon height="22" width="22" style="color: orange; min-width: 22px;" />
      </a>
    {/if}

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
