<script lang="ts">
  import { getContext } from "svelte";

  const { height, yScale, data } = getContext("LayerCake");

  const tickCount = 5;
  let ticks = $derived.by(() => {
    const elevations = ($data as any[]).map((d: any) => d.elevation);
    const minElevation = Math.min(...elevations);
    const maxElevation = Math.max(...elevations);
    const range = maxElevation - minElevation;
    const step = range / (tickCount - 1);
    return Array.from({ length: tickCount }, (_, i) =>
      Math.round(minElevation + i * step)
    );
  });
</script>

<g class="axis y-axis" data-slot="y-axis">
  {#each ticks as tick}
    {@const y = $yScale(tick)}
    <g class="tick" transform="translate(0, {y})">
      <line x1={-6} x2={0} class="stroke-border" />
      <text
        x={-12}
        text-anchor="end"
        dominant-baseline="middle"
        class="text-xs fill-muted-foreground"
      >
        {tick} m
      </text>
    </g>
  {/each}
  <line x1={0} x2={0} y1={0} y2={height} class="stroke-border" />
</g>

<style>
  :global(.y-axis) {
    user-select: none;
  }
</style>
