<script lang="ts">
  import { TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";
  import { GeoJSONSource, RawLayer, MapLibre } from "svelte-maplibre-gl";
  import { Map, LngLatBounds } from "maplibre-gl";
  import { tick } from "svelte";
  import { STARTING_LAT, STARTING_LNG } from "$lib/config";

  const ROUTE_SOURCE_ID = "route-source";
  const ROUTE_LAYER_ID = "route-layer";
  const CITY_CENTER = [STARTING_LNG, STARTING_LAT] as [number, number];

  const { route } = $props();

  let map: Map | undefined = $state.raw();
  let source = $derived(TILE_URLS[mode.current as keyof typeof TILE_URLS]);
  let fitRetry: number | undefined = $state();

  let coordinates = $derived.by(() => {
    const routeCoordinates = route?.geojson?.geometry?.coordinates ?? [];

    return routeCoordinates.filter(([lng, lat]: [number, number, number]) => {
      return (
        Number.isFinite(lng) &&
        Number.isFinite(lat) &&
        lng >= -180 &&
        lng <= 180 &&
        lat >= -90 &&
        lat <= 90
      );
    });
  });

  let initialCenter = $derived.by(() => {
    if (coordinates.length === 0) return CITY_CENTER;

    const midpoint = coordinates[Math.floor(coordinates.length / 2)];
    return [midpoint[0], midpoint[1]] as [number, number];
  });

  function getRouteBounds() {
    if (coordinates.length === 0) return null;

    const bounds = new LngLatBounds();
    coordinates.forEach(([lng, lat]: [number, number, number]) => {
      bounds.extend([lng, lat]);
    });

    return bounds;
  }

  async function fitRouteToMap() {
    const routeMap = map;
    const bounds = getRouteBounds();

    if (!routeMap || !bounds) return;

    window.clearTimeout(fitRetry);
    await tick();
    await new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));

    if (routeMap !== map) return;

    routeMap.resize();

    const canvas = routeMap.getCanvas();
    if (canvas.clientWidth === 0 || canvas.clientHeight === 0) {
      fitRetry = window.setTimeout(fitRouteToMap, 100);
      return;
    }

    if (!routeMap.loaded()) {
      routeMap.once("load", fitRouteToMap);
      return;
    }

    routeMap.fitBounds(bounds, { padding: 55, duration: 0 });
  }

  $effect(() => {
    map;
    coordinates;

    fitRouteToMap();

    return () => {
      window.clearTimeout(fitRetry);
    };
  });

  let routeGEOJSON = $derived.by(() => {
    if (coordinates.length === 0) {
      return {
        type: "FeatureCollection" as const,
        features: [],
      };
    }

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

  let endpointsGeoJSON = $derived.by(() => {
    if (coordinates.length < 2) {
      return {
        type: "FeatureCollection" as const,
        features: [],
      };
    }

    const start = coordinates[0];
    const end = coordinates[coordinates.length - 1];

    return {
      type: "FeatureCollection" as const,
      features: [
        {
          type: "Feature" as const,
          geometry: { type: "Point" as const, coordinates: [start[0], start[1]] },
          properties: { type: "start" },
        },
        {
          type: "Feature" as const,
          geometry: { type: "Point" as const, coordinates: [end[0], end[1]] },
          properties: { type: "end" },
        },
      ],
    };
  });
</script>

<MapLibre
  bind:map
  class="h-[35vh] min-h-[250px]"
  style={source}
  center={initialCenter}
  zoom={12}
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

  <GeoJSONSource data={endpointsGeoJSON} id="route-endpoints">
    <RawLayer
      id="route-start"
      source="route-endpoints"
      type="circle"
      filter={["==", ["get", "type"], "start"]}
      paint={{
        "circle-radius": 7,
        "circle-color": "#22c55e",
        "circle-stroke-color": "#ffffff",
        "circle-stroke-width": 2,
      }}
    />
    <RawLayer
      id="route-end"
      source="route-endpoints"
      type="circle"
      filter={["==", ["get", "type"], "end"]}
      paint={{
        "circle-radius": 7,
        "circle-color": "#ef4444",
        "circle-stroke-color": "#ffffff",
        "circle-stroke-width": 2,
      }}
    />
  </GeoJSONSource>
</MapLibre>

<div class="flex items-center justify-center gap-4 py-1.5 text-xs text-muted-foreground">
  <span class="flex items-center gap-1.5">
    <span class="inline-block h-2.5 w-2.5 rounded-full bg-green-500"></span>
    Start
  </span>
  <span class="flex items-center gap-1.5">
    <span class="inline-block h-2.5 w-2.5 rounded-full bg-red-500"></span>
    End
  </span>
</div>
