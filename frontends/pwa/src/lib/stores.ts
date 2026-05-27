import { CalendarDate, parseDate } from "@internationalized/date";
import { LngLatBounds, type LngLatLike, type Map } from "maplibre-gl";
import { getPastRides, getUpcomingRides, getAllRoutes, type RouteGeoJSON } from "./api";
import { addSavedRide, deleteSavedRide, getAllSavedRides, getRidesfromDB, savedRideExists, saveRidesToDB, clearAllRides, clearSavedRides, saveRoutesToDB, getRoutesFromDB } from "./db";
import { today, getLocalTimeZone, DateFormatter } from "@internationalized/date";
import { writable, derived, get } from "svelte/store";
import { SvelteMap } from "svelte/reactivity";
import { ENABLE_INSTALL_PROMPT_V2, STARTING_LAT, STARTING_LNG } from "./config";
import { type ValidatedRide, type RideData } from "./types";
import type * as GeoJSON from "geojson";
import { CITY_CODE } from "./config";
import { errorLogger } from "./errorLogger";

// Portland, OR coordinates
const FALLBACK_LAT = STARTING_LAT
const FALLBACK_LNG = STARTING_LNG

const INSTALL_MESSAGE_DISMISSED_KEY = "install-message-dismissed";
const INSTALL_MESSAGE_DISMISSED_AT_KEY = "install-message-dismissed-at";
const INSTALL_MESSAGE_DISMISS_COOLDOWN = 7 * 24 * 60 * 60 * 1000;

type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>;
  userChoice?: Promise<{ outcome: "accepted" | "dismissed"; platform: string }>;
};

export type InstallPlatform = "ios" | "android" | "desktop" | "installed" | "unsupported";

export interface InstallInfo {
  platform: InstallPlatform;
  isInstalled: boolean;
  canUseNativePrompt: boolean;
  shouldShowInstallEntry: boolean;
  title: string;
  message: string;
  primaryActionLabel: string;
}

function getNavigatorPlatform() {
  return (
    (navigator as Navigator & { userAgentData?: { platform?: string } }).userAgentData?.platform ||
    navigator.platform ||
    ""
  ).toLowerCase();
}

function isIOSBrowser(userAgent: string) {
  const platform = getNavigatorPlatform();
  const isKnownChromiumShell = /chrome|chromium|crios|edg|opr|helium/.test(userAgent);
  const hasNonApplePlatform = /android|chrome os|chromium os|linux|win/.test(platform);
  const hasIOSUserAgent = /iphone|ipad|ipod/.test(userAgent) && !/helium/.test(userAgent);
  const hasIPadDesktopModeSignals =
    platform === "macintel" &&
    navigator.maxTouchPoints > 1 &&
    !isKnownChromiumShell;

  if (hasNonApplePlatform) {
    return false;
  }

  return hasIOSUserAgent || hasIPadDesktopModeSignals;
}

// Install prompt store
export const installPromptEvent = writable<BeforeInstallPromptEvent | null>(null);
export const installMessageDismissed = writable(
  isInstallMessageDismissed()
);

function isInstallMessageDismissed() {
  if (typeof window === "undefined") {
    return false;
  }

  const dismissedAt = Number(localStorage.getItem(INSTALL_MESSAGE_DISMISSED_AT_KEY) || "0");

  if (dismissedAt > 0) {
    return Date.now() - dismissedAt < INSTALL_MESSAGE_DISMISS_COOLDOWN;
  }

  if (localStorage.getItem(INSTALL_MESSAGE_DISMISSED_KEY) === "true") {
    localStorage.removeItem(INSTALL_MESSAGE_DISMISSED_KEY);
  }

  return false;
}

