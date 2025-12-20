<script lang="ts">
  import { Map, type MapLayerMouseEvent } from "maplibre-gl";
  import { GeoJSONSource, MapLibre } from "svelte-maplibre-gl";
  import {
    currentRideStore,
    mapStore,
    rideGeoJSON,
    STARTING_ZOOM,
    TILE_URLS,
    todaysRides,
  } from "$lib/stores";
  import { mode } from "mode-watcher";
  import RideLayers from "./rideLayers.svelte";
  import ParkLayer from "./parkLayer.svelte";
  import RecenterButton from "./recenterButton.svelte";
  import LocationCards from "../locationCards.svelte";
  import RidesNotShown from "../ride/ridesNotShown.svelte";
  import { loadAllMarkersForCity } from "$lib/markers";
  import { CITY_CODE, STARTING_LAT, STARTING_LNG } from "$lib/config";
  import type { RideData } from "$lib/types";

  const SOURCE_ID = "ride-source";
  const ICON_NAME = "custom-bike-pin";
  const BIKE_PIN_ICON_PATH = "/bike-pin-icon.png";

  let mapInstance: Map | undefined = $state(undefined);
  let iconLoaded = $state(false);
  let groupMarkersLoaded = $state(false);
  let groupMarkers: Record<string, string> = $state({});
  let source = $derived(TILE_URLS[mode.current as keyof typeof TILE_URLS]);

  function navigateToRide(ride: RideData) {
    currentRideStore.setRide(ride);
    mapStore.showCurrentRide(true);
    mapStore.showNoLocationsRides(false);
    if (mapInstance) {
      mapStore.flyToSelected(mapInstance);
    }
  }

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

  // Fit map bounds to rides when they load
  // Only triggers when rides exist and change, preserving user's manual pan/zoom during background syncs
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
          const response = await mapInstance!.loadImage(BIKE_PIN_ICON_PATH);
          mapInstance!.addImage(ICON_NAME, response.data);
          iconLoaded = true;
        } catch (error) {
          iconLoaded = true; // Set to true anyway so layers render
        }
      }

      loadCustomIcon();
    }
  });

  $effect(() => {
    if (mapInstance && !groupMarkersLoaded) {
      async function loadGroupMarkers() {
        try {
          const markers = await loadAllMarkersForCity(CITY_CODE);
          groupMarkers = markers;

          // Add each marker image to the map with the group-marker- prefix
          for (const [markerKey, markerDataUrl] of Object.entries(markers)) {
            try {
              const response = await mapInstance!.loadImage(markerDataUrl);
              const imageName = `group-marker-${markerKey}`;
              mapInstance!.addImage(imageName, response.data);
            } catch (error) {
              // Fallback to bike pin icon if marker fails to load
              try {
                const bikeIconResponse = await mapInstance!.loadImage(BIKE_PIN_ICON_PATH);
                const imageName = `group-marker-${markerKey}`;
                mapInstance!.addImage(imageName, bikeIconResponse.data);
              } catch (fallbackError) {
                // Continue anyway if fallback also fails
              }
            }
          }

          groupMarkersLoaded = true;
        } catch (error) {
          // If loading markers fails entirely, mark as loaded anyway so rides display with default icon
          groupMarkersLoaded = true;
        }
      }

      loadGroupMarkers();
    }
  });
</script>

<div
  style="height: calc(100dvh - var(--header-height) - (var(--footer-height) + 35px)); width: 100%;"
>
  <MapLibre
    bind:map={mapInstance}
    class="w-full h-full"
    style={source}
    onclick={handleMapClick}
    attributionControl={false}
    center={[STARTING_LNG, STARTING_LAT]}
    zoom={STARTING_ZOOM}
  >
    {#if $rideGeoJSON && iconLoaded && groupMarkersLoaded}
      <GeoJSONSource data={$rideGeoJSON} id={SOURCE_ID}>
        <RideLayers
          sourceId={SOURCE_ID}
          defaultIconName={ICON_NAME}
          onRideClick={handleRideClick}
        />
      </GeoJSONSource>
    {/if}

    {#if mapInstance}
      <RecenterButton map={mapInstance} />
    {/if}
    <!-- <SpecialEventLayers isDarkMode={mode.current === "dark"} /> -->
    <ParkLayer isDarkMode={mode.current === "dark"} />

    <RidesNotShown />
    <LocationCards onNavigateToRide={navigateToRide} />
  </MapLibre>
</div>
