<script lang="ts">
  import { singleRideGeoJSON, TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";
  import { GeoJSONSource, MapLibre } from "svelte-maplibre-gl";
  import RideLayers from "../map/rideLayers.svelte";
  import { Map } from "maplibre-gl";
  import { loadMarkerByKey } from "$lib/markers";
  import { CITY_CODE } from "$lib/config";
  import ParkLayer from "../map/parkLayer.svelte";

  const { ride } = $props();

  const RIDE_COORDS = {
    lat: ride.lat,
    lng: ride.lng,
  };

  const SOURCE_ID = "single-ride-source";
  const ICON_NAME = "custom-bike-pin";
  const GEOAPIFY_API_URL =
    "https://api.geoapify.com/v2/icon/?type=awesome&color=%23ff0000&size=42&icon=bicycle&contentSize=15&strokeColor=%23ff0000&shadowColor=%23ff0000&contentColor=%23ffffff&noShadow&noWhiteCircle&scaleFactor=2&apiKey=d4d9d0642bfc40488a64cd3b43b4a63e";

  let map: Map | undefined = $state.raw();
  let source = $derived(TILE_URLS[mode.current as keyof typeof TILE_URLS]);
  let iconLoaded = $state(false);
  let groupMarkerLoaded = $state(false);

  // Reset icon states when ride changes
  $effect(() => {
    if (!ride) {
      iconLoaded = false;
      groupMarkerLoaded = false;
    }
  });

  $effect(() => {
    if (map && !iconLoaded) {
      async function loadDefaultIcon() {
        try {
          const response = await map!.loadImage(GEOAPIFY_API_URL);
          map!.addImage(ICON_NAME, response.data);
          const addedImage = map!.getImage(ICON_NAME);
          if (!addedImage) {
            console.warn(
              `[RideMap] ⚠ Default icon ${ICON_NAME} not found after adding`,
            );
          }
          iconLoaded = true;
        } catch (error) {
          console.error("[RideMap] Failed to load default icon: ", error);
          iconLoaded = true; // Set to true anyway so layers render
        }
      }

      loadDefaultIcon();
    }
  });

  $effect(() => {
    if (map && !groupMarkerLoaded) {
      async function loadGroupMarker() {
        try {
          if (!ride?.group_marker) {
            groupMarkerLoaded = true;
            return;
          }

          const markerDataUrl = await loadMarkerByKey(
            CITY_CODE,
            ride.group_marker,
          );
          const markerResponse = await map!.loadImage(markerDataUrl);
          const imageName = `group-marker-${ride.group_marker}`;
          map!.addImage(imageName, markerResponse.data);

          const addedImage = map!.getImage(imageName);
          if (!addedImage) {
            console.warn(
              `[RideMap] ⚠ Group marker ${imageName} not found after adding`,
            );
          }

          groupMarkerLoaded = true;
        } catch (error) {
          console.error("[RideMap] Failed to load group marker: ", error);
          groupMarkerLoaded = true; // Continue anyway
        }
      }

      loadGroupMarker();
    }
  });
</script>

<MapLibre
  bind:map
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
>
  {#if $singleRideGeoJSON}
    {#if iconLoaded && groupMarkerLoaded}
      <RideLayers sourceId={SOURCE_ID} defaultIconName={ICON_NAME} />
    {/if}
    <GeoJSONSource data={$singleRideGeoJSON} id={SOURCE_ID} />
  {/if}
  <ParkLayer isDarkMode={mode.current === "dark"} />
</MapLibre>