function detectInstallInfo(promptEvent: BeforeInstallPromptEvent | null): InstallInfo {
  if (typeof window === "undefined") {
    return {
      platform: "unsupported",
      isInstalled: false,
      canUseNativePrompt: false,
      shouldShowInstallEntry: false,
      title: "Install Cycle Scene",
      message: "Add Cycle Scene to your home screen for quick access.",
      primaryActionLabel: "Install",
    };
  }

  const userAgent = navigator.userAgent.toLowerCase();
  const isStandalone =
    (window.navigator as any).standalone === true ||
    window.matchMedia("(display-mode: standalone)").matches;
  const isIOS = isIOSBrowser(userAgent);
  const isAndroid = /android/.test(userAgent);
  const isDesktop = !isIOS && !isAndroid;

  if (isStandalone) {
    return {
      platform: "installed",
      isInstalled: true,
      canUseNativePrompt: false,
      shouldShowInstallEntry: false,
      title: "Cycle Scene is installed",
      message: "Open it from your home screen or app launcher.",
      primaryActionLabel: "Installed",
    };
  }

  if (promptEvent) {
    return {
      platform: isAndroid ? "android" : "desktop",
      isInstalled: false,
      canUseNativePrompt: true,
      shouldShowInstallEntry: true,
      title: "Install Cycle Scene",
      message: "Add Cycle Scene to your device for faster access and offline ride info.",
      primaryActionLabel: "Install",
    };
  }

  if (isIOS) {
    return {
      platform: "ios",
      isInstalled: false,
      canUseNativePrompt: false,
      shouldShowInstallEntry: true,
      title: "Add Cycle Scene to Home Screen",
      message: "Use your browser share menu to add Cycle Scene as an app.",
      primaryActionLabel: "Show steps",
    };
  }

  if (isAndroid || isDesktop) {
    return {
      platform: isAndroid ? "android" : "desktop",
      isInstalled: false,
      canUseNativePrompt: false,
      shouldShowInstallEntry: ENABLE_INSTALL_PROMPT_V2,
      title: "Install Cycle Scene",
      message: "Use your browser menu to install Cycle Scene when the install button is not available.",
      primaryActionLabel: "Show steps",
    };
  }

  return {
    platform: "unsupported",
    isInstalled: false,
    canUseNativePrompt: false,
    shouldShowInstallEntry: ENABLE_INSTALL_PROMPT_V2,
    title: "Install Cycle Scene",
    message: "Add Cycle Scene from your browser menu if your browser supports web apps.",
    primaryActionLabel: "Show steps",
  };
}

export const installInfo = derived(installPromptEvent, detectInstallInfo);

export async function promptInstallApp() {
  const event = get(installPromptEvent);

  if (!event) {
    return false;
  }

  await event.prompt();
  installPromptEvent.set(null);
  return true;
}

export function dismissInstallMessage() {
  installMessageDismissed.set(true);

  if (typeof window !== "undefined") {
    localStorage.setItem(INSTALL_MESSAGE_DISMISSED_AT_KEY, Date.now().toString());
    localStorage.removeItem(INSTALL_MESSAGE_DISMISSED_KEY);
  }
}

// Check if app is installable (not already installed)
export const isAppInstallable = derived(
  installInfo,
  ($installInfo) => $installInfo.shouldShowInstallEntry
);

// views
export const VIEW_MAP = 'map'
export const VIEW_LIST = 'list'
export const VIEW_RIDE_DETAILS = 'rideDetails'
export const VIEW_SAVED = 'saved'
export const VIEW_SETTINGS = 'settings'
export const VIEW_OTHER_RIDES = 'otherRides'
export const VIEW_DATE_PICKER = 'datePicker'

const LAST_PRIMARY_VIEW_KEY = "last-primary-view"
const PRIMARY_VIEWS = [VIEW_MAP, VIEW_LIST, VIEW_SAVED, VIEW_SETTINGS]

function getInitialPrimaryView() {
  if (typeof window === "undefined") {
    return VIEW_MAP
  }

  const storedView = localStorage.getItem(LAST_PRIMARY_VIEW_KEY)
  return PRIMARY_VIEWS.includes(storedView || "") ? storedView as string : VIEW_MAP
}

// setting sub views
export const SUB_VIEW_APPEARANCE = 'appearance'
export const SUB_VIEW_DATA = 'data'
export const SUB_VIEW_ABOUT = 'about'
export const SUB_VIEW_ADULT_ONLY_RIDES = 'adultOnlyRides'
export const SUB_VIEW_FAMILY_FRIENDLY_RIDES = 'familyFriendlyRides'
export const SUB_VIEW_COVID_SAFETY_RIDES = 'covideSafetyRides'
export const SUB_VIEW_PRIVACY_POLICY = "privacyPolicy"
export const SUB_VIEW_TERMS_OF_USE = 'termsOfUse'
export const SUB_VIEW_CHANGE_LOG = 'changeLog'
export const SUB_VIEW_CONTACT = 'contact'
export const SUB_VIEW_INSTALL = "install"
export const SUB_VIEW_SEARCH_RESULTS = "searchResults"
export const SUB_VIEWS = [
  SUB_VIEW_APPEARANCE,
  SUB_VIEW_DATA,
  SUB_VIEW_ABOUT,
  SUB_VIEW_ADULT_ONLY_RIDES,
  SUB_VIEW_FAMILY_FRIENDLY_RIDES,
  SUB_VIEW_COVID_SAFETY_RIDES,
  SUB_VIEW_PRIVACY_POLICY,
  SUB_VIEW_TERMS_OF_USE,
  SUB_VIEW_CHANGE_LOG,
  SUB_VIEW_CONTACT,
  SUB_VIEW_INSTALL,
]

