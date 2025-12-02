<script lang="ts">
  import { currentRoute, TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";
  import { GeoJSONSource, RawLayer, MapLibre } from "svelte-maplibre-gl";
  import { Map, LngLatBounds } from "maplibre-gl";
  import ParkLayer from "../map/parkLayer.svelte";

  const ROUTE_SOURCE_ID = "route-source";
  const ROUTE_LAYER_ID = "route-layer";

  const { route } = $props();

  let map: Map | undefined = $state.raw();
  let source = $derived(TILE_URLS[mode.current as keyof typeof TILE_URLS]);

  $effect(() => {
    if (map && route) {
      console.log("[RideRouteMap] Route data:", route);
      // Fit map to route bounds
      const coordinates = route.geojson.geometry.coordinates;
      if (coordinates && coordinates.length > 0) {
        console.log(
          "[RideRouteMap] Fitting bounds to coordinates:",
          coordinates.length,
        );
        const bounds = new LngLatBounds();
        coordinates.forEach((coord) => {
          bounds.extend([coord[0], coord[1]] as [number, number]);
        });

        console.log("[RideRouteMap] Calculated bounds:", bounds.toArray());

        // Use a small timeout to ensure map is ready/sized
        setTimeout(() => {
          if (map) {
            console.log("[RideRouteMap] Executing fitBounds");
            map.fitBounds(bounds, { padding: 40, duration: 800 });
          }
        }, 100);
      } else {
        console.warn("[RideRouteMap] No coordinates found in route");
      }
    }
  });

  let routeGEOJSON = $derived.by(() => {
    if (!route || !route.geojson.geometry.coordinates) {
      return {
        type: "FeatureCollection" as const,
        features: [],
      };
    }

    const coordinates = route.geojson.geometry.coordinates;
    const features = [
      {
        type: "Feature" as const,
        geometry: {
          type: "LineString" as const,
          coordinates: coordinates.map((coord: [number, number, number]) => [
            coord[0],
            coord[1],
          ]),
        },
        properties: route.geojson.properties,
      },
    ];

    return {
      type: "FeatureCollection" as const,
      features,
    };
  });
</script>

<MapLibre
  bind:map
  class="h-[55vh] min-h-[300px]"
  style={source}
  attributionControl={false}
  dragPan={false}
  dragRotate={false}
  doubleClickZoom={false}
  scrollZoom={false}
  touchZoomRotate={false}
  touchPitch={false}
  boxZoom={false}
  keyboard={false}
>
  <GeoJSONSource data={routeGEOJSON} id={ROUTE_SOURCE_ID}>
    <RawLayer
      id={ROUTE_LAYER_ID}
      source={ROUTE_SOURCE_ID}
      type="line"
      layout={{
        "line-join": "round",
        "line-cap": "round",
      }}
      paint={{
        "line-color": "#ff0000",
        "line-width": 4,
        "line-opacity": 0.8,
      }}
    />
  </GeoJSONSource>

  <ParkLayer isDarkMode={mode.current === "dark"} />
</MapLibre>
