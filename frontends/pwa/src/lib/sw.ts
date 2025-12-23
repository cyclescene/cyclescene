/// <reference lib="webworker" />
import { precacheAndRoute } from 'workbox-precaching'
import { registerRoute } from 'workbox-routing'
import { CacheFirst, NetworkFirst } from 'workbox-strategies'
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

// Cache CartoDB map tiles and resources - only cache successful responses
registerRoute(
  ({ url }) => url.hostname === 'basemaps.cartocdn.com' || url.hostname.endsWith('.basemaps.cartocdn.com'),
  new CacheFirst({
    cacheName: 'cartodb-cache',
    plugins: [
      new CacheableResponsePlugin({ statuses: [200] }), // Only cache 200 responses, not network errors
      new ExpirationPlugin({
        maxEntries: 5000,
        maxAgeSeconds: ONE_YEAR_IN_SECONDS
      })
    ]
  })
);

// Cache API responses for ride data - try network first, fall back to cache
registerRoute(
  ({ url }) => url.hostname === 'api.cyclescene.cc',
  new NetworkFirst({
    cacheName: 'api-cache',
    plugins: [
      new CacheableResponsePlugin({ statuses: [200] }), // Only cache 200 responses
      new ExpirationPlugin({ maxAgeSeconds: ONE_HOUR_IN_SECONDS * 6 })
    ]
  })
)

self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'FORCE_FOREGROUND_SYNC') {
    event.waitUntil(fetchAndNotifyUpdate('manual'))
  }
  if (event.data && event.data.type === 'SET_CITY_CODE') {
    CITY_CODE = event.data.cityCode;
    console.log('Service Worker: City code updated to', CITY_CODE);
  }
})

const RIDES_SYNC_TAG = "update-rides-30min"

self.addEventListener('periodicsync', (event: PeriodicSyncEvent) => {
  console.log('Service Worker: periodicsync event', event.tag);
  if (event.tag === RIDES_SYNC_TAG) {
    event.waitUntil(
      fetchAndNotifyUpdate('periodic').catch(err => {
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
      fetchAndNotifyUpdate('foreground').catch(err => {
        console.error('Background sync failed:', err);
        throw err;
      })
    )
  }
})

async function fetchAndNotifyUpdate(syncType: "periodic" | "manual" | "foreground" = "manual") {
  const startTime = Date.now();
  let status = "success";
  let errorMsg = "";
  let rideCount = 0;

  console.log('Service Worker: Attempting to fetch rides update', syncType);

  try {
    // Guard: only attempt if city code is set
    if (!CITY_CODE || CITY_CODE === '') {
      console.warn('Service Worker: City code not set, skipping sync');
      status = "error";
      errorMsg = "City code not set";
      await logSyncEvent(syncType, status, errorMsg, 0, Date.now() - startTime);
      return;
    }

    const url = getApiUpcomingUrl();
    console.log('Service Worker: Fetching from', url);

    const response = await fetch(url, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' }
    });

    if (!response.ok) {
      status = "error";
      const statusText = response.statusText || 'No status text';
      errorMsg = `HTTP ${response.status} ${statusText} from ${url}`;
      console.warn(`Service Worker: API returned ${response.status} ${statusText}`);
      await logSyncEvent(syncType, status, errorMsg, 0, Date.now() - startTime);
      return; // Don't throw, just return gracefully
    }

    let freshData;
    try {
      freshData = await response.json();
    } catch (parseErr) {
      status = "error";
      errorMsg = `Failed to parse API response: ${parseErr instanceof Error ? parseErr.message : String(parseErr)}`;
      console.error("Service Worker: Failed to parse API response:", parseErr);
      await logSyncEvent(syncType, status, errorMsg, 0, Date.now() - startTime);
      return;
    }
    rideCount = freshData?.rides?.length || 0;
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

    await logSyncEvent(syncType, status, errorMsg, rideCount, Date.now() - startTime);
  } catch (e) {
    status = "error";
    errorMsg = e instanceof Error ? e.message : String(e);
    console.error("Service Worker: Sync failed to fetch rides:", e);
    await logSyncEvent(syncType, status, errorMsg, 0, Date.now() - startTime);
    // Don't rethrow - just log the error
  }
}

async function logSyncEvent(
  syncType: string,
  status: string,
  errorMsg: string,
  rideCount: number,
  duration: number
) {
  try {
    const clientId = await getOrCreateClientId();
    const os = detectOS();

    await fetch("https://api.cyclescene.cc/v1/sync-logs", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        client_id: clientId,
        sync_type: syncType,
        status,
        error_msg: errorMsg,
        ride_count: rideCount,
        duration,
        city_code: CITY_CODE,
        os,
        timestamp: new Date().toISOString(),
      }),
    });
  } catch (error) {
    // Silently fail - don't block sync if logging fails
    console.error("[SW] Failed to log sync event:", error);
  }
}

function detectOS(): string {
  const userAgent = self.navigator.userAgent.toLowerCase();

  if (userAgent.includes('iphone') || userAgent.includes('ipad')) {
    return 'iOS';
  } else if (userAgent.includes('android')) {
    return 'Android';
  } else if (userAgent.includes('win')) {
    return 'Windows';
  } else if (userAgent.includes('mac')) {
    return 'macOS';
  } else if (userAgent.includes('linux')) {
    return 'Linux';
  } else if (userAgent.includes('x11')) {
    return 'Unix';
  } else {
    return 'Unknown';
  }
}

async function getOrCreateClientId(): Promise<string> {
  const cache = await caches.open("app-cache");
  const cachedResponse = await cache.match("client-id");

  if (cachedResponse) {
    const { clientId } = await cachedResponse.json();
    return clientId;
  }

  // Generate new client ID (UUID)
  const clientId = crypto.randomUUID();
  await cache.put(
    "client-id",
    new Response(JSON.stringify({ clientId }))
  );

  return clientId;
}
