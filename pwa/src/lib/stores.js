import { writable, derived } from "svelte/store";
import { getPastRides, getUpcomingRides } from "./api";
import { getRidesfromDB, saveRidesToDB } from "./db";
import { isSameDay, parseISO } from "date-fns";

const FALLBACK_LAT = 45.515232
const FALLBACK_LON = -122.6783853

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


export const currentView = writable('map')

export const currentDate = writable(new Date())

export const currentRide = writable(null)

export const rides = createRidesStore()

export const savedRideIds = writable([])

export const filteredNoAddress = derived(
    [rides, currentDate],
    ([$rides, $currentDate]) => {
        if (!$rides || !$rides.data || !$currentDate) {
            return [];
        }

        return $rides.data.filter(ride => {
            const rideDate = parseISO(ride.date);

            const lat = ride.lat?.Float64
            const lon = ride.lon?.Float64

            // if lat, lon undefined, null or == 0 lets not show them on the map
            const isLatOrLonMissing = (lat === 0 || lat === undefined || lat === null || lon === 0 || lon === undefined || lon === null)

            // if lat, lon == the fallback lets also not show them on the map
            const isFallbackCoords = (lat === FALLBACK_LAT && lon === FALLBACK_LON)

            const hasNoValidAddress = isLatOrLonMissing || isFallbackCoords

            const isSameDayAsCurrent = isSameDay($currentDate, rideDate);

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
            const rideDate = parseISO(ride.date);

            const lat = ride.lat?.Float64
            const lon = ride.lon?.Float64

            // if lat, lon are not valid lets not show them on the map
            const hasValidAddress = (
                lat !== undefined && lat !== null && lat !== 0 && lon !== undefined && lon !== null && lon !== 0 && !(lat === FALLBACK_LAT && lon === FALLBACK_LON)
            )

            const isSameDayAsCurrent = isSameDay($currentDate, rideDate);

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
            const rideDate = parseISO(ride.date);

            const isSameDayAsCurrent = isSameDay($currentDate, rideDate);

            return isSameDayAsCurrent
        });
    }
)
