<script>
    import DatePicker from "./components/datePicker.svelte";
    import { currentView } from "./lib/stores.js";
    import MapView from "./views/MapView.svelte";
    import RideView from "./views/RideView.svelte";
</script>

<main>
    <header><DatePicker /></header>
    <div class="view-container">
        {#if $currentView == "map"}
            <MapView />
        {:else if $currentView == "ride"}
            <RideView />
        {:else}
            <h1>Nothing to see here</h1>
        {/if}
    </div>
    <footer></footer>
</main>

<style>
    :root {
        --header-height: 55px;
        --footer-height: 60px;
    }

    :global(html),
    :global(body) {
        margin: 0;
        padding: 0;
        height: 100%;
        overflow: hidden;
    }

    main {
        position: relative;
        height: 100vh;
        width: 100vw;
        overflow: hidden;
    }

    header {
        width: 100vw;
        background-color: #242424;
        position: absolute;
        height: var(--header-height);
        top: 0;
        z-index: 100;
    }

    .view-container {
        position: absolute;
        top: 0;
        bottom: 0;
        left: 0;
        right: 0;
        overflow-y: auto;
        z-index: 1;
    }

    footer {
        height: var(--footer-height);
        width: 100vw;
        background-color: #242424;
        position: absolute;
        height: var(--header-height);
        bottom: 0;
        z-index: 100;
    }
</style>
