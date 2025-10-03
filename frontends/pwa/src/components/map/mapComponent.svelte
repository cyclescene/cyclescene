<script lang="ts">
  import { Map, type MapLayerMouseEvent } from "maplibre-gl";
  import { GeoJSONSource, MapLibre } from "svelte-maplibre-gl";
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
  import LocationCards from "../locationCards.svelte";

  const SOURCE_ID = "ride-source";
  const ICON_NAME = "custom-bike-pin";
  const GEOAPIFY_API_URL =
    "https://api.geoapify.com/v2/icon/?type=awesome&color=%23ff0000&size=42&icon=bicycle&contentSize=15&strokeColor=%23ff0000&shadowColor=%23ff0000&contentColor=%23ffffff&noShadow&noWhiteCircle&scaleFactor=2&apiKey=d4d9d0642bfc40488a64cd3b43b4a63e";
  // Props
  const { rides, otherRides }: { rides: RideData[]; otherRides: RideData[] } =
    $props();

  // State
  let mapInstance: Map | undefined = $state(undefined);
  let iconLoaded = $state(false);
  let source = $derived(TILE_URLS[mode.current]);

  function handleRideClick(e: MapLayerMouseEvent) {
    if (e.features && e.features.length > 0) {
      const feature = e.features[0];
      const rideId = feature.properties?.id;

      if (rideId) {
        const selectedRide = mapStore.getRideById(rideId);
        if (selectedRide) {
          mapStore.setSelectedRide(selectedRide);
          mapStore.showEventCards(true);
          mapStore.showOtherRides(false);
          if (mapInstance) {
            mapStore.flyToSelected(mapInstance);
          }
        }
      }
    }
  }

  function handleMapClick(e: MapLayerMouseEvent) {
    mapStore.clearSelectedRide();
  }

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
          const response = await mapInstance!.loadImage(GEOAPIFY_API_URL);
          mapInstance!.addImage(ICON_NAME, response.data);
          iconLoaded = true;
        } catch (error) {
          console.error("failed to load custom icon: ", error);
        }
      }

      loadCustomIcon();
    }
  });
</script>

<MapLibre
  bind:map={mapInstance}
  class="w-full h-[calc(100vh-115px)]"
  center={{ lat: STARTING_LAT, lng: STARTING_LON }}
  zoom={STARTING_ZOOM}
  style={source}
  onclick={handleMapClick}
  attributionControl={false}
>
  <GeoJSONSource data={$rideGeoJSON} id={SOURCE_ID} />
  {#if iconLoaded}
    <RideLayers
      sourceId={SOURCE_ID}
      iconName={ICON_NAME}
      onRideClick={handleRideClick}
    />
  {/if}

  {#if mapInstance}
    <RecenterButton map={mapInstance} />
  {/if}

  <LocationCards />
</MapLibre>

<style>
</style>
