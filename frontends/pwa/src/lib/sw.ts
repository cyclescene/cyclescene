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

// Cache CartoDB map tiles and resources
registerRoute(
  ({ url }) => url.hostname === 'basemaps.cartocdn.com' || url.hostname.endsWith('.basemaps.cartocdn.com'),
  new CacheFirst({
    cacheName: 'cartodb-cache',
    plugins: [
      new CacheableResponsePlugin({ statuses: [0, 200] }),
      new ExpirationPlugin({
        maxEntries: 5000,
        maxAgeSeconds: ONE_YEAR_IN_SECONDS
      })
    ]
  })
);

// Cache API responses for ride data
registerRoute(
  ({ url }) => url.hostname === 'api.cyclescene.cc',
  new NetworkOnly({
    plugins: [
      new CacheableResponsePlugin({ statuses: [0, 200] }),
      new ExpirationPlugin({ maxAgeSeconds: ONE_HOUR_IN_SECONDS * 6 })
    ]
  })
)

self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'FORCE_FOREGROUND_SYNC') {
    event.waitUntil(fetchAndNotifyUpdate())
  }
  if (event.data && event.data.type === 'SET_CITY_CODE') {
    CITY_CODE = event.data.cityCode;
    console.log('Service Worker: City code updated to', CITY_CODE);
  }
})

const RIDES_SYNC_TAG = "update-rides-6hr"

self.addEventListener('periodicsync', (event: PeriodicSyncEvent) => {
  console.log('Service Worker: periodicsync event', event.tag);
  if (event.tag === RIDES_SYNC_TAG) {
    event.waitUntil(
      fetchAndNotifyUpdate().catch(err => {
        console.error('Periodic sync failed:', err);
        throw err;
      })
    )
  }
})

self.addEventListener('sync', (event: any) => {
  console.log('Service Worker: background sync event', event.tag);
  if (event.tag === RIDES_SYNC_TAG) {
    event.waitUntil(
      fetchAndNotifyUpdate().catch(err => {
        console.error('Background sync failed:', err);
        throw err;
      })
    )
  }
})

async function fetchAndNotifyUpdate() {
  console.log('Service Worker: Attempting to fetch rides update');

  try {
    // Guard: only attempt if city code is set
    if (!CITY_CODE || CITY_CODE === '') {
      console.warn('Service Worker: City code not set, skipping sync');
      return;
    }

    const url = getApiUpcomingUrl();
    console.log('Service Worker: Fetching from', url);

    const response = await fetch(url, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' }
    });

    if (!response.ok) {
      console.warn(`Service Worker: API returned ${response.status} ${response.statusText}`);
      return; // Don't throw, just return gracefully
    }

    const freshData = await response.json();
    console.log('Service Worker: Got fresh data, notifying clients');

    self.clients.matchAll().then(clients => {
      clients.forEach(client => {
        client.postMessage({
          type: "RIDES_UPDATE_SUCCESSFULL",
          data: freshData
        })
      })
    }).catch(err => {
      console.error('Service Worker: Error notifying clients:', err);
    });
  } catch (e) {
    console.error("Service Worker: Sync failed to fetch rides:", e);
    // Don't rethrow - just log the error
  }
}
