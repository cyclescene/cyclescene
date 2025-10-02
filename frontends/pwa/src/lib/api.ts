const API_BASE_URL = import.meta.env.VITE_API_BASE_URL

/**
 * helper function to handle fetch requests and errors
 * @param {string} endpoint - specific API endpoint to call
 * @param {object} [options={}] Optional fetch options (method, headers, body)
 * @returns {Promise<any>} A promise that resolves to the JSON response data
 * @throws {Error} Throws an error if the network request fails or the server responds with error status
 */
async function apiFetch(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`

  try {
    const response = await fetch(url, options)
    if (!response.ok) {
      const errorBody = await response.text()
      throw new Error(`API request to ${endpoint} failed with status ${response.status}: ${errorBody}`)
    }

    return await response.json()
  } catch (error) {
    console.error(`Error during API fetch to ${endpoint}`, error)
    throw error
  }
}


/**
 * Fetches all upcoming rides to call the '/upcoming' endpoint
 * @returns {Promise<Array>} a promise that resolves to an array of upcoming ride objects
 */
export async function getUpcomingRides() {
  return apiFetch('/upcoming')
}


/**
 * Fetches all past rides (last 7 days) from the API
 * @returns {Promise<Array>} a promise that resolves to an array of past rides
 */
export async function getPastRides() {
  return apiFetch('/past')
}

