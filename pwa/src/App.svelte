<script>
    import "./app.css";
    import DatePicker from "./components/datePicker.svelte";
    import NavigationBar from "./components/navigationBar.svelte";
    import RideDetailsTopBar from "./components/rideDetailsTopBar.svelte";
    
    import {
        activeView,
        VIEW_LIST,
        VIEW_MAP,
        VIEW_OTHER_RIDES,
        VIEW_RIDE_DETAILS,
        VIEW_SAVED,
        VIEW_SETTINGS
    } from "./lib/stores.js";
    
    import ListView from "./views/ListView.svelte";
    import MapView from "./views/MapView.svelte";
    import OtherRidesView from "./views/OtherRidesView.svelte";
    import RideView from "./views/RideView.svelte";
</script>

<main>
    <header>
        {#if $activeView == VIEW_MAP}
            <DatePicker />
        {:else if $activeView == VIEW_RIDE_DETAILS}
            <RideDetailsTopBar />
        {:else if $activeView == VIEW_LIST}
            <DatePicker />
        {:else if $activeView == VIEW_OTHER_RIDES}
            <RideDetailsTopBar />
        {:else}
            <p>Nothing to see here</p>
        {/if}
    </header>

    <div class="view-container" class:hidden={!($activeView === VIEW_MAP)}>
        <MapView />
    </div>

    <div class="view-container" class:hidden={!($activeView === VIEW_LIST)}>
        <ListView />
    </div>

    <div
        class="view-container"
        class:hidden={!($activeView === VIEW_OTHER_RIDES)}
    >
        <OtherRidesView />
    </div>

    <div
        class="view-container"
        class:hidden={!($activeView === VIEW_RIDE_DETAILS)}
    >
        <RideView />
    </div>

    <div class="view-container" class:hidden={!($activeView === VIEW_SAVED)}>
        Saved Rides go here
    </div>

    <div class="view-container" class:hidden={!($activeView === VIEW_SETTINGS)}>
        Settings go here
    </div>

    <footer>
        <NavigationBar />
    </footer>
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

    .hidden {
        display: none;
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
