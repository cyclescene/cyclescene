<script lang="ts">
  import { currentRoute, TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";
  import { GeoJSONSource, MapLibre, LineLayer } from "svelte-maplibre-gl";
  import { Map } from "maplibre-gl";
  import type { RouteGeoJSON } from "$lib/api";
  import ParkLayer from "../map/parkLayer.svelte";

  const ROUTE_SOURCE_ID = "route-source";
  const ROUTE_LAYER_ID = "route-layer";

  let map: Map | undefined = $state.raw();
  let source = $derived(TILE_URLS[mode.current as keyof typeof TILE_URLS]);

  $effect(() => {
    if (map && $currentRoute) {
      // Fit map to route bounds
      const coordinates = $currentRoute.geometry.coordinates;
      if (coordinates && coordinates.length > 0) {
        const bounds = coordinates.reduce(
          (bounds, coord) => {
            return bounds.extend([coord[0], coord[1]]);
          },
          new (require("maplibre-gl")).LngLatBounds(
            coordinates[0].slice(0, 2),
            coordinates[0].slice(0, 2)
          )
        );

        map.fitBounds(bounds, { padding: 40, duration: 800 });
      }
    }
  });
</script>

<MapLibre
  bind:map
  class="h-[55vh] min-h-[300px]"
  dragPan={false}
  doubleClickZoom={false}
  scrollZoom={false}
  touchZoomRotate={false}
  boxZoom={false}
  attributionControl={false}
  style={source}
>
  {#if $currentRoute}
    <GeoJSONSource
      data={{
        type: "FeatureCollection",
        features: [
          {
            type: "Feature",
            geometry: $currentRoute.geometry,
            properties: {
              distance_km: $currentRoute.properties.distance_km,
              distance_mi: $currentRoute.properties.distance_mi,
            },
          },
        ],
      }}
      id={ROUTE_SOURCE_ID}
    />
    <LineLayer
      source={ROUTE_SOURCE_ID}
      id={ROUTE_LAYER_ID}
      paint={{
        "line-color": mode.current === "dark" ? "#10b981" : "#059669",
        "line-width": 3,
        "line-opacity": 0.8,
      }}
    />
  {/if}
  <ParkLayer isDarkMode={mode.current === "dark"} />
</MapLibre>
