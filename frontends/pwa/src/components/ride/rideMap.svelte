<script lang="ts">
  import { TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";
  import { MapLibre } from "svelte-maplibre-gl";

  const { ride } = $props();

  const RIDE_COORDS = {
    lat: ride.lat?.Float64,
    lon: ride.lon?.Float64,
  };

  let source = $derived(TILE_URLS[mode.current]);
</script>

<MapLibre
  class="h-[55vh] min-h-[300px]"
  center={RIDE_COORDS}
  zoom={18}
  dragPan={false}
  doubleClickZoom={false}
  scrollZoom={false}
  touchZoomRotate={false}
  boxZoom={false}
  attributionControl={false}
  style={source}
></MapLibre>
<!-- <Map options={rideMapOptions}> -->
<!--   {#if mode.current === "dark"} -->
<!--     <TileLayer url={TILE_URLS.dark} options={tileLayerOptions} /> -->
<!--   {:else} -->
<!--     <TileLayer url={TILE_URLS.light} options={tileLayerOptions} /> -->
<!--   {/if} -->
<!--   <Marker options={markerOptions} latLng={RIDE_COORDS} /> -->
<!-- </Map> -->
