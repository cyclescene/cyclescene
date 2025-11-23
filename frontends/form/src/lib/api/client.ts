import { PUBLIC_API_URL } from "$env/static/public"

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
  const url = `${PUBLIC_API_URL}${endpoint}`;

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
// IMAGE UPLOADS & SIGNED URLS
// ============================================================================

export interface SignedURLRequest {
  file_name: string;
  file_type: string;
}

export interface SignedURLResponse {
  success: boolean;
  signed_url: string;
  object_name: string;
  image_uuid: string;
  expires_at: string;
  bucket_name: string;
  error?: string;
}

export async function generateSignedUploadURL(
  fileName: string,
  fileType: string,
  entityType: string,
  cityCode: string,
): Promise<SignedURLResponse> {
  return fetchAPI<SignedURLResponse>('/v1/storage/upload-url', {
    method: 'POST',
    body: JSON.stringify({ file_name: fileName, file_type: fileType, city_code: cityCode, entity_type: entityType })
  });
}

export async function uploadImageToGCS(
  signedUrl: string,
  file: File
): Promise<void> {
  const response = await fetch(signedUrl, {
    method: 'PUT',
    body: file,
    headers: {
      'Content-Type': file.type
    }
  });

  if (!response.ok) {
    throw new Error(`Failed to upload image: ${response.statusText}`);
  }
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
  image_uuid?: string;
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
  icon_uuid?: string;
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
