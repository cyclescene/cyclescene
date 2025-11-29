<script lang="ts">
  import { SymbolLayer } from "svelte-maplibre-gl";
  import { mode } from "mode-watcher";
  import type { MapLayerMouseEvent } from "maplibre-gl";
  import { selectedRideId } from "$lib/stores";

  const DEFAULT_SIZE = 1;
  const SELECTED_SIZE = 1.2;

  const {
    sourceId,
    onRideClick,
  }: {
    sourceId: string;
    onRideClick?: (e: MapLayerMouseEvent) => void;
  } = $props();

  let selectedId = $state("");

  $effect(() => {
    selectedId = $selectedRideId;
  });
</script>

<!-- Group marker icons -->
<SymbolLayer
  id="group-marker-icons"
  source={sourceId}
  layout={{
    "icon-image": ["get", "group_marker_icon"],
    "icon-size": [
      "match",
      ["to-string", ["get", "id"]],
      selectedId,
      SELECTED_SIZE,
      DEFAULT_SIZE,
    ],
    "icon-allow-overlap": true,
  }}
  paint={{}}
  filter={["!=", ["get", "group_marker_icon"], ""]}
  onclick={onRideClick}
/>
