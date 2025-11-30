<script lang="ts">
  import { getContext } from "svelte";

  const { width, height, xScale, data } = getContext("LayerCake");

  const tickCount = 5;
  let ticks = $derived.by(() => {
    const maxDistance = Math.max(...($data as any[]).map((d: any) => d.distance));
    const step = maxDistance / (tickCount - 1);
    return Array.from({ length: tickCount }, (_, i) => i * step);
  });
</script>

<g class="axis x-axis" data-slot="x-axis">
  {#each ticks as tick}
    {@const x = $xScale(tick)}
    <g class="tick" transform="translate({x}, 0)">
      <line y1={$height} y2={$height + 6} class="stroke-border" />
      <text
        y={$height + 16}
        text-anchor="middle"
        class="text-xs fill-muted-foreground"
      >
        {(tick / 1).toFixed(1)} km
      </text>
    </g>
  {/each}
  <line
    x1={0}
    x2={$width}
    y1={$height}
    y2={$height}
    class="stroke-border"
  />
</g>

<style>
  :global(.x-axis) {
    user-select: none;
  }
</style>
