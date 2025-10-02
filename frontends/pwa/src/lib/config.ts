const CONFIG = [
  {
    "city-code": "pdx",
    "city-name": "Portland",
    "starting-coords": {
      "lat": 45.515232,
      "lon": -122.6783853
    }
  },
  {
    "city-code": "slc",
    "city-name": "Salt Lake City",
    "starting-coords": {
      "lat": 40.76078,
      "lon": -111.89105
    }
  }
]


const CITY_DATA = (() => {
  const code = import.meta.env.VITE_CITY_CODE

  const found = CONFIG.find(city => city['city-code'] === code)

  // if not found default to portland config
  if (!found) {
    console.error(`ERROR: City code '${code}' not found in cofiguration.`)
    return CONFIG[0];
  }

  return found

})()

export const CITY_CODE = CITY_DATA['city-code']
export const STARTING_LAT = CITY_DATA['starting-coords'].lat
export const STARTING_LON = CITY_DATA['starting-coords'].lon

export const FULL_CITY_CONFIG = CITY_DATA

