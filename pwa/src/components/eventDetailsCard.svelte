<script>
    import { format, parse } from "date-fns";
    import { currentRide, currentView } from "../lib/stores";

    export let events = null;
    export let visible = false;

    function formatTime(timeString) {
        if (!timeString) return "N/A";
        const parsedTime = parse(timeString, "HH:mm:ss", new Date());
        return format(parsedTime, "h:mm a");
    }

    function onCardClick(ride) {
        $currentView = "ride";
        $currentRide = ride;
    }
</script>

{#if visible && events && events.length > 0}
    <div class="event-details-card">
        {#each events as event (event.id)}
            <button class="ride-detail-item" onclick={() => onCardClick(event)}>
                <h4>{event.title}</h4>
                <p>{event.newsflash.String}</p>
                <p>{event.venue.String}</p>
                <p>{formatTime(event.starttime)}</p>
            </button>
        {/each}
    </div>
{/if}

<style>
    .event-details-card {
        position: absolute;
        bottom: 50px;
        left: 0;
        width: 100%;
        max-height: 35vh;
        background: transparent;
        color: lightgray;
        border: 5px lightgray;
        box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
        padding: 0px 20px 20px 20px;
        box-sizing: border-box;
        z-index: 1000;
        transform: translateY(0%);
        transition: transform 0.3s ease-out;
        overflow-y: auto;
        overflow-x: hidden;
    }

    .ride-detail-item {
        all: unset;
        cursor: pointer;
        border: none;
        display: block;
        box-sizing: border-box;

        width: 100%;
        border-radius: 25px;
        background-color: #242424;
        padding: 15px;
        margin-bottom: 20px;
    }
    .ride-detail-item h4 {
        margin-top: 0;
        margin-bottom: 5px;
        font-size: 1.2em;
    }
    .ride-detail-item p {
        margin-top: 0;
        margin-bottom: 5px;
        font-size: 0.9em;
    }
</style>
