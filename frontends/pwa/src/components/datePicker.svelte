<script>
  import Button from "$lib/components/ui/button/button.svelte";
  import {
    formattedDate,
    dateStore,
    navigateTo,
    VIEW_DATE_PICKER,
    activeView,
    goBackInHistory,
    mapViewStore,
  } from "../lib/stores";

  import IconChevronRight from "~icons/bxs/chevron-right";
  import IconChevronLeft from "~icons/bxs/chevron-left";

  // Function to navigate between days
  function changeDay(offset) {
    mapViewStore.clearSelectedRides();
    mapViewStore.setSelectedRides(false);
    if (offset > 0) {
      dateStore.addDays(offset);
    } else if (offset < 0) {
      dateStore.subtractDays(Math.abs(offset));
    }
  }

  function openDatePicker() {
    if ($activeView != VIEW_DATE_PICKER) {
      navigateTo(VIEW_DATE_PICKER);
    } else {
      goBackInHistory();
    }
  }
</script>

<div
  class="flex gap-5 items-center justify-center p-2.5 border-b-[1px] relative z-[500]"
>
  <Button
    disabled={false}
    variant="ghost"
    class="py-2 px-3 text-yellow-400 min-w-10"
    onclick={() => changeDay(-1)}><IconChevronLeft /></Button
  >

  <Button
    disabled={false}
    variant="ghost"
    class="text-xl text-yellow-400 grow font-bold text-center py-2 px-3"
    onclick={openDatePicker}
  >
    {$formattedDate}
  </Button>
  <Button
    disabled={false}
    variant="ghost"
    class="py-2 px-3 text-yellow-400 min-w-10"
    onclick={() => changeDay(1)}><IconChevronRight /></Button
  >
</div>
