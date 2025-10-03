<script lang="ts">
  import { rideGeoJSON, TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";
  import { GeoJSONSource, MapLibre } from "svelte-maplibre-gl";
  import RideLayers from "../map/rideLayers.svelte";

  const { ride } = $props();

  const RIDE_COORDS = {
    lat: ride.lat?.Float64,
    lon: ride.lon?.Float64,
  };

  const SOURCE_ID = "ride-source";
  const ICON_NAME = "custom-bike-pin";
  const GEOAPIFY_API_URL =
    "https://api.geoapify.com/v2/icon/?type=awesome&color=%23ff0000&size=42&icon=bicycle&contentSize=15&strokeColor=%23ff0000&shadowColor=%23ff0000&contentColor=%23ffffff&noShadow&noWhiteCircle&scaleFactor=2&apiKey=d4d9d0642bfc40488a64cd3b43b4a63e";

  let mapInstance: Map | undefined = $state(undefined);
  let source = $derived(TILE_URLS[mode.current]);
  let iconLoaded = $state(false);

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
  <GeoJSONSource data={$rideGeoJSON} id={SOURCE_ID} />
  {#if iconLoaded}
    <RideLayers sourceId={SOURCE_ID} iconName={ICON_NAME} />
  {/if}
</MapLibre>
