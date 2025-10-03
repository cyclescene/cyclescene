<script lang="ts">
  import type * as GeoJSON from "geojson";
  import type { RideData, ValidatedRide } from "$lib/types";
  import { type Snippet } from "svelte";
  import { GeoJSONSource } from "svelte-maplibre-gl";

  let {
    rides,
    sourceId,
    setValidCoords,
    children,
  }: {
    rides: RideData[];
    children: Snippet;
    sourceId: string;
    setValidCoords: (rides: ValidatedRide[]) => void;
    // bindValidCoords?: ValidatedRide[];
  } = $props();

  const validRides = $derived(
    rides
      .filter(
        (ride) =>
          ride.lon.Valid &&
          ride.lat.Valid &&
          !isNaN(ride.lon.Float64 as number) &&
          !isNaN(ride.lat.Float64 as number),
      )
      .map<ValidatedRide>((ride) => ({
        id: ride.id,
        name: ride.title,
        lat: ride.lat.Float64 as number,
        lng: ride.lon.Float64 as number,
      })),
  );

  const rideGeoJSON = $derived<GeoJSON.FeatureCollection<GeoJSON.Point, any>>({
    type: "FeatureCollection",
    features: validRides.map((coord) => ({
      type: "Feature",
      geometry: {
        type: "Point",
        coordinates: [coord.lng, coord.lat],
      },
      properties: {
        id: coord.id,
        name: coord.name,
      },
    })),
  });

  $effect(() => {
    if (setValidCoords) {
      setValidCoords(validRides);
    }
  });
</script>

<GeoJSONSource data={rideGeoJSON} id={sourceId}>
  {#if children}
    {@render children?.()}
  {/if}
</GeoJSONSource>
