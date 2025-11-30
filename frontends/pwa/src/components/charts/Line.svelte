<script lang="ts">
  import { getContext } from "svelte";
  import { line } from "d3-shape";

  const { data, xScale, yScale } = getContext("LayerCake");

  let pathData = $derived.by(() => {
    if (!$data || $data.length === 0) return "";

    const pathGenerator = line()
      .x((d: any) => $xScale(d.distance))
      .y((d: any) => $yScale(d.elevation));

    return pathGenerator($data as any) || "";
  });
</script>

<path d={pathData} class="stroke-primary" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
