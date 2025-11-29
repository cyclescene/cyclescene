import { CalendarDate, parseDate } from "@internationalized/date";
import { LngLatBounds, type LngLatLike, type Map } from "maplibre-gl"
import { getPastRides, getUpcomingRides } from "./api";
import { addSavedRide, deleteSavedRide, getAllSavedRides, getRidesfromDB, savedRideExists, saveRidesToDB, clearAllRides, clearSavedRides } from "./db";
import { today, getLocalTimeZone, DateFormatter } from "@internationalized/date";
import { writable, derived, get } from "svelte/store";
import { SvelteMap } from "svelte/reactivity";
import { STARTING_LAT, STARTING_LNG } from "./config";
import { type ValidatedRide, type RideData } from "./types";
import type * as GeoJSON from "geojson";

// Portland, OR coordinates
const FALLBACK_LAT = STARTING_LAT
const FALLBACK_LNG = STARTING_LNG

// views
export const VIEW_MAP = 'map'
export const VIEW_LIST = 'list'
export const VIEW_RIDE_DETAILS = 'rideDetails'
export const VIEW_SAVED = 'saved'
export const VIEW_SETTINGS = 'settings'
export const VIEW_OTHER_RIDES = 'otherRides'
export const VIEW_DATE_PICKER = 'datePicker'

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
]

export const TILE_URLS = {
  dark: "https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json",
  light: "https://basemaps.cartocdn.com/gl/positron-gl-style/style.json"
};

const RIDES_SYNC_TAG = "update-rides-6hr"
const SYNC_INTERVAL = 6 * 60 * 60 * 1000

