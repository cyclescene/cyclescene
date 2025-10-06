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
    SUB_VIEW_ADULT_ONLY_RIDES,
    SUB_VIEW_APPEARANCE,
    SUB_VIEW_CHANGE_LOG,
    SUB_VIEW_CONTACT,
    SUB_VIEW_COVID_SAFETY_RIDES,
    SUB_VIEW_DATA,
    SUB_VIEW_FAMILY_FRIENDLY_RIDES,
    SUB_VIEW_PRIVACY_POLICY,
    SUB_VIEW_TERMS_OF_USE,
    SUB_VIEWS,
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
  import SavedTopBar from "./components/saved/savedTopbar.svelte";
  import { ModeWatcher } from "mode-watcher";
  import SettingsView from "./views/SettingsView.svelte";
  import SettingsTopBar from "./components/settings/settingsTopBar.svelte";
  import SettingsSubTopBar from "./components/settings/settingsSubTopBar.svelte";
  import SubAppearanceView from "./views/sub/subAppearanceView.svelte";
  import SubRideListView from "./views/sub/subRideListView.svelte";
  import SubPrivacyPolicyView from "./views/sub/subPrivacyPolicyView.svelte";
  import SubAboutView from "./views/sub/subAboutView.svelte";
  import SubTermsOfServiceView from "./views/sub/subTermsOfServiceView.svelte";
  import SubChangelogView from "./views/sub/subChangelogView.svelte";
  import { SvelteSet } from "svelte/reactivity";

  onMount(() => {
    rides.init();
    rides.fetchUpcoming();

    savedRidesStore.init();
  });

  const headerMap = {
    [VIEW_MAP]: DatePicker,
    [VIEW_LIST]: DatePicker,
    [VIEW_DATE_PICKER]: DatePicker,
    [VIEW_OTHER_RIDES]: RideDetailsTopBar,
    [VIEW_RIDE_DETAILS]: RideDetailsTopBar,
    [VIEW_SAVED]: SavedTopBar,
    [VIEW_SETTINGS]: SettingsTopBar,
  };

  const SUB_VIEWS_SET = new SvelteSet(SUB_VIEWS);

  $: ActiveHeaderComponent = (() => {
    const active = $activeView;

    if (headerMap[active]) {
      return headerMap[active];
    }

    if (SUB_VIEWS_SET.has(active)) {
      return SettingsSubTopBar;
    }

    return null;
  })();

  const viewMap = {
    [VIEW_LIST]: ListView,
    [VIEW_OTHER_RIDES]: OtherRidesView,
    [VIEW_RIDE_DETAILS]: RideView,
    [VIEW_SAVED]: SavedView,
    [VIEW_SETTINGS]: SettingsView,
    [SUB_VIEW_APPEARANCE]: SubAppearanceView,
    [SUB_VIEW_TERMS_OF_USE]: SubTermsOfServiceView,
    [SUB_VIEW_PRIVACY_POLICY]: SubPrivacyPolicyView,
    [SUB_VIEW_ADULT_ONLY_RIDES]: SubRideListView,
    [SUB_VIEW_FAMILY_FRIENDLY_RIDES]: SubRideListView,
    [SUB_VIEW_COVID_SAFETY_RIDES]: SubRideListView,
    [SUB_VIEW_ABOUT]: SubAboutView,
    [SUB_VIEW_CHANGE_LOG]: SubChangelogView,
    [SUB_VIEW_DATA]: SubAboutView,
  };

  $: ActiveComponent = viewMap[$activeView];
  $: isMapVisible = $activeView === VIEW_MAP;
  $: isDatePickerVisible = $activeView === VIEW_DATE_PICKER;
</script>

<main class="flex flex-col max-h[100vh]">
  <ModeWatcher themeColors={{ dark: "black", light: "white" }} />
  <header class="shrink h-(--header-height)">
    <svelte:component this={ActiveHeaderComponent} />
  </header>

  <section class="grow view-container">
    <div class:hidden={!isMapVisible}>
      <MapView />
    </div>

    <div class:hidden={!isDatePickerVisible}>
      <DatePickerView />
    </div>

    {#if !isMapVisible && ActiveComponent}
      <div class="grow view-container">
        <svelte:component this={ActiveComponent} />
      </div>
    {/if}
  </section>
  <footer class="shrink">
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
    height: var(--header-height);
  }

  .hidden {
    display: none;
  }

  .view-container {
    position: relative;
  }

  footer {
    height: var(--footer-height);
    width: 100vw;
  }
</style>
