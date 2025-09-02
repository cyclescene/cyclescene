<script>
    import { currentView } from "./lib/stores.js";
    import MapView from "./views/MapView.svelte";
</script>

<main>
    <div class="view-container">
        {#if $currentView == "map"}
            <MapView />
        {:else}
            <h1>Nothing to see here</h1>
        {/if}
    </div>
</main>

<style>
    :root {
        --header-height: 55px;
        --footer-height: 60px;
    }

    main {
        position: relative;
        height: 100vh;
        width: 100vw;
        overflow: hidden;
    }

    .view-container {
        position: absolute;
        top: var(--header-height);
        bottom: var(--footer-height);
        left: 0;
        right: 0;
        overflow-y: auto;
    }
</style>
