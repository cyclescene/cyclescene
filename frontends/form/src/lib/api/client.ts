const API_URL = "http://localhost:8080"

export class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public data?: any
  ) {
    super(message);
    this.name = 'APIError';
  }
}

// fetchAPI fetcher function used to call the API
async function fetchAPI<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_URL}${endpoint}`;

  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    }
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new APIError(
      errorText || `HTTP ${response.status}`,
      response.status,
      errorText
    );
  }

  return response.json();
}

// ============================================================================
// TOKEN MANAGEMENT
// ============================================================================

export interface TokenResponse {
  token: string;
  expires_at: string;
}

export async function generateSubmissionToken(city: string): Promise<TokenResponse> {
  return fetchAPI<TokenResponse>('/v1/tokens/submission', {
    method: 'POST',
    body: JSON.stringify({ city })
  });
}

export interface ValidateTokenResponse {
  valid: boolean;
  city?: string;
}

export async function validateSubmissionToken(
  token: string,
  city: string
): Promise<ValidateTokenResponse> {
  return fetchAPI<ValidateTokenResponse>('/v1/tokens/validate', {
    method: 'POST',
    body: JSON.stringify({ token, city })
  });
}

// ============================================================================
// GROUP VALIDATION
// ============================================================================

export interface GroupValidationResponse {
  valid: boolean;
  name?: string;
}

export async function validateGroupCode(code: string): Promise<GroupValidationResponse> {
  return fetchAPI<GroupValidationResponse>(`/v1/groups/validate/${code.toUpperCase()}`);
}

export interface CheckCodeAvailabilityResponse {
  available: boolean;
  code?: string;
  message?: string;
}

export async function checkGroupCodeAvailability(
  code: string
): Promise<CheckCodeAvailabilityResponse> {
  return fetchAPI<CheckCodeAvailabilityResponse>('/v1/groups/check-code', {
    method: 'POST',
    body: JSON.stringify({ code })
  });
}

// ============================================================================
// RIDE SUBMISSION
// ============================================================================

export interface RideSubmissionPayload {
  title: string;
  tinytitle?: string;
  description: string;
  image_url?: string;
  audience?: string;
  ride_length?: string;
  area?: string;
  date_type: string;
  venue_name: string;
  address: string;
  location_details?: string;
  ending_location?: string;
  is_loop_ride: boolean;
  organizer_name: string;
  organizer_email: string;
  organizer_phone?: string;
  web_url?: string;
  web_name?: string;
  newsflash?: string;
  hide_email: boolean;
  hide_phone: boolean;
  hide_contact_name: boolean;
  group_code?: string;
  city: string;
  occurrences: Array<{
    start_date: string;
    start_time: string;
    event_duration_minutes?: number;
    event_time_details?: string;
  }>;
}

export interface SubmissionResponse {
  success: boolean;
  event_id?: number;
  edit_token?: string;
  message?: string;
}

export async function submitRide(
  payload: RideSubmissionPayload,
  bffToken: string
): Promise<SubmissionResponse> {
  return fetchAPI<SubmissionResponse>('/v1/rides/submit', {
    method: 'POST',
    headers: {
      'X-BFF-Token': bffToken
    },
    body: JSON.stringify(payload)
  });
}

// ============================================================================
// RIDE EDITING
// ============================================================================

export interface RideEditResponse {
  event: RideSubmissionPayload;
  is_published: boolean;
}

export async function getRideByEditToken(token: string): Promise<RideEditResponse> {
  return fetchAPI<RideEditResponse>(`/v1/rides/edit/${token}`);
}

export async function updateRide(
  token: string,
  payload: RideSubmissionPayload
): Promise<SubmissionResponse> {
  return fetchAPI<SubmissionResponse>(`/v1/rides/edit/${token}`, {
    method: 'PUT',
    body: JSON.stringify(payload)
  });
}

// ============================================================================
// GROUP REGISTRATION
// ============================================================================

export interface GroupRegistrationPayload {
  code: string;
  name: string;
  description?: string;
  city: string;
  icon_url?: string;
  web_url?: string;
}

export interface GroupResponse {
  success: boolean;
  code?: string;
  edit_token?: string;
  message?: string;
}

export async function registerGroup(
  payload: GroupRegistrationPayload,
  bffToken: string
): Promise<GroupResponse> {
  return fetchAPI<GroupResponse>('/v1/groups/register', {
    method: 'POST',
    headers: {
      'X-BFF-Token': bffToken
    },
    body: JSON.stringify(payload)
  });
}

export async function getGroupByEditToken(token: string): Promise<GroupRegistrationPayload> {
  return fetchAPI<GroupRegistrationPayload>(`/v1/groups/edit/${token}`);
}

export async function updateGroup(
  token: string,
  payload: Partial<GroupRegistrationPayload>
): Promise<GroupResponse> {
  return fetchAPI<GroupResponse>(`/v1/groups/edit/${token}`, {
    method: 'PUT',
    body: JSON.stringify(payload)
  });
}
