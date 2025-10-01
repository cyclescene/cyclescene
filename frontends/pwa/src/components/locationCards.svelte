<script>
  import { ScrollArea } from "$lib/components/ui/scroll-area/index";
  import { mapViewStore } from "$lib/stores";
  import Card from "./card.svelte";
</script>

{#if $mapViewStore.eventCardsVisible && $mapViewStore.selectedEvents && $mapViewStore.selectedEvents.length > 0}
  <div
    class="absolute bottom-[50px] mb-14 left-0 w-full bg-transparent py-0 px-5 z-[1000]"
  >
    <ScrollArea class={``} scrollbarYClasses={`hidden`}>
      <div class=" max-h-[35vh] w-full flex flex-col gap-1">
        {#each $mapViewStore.selectedEvents as ride (ride.id)}
          <Card {ride} />
        {/each}
      </div>
    </ScrollArea>
  </div>
{/if}
