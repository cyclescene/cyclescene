<script>
    import Calendar from "$lib/components/ui/calendar/calendar.svelte";
    import { currentDate } from "$lib/stores";
    import { getLocalTimeZone, today } from "@internationalized/date";

    // Function to navigate between days
    function changeDay(offset) {
        const currentStoredDate = $currentDate;
        // Create a new Date object to avoid mutating the store directly
        const newDate = new Date(currentStoredDate.getTime());
        // --- Native Date method: setDate handles day addition/subtraction ---
        newDate.setDate(newDate.getDate() + offset);

        $currentDate = newDate; // Update the store with the new Date object
    }

    let value = $state(today(getLocalTimeZone()));

    $effect(() => {
        console.log(value);
    });
</script>

<Calendar bind:value></Calendar>
