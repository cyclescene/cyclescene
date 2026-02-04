import '@testing-library/jest-dom';
import { beforeAll, afterEach, afterAll } from 'vitest';

// Setup global test environment
beforeAll(() => {
	// Add any global setup here
});

afterEach(() => {
	// Cleanup after each test
});

afterAll(() => {
	// Cleanup after all tests
});

// Mock fetch globally if needed
globalThis.fetch = globalThis.fetch || (() => Promise.resolve({
	ok: true,
	status: 200,
	json: () => Promise.resolve({})
} as Response));
