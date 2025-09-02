import { writable, derived } from "svelte/store";
import { getPastRides, getUpcomingRides } from "./api";
import { getRidesfromDB, saveRidesToDB } from "./db";
import { isSameDay, parseISO } from "date-fns";

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

export const rides = createRidesStore()

export const savedRideIds = writable([])

export const filteredRides = derived(
    [rides, currentDate],
    ([$rides, $currentDate]) => {
        // Check if we have valid data to work with
        if (!$rides || !$rides.data || !$currentDate) {
            return [];
        }

        return $rides.data.filter(ride => {
            const rideDate = parseISO(ride.date);


            return isSameDay($currentDate, rideDate);
        });
    }
);
