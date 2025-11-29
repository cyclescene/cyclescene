<script lang="ts">
  import { singleRideGeoJSON, TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";
  import { GeoJSONSource, MapLibre } from "svelte-maplibre-gl";
  import RideLayers from "../map/rideLayers.svelte";
  import GroupMarkerLayers from "../map/groupMarkerLayers.svelte";
  import { Map } from "maplibre-gl";
  import { loadMarkerByKey } from "$lib/markers";
  import { CITY_CODE } from "$lib/config";

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

  $effect(() => {
    if (map && !iconLoaded) {
      async function loadCustomIcon() {
        try {
          const response = await map!.loadImage(GEOAPIFY_API_URL);
          map!.addImage(ICON_NAME, response.data);
          iconLoaded = true;
        } catch (error) {
          console.error("failed to load custom icon: ", error);
        }
      }

      loadCustomIcon();
    }
  });

  $effect(() => {
    if (map && !groupMarkerLoaded && ride?.group_marker) {
      async function loadGroupMarker() {
        try {
          const markerDataUrl = await loadMarkerByKey(CITY_CODE, ride.group_marker);
          const response = await map!.loadImage(markerDataUrl);
          map!.addImage(`group-marker-${ride.group_marker}`, response.data);
          groupMarkerLoaded = true;
        } catch (error) {
          console.error("failed to load group marker: ", error);
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
    {#if iconLoaded}
      <RideLayers sourceId={SOURCE_ID} iconName={ICON_NAME} />
    {/if}
    {#if groupMarkerLoaded && ride?.group_marker}
      <GroupMarkerLayers sourceId={SOURCE_ID} />
    {/if}
    <GeoJSONSource data={$singleRideGeoJSON} id={SOURCE_ID} />
  {/if}
</MapLibre>
