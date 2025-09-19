import { parseDate } from "@internationalized/date";
import { getPastRides, getUpcomingRides } from "./api";
import { getRidesfromDB, saveRidesToDB } from "./db";
import { today, getLocalTimeZone, DateFormatter } from "@internationalized/date";
import { writable, derived } from "svelte/store";

// Portland, OR coordinates
const FALLBACK_LAT = 45.515232
const FALLBACK_LON = -122.6783853

// views
export const VIEW_MAP = 'map'
export const VIEW_LIST = 'list'
export const VIEW_RIDE_DETAILS = 'rideDetails'
export const VIEW_SAVED = 'saved'
export const VIEW_SETTINGS = 'settings'
export const VIEW_OTHER_RIDES = 'otherRides'
export const VIEW_DATE_PICKER = "datePicker"


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

export function navigateTo(newViewIdentifier, options = { force: false }) {
    viewStack.update(stack => {
        if (!options.force && stack[stack.length - 1] === newViewIdentifier) {
            return stack
        }
        return [...stack, newViewIdentifier]
    })

}

export function goBackInHistory() {
    viewStack.update(stack => {
        if (stack.length > 1) {
            stack.pop()
        } else {
            console.log("Cannot go further back, staying on current view.")
        }
        return stack
    })
}


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


// current ride logic
const initialRideState = null
export const currentRide = writable(initialRideState)

export function setRide(ride) {
    currentRide.set(ride)
}

export function getRide() {
    if (currentRide == initialRideState) {
        return
    } else {
        return currentRide
    }
}

export function clearRide() {
    currentRide.set(initialRideState)
}

export const rides = createRidesStore()

export const savedRideIds = writable([])

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
