<script>
  import { ScrollArea } from "$lib/components/ui/scroll-area/index";
  import { activeView, VIEW_SAVED } from "$lib/stores";
  import Card from "../card.svelte";

  export let rides = [];
</script>

<div class="absolute top-0 bottom-[45px] scroll-area w-full p-5">
  <ScrollArea class={`relative`} scrollbarYClasses={"hidden"}>
    <div class="scroll-area flex flex-col gap-2">
      {#if rides && rides.length > 0}
        {#each rides as ride (ride.id)}
          <Card {ride} />
        {/each}
      {:else if $activeView === VIEW_SAVED}
        no rides saved for today :(
      {:else}
        No rides happening today :(
      {/if}
    </div>
  </ScrollArea>
</div>

<style>
  .scroll-area {
    height: calc(100vh - var(--header-height) - var(--footer-height) - 40px);
  }
</style>
