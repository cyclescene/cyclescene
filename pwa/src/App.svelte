<script>
  import { onMount } from "svelte";
  import "./app.css";
  import DatePicker from "./components/datePicker.svelte";
  import NavigationBar from "./components/navigationBar.svelte";
  import RideDetailsTopBar from "./components/ride/rideDetailsTopBar.svelte";

  import {
    activeView,
    rides,
    savedRidesStore,
    SUB_VIEW_ABOUT,
    SUB_VIEW_APPEARANCE,
    SUB_VIEW_CONTACT,
    VIEW_DATE_PICKER,
    VIEW_LIST,
    VIEW_MAP,
    VIEW_OTHER_RIDES,
    VIEW_RIDE_DETAILS,
    VIEW_SAVED,
    VIEW_SETTINGS,
  } from "./lib/stores.js";
  import DatePickerView from "./views/DatePickerView.svelte";

  import ListView from "./views/ListView.svelte";
  import MapView from "./views/MapView.svelte";
  import OtherRidesView from "./views/OtherRidesView.svelte";
  import RideView from "./views/RideView.svelte";
  import SavedView from "./views/SavedView.svelte";
  import SavedRideTopBar from "./components/saved/savedRideTopBar.svelte";
  import { ModeWatcher } from "mode-watcher";
  import SettingsView from "./views/SettingsView.svelte";
  import SettingsTopBar from "./components/settings/settingsTopBar.svelte";
  import SettingsSubTopBar from "./components/settings/settingsSubTopBar.svelte";
  import SettingsAppearance from "./components/settings/settingsAppearance.svelte";

  onMount(() => {
    rides.init();
    rides.fetchUpcoming();

    savedRidesStore.init();
  });
</script>

<main class="flex flex-col min-h[100vh]">
  <ModeWatcher themeColors={{ dark: "black", light: "white" }} />
  <div class="shrink relative">
    <header class="shrink">
      {#if $activeView == VIEW_MAP || $activeView == VIEW_LIST || $activeView == VIEW_DATE_PICKER}
        <DatePicker />
      {:else if $activeView == VIEW_OTHER_RIDES || $activeView == VIEW_RIDE_DETAILS}
        <RideDetailsTopBar />
      {:else if $activeView == VIEW_SAVED}
        <SavedRideTopBar />
      {:else if $activeView == VIEW_SETTINGS}
        <SettingsTopBar />
      {:else if $activeView === SUB_VIEW_ABOUT || $activeView === SUB_VIEW_APPEARANCE || $activeView === SUB_VIEW_CONTACT}
        <SettingsSubTopBar />
      {/if}
    </header>
  </div>

  <div class="grow">
    <div class="map-view-container" class:hidden={!($activeView === VIEW_MAP)}>
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
      <SavedView />
    </div>

    <div class="view-container" class:hidden={!($activeView === VIEW_SETTINGS)}>
      <SettingsView />
    </div>

    <div
      class="view-container"
      class:hidden={!($activeView === VIEW_DATE_PICKER)}
    >
      <DatePickerView />
    </div>

    <div
      class="view-container"
      class:hidden={!($activeView === SUB_VIEW_APPEARANCE)}
    >
      <SettingsAppearance />
    </div>
  </div>

  <div class="shrink">
    <footer>
      <NavigationBar />
    </footer>
  </div>
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
    height: var(--header-height);
    top: 0;
    z-index: 100;
  }

  .hidden {
    display: none;
  }

  .map-view-container {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    overflow-y: auto;
    z-index: 0;
  }

  footer {
    height: var(--footer-height);
    width: 100vw;
    position: absolute;
    height: var(--header-height);
    bottom: 0;
    z-index: 100;
  }
</style>
