<script lang="ts">
  import { SymbolLayer } from "svelte-maplibre-gl";
  import { mode } from "mode-watcher";
  import type { MapLayerMouseEvent } from "maplibre-gl";
  import { selectedRideId } from "$lib/stores";

  const ICON_SIZE = 0.8;

  const {
    sourceId,
    defaultIconName,
    onRideClick,
  }: {
    sourceId: string;
    defaultIconName: string;
    onRideClick?: (e: MapLayerMouseEvent) => void;
  } = $props();

  let textColor = $derived(mode.current === "dark" ? "#ffffff" : "#000000");

  let selectedId = $state("");

  $effect(() => {
    selectedId = $selectedRideId;
  });

  $effect(() => {
    console.log(`[RideLayers] Rendered with sourceId: ${sourceId}, defaultIconName: ${defaultIconName}`);
  });
</script>

<!-- Unified marker layer: shows group marker if available, otherwise default icon -->
<SymbolLayer
  id="ride-icons"
  source={sourceId}
  layout={{
    "icon-image": [
      "case",
      ["!=", ["get", "group_marker_icon"], ""],
      ["get", "group_marker_icon"],
      defaultIconName
    ],
    "icon-size": ICON_SIZE,
    "icon-allow-overlap": true,
  }}
  paint={{
    "icon-color": "#0000ff",
  }}
  onclick={onRideClick}
/>
<SymbolLayer
  id="ride-labels"
  source={sourceId}
  layout={{
    "text-field": ["get", "name"],
    "text-font": ["Open Sans Regular"],
    "text-size": 12,
    "text-offset": [0, 2],
    "text-anchor": "top",

    "text-allow-overlap": false,
    "icon-allow-overlap": true,
    "icon-ignore-placement": true,
  }}
  paint={{
    "text-color": textColor,
    "text-halo-color": "#ffffff",
    "text-halo-width": 0.1,
  }}
  onclick={onRideClick}
/>
