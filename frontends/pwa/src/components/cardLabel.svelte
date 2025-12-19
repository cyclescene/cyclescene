<script lang="ts">
  import MapPin from "~icons/mingcute/map-pin-line";
  import Flash from "~icons/lets-icons/flash";
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

<div class="flex flex-row items-center gap-1.5">
  <Icon
    height="22"
    width="22"
    inline="true"
    style="color: orange; min-width: 22px;"
  />
  {@render children?.()}
</div>
