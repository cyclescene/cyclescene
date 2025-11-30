<script lang="ts">
  import { ChartContainer, ChartTooltip, getPayloadConfigFromPayload } from "$lib/components/ui/chart";
  import { LayerCake, Svg, Html } from "layercake";
  import AxisX from "./AxisX.svelte";
  import AxisY from "./AxisY.svelte";
  import Line from "./Line.svelte";
  import Area from "./Area.svelte";

  interface ElevationPoint {
    distance: number;
    elevation: number;
  }

  const { coordinates } = $props<{
    coordinates: [number, number, number][];
  }>();

  // Convert coordinates to elevation data with cumulative distance
  let data = $derived.by(() => {
    let cumulativeDistance = 0;
    const elevationData: ElevationPoint[] = [];

    for (let i = 0; i < coordinates.length; i++) {
      const [lon, lat, elevation] = coordinates[i];

      // Calculate distance from previous point (Haversine formula)
      if (i > 0) {
        const [prevLon, prevLat] = coordinates[i - 1];
        const R = 6371; // Earth's radius in km
        const dLat = ((lat - prevLat) * Math.PI) / 180;
        const dLon = ((lon - prevLon) * Math.PI) / 180;
        const a =
          Math.sin(dLat / 2) * Math.sin(dLat / 2) +
          Math.cos((prevLat * Math.PI) / 180) *
            Math.cos((lat * Math.PI) / 180) *
            Math.sin(dLon / 2) *
            Math.sin(dLon / 2);
        const c = 2 * Math.asin(Math.sqrt(a));
        const distance = R * c;
        cumulativeDistance += distance;
      }

      elevationData.push({
        distance: cumulativeDistance,
        elevation: elevation || 0,
      });
    }

    return elevationData;
  });

  // Calculate min/max for scaling
  let minElevation = $derived(Math.min(...data.map(d => d.elevation)));
  let maxElevation = $derived(Math.max(...data.map(d => d.elevation)));
  let maxDistance = $derived(Math.max(...data.map(d => d.distance)));
</script>

<ChartContainer config={{}} class="elevation-graph">
  {#if data.length > 0}
    <LayerCake
      data={data}
      x="distance"
      y="elevation"
      yDomain={[minElevation * 0.95, maxElevation * 1.05]}
    >
      <Svg>
        <AxisX />
        <AxisY />
        <Area />
        <Line />
      </Svg>

      <Html>
        <ChartTooltip
          formatter={(value) => {
            if (typeof value === "number") {
              return value.toFixed(0) + " m";
            }
            return value;
          }}
        />
      </Html>
    </LayerCake>
  {:else}
    <div class="flex items-center justify-center h-full text-muted-foreground">
      <p class="text-sm">No elevation data available</p>
    </div>
  {/if}
</ChartContainer>

<style>
  :global(.elevation-graph) {
    width: 100%;
    height: 300px;
  }
</style>
