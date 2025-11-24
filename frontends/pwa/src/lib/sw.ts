/// <reference lib="webworker" />
import { precacheAndRoute } from 'workbox-precaching'
import { registerRoute } from 'workbox-routing'
import { CacheFirst, NetworkOnly } from 'workbox-strategies'
import { ExpirationPlugin } from 'workbox-expiration'
import { CacheableResponsePlugin } from 'workbox-cacheable-response'

// Note: Service workers don't have access to import.meta.env
// City code should be passed from main app via postMessage during init
// For now, we'll construct the API URL from the city code if available
const API_BASE = "https://api.cyclescene.cc"
let CITY_CODE = "pdx" // fallback, will be updated via postMessage

const TILE_URLS = {
  dark: "https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json",
  light: "https://basemaps.cartocdn.com/gl/positron-gl-style/style.json"
};
const GCP_HOST = 'https:\/\/cyclescene-api-gateway-\\d+\\.us-west1\.run\.app';
const LOCAL_HOST = 'http:\/\/localhost:8080';
const RIDES_SYNC_TAG = "update-rides-6hr"

// Compute API URL dynamically based on city code
function getApiUpcomingUrl() {
  return API_BASE + "/upcoming?city=" + CITY_CODE;
}
const ONE_HOUR_IN_SECONDS = 60 * 60
const ONE_WEEK_IN_SECONDS = ONE_HOUR_IN_SECONDS * 24 * 7
const ONE_YEAR_IN_SECONDS = ONE_WEEK_IN_SECONDS * 52

declare let self: ServiceWorkerGlobalScope

precacheAndRoute(self.__WB_MANIFEST)
// --- 2. SW LIFECYCLE FOR IMMEDIATE ACTIVATION (CRITICAL FOR DEPLOYMENT) ---
self.addEventListener('install', () => {
  // Force the new Service Worker to activate immediately after install
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  // Claim all existing clients (open tabs) immediately
  event.waitUntil(self.clients.claim());
});

const tileGroup = `(https:\/\/(\\w+\.)?basemaps\.cartocdn\.com)`
// Shared plugin in for all tile sources
const mapAssetCachePlugins = [
  new CacheableResponsePlugin({ statuses: [0, 200] }),
  new ExpirationPlugin({
    maxEntries: 5000,
    maxAgeSeconds: ONE_YEAR_IN_SECONDS
  })
]

const styleCachePlugins = [
  new CacheableResponsePlugin({ statuses: [0, 200] }),
]

registerRoute(
  new RegExp(TILE_URLS.light.replace('/[.*+?^${}()|[\]\\]/g', '\\$&')),
  new CacheFirst({ cacheName: "cartodb-light-style-cache", plugins: styleCachePlugins })
)

registerRoute(
  new RegExp(TILE_URLS.dark.replace('/[.*+?^${}()|[\]\\]/g', '\\$&')),
  new CacheFirst({ cacheName: 'cartodb-dark-style-cache', plugins: styleCachePlugins })
)

registerRoute(
  // Match the host group followed by the rest of the path
  new RegExp(
    `^${tileGroup}\/gl\/.*\/\\d+\/\\d+\/\\d+\\.pbf$`,
    'i' // CRITICAL: Case-insensitive flag
  ),
  new CacheFirst({ cacheName: 'cartodb-vector-tiles-cache', plugins: mapAssetCachePlugins })
);


// --- 3. Sprites (Icons and Patterns - .png, .svg) ---
registerRoute(
  // Match the host group followed by the rest of the path
  new RegExp(
    `^${tileGroup}\/gl\/.*\/sprite.*$`,
    'i' // CRITICAL: Case-insensitive flag
  ),
  new CacheFirst({ cacheName: 'cartodb-sprites-cache', plugins: mapAssetCachePlugins })
);


// --- 4. Glyphs (Fonts - .pbf files for text) ---
registerRoute(
  // Match fonts from CartoDB (both /glyphs/ and /fonts/ paths)
  new RegExp(
    `^${tileGroup}\/(gl\/.*\/glyphs|fonts)\/.*$`,
    'i' // CRITICAL: Case-insensitive flag
  ),
  new CacheFirst({ cacheName: 'cartodb-glyphs-cache', plugins: mapAssetCachePlugins })
);

const rideDataAPIRegex = new RegExp(
  `^(${GCP_HOST}|${LOCAL_HOST})\/v1\/rides\/(upcoming|past)$`,
  'i'
)

registerRoute(
  rideDataAPIRegex,
  new NetworkOnly({
    plugins: [
      new CacheableResponsePlugin({ statuses: [0, 200] }),
      new ExpirationPlugin({ maxAgeSeconds: ONE_HOUR_IN_SECONDS * 6 })
    ]
  })
)

self.addEventListener('periodicsync', (event: PeriodicSyncEvent) => {
  if (event.tag === RIDES_SYNC_TAG) {
    event.waitUntil(fetchAndNotifyUpdate())
  }
})

self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'FORCE_FOREGROUND_SYNC') {
    event.waitUntil(fetchAndNotifyUpdate())
  }
  if (event.data && event.data.type === 'SET_CITY_CODE') {
    CITY_CODE = event.data.cityCode;
    console.log('Service Worker: City code updated to', CITY_CODE);
  }
})

self.addEventListener('sync', (event: any) => {
  if (event.tag === RIDES_SYNC_TAG) {
    event.waitUntil(fetchAndNotifyUpdate())
  }
})

async function fetchAndNotifyUpdate() {
  try {
    const url = getApiUpcomingUrl();
    const response = await fetch(url)
    const freshData = await response.json()

    self.clients.matchAll().then(clients => {
      clients.forEach(client => {
        client.postMessage({
          type: "RIDES_UPDATE_SUCCESSFULL",
          data: freshData
        })
      })
    })
  } catch (e) {
    console.error("Periodic Sync failed to fetch rides: ", e);
  }
}
