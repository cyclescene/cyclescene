<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import { mapStore } from "$lib/stores";
  import type { Map } from "maplibre-gl";
  import RecenterIcon from "~icons/material-symbols-light/recenter-rounded";

  const { map }: { map: Map } = $props();

  function handleRecenter() {
    if (map) {
      mapStore.fitMap(map);
    }
  }
</script>

<Button
  disabled={false}
  class={`text-secondary-foreground bg-background`}
  variant="primary"
  onclick={handleRecenter}
>
  <RecenterIcon style="width: 30px; height: 30px;" />
</Button>
