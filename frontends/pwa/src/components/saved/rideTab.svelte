<script lang="ts">
  import ScrollArea from "$lib/components/ui/scroll-area/scroll-area.svelte";
  import * as Tabs from "$lib/components/ui/tabs/index";
  import { savedRidesSplitByPastAndUpcoming as savedRides } from "$lib/stores";
  import Card from "../card.svelte";

  let pastRides = $savedRides.past;
  let upcomingRides = $savedRides.upcoming;
  let whichSaved = $state("upcoming");

  console.log(upcomingRides, pastRides);
</script>

<Tabs.Root
  bind:value={whichSaved}
  class="scroll-area flex-col items-center justify-center"
>
  <Tabs.List class="justify-center items-center mt-5">
    <Tabs.Trigger value="past">Past</Tabs.Trigger>
    <Tabs.Trigger value="upcoming">Upcoming</Tabs.Trigger>
  </Tabs.List>
  <Tabs.Content value="upcoming" class=" w-full px-5 pt-5">
    <ScrollArea class="">
      <div class="scroll-area">
        {#each upcomingRides as ride (ride)}
          <Card {ride} />
        {/each}
      </div>
    </ScrollArea>
  </Tabs.Content>
  <Tabs.Content value="past" class="grow w-full px-5 pt-5">
    {#each pastRides as ride (ride)}
      <Card {ride} />
    {/each}
  </Tabs.Content>
</Tabs.Root>

<style>
  .scroll-area {
    height: calc(100vh - var(--header-height) - var(--footer-height) - 90px);
  }
</style>