export const rideSearchQuery = writable("")

export const TILE_URLS = {
  dark: "https://tiles.cyclescene.cc/dark-tiles.json",
  light: "https://tiles.cyclescene.cc/light-tiles.json"
};

const RIDES_SYNC_TAG = "update-rides-30min"
const SYNC_INTERVAL = 30 * 60 * 1000

// Sync status store for tracking background sync events
interface SyncStatus {
  lastSyncTime: Date | null;
  lastSyncStatus: "success" | "error" | "syncing" | null;
  lastSyncError: string | null;
  syncCount: number;
}

function createSyncStatusStore() {
  const { subscribe, set, update } = writable<SyncStatus>({
    lastSyncTime: null,
    lastSyncStatus: null,
    lastSyncError: null,
    syncCount: 0,
  });

  // Load from localStorage on init
  if (typeof window !== "undefined") {
    const stored = localStorage.getItem("sync-status");
    if (stored) {
      try {
        const parsed = JSON.parse(stored);
        parsed.lastSyncTime = parsed.lastSyncTime ? new Date(parsed.lastSyncTime) : null;
        set(parsed);
      } catch (e) {
        console.error("Failed to parse sync status:", e);
      }
    }
  }

  function saveToLocalStorage(state: SyncStatus) {
    if (typeof window !== "undefined") {
      localStorage.setItem("sync-status", JSON.stringify(state));
    }
  }

  return {
    subscribe,
    setSyncing: () => {
      update((state) => {
        const newState = { ...state, lastSyncStatus: "syncing" as const };
        saveToLocalStorage(newState);
        return newState;
      });
    },
    setSuccess: () => {
      update((state) => {
        const newState = {
          ...state,
          lastSyncTime: new Date(),
          lastSyncStatus: "success" as const,
          lastSyncError: null,
          syncCount: state.syncCount + 1,
        };
        saveToLocalStorage(newState);
        return newState;
      });
    },
    setError: (error: string) => {
      update((state) => {
        const newState = {
          ...state,
          lastSyncTime: new Date(),
          lastSyncStatus: "error" as const,
          lastSyncError: error,
        };
        saveToLocalStorage(newState);
        return newState;
      });
    },
  };
}

export const syncStatus = createSyncStatusStore();

interface RidesStore {
  loading: boolean;
  loadingStage: string
  rideData: RideData[];
  error: string | null;
}

