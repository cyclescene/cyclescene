<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import { currentRideStore, mapStore } from "$lib/stores";
  import type { Map } from "maplibre-gl";
  import ZoomOutIcon from "~icons/material-symbols/zoom-out-map-rounded";

  const {
    map,
    highlighted = false,
    onRecentered,
  }: {
    map: Map;
    highlighted?: boolean;
    onRecentered?: () => void;
  } = $props();

  function handleRecenter() {
    if (map) {
      currentRideStore.clearRide();
      mapStore.showCurrentRide(false);
      mapStore.fitMap(map, true);
      window.setTimeout(() => onRecentered?.(), 1100);
    }
  }
</script>

<div
  class="absolute top-5 right-5 z-[1000] flex items-center justify-center"
>
  <Button
    disabled={false}
    class={`h-10 w-10 shadow-md transition-all ${
      highlighted
        ? "bg-yellow-400 text-black ring-4 ring-yellow-400/30 scale-110"
        : "bg-background text-secondary-foreground"
    }`}
    variant="primary"
    onclick={handleRecenter}
  >
    <ZoomOutIcon style="height: 28px; width: 28px;" />
  </Button>
</div>
