<script>
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

<div class="date-picker-container">
    <button class="date-nav-button" onclick={() => changeDay(-1)}>
        &lt;
    </button>

    <div class="display-date">{formattedDateForDisplay}</div>

    <button class="date-nav-button" onclick={() => changeDay(1)}> &gt; </button>
</div>

<style>
    .date-picker-container {
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 10px;
        background-color: #242424;
        border-bottom: 1px solid #ccc;
        position: relative; /* For z-index if needed */
        z-index: 500; /* Ensure it's above the map, but below modals */
    }

    .date-nav-button {
        background-color: #eee;
        border: 1px solid #ccc;
        border-radius: 4px;
        padding: 8px 12px;
        margin: 0 5px;
        color: #000;
        cursor: pointer;
        font-size: 1em;
        min-width: 40px;
    }

    .date-nav-button:hover {
        background-color: #e0e0e0;
    }

    .display-date {
        font-size: 1.2em;
        font-weight: bold;
        text-align: center;
        flex-grow: 1;
        padding: 8px 10px;
        min-width: 120px;
    }
</style>
