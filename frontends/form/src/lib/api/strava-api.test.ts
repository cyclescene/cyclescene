import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
	checkSession,
	fetchAdminClubs,
	fetchClubEvents,
	checkSessionForImport,
	logout
} from './strava';
import { RateLimitError, SessionExpiredError } from '$lib/types/strava';
import { APIError } from './client';

// Mock environment variables
vi.mock('$env/static/public', () => ({
	PUBLIC_API_URL: 'http://localhost:3000',
	PUBLIC_STRAVA_DEBUG: 'false'
}));

describe('Strava API Client', () => {
	let fetchMock: ReturnType<typeof vi.fn>;

	beforeEach(() => {
		// Mock fetch globally
		fetchMock = vi.fn();
		globalThis.fetch = fetchMock as any;
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	describe('checkSession', () => {
		it('should return true when session is valid', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => ({ clubs: [] })
			});

			const result = await checkSession();

			expect(result).toBe(true);
			expect(fetchMock).toHaveBeenCalledWith(
				'http://localhost:3000/v1/strava/admin-clubs',
				expect.objectContaining({
					credentials: 'include',
					headers: expect.objectContaining({
						'Content-Type': 'application/json'
					})
				})
			);
		});

		it('should return false when session is invalid', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 401,
				text: async () => 'Unauthorized'
			});

			const result = await checkSession();

			expect(result).toBe(false);
		});

		it('should return false on network error', async () => {
			fetchMock.mockRejectedValueOnce(new Error('Network error'));

			const result = await checkSession();

			expect(result).toBe(false);
		});
	});

	describe('fetchAdminClubs', () => {
		it('should return clubs array on success', async () => {
			const mockClubs = [
				{
					id: 123,
					name: 'Portland Riders',
					city: 'Portland',
					state: 'OR',
					country: 'USA',
					member_count: 100,
					sport_type: 'cycling',
					admin: true,
					owner: false
				},
				{
					id: 456,
					name: 'SLC Cycling',
					city: 'Salt Lake City',
					state: 'UT',
					country: 'USA',
					member_count: 50,
					sport_type: 'cycling',
					admin: false,
					owner: true
				}
			];

			fetchMock.mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => ({ clubs: mockClubs })
			});

			const clubs = await fetchAdminClubs();

			expect(clubs).toEqual(mockClubs);
			expect(clubs).toHaveLength(2);
		});

		it('should return empty array when no clubs', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => ({ clubs: [] })
			});

			const clubs = await fetchAdminClubs();

			expect(clubs).toEqual([]);
		});

		it('should throw RateLimitError on 429 response', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 429,
				text: async () =>
					JSON.stringify({
						error: 'Rate Limit Exceeded',
						retry_after_seconds: 900
					})
			});

			await expect(fetchAdminClubs()).rejects.toThrow(RateLimitError);

			try {
				await fetchAdminClubs();
			} catch (error) {
				if (error instanceof RateLimitError) {
					expect(error.retry_after_seconds).toBe(900);
				}
			}
		});

		it('should use default retry time when not provided', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 429,
				text: async () => 'Rate Limit Exceeded'
			});

			try {
				await fetchAdminClubs();
			} catch (error) {
				if (error instanceof RateLimitError) {
					expect(error.retry_after_seconds).toBe(900); // Default 15 minutes
				}
			}
		});

		it('should throw SessionExpiredError on 401 response', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 401,
				text: async () => 'Unauthorized - session expired'
			});

			await expect(fetchAdminClubs()).rejects.toThrow(SessionExpiredError);
		});

		it('should throw APIError on other errors', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 500,
				text: async () => 'Internal Server Error'
			});

			await expect(fetchAdminClubs()).rejects.toThrow(APIError);

			try {
				await fetchAdminClubs();
			} catch (error) {
				if (error instanceof APIError) {
					expect(error.status).toBe(500);
				}
			}
		});
	});

	describe('fetchClubEvents', () => {
		it('should return events array on success', async () => {
			const mockEvents = [
				{
					id: '789',
					title: 'Tuesday Night Ride',
					description: 'Weekly social ride',
					activity_type: 'Ride',
					upcoming_occurrences: ['2026-03-01T18:00:00Z'],
					zone: 'America/Los_Angeles',
					address: '123 Main St',
					start_latlng: [45.5152, -122.6784],
					club_id: 123,
					private: false,
					women_only: false,
					joined: false
				}
			];

			fetchMock.mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => ({ events: mockEvents })
			});

			const events = await fetchClubEvents(123);

			expect(events).toEqual(mockEvents);
			expect(events).toHaveLength(1);
			expect(fetchMock).toHaveBeenCalledWith(
				'http://localhost:3000/v1/strava/clubs/123/events',
				expect.any(Object)
			);
		});

		it('should return empty array when no events', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => ({ events: [] })
			});

			const events = await fetchClubEvents(123);

			expect(events).toEqual([]);
		});

		it('should handle missing events field', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => ({})
			});

			const events = await fetchClubEvents(123);

			expect(events).toEqual([]);
		});

		it('should throw SessionExpiredError on 401', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 401,
				text: async () => 'Session expired'
			});

			await expect(fetchClubEvents(123)).rejects.toThrow(SessionExpiredError);
		});

		it('should throw RateLimitError on 429', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 429,
				text: async () =>
					JSON.stringify({
						error: 'Rate limit exceeded',
						retry_after_seconds: 600
					})
			});

			await expect(fetchClubEvents(123)).rejects.toThrow(RateLimitError);
		});
	});

	describe('checkSessionForImport', () => {
		it('should return true when session is valid', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => ({ valid: true })
			});

			const result = await checkSessionForImport();

			expect(result).toBe(true);
			expect(fetchMock).toHaveBeenCalledWith(
				'http://localhost:3000/v1/strava/check-session',
				expect.any(Object)
			);
		});

		it('should return false when session check fails', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 401,
				text: async () => 'Unauthorized'
			});

			const result = await checkSessionForImport();

			expect(result).toBe(false);
		});

		it('should return false on network error', async () => {
			fetchMock.mockRejectedValueOnce(new Error('Network error'));

			const result = await checkSessionForImport();

			expect(result).toBe(false);
		});
	});

	describe('logout', () => {
		it('should call logout endpoint with POST method', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => ({ success: true })
			});

			await logout();

			expect(fetchMock).toHaveBeenCalledWith(
				'http://localhost:3000/v1/strava/logout',
				expect.objectContaining({
					method: 'POST',
					credentials: 'include',
					headers: expect.objectContaining({
						'Content-Type': 'application/json'
					})
				})
			);
		});

		it('should handle logout errors gracefully', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 500,
				text: async () => 'Server error'
			});

			await expect(logout()).rejects.toThrow();
		});
	});

	describe('Error handling edge cases', () => {
		it('should handle malformed JSON in rate limit response', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 429,
				text: async () => 'Not valid JSON'
			});

			try {
				await fetchAdminClubs();
			} catch (error) {
				if (error instanceof RateLimitError) {
					// Should use default retry time when JSON parsing fails
					expect(error.retry_after_seconds).toBe(900);
				}
			}
		});

		it('should include error message in SessionExpiredError', async () => {
			const errorMessage = 'Your session has expired. Please reconnect.';
			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 401,
				text: async () => errorMessage
			});

			try {
				await fetchAdminClubs();
			} catch (error) {
				if (error instanceof SessionExpiredError) {
					expect(error.message).toBe(errorMessage);
				}
			}
		});

		it('should handle empty error responses', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 500,
				text: async () => ''
			});

			try {
				await fetchAdminClubs();
			} catch (error) {
				if (error instanceof APIError) {
					expect(error.message).toBe('HTTP 500');
				}
			}
		});
	});

	describe('Request configuration', () => {
		it('should include credentials in all requests', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => ({ clubs: [] })
			});

			await fetchAdminClubs();

			expect(fetchMock).toHaveBeenCalledWith(
				expect.any(String),
				expect.objectContaining({
					credentials: 'include'
				})
			);
		});

		it('should set Content-Type header', async () => {
			fetchMock.mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => ({ clubs: [] })
			});

			await fetchAdminClubs();

			expect(fetchMock).toHaveBeenCalledWith(
				expect.any(String),
				expect.objectContaining({
					headers: expect.objectContaining({
						'Content-Type': 'application/json'
					})
				})
			);
		});
	});
});
