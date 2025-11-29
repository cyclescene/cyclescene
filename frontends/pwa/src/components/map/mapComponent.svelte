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
  import ParkLayer from "./parkLayer.svelte";
  import SpecialEventLayers from "./specialEventLayers.svelte";
  import RecenterButton from "./recenterButton.svelte";
  import LocationCards from "../locationCards.svelte";
  import RidesNotShown from "../ride/ridesNotShown.svelte";
  import { loadAllMarkersForCity } from "$lib/markers";
  import { CITY_CODE } from "$lib/config";

  const SOURCE_ID = "ride-source";
  const ICON_NAME = "custom-bike-pin";
  const GEOAPIFY_API_URL =
    "https://api.geoapify.com/v2/icon/?type=awesome&color=%23ff0000&size=42&icon=bicycle&contentSize=15&strokeColor=%23ff0000&shadowColor=%23ff0000&contentColor=%23ffffff&noShadow&noWhiteCircle&scaleFactor=2&apiKey=d4d9d0642bfc40488a64cd3b43b4a63e";

  let mapInstance: Map | undefined = $state(undefined);
  let iconLoaded = $state(false);
  let groupMarkersLoaded = $state(false);
  let groupMarkers: Record<string, string> = $state({});
  let source = $derived(TILE_URLS[mode.current as keyof typeof TILE_URLS]);

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
          console.log(`[MapComponent] Loading default icon: ${ICON_NAME}`);
          const response = await mapInstance!.loadImage(GEOAPIFY_API_URL);
          mapInstance!.addImage(ICON_NAME, response.data);
          console.log(`[MapComponent] Successfully added image to map: ${ICON_NAME}`);
          iconLoaded = true;
        } catch (error) {
          console.error("failed to load custom icon: ", error);
        }
      }

      loadCustomIcon();
    }
  });

  $effect(() => {
    if (mapInstance && !groupMarkersLoaded) {
      async function loadGroupMarkers() {
        try {
          console.log(`[MapComponent] Starting to load group markers for city: ${CITY_CODE}`);
          const markers = await loadAllMarkersForCity(CITY_CODE);
          groupMarkers = markers;

          console.log(`[MapComponent] Loaded ${Object.keys(markers).length} group markers from spritesheet`);
          console.log(`[MapComponent] Marker keys:`, Object.keys(markers));

          // Add each marker image to the map with the group-marker- prefix
          for (const [markerKey, markerDataUrl] of Object.entries(markers)) {
            try {
              console.log(`[MapComponent] Loading group marker: group-marker-${markerKey}`);
              const response = await mapInstance!.loadImage(markerDataUrl);
              const imageName = `group-marker-${markerKey}`;
              mapInstance!.addImage(imageName, response.data);
              console.log(`[MapComponent] ✓ Successfully added image to map: ${imageName}`);
            } catch (error) {
              console.error(`[MapComponent] ✗ Failed to load group marker image for ${markerKey}:`, error);
            }
          }

          groupMarkersLoaded = true;
          console.log(`[MapComponent] Group markers loading complete. Total images in map:`, Object.keys(markers).length);
        } catch (error) {
          console.error("[MapComponent] Failed to load group markers: ", error);
          // Continue anyway - rides can still be displayed with default icon
        }
      }

      loadGroupMarkers();
    }
  });
</script>

<div
  style="height: calc(100dvh - var(--header-height) - var(--footer-height)); width: 100%;"
>
  <MapLibre
    bind:map={mapInstance}
    class="w-full h-full"
    style={source}
    onclick={handleMapClick}
    attributionControl={false}
  >
    {#if $rideGeoJSON}
      {#if iconLoaded && groupMarkersLoaded}
        <RideLayers
          sourceId={SOURCE_ID}
          defaultIconName={ICON_NAME}
          onRideClick={handleRideClick}
        />
      {/if}
      <GeoJSONSource data={$rideGeoJSON} id={SOURCE_ID} />
    {/if}

    {#if mapInstance}
      <RecenterButton map={mapInstance} />
    {/if}
    <SpecialEventLayers isDarkMode={mode.current === "dark"} />
    <ParkLayer isDarkMode={mode.current === "dark"} />

    <RidesNotShown />
    <LocationCards />
  </MapLibre>
</div>
