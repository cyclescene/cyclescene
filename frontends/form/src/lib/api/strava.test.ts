import { describe, it, expect } from 'vitest';
import { formatRetryTime } from './strava';

/**
 * Note: getImportWebSocketUrl() uses PUBLIC_API_URL from environment variables,
 * which are set at build time. Testing it would require mocking the entire
 * $env/static/public module. Since it's a simple string transformation,
 * we skip testing it here and rely on integration tests.
 */

describe('Strava API Utilities', () => {
	describe('formatRetryTime', () => {
		it('should format seconds as rounded-up minutes', () => {
			// The function uses Math.ceil(seconds / 60) to round up to minutes
			expect(formatRetryTime(30)).toBe('1 minute'); // 30s = 0.5m, rounds to 1
			expect(formatRetryTime(1)).toBe('1 minute'); // 1s = 0.017m, rounds to 1
		});

		it('should format exact minutes correctly', () => {
			expect(formatRetryTime(60)).toBe('1 minute');
			expect(formatRetryTime(120)).toBe('2 minutes');
			expect(formatRetryTime(180)).toBe('3 minutes');
		});

		it('should round up partial minutes', () => {
			expect(formatRetryTime(90)).toBe('2 minutes'); // 1.5m rounds to 2
			expect(formatRetryTime(125)).toBe('3 minutes'); // 2.08m rounds to 3
			expect(formatRetryTime(61)).toBe('2 minutes'); // 1.017m rounds to 2
		});

		it('should handle edge cases', () => {
			expect(formatRetryTime(0)).toBe('0 minutes');
			expect(formatRetryTime(900)).toBe('15 minutes'); // Common rate limit duration
		});

		it('should pluralize correctly', () => {
			expect(formatRetryTime(59)).toBe('1 minute'); // singular
			expect(formatRetryTime(61)).toBe('2 minutes'); // plural
		});
	});
});
