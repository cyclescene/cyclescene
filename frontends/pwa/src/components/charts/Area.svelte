<script lang="ts">
  import { getContext } from "svelte";
  import { area } from "d3-shape";

  const { data, xScale, yScale, height } = getContext("LayerCake");

  let pathData = $derived.by(() => {
    if (!$data || $data.length === 0) return "";

    const areaGenerator = area()
      .x((d: any) => $xScale(d.distance))
      .y0($height)
      .y1((d: any) => $yScale(d.elevation));

    return areaGenerator($data as any) || "";
  });
</script>

<path
  d={pathData}
  class="fill-primary/10"
  stroke="none"
  opacity="0.6"
/>
