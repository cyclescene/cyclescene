<script lang="ts">
  import { TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";
  import {
    GlobeControl,
    MapLibre,
    NavigationControl,
    ScaleControl,
  } from "svelte-maplibre-gl";

  export let ride;

  const RIDE_COORDS = {
    lat: ride.lat?.Float64,
    lon: ride.lon?.Float64,
  };

  console.log(`${TILE_URLS.dark}`);
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
  style="https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json"
></MapLibre>
<!-- <Map options={rideMapOptions}> -->
<!--   {#if mode.current === "dark"} -->
<!--     <TileLayer url={TILE_URLS.dark} options={tileLayerOptions} /> -->
<!--   {:else} -->
<!--     <TileLayer url={TILE_URLS.light} options={tileLayerOptions} /> -->
<!--   {/if} -->
<!--   <Marker options={markerOptions} latLng={RIDE_COORDS} /> -->
<!-- </Map> -->
