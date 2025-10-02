import { parseDate } from "@internationalized/date";
import { getPastRides, getUpcomingRides } from "./api";
import { addSavedRide, deleteSavedRide, getAllSavedRides, getRidesfromDB, savedRideExists, saveRidesToDB } from "./db";
import { today, getLocalTimeZone, DateFormatter } from "@internationalized/date";
import { writable, derived, get } from "svelte/store";
import { SvelteMap, SvelteSet } from "svelte/reactivity";
import { STARTING_LAT, STARTING_LON } from "./config";



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

const SUB_VIEWS_SET = new SvelteSet(SUB_VIEWS)



export const TILE_URLS = {
  dark: "https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json",
  light: "https://basemaps.cartocdn.com/gl/positron-gl-style/style.json"
};



function createRidesStore() {
  const { subscribe, set, update } = writable({
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
  const { subscribe, set } = writable({
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
    saveRide: async (ride) => {
      try {
        await addSavedRide(ride)
        const savedRides = await getAllSavedRides()
        set({ loading: false, data: savedRides, error: null })

      } catch (e) {
        set({ loading: false, data: [], error: "Could not load saved rides" })
      }
    },
    deleteRide: async (rideID) => {
      try {
        await deleteSavedRide(rideID)
        const savedRides = await getAllSavedRides()
        set({ loading: false, data: savedRides, error: null })
      } catch (e) {
        set({ loading: false, data: [], error: "Could not load saved rides" })
      }

    },
    isRideSaved: async (rideID) => {
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
    let ridesByDate = new SvelteMap()
    if (!$savedRides.data && $savedRides.data.length === 0) {
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
      ridesByDate.get(key).rides.push(ride);
    });
    return Array.from(ridesByDate.values()).sort((a, b) => a.date.compare(b.date));
  }
)

export const selectedSaveRidesNagivationDate = writable(today(getLocalTimeZone()))
export const allSavedRidesNavigationDates = derived(
  [savedRidesGroupedByDate],
  ([$savedRidesGroupedByDate]) => {
    const uniqueDatesMap = new SvelteMap()

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

export const viewStack = writable([VIEW_MAP])
export const activeView = writable(VIEW_MAP)

viewStack.subscribe(stack => {
  if (stack.length > 0) {
    activeView.set(stack[stack.length - 1])
  } else {
    activeView.set(VIEW_MAP)
    viewStack.set([VIEW_MAP])
  }
})

// sets the next view on top of a stack to be able to return to the view a user was before
export function navigateTo(newViewIdentifier, options = { force: false }) {
  viewStack.update(stack => {
    if (!options.force && stack[stack.length - 1] === newViewIdentifier) {
      return stack
    }

    return [...stack, newViewIdentifier]
  })

}

export function jumpToView(targetViewIdentifier) {
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
export const currentDate = writable(initialDate)

export const dateStore = {
  subscribe: currentDate.subscribe,
  setToday: () => {
    currentDate.set(today(getLocalTimeZone()))
  },
  addDays: (offset) => {
    currentDate.update((currentStoredDate) => {
      if (!currentStoredDate) {
        console.error("dateStore.addDays: current store was undefined/null. " +
          "initializing to today and applying offset");
        return today(getLocalTimeZone()).add({ days: offset })
      }
      return currentStoredDate.add({ days: offset })
    })
  },
  subtractDays: (offset) => {
    currentDate.update((currentStoredDate) => {
      if (!currentStoredDate) {
        console.error("dateStore.addDays: current store was undefined/null. " +
          "initializing to today and applying offset");
        return today(getLocalTimeZone()).add({ days: offset })
      }
      return currentStoredDate.subtract({ days: offset })
    })
  },
  setSpecificDate: (date) => {
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
export const currentRide = writable(initialRideState)

export const currentRideStore = {
  subscribe: currentRide.subscribe,
  setRide: function(ride) {
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
function createMapViewStore() {
  const { subscribe, update } = writable({
    eventCardsVisible: false,
    otherRidesVisible: false,
    selectedEvent: null,
    otherRides: []
  })

  return {
    subscribe: subscribe,
    showEventCards: (bool) => {
      update(store => ({
        ...store,
        eventCardsVisible: bool
      }))
    },
    showOtherRides: (bool) => {
      update(store => ({
        ...store,
        otherRidesVisible: bool
      }))
    },
    setSelectedRide: (ride) => {
      update(store => ({
        ...store,
        selectedEvent: ride
      }))
    },
    clearSelectedRides: () => {
      update(store => ({
        ...store,
        selectedEvent: null
      }))
    },
    setOtherRides: (rides) => {
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
    }
  }
}

export const mapViewStore = createMapViewStore()
