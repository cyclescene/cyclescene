<script lang="ts">
  import { MapLibre } from "svelte-maplibre-gl";
  import { STARTING_LAT, STARTING_LON } from "$lib/config";
  import { TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";

  console.log(TILE_URLS);

  let source = $derived(TILE_URLS[mode.current]);

  console.log(source);
</script>

<div class="map-container">
  <MapLibre
    class="w-full h-[calc(100vh-115px)] mt-[60px] mb-[55px]"
    center={{ lat: STARTING_LAT, lng: STARTING_LON }}
    zoom={13}
    style={source}
    attributionControl={false}
  ></MapLibre>
</div>

<style>
  .map-container {
    height: calc(100% - 115px);
    width: 100%;
    margin-top: 60px;
    margin-bottom: 50px;
  }
</style>
