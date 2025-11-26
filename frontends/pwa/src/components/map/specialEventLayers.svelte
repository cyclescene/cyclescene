<script lang="ts">
  import { onMount } from "svelte";
  import { today, getLocalTimeZone } from "@internationalized/date";
  import PrideLayer from "./events/prideLayer.svelte";
  import HalloweenLayer from "./events/halloweenLayer.svelte";
  import ChristmasLayer from "./events/christmasLayer.svelte";

  interface Props {
    isDarkMode: boolean;
  }

  let { isDarkMode }: Props = $props();

  // Event constants
  const EVENT_PRIDE = "pride";
  const EVENT_HALLOWEEN = "halloween";
  const EVENT_CHRISTMAS = "christmas";

  // Determine current event based on month
  function getCurrentEvent(): string | null {
    const currentDate = today(getLocalTimeZone());
    const month = currentDate.month;
    if (month === 6) return EVENT_PRIDE; // June
    if (month === 10) return EVENT_HALLOWEEN; // October
    if (month === 12) return EVENT_CHRISTMAS; // December
    return null;
  }

  let currentEvent = $state(getCurrentEvent());

  onMount(() => {
    // Update event daily in case the component stays mounted across midnight
    const checkEventInterval = setInterval(() => {
      currentEvent = getCurrentEvent();
    }, 1000 * 60 * 60); // Check every hour

    return () => clearInterval(checkEventInterval);
  });

  // Component map
  const eventLayerMap: Record<string, any> = {
    [EVENT_PRIDE]: PrideLayer,
    [EVENT_HALLOWEEN]: HalloweenLayer,
    [EVENT_CHRISTMAS]: ChristmasLayer,
  };

  const CurrentEventComponent = eventLayerMap[currentEvent || ""];
</script>

{#if CurrentEventComponent}
  <svelte:component this={CurrentEventComponent} isDarkMode={isDarkMode} />
{/if}
