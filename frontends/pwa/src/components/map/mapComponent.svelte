<script lang="ts">
  import { Map } from "maplibre-gl";
  import { CustomControl, GeoJSONSource, MapLibre } from "svelte-maplibre-gl";
  import { STARTING_LAT, STARTING_LON } from "$lib/config";
  import {
    mapStore,
    rideGeoJSON,
    STARTING_ZOOM,
    TILE_URLS,
    validRides,
  } from "$lib/stores";
  import { mode } from "mode-watcher";
  import type { RideData } from "$lib/types";
  import RideLayers from "./rideLayers.svelte";
  import RecenterButton from "./recenterButton.svelte";

  const SOURCE_ID = "ride-source";
  const ICON_NAME = "custom-circle-pin";
  const GEOAPIFY_API_URL =
    "https://api.geoapify.com/v2/icon/?type=awesome&color=%23ff0000&size=42&icon=bicycle&contentSize=15&strokeColor=%23ff0000&shadowColor=%23ff0000&contentColor=%23ffffff&noShadow&noWhiteCircle&scaleFactor=2&apiKey=";
  const GEOAPIFY_API_KEY = "d4d9d0642bfc40488a64cd3b43b4a63e";
  const CUSTOM_ICON = GEOAPIFY_API_URL + GEOAPIFY_API_KEY;
  // Props
  const { rides, otherRides }: { rides: RideData[]; otherRides: RideData[] } =
    $props();

  // State
  let mapInstance: Map | undefined = $state(undefined);
  let iconLoaded = $state(false);

  $effect(() => {
    mapStore.setPrimaryRides(rides);
  });

  $effect(() => {
    if (mapInstance && $validRides) {
      mapStore.fitMap(mapInstance);
    }
  });

  $effect(() => {
    if (mapInstance && !iconLoaded) {
      async function loadCustomIcon() {
        try {
          const response = await mapInstance!.loadImage(CUSTOM_ICON);
          console.log(CUSTOM_ICON);
          mapInstance!.addImage(ICON_NAME, response.data);
          iconLoaded = true;
        } catch (error) {
          console.error("failed to load custom icon: ", error);
        }
      }

      loadCustomIcon();
    }
  });

  let source = $derived(TILE_URLS[mode.current]);
</script>

<div class="map-container">
  <MapLibre
    bind:map={mapInstance}
    class="w-full h-[calc(100vh-115px)] mt-[60px] mb-[55px]"
    center={{ lat: STARTING_LAT, lng: STARTING_LON }}
    zoom={STARTING_ZOOM}
    style={source}
    attributionControl={false}
  >
    <GeoJSONSource data={$rideGeoJSON} id={SOURCE_ID} />
    {#if iconLoaded}
      <RideLayers sourceId={SOURCE_ID} iconName={ICON_NAME} />
    {/if}

    {#if mapInstance}
      <CustomControl>
        <RecenterButton map={mapInstance} />
      </CustomControl>
    {/if}
  </MapLibre>
</div>

<style>
  .map-container {
    height: calc(100% - 115px);
    width: 100%;
    margin-top: 60px;
    margin-bottom: 50px;
  }
</style>