function createRidesStore() {
  const { subscribe, set, update } = writable<{
    loading: boolean,
    rideData: RideData[],
    error: String | null
  }>({
    loading: true,
    rideData: [],
    error: null
  })


  function updateStoreAndDB(freshRides: RideData[]) {
    saveRidesToDB(freshRides).then(() => {
      getRidesfromDB()
        .then((rides) => (set({ loading: false, rideData: rides, error: null })))
        .catch(e => {
          update((store) => ({
            ...store, loading: false, error: `${e}`
          }))
        })
    })
      .catch(err => {
        console.error("Failed to save fresh rides to IDB: ", err);
      })
  }

  if ('serviceWorker' in navigator) {
    const swMessageListener = (event: MessageEvent) => {
      const data = event.data
      if (data.type === "RIDES_UPDATE_SUCCESSFULL" && data.data) {
        updateStoreAndDB(data.data)
      }
    }
    navigator.serviceWorker.addEventListener('message', swMessageListener)

  }

  return {
    subscribe,
    init: async () => {
      try {
        let cachedRides = await getRidesfromDB()
        // only do a manual fetch if cached rides are empty
        if (cachedRides.length === 0) {
          const upcomingRides = await getUpcomingRides()
          const pastRides = await getPastRides()
          const freshRides = [...upcomingRides, ...pastRides]
          await saveRidesToDB(freshRides)
          // Update store with fresh rides instead of reloading page
          cachedRides = freshRides
        }

        set({ loading: false, rideData: cachedRides, error: null })
      } catch (err) {
        update(store => ({ ...store, loading: false, error: "Unable to get idb ride data" }))
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
    refetch: async () => {
      try {
        update(store => ({ ...store, loading: true }))
        const cachedRides = await getRidesfromDB()
        set({ loading: false, rideData: cachedRides, error: null })
      } catch (e) {
        update(() => ({ loading: false, rideData: [], error: `${e}` }))
        console.error(e);
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
        const freshRides = [...upcomingRides, ...pastRides]
        // Save to IndexedDB
        await saveRidesToDB(freshRides)
        // Update the store
        set({ loading: false, rideData: freshRides, error: null })
      } catch (e) {
        update(() => ({ loading: false, rideData: [], error: `Failed to refresh rides: ${e}` }))
        console.error(e);
      }
    }
  }
}

export function triggerForegroundSync() {
  if ('serviceWorker' in navigator && navigator.serviceWorker.controller) {
    navigator.serviceWorker.controller.postMessage({
      type: "FORCE_FOREGROUND_SYNC"
    })

    console.log("Forground sync requested by user");

  } else {
    alert("Cannot connect to the backgoung worker. Please check PWA installation")
  }
}

interface SavedRideStore {
  loading: boolean;
  data: RideData[];
  error: string | null
}

function createSavedRideStore() {
  const { subscribe, set } = writable<SavedRideStore>({
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
        console.error(e);
        set({ loading: false, data: [], error: "Could not load saved rides" })
      }
    },
    saveRide: async (ride: RideData) => {
      try {
        await addSavedRide(ride)
        const savedRides = await getAllSavedRides()
        set({ loading: false, data: savedRides, error: null })

      } catch (e) {
        set({ loading: false, data: [], error: "Could not load saved rides" })
      }
    },
    deleteRide: async (rideID: string) => {
      try {
        await deleteSavedRide(rideID)
        const savedRides = await getAllSavedRides()
        set({ loading: false, data: savedRides, error: null })
      } catch (e) {
        set({ loading: false, data: [], error: "Could not load saved rides" })
      }

    },
    isRideSaved: async (rideID: string) => {
      try {
        const exists = await savedRideExists(rideID)
        return exists
      } catch (e) {
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
      console.log("Past?");
      rides.upcoming.push(ride)
    } else {
      console.log("UPCOMING");
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

export const viewStack = writable<string[]>([VIEW_MAP])
export const activeView = writable<string>(VIEW_MAP)

viewStack.subscribe(stack => {
  if (stack.length > 0) {
    activeView.set(stack[stack.length - 1])
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
      return stack.slice(0, index + 1)
    }
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
    currentDate.set(today(getLocalTimeZone()))
  },
  addDays: (offset: number) => {
    currentDate.update((currentStoredDate) => {
      if (!currentStoredDate) {
        console.error("dateStore.addDays: current store was undefined/null. " +
          "initializing to today and applying offset");
        return today(getLocalTimeZone()).add({ days: offset })
      }
      return currentStoredDate.add({ days: offset })
    })
  },
  subtractDays: (offset: number) => {
    currentDate.update((currentStoredDate) => {
      if (!currentStoredDate) {
        console.error("dateStore.addDays: current store was undefined/null. " +
          "initializing to today and applying offset");
        return today(getLocalTimeZone()).add({ days: offset })
      }
      return currentStoredDate.subtract({ days: offset })
    })
  },
  setSpecificDate: (date: CalendarDate) => {
    currentDate.set(date)
  }
}


export const formattedDate = derived([currentDate], ([$currentDate]) => {
  if (!$currentDate) {
    console.warn("formattedDate derived store received undefined/null currentDate.");
    return "LoadingDate"
  }



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
    lng: ride.lng as number
  }))
})

export const rideGeoJSON = derived(
  [validRides, ridesWithLocations], ($rides, $allRides) => {

    const seenCoords: Record<string, number> = {}
    const features = new Array($rides.length)

    for (let i = 0; i < $rides.length; i++) {
      const ride = $rides[i]
      // Find the original ride to get group_marker info
      const originalRide = $allRides.find(r => r.id === ride.id);
      const groupMarker = originalRide?.group_marker ? `group-marker-${originalRide.group_marker}` : "";

      let lng = ride.lng
      let lat = ride.lat

      let key = `${lng}_${lat}`
      let dupCount = seenCoords[key] ?? 0

      if (dupCount > 0) {
        const offset = dupCount * 0.0009
        lat += offset
        lng += offset
        key = `${lng}_${lat}`
      }

      seenCoords[key] = dupCount + 1

      features[i] = {
        type: "Feature",
        geometry: { type: "Point", coordinates: [lng, lat] },
        properties: { id: ride.id, name: ride.name, group_marker_icon: groupMarker }
      }
    }

    return {
      type: "FeatureCollection",
      features
    } as GeoJSON.FeatureCollection<GeoJSON.Point, any>;
  }
)

export const STARTING_ZOOM = 12
export const SINGLE_RIDE_ZOOM = 16


function createMapStore() {
  const { subscribe, update } = rawMapStore

  const fitMaptoRides = (map: Map, rides: ValidatedRide[]) => {
    if (!map || !rides) return;

    if (rides.length === 0) {
      map.flyTo({
        center: [STARTING_LNG, STARTING_LAT],
        zoom: STARTING_ZOOM,
        essential: true,
        duration: 1000,
      });
      return;
    }

    if (rides.length === 1) {
      map.flyTo({
        center: rides[0],
        zoom: SINGLE_RIDE_ZOOM,
        essential: true,
        duration: 1000,
      });

      return;
    }

    const bounds = new LngLatBounds();
    rides.forEach((ride) => bounds.extend(ride));
    map.fitBounds(bounds, { padding: 100, duration: 800 });
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
    fitMap: (mapInstance: Map) => {
      const currentValidCoords = get(validRides)

      if (currentValidCoords.length === 0) {
        fitMaptoRides(mapInstance, currentValidCoords)
      }
      setTimeout(
        () => {
          fitMaptoRides(mapInstance, currentValidCoords)
        }, 50
      )
    },
    flyToSelected: (mapInstance: Map) => {
      if (!mapInstance) return

      update(store => ({ ...store, isPreformingSpecificAction: true }))
      const selected = get(currentRide)


      if (selected) {
        const coords: LngLatLike = [selected.lng as number, selected.lat as number]
        mapInstance.flyTo({
          zoom: SINGLE_RIDE_ZOOM,
          center: coords,
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

    const groupMarker = $currentRide.group_marker ? `group-marker-${$currentRide.group_marker}` : "";

    return {
      type: "FeatureCollection",
      features: [{
        type: "Feature",
        geometry: { type: "Point", coordinates: [lng, lat] },
        properties: { id: $currentRide.id, name: $currentRide.title, group_marker_icon: groupMarker }
      }]
    } as GeoJSON.FeatureCollection<GeoJSON.Point, any>;
  }
)

export const mapStore = createMapStore()
