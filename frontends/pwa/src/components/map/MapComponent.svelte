<script lang="ts">
  import { LngLatBounds, Map } from "maplibre-gl";
  import type { LngLatLike } from "maplibre-gl";
  import { MapLibre, Marker, Popup } from "svelte-maplibre-gl";
  import { STARTING_LAT, STARTING_LON } from "$lib/config";
  import { TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";

  const ORIGINAL_MAP_ZOOM = 12;
  const COLLISION_THRESHOLD_PIXELS = 50;
  const SINGLE_RIDE_ZOOM = 15.5;

  // Props
  const { rides, otherRides } = $props();

  // State
  let mapInstance: Map | undefined = $state(undefined);

  const todaysRides = $derived(
    rides.filter(
      (ride) =>
        ride.lon?.Float64 &&
        ride.lat?.Float64 &&
        !isNaN(ride.lon.Float64) &&
        !isNaN(ride.lat.Float64),
    ) || [],
  );

  const validCoords: LngLatLike[] = $derived(
    rides
      .filter(
        (ride) =>
          ride.lon?.Float64 &&
          ride.lat?.Float64 &&
          !isNaN(ride.lon.Float64) &&
          !isNaN(ride.lat.Float64),
      )
      .map((ride) => ({
        lat: ride.lat.Float64,
        lng: ride.lon.Float64,
      })),
  );

  // Map Controls
  const fitAllMarkers = (map: Map) => {
    const coords = validCoords;

    if (!map || !coords) return;

    if (coords.length === 0) {
      map.flyTo({
        center: [STARTING_LON, STARTING_LAT],
        zoom: ORIGINAL_MAP_ZOOM,
        essential: true,
        duration: 800,
      });
      return;
    }

    if (coords.length === 1) {
      map.flyTo({
        center: coords[0],
        zoom: SINGLE_RIDE_ZOOM,
        essential: true,
        duration: 800,
      });

      return;
    }

    const bounds = new LngLatBounds();
    coords.forEach((coord) => bounds.extend([coord.lng, coord.lat]));
    map.fitBounds(bounds, { padding: 100, duration: 800 });
  };

  // Event Handlers

  // Side Effects
  $effect(() => {
    if (mapInstance) {
      fitAllMarkers(mapInstance);
    }
  });

  let source = $derived(TILE_URLS[mode.current]);
</script>

<div class="map-container">
  <MapLibre
    bind:map={mapInstance}
    class="w-full h-[calc(100vh-115px)] mt-[60px] mb-[55px]"
    style={source}
    attributionControl={false}
  >
    {#each todaysRides as ride (ride.id)}
      <Marker lnglat={[ride.lon?.Float64, ride.lat?.Float64]}>
        <Popup>
          {ride.title}
        </Popup>
      </Marker>
    {/each}
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
