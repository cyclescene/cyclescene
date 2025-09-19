<script>
    import Button from "$lib/components/ui/button/button.svelte";
    import {
        formattedDate,
        dateStore,
        navigateTo,
        VIEW_DATE_PICKER,
        activeView,
        goBackInHistory,
    } from "../lib/stores";

    // Function to navigate between days
    function changeDay(offset) {
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
    class="flex items-center justify-center p-2.5 bg-black text-white border-b-[1px] relative z-[500]"
>
    <Button
        class="bg-black border-2 border-white py-2 px-3 min-w-10"
        onclick={() => changeDay(-1)}>&lt;</Button
    >

    <button
        class="text-2xl grow font-bold text-center py-2 px-3"
        onclick={openDatePicker}
    >
        {$formattedDate}
    </button>
    <Button
        class="bg-black border-2 border-white py-2 px-3 min-w-10"
        onclick={() => changeDay(1)}>&gt;</Button
    >
</div>