function createRidesStore() {
  const { subscribe, set, update } = writable<RidesStore>({
    loading: true,
    loadingStage: "",
    rideData: [],
    error: null
  })

  function uniqueRidesById(rides: RideData[]) {
    return Array.from(new globalThis.Map(rides.map((ride) => [ride.id, ride])).values())
  }

  function updateStoreAndDB(freshRides: RideData[]) {
    freshRides = uniqueRidesById(freshRides)
    saveRidesToDB(freshRides).then(() => {
      getRidesfromDB()
        .then((rides) => (set({ loading: false, rideData: rides, loadingStage: "", error: null })))
        .catch(e => {
          update((store) => ({
            ...store, loading: false, error: `${e}`
          }))
        })
    })
      .catch(() => {
        // Failed to save rides
      })
  }

  if ('serviceWorker' in navigator) {
    const swMessageListener = (event: MessageEvent) => {
      const data = event.data
      if (data.type === "RIDES_UPDATE_SUCCESSFULL" && data.data) {
        syncStatus.setSyncing();
        try {
          updateStoreAndDB(data.data)
          syncStatus.setSuccess();
        } catch (error) {
          syncStatus.setError(error instanceof Error ? error.message : String(error));
        }
      }
    }
    navigator.serviceWorker.addEventListener('message', swMessageListener)

  }

  return {
    subscribe,
    init: async () => {
      update(store => ({ ...store, loading: true, loadingStage: "Loading rides from cache..." }))
      try {
        // Load from IndexedDB immediately
        let cachedRides = await getRidesfromDB()

        // Show cached data immediately
        update(store => ({ ...store, loading: false, loadingStage: "", rideData: cachedRides, error: null }))

        // Only fetch from API if cache is empty
        if (cachedRides.length === 0 && window.navigator.onLine === true) {
          update(store => ({ ...store, loading: true, loadingStage: "Fetching rides from API..." }))
          try {
            const upcomingRides = await getUpcomingRides()
            const pastRides = await getPastRides()

            const freshRides = uniqueRidesById([...upcomingRides, ...pastRides])

            update(store => ({ ...store, loading: true, loadingStage: "Saving rides to cache..." }))
            await saveRidesToDB(freshRides)
            const newRides = await getRidesfromDB()
            update(store => ({ ...store, loading: false, loadingStage: "", rideData: newRides, error: null }))
          } catch (apiErr) {
            console.error("Failed to fetch from API:", apiErr)
            update(store => ({ ...store, loading: false, loadingStage: "", error: "Unable to fetch ride data from API" }))
          }
        }
      } catch (err) {
        // Still try to show cached rides even if there was an error
        const cachedRidesOnError = await getRidesfromDB().catch(() => [])
        update(store => ({ ...store, loading: false, loadingStage: "", rideData: cachedRidesOnError, error: "Unable to load rides" }))
      }

      if ('serviceWorker' in navigator) {
        navigator.serviceWorker.ready
          .then(registration => {
            // Use Periodic Sync for browsers that support it (Chrome, Firefox)
            if ('PeriodicSyncManager' in self) {
              return registration.periodicSync.register(
                RIDES_SYNC_TAG,
                { minInterval: SYNC_INTERVAL }
              )
            }
            // Use Background Sync API for Apple devices
            else if ('SyncManager' in self) {
              return registration.sync.register(RIDES_SYNC_TAG)
            }
          })
      }
    },
    clearAndRefreshRides: async () => {
      try {
        update(store => ({ ...store, loading: true }))
        // Clear the rides from IndexedDB
        await clearAllRides()
        // Fetch fresh data from API
        const upcomingRides = await getUpcomingRides()
        const pastRides = await getPastRides()
        const freshRides = uniqueRidesById([...upcomingRides, ...pastRides])
        // Save to IndexedDB
        await saveRidesToDB(freshRides)
        // Update the store
        set({ loading: false, loadingStage: "", rideData: freshRides, error: null })
      } catch (e) {
        update(() => ({ loading: false, loadingStage: "", rideData: [], error: `Failed to refresh rides: ${e}` }))
      }
    }
  }
}

export function triggerForegroundSync() {
  if ('serviceWorker' in navigator && navigator.serviceWorker.controller) {
    navigator.serviceWorker.controller.postMessage({
      type: "FORCE_FOREGROUND_SYNC"
    })
  }
}

interface SavedRideStore {
  loading: boolean;
  data: RideData[];
  error: string | null
}

function createSavedRideStore() {
  const { subscribe, set, update } = writable<SavedRideStore>({
    loading: true,
    data: [],
    error: null
  })
  return {
    subscribe,
    init: async () => {
      try {
        const cachedRides = await getAllSavedRides()
        set({ loading: false, data: cachedRides, error: null })
      } catch (e) {
        set({ loading: false, data: [], error: "Could not load saved rides" })
      }
    },
    saveRide: async (ride: RideData) => {
      try {
        await addSavedRide(ride)
        const savedRides = await getAllSavedRides()
        set({ loading: false, data: savedRides, error: null })

      } catch (e) {
        errorLogger.logError('db_error', e instanceof Error ? e : new Error(String(e)), {
          component: 'stores.ts',
          action: 'saveRide',
          additionalData: { rideId: ride.id }
        })
        update((state) => ({ ...state, loading: false, error: "Could not save ride" }))
        throw e
      }
    },
    deleteRide: async (rideID: string) => {
      try {
        await deleteSavedRide(rideID)
        const savedRides = await getAllSavedRides()
        set({ loading: false, data: savedRides, error: null })
      } catch (e) {
        errorLogger.logError('db_error', e instanceof Error ? e : new Error(String(e)), {
          component: 'stores.ts',
          action: 'deleteRide',
          additionalData: { rideId: rideID }
        })
        update((state) => ({ ...state, loading: false, error: "Could not delete ride" }))
        throw e
      }

    },
    isRideSaved: async (rideID: string) => {
      try {
        const exists = await savedRideExists(rideID)
        return exists
      } catch (e) {
        errorLogger.logError('db_error', e instanceof Error ? e : new Error(String(e)), {
          component: 'stores.ts',
          action: 'isRideSaved',
          additionalData: { rideId: rideID }
        })
        return false
      }
    },
    clearAll: async () => {
      try {
        await clearSavedRides()
        set({ loading: false, data: [], error: null })
      } catch (e) {
        set({ loading: false, data: [], error: "Could not clear saved rides" })
      }
    }
  }
}

export const savedRidesStore = createSavedRideStore()

