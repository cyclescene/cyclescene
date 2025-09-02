import { openDB } from 'idb'

const DB_NAME = 'bike-bae-db'
const RIDES_STORE_NAME = 'rides'
const DB_VERSION = 1

// Initialize the database and create the object store if it doesn't exist
const dbPromise = openDB(DB_NAME, DB_VERSION, {
    upgrade(db) {
        if (!db.objectStoreNames.contains(RIDES_STORE_NAME)) {
            db.createObjectStore(RIDES_STORE_NAME, { keyPath: 'id' })
        }
    }
})

/**
 * Saves an array of ride objects to IndexedDB, orverwriting existing data.
 * @param {Array<Object>} rides - array of rides to save
 */
export async function saveRidesToDB(rides) {
    const db = await dbPromise
    const tx = db.transaction(RIDES_STORE_NAME, "readwrite")
    await tx.store.clear()
    await Promise.all(rides.map(ride => tx.store.add(ride)))
    await tx.done
}

/**
 * Retrieves all rides from IndexedDB
 * @returns {Promise<Array<Object>>} - an array of ride objects
 */
export async function getRidesfromDB() {
    const db = await dbPromise
    return await db.getAll(RIDES_STORE_NAME)
}
