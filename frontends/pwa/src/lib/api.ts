import type { RideData } from "./types"

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL
const CITY_CODE = import.meta.env.VITE_CITY_CODE

async function apiFetch(endpoint: string, options: RequestInit = {}): Promise<RideData[]> {
  const url = `${API_BASE_URL}/v1/rides${endpoint}?city=${CITY_CODE}`
  console.log(`[apiFetch] Fetching from: ${url}`)
  console.log(`[apiFetch] API_BASE_URL: ${API_BASE_URL}, CITY_CODE: ${CITY_CODE}`)

  try {
    const response = await fetch(url, options)
    console.log(`[apiFetch] Response status: ${response.status}`)
    if (!response.ok) {
      const errorBody = await response.text()
      throw new Error(`API request to ${endpoint} failed with status ${response.status}: ${errorBody}`)
    }

    const data = await response.json()
    console.log(`[apiFetch] Got ${data.length} items from ${endpoint}`)
    return data
  } catch (error) {
    console.error(`Error during API fetch to ${endpoint}`, error)
    throw error
  }
}


export async function getUpcomingRides(): Promise<RideData[]> {
  return apiFetch('/upcoming')
}

export async function getPastRides(): Promise<RideData[]> {
  return apiFetch('/past')
}

