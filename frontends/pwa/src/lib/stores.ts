import { CalendarDate, parseDate } from "@internationalized/date";
import { LngLatBounds, type Map } from "maplibre-gl"
import { getPastRides, getUpcomingRides } from "./api";
import { addSavedRide, deleteSavedRide, getAllSavedRides, getRidesfromDB, savedRideExists, saveRidesToDB } from "./db";
import { today, getLocalTimeZone, DateFormatter } from "@internationalized/date";
import { writable, derived, get } from "svelte/store";
import { SvelteMap } from "svelte/reactivity";
import { STARTING_LAT, STARTING_LON } from "./config";
import { type ValidatedRide, type RideData } from "./types";
import type * as GeoJSON from "geojson";



// Portland, OR coordinates
const FALLBACK_LAT = STARTING_LAT
const FALLBACK_LON = STARTING_LON

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

function createRidesStore() {
  const { subscribe, set, update } = writable<{
    loading: boolean,
    data: RideData[],
    error: String | null

  }>({
    loading: true,
    data: [],
    error: null
  })

  return {
    subscribe,
    init: async () => {
      try {
        const cachedRides = await getRidesfromDB()
        set({ loading: false, data: cachedRides, error: null })
      } catch (error) {
        set({ loading: false, data: [], error: "Could not load cached rides" })
      }
    },
    fetchUpcoming: async () => {
      try {
        const freshRides = await getUpcomingRides();
        set({ loading: false, data: freshRides, error: null })
        await saveRidesToDB(freshRides)
      } catch (error) {
        update(store => ({ ...store, loading: false, error: `API fetch failed, relying on cached data: ${error}` }))
      }
    },
    fetchPast: async () => {
      try {
        const freshRides = await getPastRides();
        set({ loading: false, data: freshRides, error: null })
        await saveRidesToDB(freshRides)
      } catch (error) {
        update(store => ({ ...store, loading: false, error: `API fetch failed, relying on cached data: ${error}` }))
      }
    }
  }
}

