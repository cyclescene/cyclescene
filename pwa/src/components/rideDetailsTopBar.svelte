<script>
    import { clearRide, goBackInHistory } from "../lib/stores";

    function handleGoBack() {
        goBackInHistory();
        clearRide();
    }
</script>

<div class="flex justify-center items-center p-2.5 bg-black z-[500] text-white">
    <button class="date-nav-button" onclick={handleGoBack}> Back </button>

    <div class="grow font-bold py-2 px-2.5 text-center text-xl">
        Ride Details
    </div>

    <button class="date-nav-button">Share</button>
</div>

<style>
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
</style>
