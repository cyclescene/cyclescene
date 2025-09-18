<script>
    import Button from "$lib/components/ui/button/button.svelte";
    import { currentDate } from "../lib/stores";
    import { format, isToday, isTomorrow, isYesterday } from "date-fns";

    let formattedDateForDisplay = "";

    $: {
        if ($currentDate) {
            if (isToday($currentDate)) {
                formattedDateForDisplay = "Today";
            } else if (isTomorrow($currentDate)) {
                formattedDateForDisplay = "Tomorrow";
            } else if (isYesterday($currentDate)) {
                formattedDateForDisplay = "Yesterday";
            } else {
                formattedDateForDisplay = format($currentDate, "eee, MMM d");
            }
        } else {
            formattedDateForDisplay = "";
        }
    }

    // Function to navigate between days
    function changeDay(offset) {
        const currentStoredDate = $currentDate;
        // Create a new Date object to avoid mutating the store directly
        const newDate = new Date(currentStoredDate.getTime());
        // --- Native Date method: setDate handles day addition/subtraction ---
        newDate.setDate(newDate.getDate() + offset);

        $currentDate = newDate; // Update the store with the new Date object
    }
</script>

<div
    class="flex items-center justify-center p-2.5 bg-black text-white border-b-[1px] relative z-[500]"
>
    <Button
        class="bg-black border-2 border-white py-2 px-3 min-w-10"
        onclick={() => changeDay(-1)}>&lt;</Button
    >

    <div class="text-2xl font-bold text-center grow py-2 px-3 min-w-[120px]">
        {formattedDateForDisplay}
    </div>

    <Button
        class="bg-black border-2 border-white py-2 px-3 min-w-10"
        onclick={() => changeDay(1)}>&gt;</Button
    >
</div>