interface RoutesStore {
  loading: boolean;
  routes: SvelteMap<string, RouteGeoJSON>;
  error: string | null;
}

function createRoutesStore() {
  const { subscribe, set, update } = writable<RoutesStore>({
    loading: true,
    routes: new SvelteMap(),
    error: null
  })

  return {
    subscribe,
    init: async () => {
      try {
        let cachedRoutes: RouteGeoJSON[] = []
        if (window.navigator.onLine === true) {
          const freshRoutes = await getAllRoutes()
          await saveRoutesToDB(freshRoutes)

          cachedRoutes = await getRoutesFromDB()
        } else {
          cachedRoutes = await getRoutesFromDB()
        }

        // Convert to Map for quick lookup by ID
        const routesMap = new SvelteMap(cachedRoutes.map(route => [route.id, route]))
        set({ loading: false, routes: routesMap, error: null })
      } catch (err) {
        update(store => ({ ...store, loading: false, error: `${err}` }))
      }
    }
  }
}

const routesStoreInternal = createRoutesStore()
export const routesStore = routesStoreInternal

// Helper function to get a route by ID from the store
export const getRouteById = (routeId: string): RouteGeoJSON | undefined => {
  let result: RouteGeoJSON | undefined
  routesStoreInternal.subscribe(store => {
    result = store.routes.get(routeId)
  })()
  return result
}

export const allSavedRides = derived(
  savedRidesStore,
  ($savedRides) => {
    if (!$savedRides || !$savedRides.data) {
      return [];
    }


    return $savedRides.data
  }
)

export const savedRidesSplitByPastAndUpcoming = derived(allSavedRides, ($rides) => {
  const rides: { upcoming: RideData[]; past: RideData[] } = { upcoming: [], past: [] }

  const todaysDate = today(getLocalTimeZone())

  for (let i = 0; i < $rides.length; i++) {
    const ride = $rides[i]
    const rideDate = parseDate(ride.date)


    if (todaysDate.compare(rideDate) <= 0) {
      rides.upcoming.push(ride)
    } else {
      rides.past.push(ride)
    }
  }

  return rides
})


export const savedRidesGroupedByDate = derived(
  [savedRidesStore],
  ([$savedRides]) => {
    let ridesByDate = new SvelteMap<string, {
      date: CalendarDate
      rides: RideData[]
    }>()
    const numOfRides = $savedRides.data.length
    if (!$savedRides.data && numOfRides === 0) {
      return []
    }
    $savedRides.data.forEach((ride) => {
      const calendarDate = parseDate(ride.date)
      const key = calendarDate.toString()
      if (!ridesByDate.has(key)) {
        ridesByDate.set(key, {
          date: calendarDate,
          rides: []
        })
      }
      ridesByDate.get(key)!.rides.push(ride);
    });
    return Array.from(ridesByDate.values()).sort((a, b) => a.date.compare(b.date));
  }
)

export const selectedSaveRidesNagivationDate = writable(today(getLocalTimeZone()))
export const allSavedRidesNavigationDates = derived(
  [savedRidesGroupedByDate],
  ([$savedRidesGroupedByDate]) => {
    const uniqueDatesMap = new SvelteMap<string, CalendarDate>()
    $savedRidesGroupedByDate.forEach(group => {
      uniqueDatesMap.set(group.date.toString(), group.date)
    })

    const todaysDate = today(getLocalTimeZone())
    uniqueDatesMap.set(todaysDate.toString(), todaysDate)

    return Array.from(uniqueDatesMap.values()).sort((a, b) => a.compare(b))
  }
)

export const savedRidesForSelectedDay = derived(
  [savedRidesGroupedByDate, selectedSaveRidesNagivationDate],
  ([$groupedRides, $selectedDate]) => {
    if (!$groupedRides || $groupedRides.length === 0) {
      return []
    }
    const dayGroup = $groupedRides.find(group => group.date.compare($selectedDate) === 0)
    return dayGroup ? dayGroup.rides : []
  }
)




// NAVIGATION STORE

const initialPrimaryView = getInitialPrimaryView()

export const viewStack = writable<string[]>([initialPrimaryView])
export const activeView = writable<string>(initialPrimaryView)

viewStack.subscribe(stack => {
  if (stack.length > 0) {
    const nextActiveView = stack[stack.length - 1]
    activeView.set(nextActiveView)

    if (typeof window !== "undefined" && PRIMARY_VIEWS.includes(nextActiveView)) {
      localStorage.setItem(LAST_PRIMARY_VIEW_KEY, nextActiveView)
    }
  } else {
    activeView.set(VIEW_MAP)
    viewStack.set([VIEW_MAP])
  }
})

