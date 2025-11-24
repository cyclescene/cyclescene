<script lang="ts">
  import { Map, type MapLayerMouseEvent } from "maplibre-gl";
  import { GeoJSONSource, MapLibre } from "svelte-maplibre-gl";
  import {
    currentRideStore,
    mapStore,
    rideGeoJSON,
    TILE_URLS,
    todaysRides,
  } from "$lib/stores";
  import { mode } from "mode-watcher";
  import RideLayers from "./rideLayers.svelte";
  import RecenterButton from "./recenterButton.svelte";
  import LocationCards from "../locationCards.svelte";
  import RidesNotShown from "../ride/ridesNotShown.svelte";

  const SOURCE_ID = "ride-source";
  const ICON_NAME = "custom-bike-pin";
  const GEOAPIFY_API_URL =
    "https://api.geoapify.com/v2/icon/?type=awesome&color=%23ff0000&size=42&icon=bicycle&contentSize=15&strokeColor=%23ff0000&shadowColor=%23ff0000&contentColor=%23ffffff&noShadow&noWhiteCircle&scaleFactor=2&apiKey=d4d9d0642bfc40488a64cd3b43b4a63e";

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
          currentRideStore.setRide(selectedRide);
          mapStore.showCurrentRide(true);
          mapStore.showNoLocationsRides(false);
          if (mapInstance) {
            mapStore.flyToSelected(mapInstance);
          }
        }
      }
    }
  }

  function handleMapClick(_: MapLayerMouseEvent) {
    currentRideStore.clearRide();
  }

  $effect(() => {
    if (!mapInstance || $todaysRides.length === 0) {
      return;
    }
    mapStore.fitMap(mapInstance);
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

<div style="height: calc(100vh - var(--header-height) - var(--footer-height)); width: 100%;">
  <MapLibre
    bind:map={mapInstance}
    class="w-full h-full"
    style={source}
    onclick={handleMapClick}
    attributionControl={false}
  >
  {#if $rideGeoJSON}
    {#if iconLoaded}
      <RideLayers
        sourceId={SOURCE_ID}
        iconName={ICON_NAME}
        onRideClick={handleRideClick}
      />
    {/if}
    <GeoJSONSource data={$rideGeoJSON} id={SOURCE_ID} />
  {/if}

  {#if mapInstance}
    <RecenterButton map={mapInstance} />
  {/if}

  <RidesNotShown />
  <LocationCards />
  </MapLibre>
</div>
