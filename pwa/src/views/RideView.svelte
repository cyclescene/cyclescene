<script>
    import { currentRide, currentView } from "../lib/stores";

    function onButtonClick() {
        $currentView = "map";
        $currentRide = null;
    }
</script>

<div class="ride-details-container">
    <h1>{$currentRide.title}</h1>
    <h2>{$currentRide.newsflash.String}</h2>
    <p>{$currentRide.details.String}</p>
    <button onclick={onButtonClick}>close</button>
</div>

<style>
    .ride-details-container {
        padding: 20px;
        background: transparent;
        color: lightgray;
        border: 5px 5px lightgray;
        box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
        box-sizing: border-box;
        z-index: 1000;
        transform: translateY(0%);
        transition: transform 0.3s ease-out;
        overflow-y: auto;
    }

    h1 {
        margin-top: 60px;
    }
</style>
