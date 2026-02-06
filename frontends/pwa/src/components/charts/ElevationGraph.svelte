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
                distance: Math.round(cumulativeDistance * 100) / 100,
                elevation: Math.round(validElevation),
            });
        }

        return elevationData;
    });

    const chartConfig = {
        elevation: { label: "Elevation (m)", color: "var(--chart-1)" },
    } satisfies Chart.ChartConfig;
</script>

<div class="relative">
    <span
        class="absolute -left-7 top-1/2 -translate-y-1/2 -rotate-90 text-[10px] text-muted-foreground whitespace-nowrap"
        >Elevation (m)</span
    >
    <Chart.Container config={chartConfig} class="elevation-graph ml-4">
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
                        format: (v) => {
                            if (v === 0) return "";
                            return v.toFixed(1);
                        },
                    },
                    yAxis: {
                        format: (v) => {
                            if (v === 0) return "";
                            return v.toFixed(0);
                        },
                    },
                }}
            >
                {#snippet tooltip()}
                    <Chart.Tooltip labelFormatter={(v) => `${v} km`} />
                {/snippet}
            </AreaChart>
        {:else}
            <div
                class="flex items-center justify-center h-full text-muted-foreground"
            >
                <p class="text-sm">No elevation data available</p>
            </div>
        {/if}
    </Chart.Container>
    <span
        class="block text-center text-[10px] text-muted-foreground ml-4 mt-1"
        >Distance (km)</span
    >
</div>

<style>
    :global(.elevation-graph) {
        width: 100%;
        height: 200px;
    }
</style>
