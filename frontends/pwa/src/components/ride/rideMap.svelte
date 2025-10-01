<script>
  import { TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";
  import { Map, TileLayer, Marker, Popup, Tooltip } from "sveaflet";

  export let ride;

  const RIDE_COORDS = [ride.lat?.Float64, ride.lon?.Float64];

  const rideMapOptions = {
    center: RIDE_COORDS,
    zoom: 18,
    zoomControl: false,
    dragging: false,
    boxZoom: false,
    scrollWheelZoom: false,
    doubleClickZoom: false,
    touchZoom: false,
  };

  const markerOptions = {
    permanent: true,
  };

  const tileLayerOptions = {
    attribution:
      "Map tiles by Carto, under CC BY 3.0. Data by OpenStreetMap, under ODbL.",
  };
</script>

<Map options={rideMapOptions}>
  {#if mode.current === "dark"}
    <TileLayer url={TILE_URLS.dark} options={tileLayerOptions} />
  {:else}
    <TileLayer url={TILE_URLS.light} options={tileLayerOptions} />
  {/if}
  <Marker options={markerOptions} latLng={RIDE_COORDS} />
</Map>
