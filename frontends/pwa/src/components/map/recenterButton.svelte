<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import { mapStore } from "$lib/stores";
  import type { Map } from "maplibre-gl";
  import ZoomOutIcon from "~icons/material-symbols/zoom-out-map-rounded";

  const { map }: { map: Map } = $props();

  function handleRecenter() {
    if (map) {
      mapStore.clearSelectedRide();
    }
  }
</script>

<div
  class="absolute top-5 right-5 z-[1000] p-3 bg-red-50 h-5 w-5 flex items-center justify-center"
>
  <Button
    disabled={false}
    class={`text-secondary-foreground bg-background h-10 w-10`}
    variant="primary"
    onclick={handleRecenter}
  >
    <ZoomOutIcon style="height: 28px; width: 28px;" />
  </Button>
</div>
