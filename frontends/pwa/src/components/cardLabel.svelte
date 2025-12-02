<script lang="ts">
  import MapPin from "~icons/mingcute/map-pin-line";
  import Flash from "~icons/lets-icons/flash-light";
  import Clock from "~icons/hugeicons/clock-01";
  import type { Component } from "svelte";

  let {
    label,
    children,
  }: {
    label: string;
    children?: any;
  } = $props();

  let Icon: Component;
  switch (label) {
    case "newsflash":
      Icon = Flash;
      break;
    case "startTime":
      Icon = Clock;
      break;
    case "venue":
      Icon = MapPin;
      break;
    case "date":
      Icon = Clock;
      break;
    default:
      Icon = MapPin; // fallback icon
  }
</script>

<div
  class="flex flex-row items-center
  gap-2"
>
  <Icon height="18" width="18" inlinn="true" style="color: orange;" />
  {@render children?.()}
</div>