// sets the next view on top of a stack to be able to return to the view a user was before
export function navigateTo(newViewIdentifier: string, options = { force: false }) {
  viewStack.update(stack => {
    if (!options.force && stack[stack.length - 1] === newViewIdentifier) {
      return stack
    }

    return [...stack, newViewIdentifier]
  })

}

export function jumpToView(targetViewIdentifier: string) {
  viewStack.update(stack => {
    const index = stack.lastIndexOf(targetViewIdentifier)
    if (index !== -1) {
      return stack.slice(0, index + 1) as string[]
    }
    return stack
  })

  return [VIEW_MAP]
}

// go back to the previous View
export function goBackInHistory() {
  viewStack.update(stack => {
    if (stack.length > 1) {
      stack.pop()
    }
    return stack
  })
}


// DATE STORE

const initialDate = today(getLocalTimeZone())
export const currentDate = writable<CalendarDate>(initialDate)

export const dateStore = {
  subscribe: currentDate.subscribe,
  setToday: () => {
    clearRideForDateChange()
    currentDate.set(today(getLocalTimeZone()))
  },
  addDays: (offset: number) => {
    clearRideForDateChange()
    currentDate.update((currentStoredDate) => {
      if (!currentStoredDate || typeof currentStoredDate.add !== 'function') {
        return today(getLocalTimeZone()).add({ days: offset })
      }
      return currentStoredDate.add({ days: offset })
    })
  },
  subtractDays: (offset: number) => {
    clearRideForDateChange()
    currentDate.update((currentStoredDate) => {
      if (!currentStoredDate || typeof currentStoredDate.subtract !== 'function') {
        return today(getLocalTimeZone()).subtract({ days: offset })
      }
      return currentStoredDate.subtract({ days: offset })
    })
  },
  setSpecificDate: (date: CalendarDate | null) => {
    clearRideForDateChange()
    if (date && typeof date.add === 'function' && typeof date.subtract === 'function') {
      currentDate.set(date)
    } else if (!date) {
      currentDate.set(today(getLocalTimeZone()))
    }
  }
}

function clearRideForDateChange() {
  currentRide.set(initialRideState)
  rawMapStore.update(store => ({
    ...store,
    showCurrentRide: false
  }))
}


export const formattedDate = derived([currentDate], ([$currentDate]) => {
  if (!$currentDate) {
    return "LoadingDate"
  }

  // Guard against invalid CalendarDate objects
  try {
    const dateFormatter = new DateFormatter("en-US", {
      weekday: "short",
      month: "short",
      day: "numeric",
    });

    const todaysDate = today(getLocalTimeZone());
    const tomorrowsDate = todaysDate?.add({ days: 1 });
    const yesterdaysDate = todaysDate?.subtract({ days: 1 });
    if ($currentDate.compare(todaysDate) === 0) {
      return "Today";
    } else if ($currentDate.compare(tomorrowsDate) === 0) {
      return "Tomorrow";
    } else if ($currentDate.compare(yesterdaysDate) === 0) {
      return "Yesterday";
    } else {
      return dateFormatter.format($currentDate.toDate(getLocalTimeZone()));
    }
  } catch (e) {
    return "LoadingDate";
  }
})


// CURRENT RIDE STORE
const initialRideState = null
export const currentRide = writable<RideData | null>(initialRideState)

export const currentRideStore = {
  subscribe: currentRide.subscribe,
  setRide: function(ride: RideData) {
    currentRide.set(ride)
  },
  getRide: function() {
    if (get(currentRide) === initialRideState) {
      return initialRideState
    } else {
      return get(currentRide)
    }
  },

  clearRide: function clearRide() {
    currentRide.set(initialRideState)
  }
}

export const currentRoute = derived(
  [currentRide, routesStore],
  ([$currentRide, $routesStore]) => {
    if ($currentRide && $currentRide.route_id) {
      return $routesStore.routes.get($currentRide.route_id)
    }
    return null
  }
)

// ALL RIDES STORE
export const rides = createRidesStore()
export const ridesWithoutLocations = derived(
  [rides, currentDate],
  ([$rides, $currentDate]) => {
    if (!$rides || !$rides.rideData || !$currentDate) {
      return [];
    }

    return $rides.rideData.filter(ride => {
      const rideDate = parseDate(ride.date);

      const lat = ride.lat
      const lon = ride.lng

      // if lat, lon undefined, null or == 0 lets not show them on the map
      const isLatOrLonMissing = (lat === 0 || lat === undefined || lat === null || lon === 0 || lon === undefined || lon === null)

      // if lat, lon == the fallback lets also not show them on the map
      const isFallbackCoords = (lat === FALLBACK_LAT && lon === FALLBACK_LNG)

      const hasNoValidAddress = isLatOrLonMissing || isFallbackCoords

      const isSameDayAsCurrent = $currentDate.compare(rideDate) === 0

      return isSameDayAsCurrent && hasNoValidAddress
    })
  }
)