function createSavedRideStore() {
  const { subscribe, set } = writable<{
    loading: boolean,
    data: RideData[],
    error: String | null

  }>({
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
    deleteRide: async (rideID: String) => {
      try {
        await deleteSavedRide(rideID)
        const savedRides = await getAllSavedRides()
        set({ loading: false, data: savedRides, error: null })
      } catch (e) {
        set({ loading: false, data: [], error: "Could not load saved rides" })
      }

    },
    isRideSaved: async (rideID: String) => {
      try {
        const exists = await savedRideExists(rideID)
        return exists
      } catch (e) {
      }
    }
  }
}

export const savedRidesStore = createSavedRideStore()
export const allSavedRides = derived(
  [savedRidesStore],
  ([$savedRides]) => {
    if (!$savedRides || !$savedRides.data) {
      return [];
    }


    return $savedRides.data
  }
)

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
    if (currentRide == initialRideState) {
      return null
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
export const filteredNoAddress = derived(
  [rides, currentDate],
  ([$rides, $currentDate]) => {
    if (!$rides || !$rides.data || !$currentDate) {
      return [];
    }

    return $rides.data.filter(ride => {
      const rideDate = parseDate(ride.date);

      const lat = ride.lat?.Float64
      const lon = ride.lon?.Float64

      // if lat, lon undefined, null or == 0 lets not show them on the map
      const isLatOrLonMissing = (lat === 0 || lat === undefined || lat === null || lon === 0 || lon === undefined || lon === null)

      // if lat, lon == the fallback lets also not show them on the map
      const isFallbackCoords = (lat === FALLBACK_LAT && lon === FALLBACK_LON)

      const hasNoValidAddress = isLatOrLonMissing || isFallbackCoords

      const isSameDayAsCurrent = $currentDate.compare(rideDate) === 0

      return isSameDayAsCurrent && hasNoValidAddress
    })
  }
)

export const filteredRides = derived(
  [rides, currentDate],
  ([$rides, $currentDate]) => {
    if (!$rides || !$rides.data || !$currentDate) {
      return [];
    }

    return $rides.data.filter(ride => {
      const rideDate = parseDate(ride.date);

      const lat = ride.lat?.Float64
      const lon = ride.lon?.Float64

      // if lat, lon are not valid lets not show them on the map
      const hasValidAddress = (
        lat !== undefined && lat !== null && lat !== 0 && lon !== undefined && lon !== null && lon !== 0 && !(lat === FALLBACK_LAT && lon === FALLBACK_LON)
      )

      const isSameDayAsCurrent = $currentDate.compare(rideDate) === 0

      return isSameDayAsCurrent && hasValidAddress
    });
  }
);

export const allRides = derived(
  [rides, currentDate],
  ([$rides, $currentDate]) => {
    if (!$rides || !$rides.data || !$currentDate) {
      return [];
    }

    return $rides.data.filter(ride => {
      const rideDate = parseDate(ride.date);

      const isSameDayAsCurrent = $currentDate.compare(rideDate) === 0

      return isSameDayAsCurrent
    });
  }
)

export const allUpcomingAdultOnlyRides = derived([rides], ([$rides]) => {
  if (!$rides || !$rides.data) {
    return [];
  }


  return $rides.data.filter(ride => {
    const rideDate = parseDate(ride.date)
    //
    const isTodayOrUpcoming = initialDate.compare(rideDate) <= 0
    const isAdultsOnlyRide = ride.audience === "A"

    return isAdultsOnlyRide && isTodayOrUpcoming
  })

})

export const allUpcomingFamilyFriendlyRides = derived([rides], ([$rides]) => {
  if (!$rides || !$rides.data) {
    return [];
  }


  return $rides.data.filter(ride => {
    const rideDate = parseDate(ride.date)
    //
    const isTodayOrUpcoming = initialDate.compare(rideDate) <= 0
    const isFamilyFriendlyRide = ride.audience === "F"

    return isFamilyFriendlyRide && isTodayOrUpcoming
  })

})

export const allUpcomingCovidSafetyRides = derived([rides], ([$rides]) => {
  if (!$rides || !$rides.data) {
    return [];
  }


  return $rides.data.filter(ride => {
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
  eventCardsVisible: boolean,
  otherRidesVisible: boolean,
  selectedEvent?: RideData | null,
  primaryRides: RideData[],
  otherRides: RideData[]
}
//
const rawMapStore = writable<MapViewStore>()

export const validRides = derived(rawMapStore, ($state) => {
  return $state.primaryRides.filter(ride => ride.lon.Valid && ride.lat.Valid && !isNaN(ride.lat.Float64 as number) && !isNaN(ride.lon.Float64 as number)).map<ValidatedRide>(ride => ({
    id: ride.id,
    name: ride.title,
    lat: ride.lat.Float64,
    lng: ride.lon.Float64
  }))
})

export const rideGeoJSON = derived(
  validRides, ($rides) => {
    const geoJson = {
      type: "FeatureCollection",
      features: $rides.map(coord => ({
        type: "Feature" as const,
        geometry: {
          type: "Point" as const,
          coordinates: [coord.lng, coord.lat]
        },
        properties: {
          id: coord.id,
          name: coord.name
        }
      }))
    }

    return geoJson as GeoJSON.FeatureCollection<GeoJSON.Point, any>;
  }
)

export const STARTING_ZOOM = 12
const SINGLE_RIDE_ZOOM = 16

function createMapStore() {
  const { subscribe, update } = rawMapStore

  const fitMaptoRides = (map: Map, coords: ValidatedRide[]) => {
    if (!map || !coords) return;

    if (coords.length === 0) {
      map.flyTo({
        center: [STARTING_LON, STARTING_LAT],
        zoom: STARTING_ZOOM,
        essential: true,
        duration: 800,
      });
      return;
    }

    if (coords.length === 1) {
      map.flyTo({
        center: coords[0],
        zoom: SINGLE_RIDE_ZOOM,
        essential: true,
        duration: 800,
      });

      return;
    }

    const bounds = new LngLatBounds();
    coords.forEach((coord) => bounds.extend(coord));
    map.fitBounds(bounds, { padding: 100, duration: 800 });
  }

  return {
    subscribe: subscribe,
    showEventCards: (bool: boolean) => {
      update(store => ({
        ...store,
        eventCardsVisible: bool
      }))
    },
    showOtherRides: (bool: boolean) => {
      update(store => ({
        ...store,
        otherRidesVisible: bool
      }))
    },
    setSelectedRide: (ride: RideData) => {
      update(store => ({
        ...store,
        selectedEvent: ride
      }))
    },
    clearSelectedRide: () => {
      update(store => ({
        ...store,
        selectedEvent: null
      }))
    },
    setOtherRides: (rides: RideData[]) => {
      update(store => ({
        ...store,
        otherRides: rides
      }))
    },
    clearOtherRides: () => {
      update(store => ({
        ...store,
        otherRides: []
      }))
    },
    setPrimaryRides: (rides: RideData[]) => {
      update(store => ({
        ...store,
        primaryRides: rides
      }))
    },
    fitMap: (mapInstance: Map) => {
      const currentValidCoords = get(validRides)
      fitMaptoRides(mapInstance, currentValidCoords)
    }
  }
}

export const mapStore = createMapStore()
