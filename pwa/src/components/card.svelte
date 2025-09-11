<script>
    import { format, parse } from "date-fns";
    import { currentRide, navigateTo, VIEW_RIDE_DETAILS } from "../lib/stores";
    export let ride = {};

    function formatTime(timeString) {
        if (!timeString) return "N/A";
        const parsedTime = parse(timeString, "HH:mm:ss", new Date());
        return format(parsedTime, "h:mm a");
    }

    function onCardClick(ride) {
        navigateTo(VIEW_RIDE_DETAILS);
        $currentRide = ride;
    }
</script>

<div class="card-container">
    <button class="card-item" onclick={() => onCardClick(ride)}>
        <h4>
            {ride.title}
        </h4>
        {#if ride.newsflash.Valid}
            <p>
                {ride.newsflash.String}
            </p>
        {/if}
        <p>
            {ride.venue.String}
        </p>
        <p>{formatTime(ride.starttime)}</p>
    </button>
</div>

<style>
    .card-container {
        background-color: black;
        width: 100%;
        border: 2px solid white;
        box-sizing: border-box;
        border-radius: 20px;
        padding: 20px;
    }

    .card-container:not(:first-child) {
        margin-top: 15px;
    }

    .card-item {
        all: unset;
        cursor: pointer;
        border: none;
        display: block;
        box-sizing: border-box;

        width: 100%;
    }

    h4 {
        font-size: 1.2em;
        margin: 0;
    }
</style>
