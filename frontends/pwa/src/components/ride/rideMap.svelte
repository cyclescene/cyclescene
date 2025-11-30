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

  $effect(() => {
    if (map && !iconLoaded) {
      async function loadIcons() {
        try {
          console.log(`[RideMap] Loading icons for ride: ${ride?.title}`);

          // Load default bike pin icon
          console.log(`[RideMap] Loading default icon: ${ICON_NAME}`);
          const response = await map!.loadImage(GEOAPIFY_API_URL);
          map!.addImage(ICON_NAME, response.data);
          console.log(
            `[RideMap] ✓ Successfully added default icon to map: ${ICON_NAME}`,
          );

          // Load group marker if ride has one
          if (ride?.group_marker) {
            try {
              console.log(
                `[RideMap] Loading group marker: ${ride.group_marker}`,
              );
              const markerDataUrl = await loadMarkerByKey(
                CITY_CODE,
                ride.group_marker,
              );
              const markerResponse = await map!.loadImage(markerDataUrl);
              const imageName = `group-marker-${ride.group_marker}`;
              map!.addImage(imageName, markerResponse.data);
              console.log(
                `[RideMap] ✓ Successfully added group marker to map: ${imageName}`,
              );
            } catch (error) {
              console.error(
                `[RideMap] ✗ Failed to load group marker: ${error}`,
              );
            }
          } else {
            console.log(
              `[RideMap] No group marker for this ride, using default icon only`,
            );
          }

          iconLoaded = true;
        } catch (error) {
          console.error("[RideMap] Failed to load icons: ", error);
        }
      }

      loadIcons();
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
    {#if iconLoaded}
      <RideLayers sourceId={SOURCE_ID} defaultIconName={ICON_NAME} />
    {/if}
    <GeoJSONSource data={$singleRideGeoJSON} id={SOURCE_ID} />
  {/if}
  <ParkLayer isDarkMode={mode.current === "dark"} />
</MapLibre>