export const ridesWithLocations = derived(
  [rides, currentDate],
  ([$rides, $currentDate]) => {
    if (!$rides || !$rides.rideData || !$currentDate) {
      return [];
    }

    return $rides.rideData.filter(ride => {
      const rideDate = parseDate(ride.date);

      const lat = ride.lat
      const lon = ride.lng

      // if lat, lon are not valid lets not show them on the map
      const hasValidAddress = (
        lat !== undefined && lat !== null && lat !== 0 && lon !== undefined && lon !== null && lon !== 0 && !(lat === FALLBACK_LAT && lon === FALLBACK_LNG)
      )

      const isSameDayAsCurrent = $currentDate.compare(rideDate) === 0

      return isSameDayAsCurrent && hasValidAddress
    });
  }
);

export const todaysRides = derived(
  [rides, currentDate],
  ([$rides, $currentDate]) => {
    if (!$rides || !$rides.rideData || !$currentDate) {
      return [];
    }

    return $rides.rideData.filter(ride => {
      const rideDate = parseDate(ride.date);

      const isSameDayAsCurrent = $currentDate.compare(rideDate) === 0

      return isSameDayAsCurrent
    });
  }
)

export const allUpcomingAdultOnlyRides = derived([rides], ([$rides]) => {
  if (!$rides || !$rides.rideData) {
    return [];
  }


  return $rides.rideData.filter(ride => {
    const rideDate = parseDate(ride.date)
    //
    const isTodayOrUpcoming = initialDate.compare(rideDate) <= 0
    const isAdultsOnlyRide = ride.audience === "A"

    return isAdultsOnlyRide && isTodayOrUpcoming
  })

})

export const allUpcomingFamilyFriendlyRides = derived([rides], ([$rides]) => {
  if (!$rides || !$rides.rideData) {
    return [];
  }


  return $rides.rideData.filter(ride => {
    const rideDate = parseDate(ride.date)
    //
    const isTodayOrUpcoming = initialDate.compare(rideDate) <= 0
    const isFamilyFriendlyRide = ride.audience === "F"

    return isFamilyFriendlyRide && isTodayOrUpcoming
  })

})

export const allUpcomingCovidSafetyRides = derived([rides], ([$rides]) => {
  if (!$rides || !$rides.rideData) {
    return [];
  }


  return $rides.rideData.filter(ride => {
    const rideDate = parseDate(ride.date)
    //
    const isTodayOrUpcoming = initialDate.compare(rideDate) <= 0
    const isCovidSafetyRide = ride.safetyplan

    return isCovidSafetyRide && isTodayOrUpcoming
  })

})


// MAP STORE
//
interface MapViewStore {
  showCurrentRide: boolean
  showNoLocationRideCard: boolean
}
//
const rawMapStore = writable<MapViewStore>({
  showCurrentRide: false,
  showNoLocationRideCard: false
})

export const validRides = derived([ridesWithLocations], ([$rides]) => {
  return $rides.filter(ride => !isNaN(ride.lat as number) && !isNaN(ride.lng as number)).map<ValidatedRide>(ride => ({
    id: ride.id,
    name: ride.title,
    lat: ride.lat as number,
    lng: ride.lng as number,
    marker_key: ride.group_marker
  }))
})

export const rideGeoJSON = derived(
  validRides, ($rides) => {

    const seenCoords: Record<string, number> = {}
    const features = new Array($rides.length)

    for (let i = 0; i < $rides.length; i++) {
      const ride = $rides[i]
      let lng = ride.lng
      let lat = ride.lat

      let key = `${lng}_${lat}`
      let dupCount = seenCoords[key] ?? 0

      if (dupCount > 0) {
        const offset = offsetDuplicateCoordinate(lat, dupCount)
        lat += offset.lat
        lng += offset.lng
      }

      seenCoords[key] = dupCount + 1

      const properties: any = {
        id: ride.id,
        name: ride.name,
      };

      if (ride.marker_key) {
        properties.group_marker_icon = `group-marker-${ride.marker_key}`;
      }

      features[i] = {
        type: "Feature",
        geometry: { type: "Point", coordinates: [lng, lat] },
        properties
      }
    }

    return {
      type: "FeatureCollection",
      features
    } as GeoJSON.FeatureCollection<GeoJSON.Point, any>;
  }
)

