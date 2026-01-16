<script lang="ts">
  // Cycle Scene PWA - Main Application
  import "./app.css";
  import { errorLogger } from "./lib/errorLogger";
  import DatePicker from "./components/datePicker.svelte";
  import NavigationBar from "./components/navigationBar.svelte";
  import RideDetailsTopBar from "./components/ride/rideDetailsTopBar.svelte";

  import {
    activeView,
    installPromptEvent,
    rides,
    routesStore,
    savedRidesStore,
    syncStatus,
    triggerForegroundSync,
    VIEW_DATE_PICKER,
    VIEW_LIST,
    VIEW_MAP,
    VIEW_OTHER_RIDES,
    VIEW_RIDE_DETAILS,
    VIEW_SAVED,
    VIEW_SETTINGS,
    // SUB_VIEW_CONTACT,
    SUB_VIEW_COVID_SAFETY_RIDES,
    SUB_VIEW_DATA,
    SUB_VIEW_FAMILY_FRIENDLY_RIDES,
    SUB_VIEW_INSTALL,
    SUB_VIEW_PRIVACY_POLICY,
    SUB_VIEW_TERMS_OF_USE,
    SUB_VIEWS,
    SUB_VIEW_ABOUT,
    SUB_VIEW_ADULT_ONLY_RIDES,
    SUB_VIEW_APPEARANCE,
    SUB_VIEW_CHANGE_LOG,
  } from "./lib/stores.js";
  import DatePickerView from "./views/DatePickerView.svelte";

  import { ModeWatcher } from "mode-watcher";
  import { SvelteSet } from "svelte/reactivity";
  import SavedTopBar from "./components/saved/savedTopbar.svelte";
  import SettingsSubTopBar from "./components/settings/settingsSubTopBar.svelte";
  import SettingsTopBar from "./components/settings/settingsTopBar.svelte";
  import ListView from "./views/ListView.svelte";
  import MapView from "./views/MapView.svelte";
  import OtherRidesView from "./views/OtherRidesView.svelte";
  import RideView from "./views/RideView.svelte";
  import SavedView from "./views/SavedView.svelte";
  import SettingsView from "./views/SettingsView.svelte";
  import SubAboutView from "./views/sub/subAboutView.svelte";
  import SubAppearanceView from "./views/sub/subAppearanceView.svelte";
  import SubChangelogView from "./views/sub/subChangelogView.svelte";
  import SubDataView from "./views/sub/subDataView.svelte";
  import SubPrivacyPolicyView from "./views/sub/subPrivacyPolicyView.svelte";
  import SubRideListView from "./views/sub/subRideListView.svelte";
  import SubTermsOfServiceView from "./views/sub/subTermsOfServiceView.svelte";
  import SubInstallView from "./views/sub/subInstallView.svelte";

  // Initialization effect (runs once on mount)
  let initialized = $state(false);

  $effect(() => {
    if (initialized) return;

    (async () => {
      // Set dynamic page title based on city
      const cityCode = import.meta.env.VITE_CITY_CODE || "pdx";
      const cityNames = {
        pdx: "Portland",
        slc: "Salt Lake City",
      };
      const cityName = cityNames[cityCode] || cityCode.toUpperCase();
      document.title = `Cycle Scene - ${cityName}`;

      await rides.init();
      await savedRidesStore.init();
      await routesStore.init();

      // Tell service worker the city code (non-blocking)
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.ready
          .then((registration) => {
            if (registration.active) {
              registration.active.postMessage({
                type: "SET_CITY_CODE",
                cityCode: cityCode,
              });
            }
          })
          .catch((err) => {
            console.error("Service worker error:", err);
          });
      }

      initialized = true;
    })();
  });

  // Automatic sync on app focus (critical for iOS, helpful for all platforms)
  $effect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        // Check if data is stale (> 30 minutes old)
        const lastSync = $syncStatus.lastSyncTime;
        const isStale =
          !lastSync || Date.now() - lastSync.getTime() > 30 * 60 * 1000;

        if (isStale && navigator.onLine) {
          console.log("[App] Data is stale, triggering sync...");
          triggerForegroundSync();
        }
      }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  });

  const SUB_VIEWS_SET = new SvelteSet(SUB_VIEWS);

  let ActiveHeader = $derived(
    $activeView === VIEW_MAP ||
      $activeView === VIEW_DATE_PICKER ||
      $activeView === VIEW_LIST
      ? DatePicker
      : $activeView === VIEW_OTHER_RIDES || $activeView === VIEW_RIDE_DETAILS
        ? RideDetailsTopBar
        : $activeView === VIEW_SAVED
          ? SavedTopBar
          : $activeView === VIEW_SETTINGS
            ? SettingsTopBar
            : DatePicker,
  );

  // Sub-view header for settings sub-views
  let ActiveSubHeader = $derived(
    SUB_VIEWS_SET.has($activeView) ? SettingsSubTopBar : null,
  );

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
    [SUB_VIEW_DATA]: SubDataView,
    [SUB_VIEW_INSTALL]: SubInstallView,
  };

  // Derived reactive values
  let ActiveComponent = $derived(viewMap[$activeView]);
  let isMapVisible = $derived($activeView === VIEW_MAP);
  let isDatePickerVisible = $derived($activeView === VIEW_DATE_PICKER);
  let ridesLoading = $derived($rides.loading);
  let ridesLoadingStage = $derived($rides.loadingStage);
</script>

<svelte:window
  onappinstalled={() => {
    console.log("[App] App installed!");
    installPromptEvent.set(null);
  }}
  onerror={(event) => {
    errorLogger.logError('unexpected_error', event.error || new Error(event.message), {
      component: 'global',
      action: 'uncaught_error'
    });
  }}
  onunhandledrejection={(event) => {
    errorLogger.logError('unexpected_error', new Error(String(event.reason)), {
      component: 'global',
      action: 'unhandled_promise_rejection'
    });
  }}
/>

<main class="flex flex-col">
  <ModeWatcher themeColors={{ dark: "black", light: "white" }} />
  {#if ridesLoading}
    <div class="loading-overlay">
      <div class="loading-spinner">
        <div class="spinner"></div>
        <p>{ridesLoadingStage}</p>
      </div>
    </div>
  {/if}
  <header class="shrink">
    {#if ActiveSubHeader}
      <ActiveSubHeader />
    {:else}
      <ActiveHeader />
    {/if}
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
        <ActiveComponent />
      </div>
    {/if}
  </section>
  <footer class="shrink">
    <NavigationBar />
  </footer>
</main>

<style>
  :global(html),
  :global(body) {
    margin: 0;
    padding: 0;
    height: 100%;
    overflow: hidden;
  }

  main {
    display: flex;
    flex-direction: column;
    height: 100dvh;
    width: 100vw;
    overflow: hidden;
    padding-top: env(safe-area-inset-top);
    box-sizing: border-box;
  }

  header {
    flex-shrink: 0;
    width: 100%;
    height: var(--header-height);
    overflow: hidden;
  }

  .hidden {
    display: none;
  }

  section {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .view-container {
    position: relative;
    overflow: hidden;
    flex: 1;
    min-height: 0;
  }

  footer {
    flex-shrink: 0;
    width: 100%;
    height: calc(var(--footer-height) + 35px);
    overflow: hidden;
  }

  .loading-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
  }

  .loading-spinner {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid rgba(255, 255, 255, 0.3);
    border-top: 4px solid white;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  .loading-spinner p {
    color: white;
    font-size: 1rem;
    margin: 0;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
