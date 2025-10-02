<script>
  import { ScrollArea } from "$lib/components/ui/scroll-area/index";
  import { mapViewStore } from "$lib/stores";
  import Card from "./card.svelte";

  $: ride = $mapViewStore.selectedEvent;
</script>

{#if $mapViewStore.eventCardsVisible && $mapViewStore.selectedEvent}
  <div
    class="absolute bottom-[50px] mb-14 left-0 w-full bg-transparent py-0 px-5 z-[1000]"
  >
    <ScrollArea class={``} scrollbarYClasses={`hidden`}>
      <div class=" max-h-[35vh] w-full flex flex-col gap-1">
        <Card {ride} />
      </div>
    </ScrollArea>
  </div>
{/if}
