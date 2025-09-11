<script>
    import { currentRide, goBackInHistory } from "../lib/stores";

    function handleGoBack() {
        goBackInHistory();
        $currentRide = null;
    }
</script>

<div class="ride-details-top-bar-container">
    <button class="date-nav-button" onclick={handleGoBack}> Back </button>

    <div class="display-date">Ride Details</div>

    <button class="date-nav-button"> Share </button>
</div>

<style>
    .ride-details-top-bar-container {
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
