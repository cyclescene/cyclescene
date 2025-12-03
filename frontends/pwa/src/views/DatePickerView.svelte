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

  // Update store when user selects a date
  const handleDateSelect = (selectedDate) => {
    if (selectedDate) {
      dateStore.setSpecificDate(selectedDate);
      currentRideStore.clearRide();
      mapStore.showCurrentRide(false);
      goBackInHistory();
    }
  };

  // Watch for value changes and handle selection
  $effect(() => {
    if (value) {
      handleDateSelect(value);
    }
  });

  function onClickToday() {
    value = todaysDate;
  }
</script>

<div class="date-picker-container">
  <div class="date-picker-content">
    <Calendar bind:value class="p-5 rounded-xl text-2xl" />

    <Button
      disabled={false}
      variant="ghost"
      onclick={onClickToday}
      class="text-yellow-500 text-xl mt-10">Today</Button
    >
  </div>
</div>

<style>
  .date-picker-container {
    height: 100%;
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 1.25rem;
    padding-bottom: calc(var(--footer-height) + env(safe-area-inset-bottom) + 10px);
    overflow: hidden;
  }

  .date-picker-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }
</style>
