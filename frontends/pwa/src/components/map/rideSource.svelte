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
        (ride) => !isNaN(ride.lng as number) && !isNaN(ride.lat as number),
      )
      .map<ValidatedRide>((ride) => ({
        id: ride.id,
        name: ride.title,
        lat: ride.lat as number,
        lng: ride.lng as number,
      })),
  );

  const rideGeoJSON = $derived<GeoJSON.FeatureCollection<GeoJSON.Point, any>>({
    type: "FeatureCollection",
    features: validRides.map((coord) => {
      // Find the original ride to get group_marker info
      const originalRide = rides.find(r => r.id === coord.id);
      const groupMarker = originalRide?.group_marker ? `group-marker-${originalRide.group_marker}` : "";

      return {
        type: "Feature",
        geometry: {
          type: "Point",
          coordinates: [coord.lng, coord.lat],
        },
        properties: {
          id: coord.id,
          name: coord.name,
          group_marker_icon: groupMarker,
        },
      };
    }),
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
