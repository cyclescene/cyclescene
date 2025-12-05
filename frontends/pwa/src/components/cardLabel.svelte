<script lang="ts">
  import MapPin from "~icons/mingcute/map-pin-line";
  import Flash from "~icons/lets-icons/flash-light";
  import Clock from "~icons/hugeicons/clock-01";
  import { type Component } from "svelte";

  let {
    label,
    children,
  }: {
    label: string;
    children?: any;
  } = $props();

  let Icon: Component = $derived(
    label === "newsflash"
      ? Flash
      : label === "startTime"
        ? Clock
        : label === "venue"
          ? MapPin
          : label === "date"
            ? Clock
            : MapPin, // fallback
  );
</script>

<div
  class="flex flex-row items-center
  gap-2"
>
  <Icon height="18" width="18" inlinn="true" style="color: orange;" />
  {@render children?.()}
</div>
