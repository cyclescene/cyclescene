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
    this.results = [];
    this.reconnectAttempts = 0;
    this.importCompleted = false;

    try {
      const url = getImportWebSocketUrl();
      debugLog("WebSocket URL", url);

      this.ws = new WebSocket(url);
      this.setState("connecting");

      this.ws.onopen = () => {
        debugLog("WebSocket connected");
        this.setState("connected");
        this.reconnectAttempts = 0;

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
          debugLog(`Reconnecting (attempt ${this.reconnectAttempts})`);
          setTimeout(() => {
            this.connect(organizerEmail, events, groupCode);
          }, 1000 * this.reconnectAttempts);
          return;
        }

        this.setState("disconnected");

        // If connection failed without completing, report error
        if (event.code !== 1000 && this.results.length === 0) {
          this.options.onError(
            event.reason || "Connection lost. Please try again."
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
   * Handle incoming WebSocket messages
   */
  private handleMessage(message: ProgressMessage): void {
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
   * Close the WebSocket connection
   */
  close(): void {
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
