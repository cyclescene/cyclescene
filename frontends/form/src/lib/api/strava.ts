// ============================================================================
// Strava API Client
// ============================================================================

import { PUBLIC_API_URL, PUBLIC_STRAVA_DEBUG } from "$env/static/public";
import { APIError } from "./client";
import {
  RateLimitError,
  SessionExpiredError,
  type AdminClubsResponse,
  type ClubEventsResponse,
  type AuthInitiateResponse,
  type StravaClub,
  type StravaGroupEvent,
} from "$lib/types/strava";

// Debug logging helper
function debugLog(message: string, data?: unknown) {
  if (typeof window !== "undefined" && PUBLIC_STRAVA_DEBUG === "true") {
    console.log(`[Strava] ${message}`, data ?? "");
  }
}

// Fetch wrapper for Strava API calls (includes credentials for session cookie)
async function fetchStravaAPI<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${PUBLIC_API_URL}${endpoint}`;

  debugLog(`API Request: ${options.method || "GET"} ${endpoint}`);

  const response = await fetch(url, {
    ...options,
    credentials: "include", // Important: include session cookie
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  if (!response.ok) {
    const errorText = await response.text();
    debugLog(`API Error: ${response.status}`, errorText);

    // Handle rate limit (429)
    if (response.status === 429) {
      let retryAfterSeconds = 900; // Default 15 minutes
      try {
        const errorData = JSON.parse(errorText);
        if (errorData.retry_after_seconds) {
          retryAfterSeconds = errorData.retry_after_seconds;
        }
      } catch {
        // If parsing fails, use default
      }
      throw new RateLimitError(
        errorText || "Rate limit exceeded",
        retryAfterSeconds
      );
    }

    // Handle session expired (401)
    if (response.status === 401) {
      throw new SessionExpiredError(
        errorText || "Your session expired. Please reconnect to Strava."
      );
    }

    throw new APIError(
      errorText || `HTTP ${response.status}`,
      response.status,
      errorText
    );
  }

  const data = await response.json();
  debugLog(`API Response:`, data);
  return data;
}

// ============================================================================
// OAuth Functions
// ============================================================================

/**
 * Initiates OAuth flow by opening a popup window
 * @param city - City code for the import (e.g., "pdx")
 * @returns Promise that resolves when auth completes (via postMessage)
 */
export function initiateAuth(city: string): Promise<void> {
  return new Promise((resolve, reject) => {
    debugLog("Initiating OAuth", { city });

    // Calculate popup position (centered)
    const width = 600;
    const height = 700;
    const left = window.screenX + (window.outerWidth - width) / 2;
    const top = window.screenY + (window.outerHeight - height) / 2;

    // Open popup to backend auth endpoint
    const authUrl = `${PUBLIC_API_URL}/v1/strava/auth/initiate?city=${encodeURIComponent(city)}`;
    const popup = window.open(
      authUrl,
      "strava-auth",
      `width=${width},height=${height},left=${left},top=${top},scrollbars=yes`
    );

    if (!popup) {
      debugLog("Popup blocked");
      reject(new Error("Please allow popups to connect with Strava"));
      return;
    }

    // Listen for completion message from popup
    const handleMessage = (event: MessageEvent) => {
      // Verify origin (should be our form domain)
      if (event.data?.type === "strava-auth-complete") {
        debugLog("OAuth complete via postMessage");
        window.removeEventListener("message", handleMessage);
        clearInterval(pollTimer);
        resolve();
      } else if (event.data?.type === "strava-auth-error") {
        debugLog("OAuth error via postMessage", event.data.error);
        window.removeEventListener("message", handleMessage);
        clearInterval(pollTimer);

        // Check if user denied authorization
        const errorMsg = event.data.error || "";
        if (errorMsg.includes("access_denied") || errorMsg.includes("cancelled")) {
          reject(new Error("Authorization cancelled. You can try again anytime."));
        } else {
          reject(new Error(errorMsg || "Strava authorization failed"));
        }
      }
    };

    window.addEventListener("message", handleMessage);

    // Also poll for popup closure (fallback if postMessage fails)
    const pollTimer = setInterval(() => {
      if (popup.closed) {
        debugLog("Popup closed (polling detected)");
        clearInterval(pollTimer);
        window.removeEventListener("message", handleMessage);
        // Don't reject - popup might have closed after success
        // We'll check session validity instead
        checkSession()
          .then((valid) => {
            if (valid) {
              resolve();
            } else {
              // Session invalid after popup closed - user likely cancelled
              reject(new Error("Authorization cancelled. You can try again anytime."));
            }
          })
          .catch(() => {
            // Session check failed - user likely cancelled
            reject(new Error("Authorization cancelled. You can try again anytime."));
          });
      }
    }, 500);

    // Timeout after 5 minutes
    setTimeout(() => {
      clearInterval(pollTimer);
      window.removeEventListener("message", handleMessage);
      if (!popup.closed) {
        popup.close();
      }
      reject(new Error("Strava authorization timed out"));
    }, 5 * 60 * 1000);
  });
}

/**
 * Check if the current session is valid
 * @returns true if session cookie is valid, false otherwise
 */
export async function checkSession(): Promise<boolean> {
  try {
    debugLog("Checking session validity");
    // Try to fetch admin clubs - will fail if session invalid
    await fetchStravaAPI<AdminClubsResponse>("/v1/strava/admin-clubs");
    debugLog("Session valid");
    return true;
  } catch (error) {
    debugLog("Session invalid", error);
    return false;
  }
}

/**
 * Fetch clubs where the user is an admin/owner
 * @returns Array of admin clubs
 */
export async function fetchAdminClubs(): Promise<StravaClub[]> {
  debugLog("Fetching admin clubs");
  const response = await fetchStravaAPI<AdminClubsResponse>(
    "/v1/strava/admin-clubs"
  );
  debugLog(`Found ${response.clubs?.length || 0} admin clubs`);
  return response.clubs || [];
}

/**
 * Fetch events for a specific club
 * @param clubId - Strava club ID
 * @returns Array of group events
 */
export async function fetchClubEvents(
  clubId: number
): Promise<StravaGroupEvent[]> {
  debugLog("Fetching club events", { clubId });
  const response = await fetchStravaAPI<ClubEventsResponse>(
    `/v1/strava/clubs/${clubId}/events`
  );
  debugLog(`Found ${response.events?.length || 0} events for club ${clubId}`);
  return response.events || [];
}

/**
 * Check session validity before import (allows backend to refresh token if needed)
 * @returns true if session is valid and ready for import
 */
export async function checkSessionForImport(): Promise<boolean> {
  try {
    debugLog("Checking session before import");
    await fetchStravaAPI<{ valid: boolean }>("/v1/strava/check-session");
    debugLog("Session valid for import");
    return true;
  } catch (error) {
    debugLog("Session check failed", error);
    return false;
  }
}

/**
 * Logout and clear the Strava session
 */
export async function logout(): Promise<void> {
  debugLog("Logging out");
  await fetchStravaAPI<{ success: boolean }>("/v1/strava/logout", {
    method: "POST",
  });
  debugLog("Logout complete");
}

/**
 * Format seconds into a human-readable time string
 * @param seconds - Number of seconds
 * @returns Formatted string like "5 minutes" or "1 minute"
 */
export function formatRetryTime(seconds: number): string {
  const minutes = Math.ceil(seconds / 60);
  return `${minutes} minute${minutes !== 1 ? 's' : ''}`;
}

// ============================================================================
// WebSocket URL Helper
// ============================================================================

/**
 * Get the WebSocket URL for the import endpoint
 * @returns WebSocket URL
 */
export function getImportWebSocketUrl(): string {
  // Convert http(s) to ws(s)
  const wsProtocol = PUBLIC_API_URL.startsWith("https") ? "wss" : "ws";
  const wsHost = PUBLIC_API_URL.replace(/^https?:\/\//, "");
  return `${wsProtocol}://${wsHost}/v1/strava/import`;
}
