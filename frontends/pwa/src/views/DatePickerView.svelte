<script>
  import Button from "$lib/components/ui/button/button.svelte";
  import Calendar from "$lib/components/ui/calendar/calendar.svelte";
  import {
    currentRideStore,
    dateStore,
    goBackInHistory,
    mapStore,
  } from "$lib/stores";
  import { getLocalTimeZone, today } from "@internationalized/date";

  const todaysDate = today(getLocalTimeZone());
  let value = $state(todaysDate);

  $effect(() => {
    dateStore.setSpecificDate(value);
    currentRideStore.clearRide();
    mapStore.showCurrentRide(false);
    goBackInHistory();
  });

  function onClickToday() {
    dateStore.setSpecificDate(todaysDate);
    goBackInHistory();
  }
</script>

<div
  class="date-picker absolute top-0 bottom-[75px] min-h-[calc(100vh_-_115px)] w-full p-5 flex flex-col items-center justify-center z-[1000]"
>
  <Calendar bind:value class="p-5 rounded-xl text-2xl" />

  <Button
    disabled={false}
    variant="ghost"
    onclick={onClickToday}
    class="text-yellow-500 text-xl mt-10">Today</Button
  >
</div>
