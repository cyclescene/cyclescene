<script lang="ts">
  import { AreaChart } from "layerchart";
  import { curveLinear } from "d3-shape";
  import * as Chart from "$lib/components/ui/chart/index.js";

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
      const coord = coordinates[i];
      const lon = coord[0];
      const lat = coord[1];
      const elevation = coord[2];

      // Skip if elevation is missing, null, undefined, or NaN
      const validElevation =
        elevation != null && !isNaN(elevation) ? elevation : 0;

      // Calculate distance from previous point (Haversine formula)
      if (i > 0) {
        const prevCoord = coordinates[i - 1];
        const prevLon = prevCoord[0];
        const prevLat = prevCoord[1];
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
        elevation: validElevation,
      });
    }

    return elevationData;
  });

  const chartConfig = {
    elevation: { label: "Elevation", color: "var(--chart-1)" },
  } satisfies Chart.ChartConfig;
</script>

<Chart.Container config={chartConfig} class="elevation-graph">
  {#if data.length > 0}
    <AreaChart
      {data}
      x="distance"
      series={[
        {
          key: "elevation",
          label: "Elevation (m)",
          color: chartConfig.elevation.color,
        },
      ]}
      props={{
        area: {
          curve: curveLinear,
          "fill-opacity": 0.4,
          line: { class: "stroke-1" },
        },
        xAxis: {
          labelPlacement: "middle",
          format: (v: number) => {
            // Skip showing label for 0
            if (v === 0) return "";
            return (v / 1).toFixed(1);
          },
          label: "Distance (km)",
        },
        yAxis: {
          labelPlacement: "middle",
          format: (v: number) => {
            // Skip showing label for 0
            if (v === 0) return "";
            return (v / 1).toFixed(0);
          },
          label: "Elevation (m)",
        },
      }}
    >
      {#snippet tooltip()}
        <Chart.Tooltip hideLabel />
      {/snippet}
    </AreaChart>
  {:else}
    <div class="flex items-center justify-center h-full text-muted-foreground">
      <p class="text-sm">No elevation data available</p>
    </div>
  {/if}
</Chart.Container>

<style>
  :global(.elevation-graph) {
    width: 100%;
    height: 100px;
  }
</style>
