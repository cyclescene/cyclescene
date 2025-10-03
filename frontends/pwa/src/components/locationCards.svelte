<script lang="ts">
  import { ScrollArea } from "$lib/components/ui/scroll-area/index";
  import { mapStore } from "$lib/stores";
  import type { RideData } from "$lib/types";
  import Card from "./card.svelte";

  let ride = $state<RideData | null>();

  $effect(() => {
    ride = $mapStore.selectedEvent;
  });
</script>

{#if $mapStore.eventCardsVisible && $mapStore.selectedEvent}
  <div
    class="absolute bottom-[-20px] mb-14 left-0 w-full bg-transparent py-0 px-5 z-[1000]"
  >
    <ScrollArea class={``} scrollbarYClasses={`hidden`}>
      <div class=" max-h-[35vh] w-full flex flex-col gap-1">
        <Card {ride} />
      </div>
    </ScrollArea>
  </div>
{/if}
