// ============================================================================
// Strava WebSocket Client Helper
// ============================================================================

import { PUBLIC_STRAVA_DEBUG } from "$env/static/public";
import { getImportWebSocketUrl } from "$lib/api/strava";
import type {
  ImportRequest,
  ProgressMessage,
  ImportResult,
  EventImportConfig,
} from "$lib/types/strava";

// Debug logging helper
function debugLog(message: string, data?: unknown) {
  if (typeof window !== "undefined" && PUBLIC_STRAVA_DEBUG === "true") {
    console.log(`[Strava WS] ${message}`, data ?? "");
  }
}

// Get session ID from cookie (only works in dev mode where HttpOnly=false)
function getSessionIdFromCookie(): string | undefined {
  if (typeof document === "undefined") return undefined;
  const match = document.cookie.match(/(?:^|;\s*)strava_session_id=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : undefined;
}

export type WebSocketState = "connecting" | "connected" | "disconnected" | "error";

export interface StravaWebSocketOptions {
  onProgress: (message: ProgressMessage) => void;
  onComplete: (results: ImportResult[]) => void;
  onError: (error: string) => void;
  onStateChange?: (state: WebSocketState) => void;
}

export class StravaImportWebSocket {
  private ws: WebSocket | null = null;
  private options: StravaWebSocketOptions;
  private state: WebSocketState = "disconnected";
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 3;
  private results: ImportResult[] = [];
  private importCompleted = false; // Track if import finished (don't reconnect after)
  private manualStop = false; // Track if user manually stopped the import
  private lastRequest: { email: string; events: EventImportConfig[]; groupCode?: string } | null = null;
  private heartbeatTimer: number | null = null; // 60-second heartbeat timeout
  private activityTimer: number | null = null; // 15-second activity warning
  private hasActivity = false; // Track if any activity occurred

  constructor(options: StravaWebSocketOptions) {
    this.options = options;
  }

  /**
   * Connect to the WebSocket and start the import
   * @param organizerEmail - Email for edit links
   * @param events - Events to import with their configs
   * @param groupCode - Optional group code to associate all rides with
   */
  connect(organizerEmail: string, events: EventImportConfig[], groupCode?: string): void {
    debugLog("Connecting to WebSocket");

    // Store request for potential retry
    this.lastRequest = { email: organizerEmail, events, groupCode };

    // Reset state for new connection (but preserve results if retrying)
    this.reconnectAttempts = 0;
    this.importCompleted = false;
    this.manualStop = false;

    try {
      const url = getImportWebSocketUrl();
      debugLog("WebSocket URL", url);

      this.ws = new WebSocket(url);
      this.setState("connecting");

      this.ws.onopen = () => {
        debugLog("WebSocket connected");
        this.setState("connected");
        this.reconnectAttempts = 0;

        // Start timeout timers
        this.startTimeouts();

        // Send import request
        const request: ImportRequest = {
          organizer_email: organizerEmail,
          events: events,
        };
        if (groupCode) {
          request.group_code = groupCode;
        }
        // Try to read session from cookie (works in dev mode where HttpOnly=false)
        const sessionId = getSessionIdFromCookie();
        if (sessionId) {
          request.session_id = sessionId;
          debugLog("Including session_id from cookie");
        }
        debugLog("Sending import request", request);
        this.ws?.send(JSON.stringify(request));
      };

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as ProgressMessage;
          debugLog("Received message", message);
          this.handleMessage(message);
        } catch (error) {
          debugLog("Failed to parse message", error);
        }
      };

      this.ws.onerror = (error) => {
        debugLog("WebSocket error", error);
        this.setState("error");
        // Don't call onError here - wait for onclose to get more info
      };

      this.ws.onclose = (event) => {
        debugLog("WebSocket closed", { code: event.code, reason: event.reason });

        // Don't reconnect if manually stopped by user
        if (this.manualStop) {
          this.setState("disconnected");
          debugLog("Import stopped by user");

          // Log manual stop event
          console.log(
            JSON.stringify({
              event: "strava_import_stopped",
              total_events: events.length,
              completed: this.results.filter(r => r.success).length,
              stopped_by_user: true,
              timestamp: new Date().toISOString(),
            })
          );
          return;
        }

        // Don't reconnect if import already completed successfully
        if (this.importCompleted) {
          this.setState("disconnected");
          return;
        }

        // If we haven't completed and haven't hit max reconnects, try to reconnect
        if (
          this.state !== "disconnected" &&
          this.reconnectAttempts < this.maxReconnectAttempts &&
          event.code !== 1000 // Normal closure
        ) {
          this.reconnectAttempts++;
          debugLog(`Reconnecting (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

          // Notify state change with reconnect attempt info
          this.setState("connecting");

          setTimeout(() => {
            this.connect(organizerEmail, events, groupCode);
          }, 1000 * this.reconnectAttempts);
          return;
        }

        this.setState("disconnected");

        // If connection failed without completing, report error
        if (event.code !== 1000 && this.results.length === 0) {
          this.options.onError(
            event.reason || "Connection lost. Please check your internet and try again."
          );
        }
      };
    } catch (error) {
      debugLog("Failed to create WebSocket", error);
      this.setState("error");
      this.options.onError("Failed to connect. Please try again.");
    }
  }

  /**
   * Start timeout timers
   */
  private startTimeouts(): void {
    // Clear any existing timers
    this.clearTimeouts();
    this.hasActivity = false;

    // 15-second activity warning
    this.activityTimer = window.setTimeout(() => {
      if (!this.hasActivity) {
        debugLog("No activity for 15 seconds, showing warning");
        // Notify via progress callback with a special message
        this.options.onProgress({
          type: "progress",
          message: "Taking longer than expected...",
        });
      }
    }, 15000);

    // 60-second heartbeat timeout
    this.heartbeatTimer = window.setTimeout(() => {
      debugLog("Heartbeat timeout (60s), triggering reconnect");
      this.options.onError("Connection timeout. Attempting to reconnect...");
      this.ws?.close();
    }, 60000);
  }

  /**
   * Reset timeout timers (called when activity is detected)
   */
  private resetTimeouts(): void {
    this.hasActivity = true;
    this.startTimeouts();
  }

  /**
   * Clear timeout timers
   */
  private clearTimeouts(): void {
    if (this.heartbeatTimer) {
      clearTimeout(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
    if (this.activityTimer) {
      clearTimeout(this.activityTimer);
      this.activityTimer = null;
    }
  }

  /**
   * Handle incoming WebSocket messages
   */
  private handleMessage(message: ProgressMessage): void {
    // Reset timeouts on any message
    this.resetTimeouts();

    switch (message.type) {
      case "heartbeat":
        debugLog("Heartbeat received");
        // Heartbeat - just keep connection alive, no UI update needed
        break;

      case "progress":
        // Progress update for a specific event/step
        this.options.onProgress(message);
        break;

      case "complete":
        // Single event completed
        this.options.onProgress(message);
        // Store result
        if (message.strava_event_id) {
          this.results.push({
            strava_event_id: message.strava_event_id,
            title: message.event_title || "Unknown Event",
            success: message.success ?? false,
            cyclescene_event_id: message.cyclescene_event_id,
            edit_token: message.edit_token,
            edit_url: message.edit_url,
            error: message.error ?? undefined,
          });
        }
        break;

      case "done":
        // All events processed
        debugLog("Import complete", message);
        this.importCompleted = true; // Mark as done to prevent reconnect
        this.setState("disconnected");
        // Use results from done message if available, otherwise use accumulated results
        const finalResults = message.results || this.results;
        this.options.onComplete(finalResults);
        this.close();
        break;

      case "error":
        // Fatal error
        debugLog("Import error", message);
        this.importCompleted = true; // Don't reconnect after fatal error
        this.options.onError(message.message || "Import failed");
        this.close();
        break;
    }
  }

  /**
   * Set and notify state change
   */
  private setState(state: WebSocketState): void {
    this.state = state;
    this.options.onStateChange?.(state);
  }

  /**
   * Get current connection state
   */
  getState(): WebSocketState {
    return this.state;
  }

  /**
   * Get current reconnect attempt count
   */
  getReconnectAttempts(): { current: number; max: number } {
    return {
      current: this.reconnectAttempts,
      max: this.maxReconnectAttempts,
    };
  }

  /**
   * Get completed results (useful for retry with partial results)
   */
  getCompletedResults(): ImportResult[] {
    return this.results;
  }

  /**
   * Stop the import manually (user-initiated)
   */
  stop(): void {
    debugLog("Stopping import (user initiated)");
    this.manualStop = true;
    this.close();
  }

  /**
   * Manually retry the import (unlimited retries allowed)
   * Preserves already-completed results
   */
  manualRetry(): void {
    if (!this.lastRequest) {
      debugLog("No previous request to retry");
      return;
    }

    debugLog("Manual retry initiated", {
      previousResults: this.results.length,
      totalEvents: this.lastRequest.events.length,
    });

    // Log retry event
    const completedCount = this.results.filter(r => r.success).length;
    const failedCount = this.results.filter(r => !r.success).length;
    console.log(
      JSON.stringify({
        event: "strava_import_retry",
        attempt_number: this.reconnectAttempts + 1,
        total_events: this.lastRequest.events.length,
        completed_before_failure: completedCount,
        failed_before_retry: failedCount,
        timestamp: new Date().toISOString(),
      })
    );

    // Reset reconnect counter for fresh attempts
    this.reconnectAttempts = 0;
    this.importCompleted = false;
    this.manualStop = false;

    // Reconnect
    this.connect(
      this.lastRequest.email,
      this.lastRequest.events,
      this.lastRequest.groupCode
    );
  }

  /**
   * Close the WebSocket connection
   */
  close(): void {
    this.clearTimeouts();
    if (this.ws) {
      debugLog("Closing WebSocket");
      this.ws.close(1000, "Client closed");
      this.ws = null;
    }
    this.setState("disconnected");
  }
}

/**
 * Create and connect a WebSocket for importing events
 * Convenience function for one-shot imports
 */
export function createImportWebSocket(
  options: StravaWebSocketOptions
): StravaImportWebSocket {
  return new StravaImportWebSocket(options);
}
