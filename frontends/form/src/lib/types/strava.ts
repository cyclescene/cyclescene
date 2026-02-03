// ============================================================================
// Strava Import Types
// ============================================================================

// ============================================================================
// API Response Types (from backend)
// ============================================================================

export interface StravaClub {
  id: number;
  name: string;
  profile_medium: string;
  city: string;
  state: string;
  country: string;
  member_count: number;
  is_admin: boolean;
  is_owner: boolean;
}

export interface StravaGroupEvent {
  id: string; // String to avoid JavaScript precision loss with large int64 IDs
  title: string;
  description: string;
  activity_type: string;
  upcoming_occurrences: string[]; // ISO 8601 timestamps
  zone: string; // IANA timezone
  address: string;
  start_latlng: [number, number]; // [lat, lng]
  route_id: number | null;
  route: StravaRoute | null;
  skill_levels: string | null;
  terrain: string | null;
  women_only: boolean;
  private: boolean;
  club_id: number;
}

export interface StravaRoute {
  id: number;
  name: string;
  distance: number; // meters
  elevation_gain: number; // meters
}

// ============================================================================
// WebSocket Message Types (match backend exactly)
// ============================================================================

// Client -> Server
export interface ImportRequest {
  session_id?: string; // Optional - backend reads from HttpOnly cookie if not provided
  organizer_email: string;
  group_code?: string; // Optional 4-char group code to associate all imported rides with
  events: EventImportConfig[];
}

export interface EventImportConfig {
  strava_event_id: string; // MUST be string (int64 precision issue)
  club_id: number;
  overrides?: EventOverrides;
}

export interface EventOverrides {
  audience?: string; // "G", "F", "A", "E"
  image_url?: string;
  event_duration_minutes?: number;
}

// Server -> Client
export type ProgressMessageType =
  | "heartbeat"
  | "progress"
  | "complete"
  | "done"
  | "error";

export type ProgressStep = "fetching" | "coordinates" | "route" | "database";

export type ProgressStatus = "in_progress" | "success" | "error";

export interface ProgressMessage {
  type: ProgressMessageType;
  event_index?: number;
  total_events?: number;
  strava_event_id?: string;
  event_title?: string;
  step?: ProgressStep;
  status?: ProgressStatus;
  message?: string;
  // Complete message fields
  cyclescene_event_id?: number;
  edit_token?: string;
  edit_url?: string;
  success?: boolean;
  error?: string | null;
  // Done message fields
  total_imported?: number;
  total_failed?: number;
  summary_email_sent?: boolean;
  results?: ImportResult[];
}

export interface ImportResult {
  strava_event_id: string;
  title: string;
  success: boolean;
  cyclescene_event_id?: number;
  edit_token?: string;
  edit_url?: string;
  error?: string;
}

// ============================================================================
// UI State Types
// ============================================================================

export type ImportStep =
  | "connect"
  | "email"
  | "select"
  | "importing"
  | "complete";

export interface StravaState {
  authenticated: boolean;
  step: ImportStep;
  organizerEmail: string;
  adminClubs: StravaClub[];
  clubEvents: Map<number, StravaGroupEvent[]>; // clubId -> events
  selectedEvents: Map<string, EventImportConfig>; // eventId -> config
  expandedClubs: Set<number>;
  expandedEvents: Set<string>; // Events with customize panel open
  importProgress: Map<string, ProgressMessage>; // eventId -> latest progress
  importResults: ImportResult[];
  error: string | null;
}

export interface ClubWithEvents extends StravaClub {
  events?: StravaGroupEvent[];
  loading?: boolean;
  error?: string;
}

// ============================================================================
// API Response Wrappers
// ============================================================================

export interface AdminClubsResponse {
  clubs: StravaClub[];
}

export interface ClubEventsResponse {
  events: StravaGroupEvent[];
}

export interface AuthInitiateResponse {
  auth_url: string;
}

// ============================================================================
// Error Classes
// ============================================================================

/**
 * Rate limit error (HTTP 429)
 */
export class RateLimitError extends Error {
  retry_after_seconds: number;

  constructor(message: string, retryAfterSeconds: number) {
    super(message);
    this.name = "RateLimitError";
    this.retry_after_seconds = retryAfterSeconds;
  }
}

/**
 * Session expired error (HTTP 401 with refresh token failure)
 */
export class SessionExpiredError extends Error {
  constructor(message: string = "Your session expired. Please reconnect to Strava.") {
    super(message);
    this.name = "SessionExpiredError";
  }
}

// ============================================================================
// Constants
// ============================================================================

export const AUDIENCE_OPTIONS = [
  { value: "G", label: "G - All Ages" },
  { value: "F", label: "F - Family Friendly" },
  { value: "A", label: "A - Adult (21+)" },
  { value: "E", label: "E - Explicit/Adult Content" },
] as const;

export const DURATION_OPTIONS = [
  { value: 30, label: "30 minutes" },
  { value: 60, label: "1 hour" },
  { value: 90, label: "1.5 hours" },
  { value: 120, label: "2 hours" },
  { value: 150, label: "2.5 hours" },
  { value: 180, label: "3 hours" },
  { value: 240, label: "4 hours" },
  { value: 300, label: "5 hours" },
  { value: 360, label: "6 hours" },
] as const;

// Default values for Strava imports
export const STRAVA_IMPORT_DEFAULTS = {
  audience: "G",
  duration: null, // Will use route-based estimate if available
} as const;
