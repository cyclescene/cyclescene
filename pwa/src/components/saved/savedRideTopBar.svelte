<script>
  import Button from "$lib/components/ui/button/button.svelte";
  import {
    allSavedRidesNavigationDates,
    savedRidesStore,
    selectedSaveRidesNagivationDate,
  } from "$lib/stores";
  import IconChevronRight from "~icons/bxs/chevron-right";
  import IconChevronLeft from "~icons/bxs/chevron-left";
  import { formatDate } from "$lib/utils";

  function goToPreviousDay() {
    const currentIndex = $allSavedRidesNavigationDates.findIndex(
      (d) => d.compare($selectedSaveRidesNagivationDate) === 0,
    );
    if (currentIndex > 0) {
      selectedSaveRidesNagivationDate.set(
        $allSavedRidesNavigationDates[currentIndex - 1],
      );
    }
  }

  function goToNextDay() {
    const currentIndex = $allSavedRidesNavigationDates.findIndex(
      (d) => d.compare($selectedSaveRidesNagivationDate) === 0,
    );
    if (currentIndex < $allSavedRidesNavigationDates.length - 1) {
      selectedSaveRidesNagivationDate.set(
        $allSavedRidesNavigationDates[currentIndex + 1],
      );
    }
  }
</script>

<div
  class="flex items-center justify-center gap-5 p-2.5 border-b-[1px] relative z-[500]"
>
  <Button
    disabled={false}
    variant="secondary"
    class="py-2 px-3 text-yellow-400 min-w-10"
    onclick={goToPreviousDay}><IconChevronLeft /></Button
  >

  <Button
    disabled={false}
    variant="secondary"
    class="text-xl text-yellow-400 grow font-bold text-center py-2 px-3"
  >
    {formatDate($selectedSaveRidesNagivationDate.toString())}
  </Button>
  <Button
    disabled={false}
    variant="secondary"
    class="py-2 px-3 text-yellow-400 min-w-10"
    onclick={goToNextDay}><IconChevronRight /></Button
  >
</div>
