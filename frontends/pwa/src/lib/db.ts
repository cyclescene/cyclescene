import { openDB } from "idb"
import type { RideData } from "./types"
import type { RouteGeoJSON } from "./api"

const DB_NAME = 'cycle-scene-pdx'
const ALLRIDES_STORE_NAME = 'rides'
const SAVED_RIDES_STORE_NAME = "saved"
const SAVED_RIDES_STORE_DATE_INDEX = "dateIndex"
const ROUTES_STORE_NAME = "routes"
const DB_VERSION = 3

// Initialize the database and create the object store if it doesn't exist
// LOGIC FOR ALL RIDES ////////////////////
const dbPromise = openDB(DB_NAME, DB_VERSION, {
  upgrade(db) {
    if (!db.objectStoreNames.contains(ALLRIDES_STORE_NAME)) {
      db.createObjectStore(ALLRIDES_STORE_NAME, { keyPath: 'id' })
    }
    if (!db.objectStoreNames.contains(SAVED_RIDES_STORE_NAME)) {
      const savedRidesStore = db.createObjectStore(SAVED_RIDES_STORE_NAME, { keyPath: 'id' })
      // dateIndex needed to be able to sort by the ride date
      savedRidesStore.createIndex(SAVED_RIDES_STORE_DATE_INDEX, "date", { unique: false })
    }
    if (!db.objectStoreNames.contains(ROUTES_STORE_NAME)) {
      db.createObjectStore(ROUTES_STORE_NAME, { keyPath: 'id' })
    }
  }
})

export async function saveRidesToDB(rides: RideData[]) {
  const db = await dbPromise
  const tx = db.transaction(ALLRIDES_STORE_NAME, "readwrite")
  await tx.objectStore(ALLRIDES_STORE_NAME).clear()
  const results = await Promise.allSettled(rides.map(ride => tx.store.put(ride)))
  await tx.done
}

export async function getRidesfromDB(): Promise<RideData[]> {
  const db = await dbPromise
  return await db.getAll(ALLRIDES_STORE_NAME)
}


// LOGIC for user instance saved rides ////////////////////////////////

// saves a ride to the saved ride store
export async function addSavedRide(ride: RideData) {
  const db = await dbPromise
  const tx = db.transaction(SAVED_RIDES_STORE_NAME, "readwrite")
  try {
    await tx.store.put(ride)
  } catch (e) {
    // Error adding saved ride
  }
  await tx.done
}

// remove a saved ride
export async function deleteSavedRide(rideID: string) {
  const db = await dbPromise
  const tx = db.transaction(SAVED_RIDES_STORE_NAME, "readwrite")
  try {
    await tx.store.delete(rideID)
  } catch (e) {
    // Error deleting saved ride
  }
  await tx.done
}

// get all saved rides
export async function getAllSavedRides() {
  const db = await dbPromise
  const tx = db.transaction(SAVED_RIDES_STORE_NAME)
  let objectStore = tx.objectStore(SAVED_RIDES_STORE_NAME)
  let index = objectStore.index(SAVED_RIDES_STORE_DATE_INDEX)

  return await index.getAll()
}

// in case a user want to clear their saved rides
export async function clearSavedRides() {
  const db = await dbPromise
  return await db.clear(SAVED_RIDES_STORE_NAME)

}

export async function savedRideExists(rideId: string) {
  const db = await dbPromise
  const tx = db.transaction(SAVED_RIDES_STORE_NAME)
  let objectStore = tx.objectStore(SAVED_RIDES_STORE_NAME)
  const result = await objectStore.get(rideId)
  return result !== undefined
}

// clear all rides from the allrides store
export async function clearAllRides() {
  const db = await dbPromise
  return await db.clear(ALLRIDES_STORE_NAME)
}

// LOGIC FOR ROUTES ////////////////////////////////

export async function saveRoutesToDB(routes: RouteGeoJSON[]) {
  const db = await dbPromise
  const tx = db.transaction(ROUTES_STORE_NAME, "readwrite")
  await tx.objectStore(ROUTES_STORE_NAME).clear()
  const results = await Promise.allSettled(routes.map(route => tx.store.put(route)))
  await tx.done
}

export async function getRoutesFromDB(): Promise<RouteGeoJSON[]> {
  const db = await dbPromise
  return await db.getAll(ROUTES_STORE_NAME)
}

export async function getRouteFromDB(routeId: string): Promise<RouteGeoJSON | undefined> {
  const db = await dbPromise
  return await db.get(ROUTES_STORE_NAME, routeId)
}

export async function clearAllRoutes() {
  const db = await dbPromise
  return await db.clear(ROUTES_STORE_NAME)
}
