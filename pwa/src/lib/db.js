import { openDB } from 'idb'

const DB_NAME = 'cycle-scene-pdx'
const ALLRIDES_STORE_NAME = 'rides'
const SAVED_RIDES_STORE_NAME = "saved"
const SAVED_RIDES_STORE_DATE_INDEX = "dateIndex"
const DB_VERSION = 2

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
    }
})

/**
 * Saves an array of ride objects to IndexedDB, orverwriting existing data.
 * @param {Array<Object>} rides - array of rides to save
 */
export async function saveRidesToDB(rides) {
    const db = await dbPromise
    const tx = db.transaction(ALLRIDES_STORE_NAME, "readwrite")
    await Promise.all(rides.map(ride => tx.store.put(ride))).catch(e => console.error(e))
    await tx.done
}

/**
 * Retrieves all rides from IndexedDB
 * @returns {Promise<Array<Object>>} - an array of ride objects
 */
export async function getRidesfromDB() {
    const db = await dbPromise
    return await db.getAll(ALLRIDES_STORE_NAME)
}


// LOGIC for user instance saved rides ////////////////////////////////

// saves a ride to the saved ride store
export async function addSavedRide(ride) {
    const db = await dbPromise
    const tx = db.transaction(SAVED_RIDES_STORE_NAME, "readwrite")
    try {
        await tx.store.put(ride)

    } catch (e) {
        console.error(e);
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