function offsetDuplicateCoordinate(lat: number, duplicateIndex: number) {
  const metersPerDegree = 111_320
  const angle = duplicateIndex * 2.399963229728653 // golden angle in radians
  const ring = Math.floor((duplicateIndex - 1) / 8)
  const radiusMeters = 8 + ring * 6
  const latOffset = (Math.sin(angle) * radiusMeters) / metersPerDegree
  const lngMetersPerDegree = metersPerDegree * Math.max(Math.cos(lat * Math.PI / 180), 0.01)
  const lngOffset = (Math.cos(angle) * radiusMeters) / lngMetersPerDegree

  return { lat: latOffset, lng: lngOffset }
}

export const STARTING_ZOOM = 12
export const SINGLE_RIDE_ZOOM = 16
export const MAX_FIT_BOUNDS_ZOOM = 15


function createMapStore() {
  const { subscribe, update } = rawMapStore

  const fitMaptoRides = (map: Map, rides: ValidatedRide[], resetOrientation = false) => {
    if (!map || !rides) return;

    const orientation = resetOrientation ? { bearing: 0, pitch: 0 } : {}

    if (rides.length === 0) {
      map.flyTo({
        center: [STARTING_LNG, STARTING_LAT],
        zoom: STARTING_ZOOM,
        essential: true,
        duration: 1000,
        ...orientation,
      });
      return;
    }

    if (rides.length === 1) {
      map.flyTo({
        center: rides[0],
        zoom: SINGLE_RIDE_ZOOM,
        essential: true,
        duration: 1000,
        ...orientation,
      });

      return;
    }

    const bounds = new LngLatBounds();
    rides.forEach((ride) => bounds.extend(ride));
    map.fitBounds(bounds, { padding: 100, duration: 800, maxZoom: MAX_FIT_BOUNDS_ZOOM, ...orientation });
  }

  return {
    subscribe: subscribe,
    showCurrentRide: (bool: boolean) => {
      update(store => ({
        ...store,
        showCurrentRide: bool
      }))
    },
    showNoLocationsRides: (bool: boolean) => {
      update(store => ({
        ...store,
        showNoLocationRideCard: bool
      }))
    },
    getRideById: (rideId: string) => {
      return get(todaysRides).filter(ride => ride.id === rideId)[0]
    },
    fitMap: (mapInstance: Map, resetOrientation = false) => {
      const currentValidCoords = get(validRides)

      if (currentValidCoords.length === 0) {
        fitMaptoRides(mapInstance, currentValidCoords, resetOrientation)
      }
      setTimeout(
        () => {
          fitMaptoRides(mapInstance, currentValidCoords, resetOrientation)
        }, 50
      )
    },
    flyToSelected: (mapInstance: Map, offset: [number, number] = [0, 0]) => {
      if (!mapInstance) return

      update(store => ({ ...store, isPreformingSpecificAction: true }))
      const selected = get(currentRide)


      if (selected) {
        const coords: LngLatLike = [selected.lng as number, selected.lat as number]
        mapInstance.flyTo({
          zoom: SINGLE_RIDE_ZOOM,
          center: coords,
          offset,
          duration: 900,
          essential: true
        })
      }
    }
  }
}

export const selectedRideId = derived(currentRide, ($ride) => $ride && $ride.id || "")

// GeoJSON for a single ride without offset applied (for individual ride maps)
export const singleRideGeoJSON = derived(
  currentRide, ($currentRide) => {
    if (!$currentRide) {
      return {
        type: "FeatureCollection",
        features: []
      } as GeoJSON.FeatureCollection<GeoJSON.Point, any>;
    }

    const lat = $currentRide.lat as number
    const lng = $currentRide.lng as number

    // Only create a feature if coordinates are valid
    if (!lat || !lng || isNaN(lat) || isNaN(lng)) {
      return {
        type: "FeatureCollection",
        features: []
      } as GeoJSON.FeatureCollection<GeoJSON.Point, any>;
    }

    const properties: any = {
      id: $currentRide.id,
      name: $currentRide.title,
    };

    if ($currentRide.group_marker) {
      properties.group_marker_icon = `group-marker-${$currentRide.group_marker}`;
    }

    return {
      type: "FeatureCollection",
      features: [{
        type: "Feature",
        geometry: { type: "Point", coordinates: [lng, lat] },
        properties
      }]
    } as GeoJSON.FeatureCollection<GeoJSON.Point, any>;
  }
)

export const mapStore = createMapStore()
