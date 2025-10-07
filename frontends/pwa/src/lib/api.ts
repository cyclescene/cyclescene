import type { RideData } from "./types"

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL
const CITY_CODE = import.meta.env.VITE_CITY_CODE

async function apiFetch(endpoint: string, options: RequestInit = {}): Promise<RideData[]> {
  const url = `${API_BASE_URL}${endpoint}?city=${CITY_CODE}`

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


export async function getUpcomingRides(): Promise<RideData[]> {
  return apiFetch('/upcoming')
}

export async function getPastRides(): Promise<RideData[]> {
  return apiFetch('/past')
}

