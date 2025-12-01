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
    if (!map || !route) return;

    const fitMapToRoute = () => {
      if (!map || !route) return;

      const coordinates = route.geojson.geometry.coordinates;
      if (!coordinates || coordinates.length === 0) {
        console.warn("[RideRouteMap] No coordinates found in route");
        return;
      }

      const bounds = new LngLatBounds();
      coordinates.forEach((coord: [number, number, number]) => {
        bounds.extend([coord[0], coord[1]] as [number, number]);
      });

      console.log("[RideRouteMap] Fitting bounds:", bounds.toArray());

      // Use fitBounds with reasonable padding
      map.fitBounds(bounds, {
        padding: { top: 50, bottom: 50, left: 50, right: 50 },
        duration: 800,
        maxZoom: 16,
      });
    };

    // Wait for the source to be loaded on the map
    const onSourceData = (e: any) => {
      if (e.sourceId === ROUTE_SOURCE_ID && e.isSourceLoaded) {
        console.log("[RideRouteMap] Route source loaded, fitting bounds");
        fitMapToRoute();
        map.off("sourcedata", onSourceData);
      }
    };

    // Check if style is loaded first
    if (map.isStyleLoaded()) {
      // Small delay to ensure GeoJSONSource is added
      setTimeout(fitMapToRoute, 100);
    } else {
      map.once("style.load", () => {
        setTimeout(fitMapToRoute, 100);
      });
    }

    // Also listen for source data changes
    map.on("sourcedata", onSourceData);

    return () => {
      map?.off("sourcedata", onSourceData);
    };
  });

  let routeGEOJSON = $derived.by(() => {
    if (!route || !route.geojson.geometry.coordinates) {
      console.warn("[RideRouteMap] Invalid route or missing coordinates");
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

    console.log("[RideRouteMap] Generated GeoJSON:", {
      type: "FeatureCollection",
      features,
    });

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
