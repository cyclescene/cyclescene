# Milestone 6: Testing & Documentation - DETAILED PLAN

**Status:** Ready to implement
**Created:** 2026-02-03
**Last Updated:** 2026-02-03

---

## Executive Summary

This milestone completes the Strava import feature by:
1. **Backend Testing:** Completing test coverage for handlers, WebSocket, and monitoring (5 gaps identified)
2. **Frontend Testing:** Building complete test infrastructure from scratch (11 files, ~2,055 lines of code)
3. **Documentation:** Adding Strava setup instructions to 4 README files
4. **Bug Fix:** Correcting invalid Go version in go.mod

**Current Test Coverage:**
- **Backend:** ~70% (needs completion)
- **Frontend:** 0% (needs full setup + tests)

---

## PHASE 1: Critical Bug Fix (5 minutes)

### 1.1 Fix go.mod Version Format

**Problem:** Invalid Go version causing build errors
**File:** `backend/go.mod:3`
**Current:** `go 1.24.0` ❌
**Correct:** `go 1.24` ✅

```bash
# Fix
cd backend
sed -i '' 's/go 1.24.0/go 1.24/' go.mod
go mod tidy
```

**Validation:**
```bash
go list  # Should not error
go build ./...
```

---

## PHASE 2: Backend Testing Completion (4-6 hours)

### 2.1 Complete Existing Test Files

#### 2.1.1 handler_test.go - Add Missing Tests

**File:** `backend/internal/api/strava/handler_test.go` (394 lines → ~500 lines)

**Current Status:** Partial coverage with incomplete tests

**Tests to Add:**

1. **Complete `TestGetAdminClubs_WithSession`** (currently marked incomplete)
   - Mock service to return clubs
   - Verify JSON serialization
   - Test int64 → string ID conversion (JavaScript precision safety)

2. **Add `TestGetClubEvents_WithSession_Success`**
   - Mock service to return events
   - Verify filtering logic
   - Test error handling

3. **Add `TestCheckSession_Valid`**
   - Test session validation
   - Verify JSON response structure

4. **Add `TestCheckSession_Expired`**
   - Test expired session handling
   - Verify error response

**Estimated Addition:** ~100 lines

---

#### 2.1.2 websocket_test.go - Complete Integration Tests

**File:** `backend/internal/api/strava/websocket_test.go` (463 lines → ~650 lines)

**Current Status:** Integration tests marked as skippable

**Tests to Unskip/Complete:**

1. **`TestWebSocketHandler_Integration`** (currently skipped)
   - Set up WebSocket test client
   - Send ImportRequest
   - Verify progress messages
   - Test complete/done flow
   - **Dependencies:** `gorilla/websocket` test utilities

2. **Add `TestImportSingleEvent_Success`**
   - Mock service methods
   - Test full import pipeline: fetch → coordinates → route → database
   - Verify result structure

3. **Add `TestImportSingleEvent_RateLimitError`**
   - Mock 429 response
   - Verify error message sent to client

4. **Add `TestImportSingleEvent_DuplicateError`**
   - Mock database duplicate error
   - Verify graceful handling

5. **Add `TestImportMultipleEvents_PartialFailure`**
   - Test 3 events: 2 succeed, 1 fails
   - Verify results array has all 3 entries
   - Verify error messages

6. **Add `TestConcurrentImports_MaxLimit`**
   - Test athlete with max concurrent imports
   - Verify rejection message

7. **Add `TestHeartbeat_Reconnect`**
   - Simulate heartbeat timeout
   - Verify cleanup logic

**Estimated Addition:** ~180 lines

---

### 2.2 Create New Test File

#### 2.2.1 monitoring_test.go - NEW FILE

**File:** `backend/internal/strava/monitoring_test.go` (NEW, ~200 lines)

**Current Status:** ❌ monitoring.go has ZERO test coverage (152 lines untested)

**Critical Functions to Test:**

1. **`TestLogAPICall_Success`**
   - Mock database insert
   - Verify all 16 parameters logged correctly
   - Test with rate limit headers (Limit + Usage + Next)

2. **`TestLogAPICall_DatabaseError`**
   - Mock database failure
   - Verify error logged but doesn't crash

3. **`TestCheckRateLimitWarning_AboveThreshold`**
   - Test 80% usage → warning logged
   - Test 90% usage → warning logged
   - Verify log message format

4. **`TestCheckRateLimitWarning_BelowThreshold`**
   - Test 50% usage → no warning
   - Verify no log output

5. **`TestCheckRateLimitWarning_NoHeaders`**
   - Test when headers are 0
   - Verify no division by zero

6. **`TestAPICallMetrics_AllFields`**
   - Test metrics struct creation
   - Verify all fields populated correctly

**Test Setup Needed:**
```go
// Mock repository interface
type mockMonitoringRepo struct {
    logAPICallFunc func(ctx context.Context, metrics strava.APICallMetrics) error
}

func (m *mockMonitoringRepo) LogAPICall(ctx context.Context, metrics strava.APICallMetrics) error {
    if m.logAPICallFunc != nil {
        return m.logAPICallFunc(ctx, metrics)
    }
    return nil
}
```

**Estimated:** ~200 lines

---

### 2.3 Backend Testing Summary

| File | Current Status | Action | Est. Lines Added |
|------|---------------|--------|------------------|
| handler_test.go | Partial (394 lines) | Complete 4 tests | +100 |
| websocket_test.go | Partial (463 lines) | Complete 7 tests | +180 |
| monitoring_test.go | ❌ Missing | Create new file | +200 |
| **Total** | | | **+480 lines** |

**Backend Test Commands:**
```bash
cd backend

# Run all Strava tests
go test ./internal/strava/... -v -cover
go test ./internal/api/strava/... -v -cover

# Check coverage
go test ./internal/strava/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Expected: >90% coverage
```

---

## PHASE 3: Frontend Testing Infrastructure Setup (2-3 hours)

### 3.1 Install Testing Framework

**File:** `frontends/form/package.json`

**Current State:** NO test framework installed

**Framework Choice:** Vitest (recommended for SvelteKit)

#### 3.1.1 Install Dependencies

```bash
cd frontends/form

# Core testing
pnpm add -D vitest @vitest/ui

# Svelte testing utilities
pnpm add -D @testing-library/svelte @testing-library/jest-dom

# Component testing
pnpm add -D @sveltejs/vite-plugin-svelte-testing

# DOM environment
pnpm add -D jsdom happy-dom

# Mocking utilities
pnpm add -D @vitest/spy
```

#### 3.1.2 Add Test Scripts to package.json

```json
{
  "scripts": {
    "test": "vitest run",
    "test:watch": "vitest",
    "test:ui": "vitest --ui",
    "test:coverage": "vitest run --coverage"
  }
}
```

#### 3.1.3 Create Vitest Config

**File:** `frontends/form/vitest.config.ts` (NEW)

```typescript
import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
  plugins: [sveltekit(), svelte({ hot: !process.env.VITEST })],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/setupTests.ts'],
    include: ['src/**/*.{test,spec}.{js,ts}'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/setupTests.ts',
      ]
    }
  }
});
```

#### 3.1.4 Create Test Setup File

**File:** `frontends/form/src/setupTests.ts` (NEW)

```typescript
import { expect, afterEach } from 'vitest';
import { cleanup } from '@testing-library/svelte';
import '@testing-library/jest-dom';

// Cleanup after each test
afterEach(() => {
  cleanup();
});

// Mock environment variables
process.env.PUBLIC_API_URL = 'http://localhost:8080';
process.env.PUBLIC_STRAVA_DEBUG = 'false';

// Global test utilities
global.fetch = vi.fn();
```

**Estimated Time:** 1 hour

---

### 3.2 Create Test Utilities

#### 3.2.1 Mock Factories

**File:** `frontends/form/src/lib/test-utils/strava-mocks.ts` (NEW, ~150 lines)

```typescript
import type {
  StravaClub,
  StravaGroupEvent,
  StravaRoute,
  AdminClubsResponse,
  ClubEventsResponse
} from '$lib/types/strava';

export function mockClub(overrides?: Partial<StravaClub>): StravaClub {
  return {
    id: '123456',
    name: 'Test Cycling Club',
    city: 'Portland',
    state: 'OR',
    country: 'USA',
    activity_types: ['cycling'],
    member_count: 150,
    url: 'https://strava.com/clubs/123456',
    ...overrides
  };
}

export function mockEvent(overrides?: Partial<StravaGroupEvent>): StravaGroupEvent {
  return {
    id: '789012',
    title: 'Weekly Group Ride',
    description: 'Casual social ride',
    club_id: '123456',
    address: 'Portland, OR',
    organized_by_athlete_id: '1111',
    route_id: '999',
    upcoming_occurrences: ['2026-02-10T18:00:00Z'],
    private: false,
    skill_levels: 'casual',
    terrain: 'mostly_flat',
    last_activity_at: '2026-02-03T12:00:00Z',
    route: mockRoute(),
    ...overrides
  };
}

export function mockRoute(overrides?: Partial<StravaRoute>): StravaRoute {
  return {
    id: '999',
    name: 'Test Route',
    distance: 16093.4, // 10 miles in meters
    elevation_gain: 100,
    ...overrides
  };
}

export function mockAdminClubsResponse(): AdminClubsResponse {
  return {
    clubs: [mockClub(), mockClub({ id: '654321', name: 'Another Club' })]
  };
}

export function mockClubEventsResponse(): ClubEventsResponse {
  return {
    events: [mockEvent(), mockEvent({ id: '789013', title: 'Morning Ride' })]
  };
}

// WebSocket message mocks
export function mockProgressMessage(type: 'heartbeat' | 'progress' | 'complete' | 'done' | 'error') {
  const base = {
    heartbeat: { type: 'heartbeat' },
    progress: {
      type: 'progress',
      event_id: '789012',
      event_title: 'Weekly Group Ride',
      step: 'fetching',
      status: 'in_progress'
    },
    complete: {
      type: 'complete',
      event_id: '789012',
      event_title: 'Weekly Group Ride',
      ride_id: 'abc123',
      edit_url: 'https://form.cyclescene.cc/edit/abc123'
    },
    done: {
      type: 'done',
      total: 2,
      successful: 2,
      failed: 0
    },
    error: {
      type: 'error',
      event_id: '789012',
      event_title: 'Weekly Group Ride',
      error: 'Rate limit exceeded'
    }
  };
  return base[type];
}
```

#### 3.2.2 Mock WebSocket

**File:** `frontends/form/src/lib/test-utils/mock-websocket.ts` (NEW, ~100 lines)

```typescript
import { vi } from 'vitest';

export class MockWebSocket {
  url: string;
  readyState: number = WebSocket.CONNECTING;
  onopen: ((ev: Event) => any) | null = null;
  onmessage: ((ev: MessageEvent) => any) | null = null;
  onerror: ((ev: Event) => any) | null = null;
  onclose: ((ev: CloseEvent) => any) | null = null;

  constructor(url: string) {
    this.url = url;
    setTimeout(() => this.simulateOpen(), 0);
  }

  send = vi.fn();
  close = vi.fn(() => {
    this.readyState = WebSocket.CLOSED;
    if (this.onclose) {
      this.onclose(new CloseEvent('close'));
    }
  });

  simulateOpen() {
    this.readyState = WebSocket.OPEN;
    if (this.onopen) {
      this.onopen(new Event('open'));
    }
  }

  simulateMessage(data: any) {
    if (this.onmessage) {
      this.onmessage(new MessageEvent('message', { data: JSON.stringify(data) }));
    }
  }

  simulateError() {
    if (this.onerror) {
      this.onerror(new Event('error'));
    }
  }

  simulateClose() {
    this.close();
  }
}

export function setupMockWebSocket() {
  global.WebSocket = MockWebSocket as any;
}
```

#### 3.2.3 Fetch Mocks

**File:** `frontends/form/src/lib/test-utils/fetch-mocks.ts` (NEW, ~80 lines)

```typescript
import { vi } from 'vitest';

export function mockFetchSuccess(data: any, status = 200) {
  return vi.fn().mockResolvedValue({
    ok: true,
    status,
    json: async () => data,
    headers: new Headers()
  });
}

export function mockFetchError(status: number, message: string) {
  return vi.fn().mockResolvedValue({
    ok: false,
    status,
    statusText: message,
    json: async () => ({ error: message })
  });
}

export function mockFetchRateLimit(retryAfter = 60) {
  return vi.fn().mockResolvedValue({
    ok: false,
    status: 429,
    statusText: 'Too Many Requests',
    json: async () => ({
      error: 'Rate limit exceeded',
      retry_after_seconds: retryAfter
    })
  });
}

export function setupFetchMock() {
  global.fetch = vi.fn();
}

export function resetFetchMock() {
  vi.restoreAllMocks();
}
```

**Total Utility Files:** 3 files, ~330 lines

---

## PHASE 4: Frontend Unit Tests (4-6 hours)

### 4.1 Core Utility Tests

#### 4.1.1 Error Classes Tests

**File:** `frontends/form/src/lib/types/strava.test.ts` (NEW, ~120 lines)

**Tests:**
- `TestRateLimitError_Construction`
- `TestRateLimitError_Properties`
- `TestSessionExpiredError_Construction`
- `TestSessionExpiredError_Message`

```typescript
import { describe, it, expect } from 'vitest';
import { RateLimitError, SessionExpiredError } from './strava';

describe('RateLimitError', () => {
  it('should create error with retry_after_seconds', () => {
    const error = new RateLimitError(60);
    expect(error.message).toBe('Rate limit exceeded');
    expect(error.retry_after_seconds).toBe(60);
    expect(error).toBeInstanceOf(Error);
  });

  it('should format message with retry time', () => {
    const error = new RateLimitError(120);
    expect(error.toString()).toContain('Rate limit exceeded');
  });
});

describe('SessionExpiredError', () => {
  it('should create error with correct message', () => {
    const error = new SessionExpiredError();
    expect(error.message).toBe('Session expired');
    expect(error).toBeInstanceOf(Error);
  });
});
```

#### 4.1.2 API Client Tests

**File:** `frontends/form/src/lib/api/strava.test.ts` (NEW, ~400 lines)

**Tests:**
1. `TestInitiateAuth_Success` - OAuth popup flow
2. `TestInitiateAuth_UserCancels` - User closes popup
3. `TestInitiateAuth_Timeout` - 5min timeout
4. `TestCheckSession_Valid` - Valid session
5. `TestCheckSession_Expired` - Expired session (401)
6. `TestFetchAdminClubs_Success` - Returns clubs
7. `TestFetchAdminClubs_RateLimit` - 429 error with retry
8. `TestFetchClubEvents_Success` - Returns events
9. `TestFetchClubEvents_Empty` - No events
10. `TestCheckSessionForImport_Valid` - Pre-import check
11. `TestLogout_Success` - Session cleared
12. `TestFormatRetryTime_Seconds` - "60 seconds"
13. `TestFormatRetryTime_Minutes` - "2 minutes"
14. `TestGetImportWebSocketUrl_Production` - Correct URL

```typescript
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  initiateAuth,
  checkSession,
  fetchAdminClubs,
  formatRetryTime,
  RateLimitError,
  SessionExpiredError
} from './strava';
import { mockFetchSuccess, mockFetchError, mockFetchRateLimit } from '../test-utils/fetch-mocks';
import { mockAdminClubsResponse } from '../test-utils/strava-mocks';

describe('strava API client', () => {
  beforeEach(() => {
    global.fetch = vi.fn();
    global.window.open = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('fetchAdminClubs', () => {
    it('should return clubs on success', async () => {
      const mockData = mockAdminClubsResponse();
      global.fetch = mockFetchSuccess(mockData);

      const result = await fetchAdminClubs();

      expect(result).toEqual(mockData.clubs);
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/strava/clubs/admin'),
        expect.objectContaining({ credentials: 'include' })
      );
    });

    it('should throw RateLimitError on 429', async () => {
      global.fetch = mockFetchRateLimit(120);

      await expect(fetchAdminClubs()).rejects.toThrow(RateLimitError);
      await expect(fetchAdminClubs()).rejects.toThrow('Rate limit exceeded');
    });

    it('should throw SessionExpiredError on 401', async () => {
      global.fetch = mockFetchError(401, 'Unauthorized');

      await expect(fetchAdminClubs()).rejects.toThrow(SessionExpiredError);
    });
  });

  describe('formatRetryTime', () => {
    it('should format seconds', () => {
      expect(formatRetryTime(45)).toBe('45 seconds');
    });

    it('should format minutes', () => {
      expect(formatRetryTime(120)).toBe('2 minutes');
    });

    it('should format single minute', () => {
      expect(formatRetryTime(60)).toBe('1 minute');
    });
  });
});
```

#### 4.1.3 WebSocket Client Tests

**File:** `frontends/form/src/lib/utils/strava-websocket.test.ts` (NEW, ~500 lines)

**Tests:**
1. `TestWebSocketClient_Connect` - Connection established
2. `TestWebSocketClient_SendRequest` - Sends ImportRequest
3. `TestWebSocketClient_ReceiveHeartbeat` - Heartbeat handling
4. `TestWebSocketClient_ReceiveProgress` - Progress updates
5. `TestWebSocketClient_ReceiveComplete` - Completion messages
6. `TestWebSocketClient_ReceiveDone` - Final done message
7. `TestWebSocketClient_ReceiveError` - Error handling
8. `TestWebSocketClient_Reconnect` - Automatic reconnection
9. `TestWebSocketClient_MaxReconnectAttempts` - Stops after 3
10. `TestWebSocketClient_HeartbeatTimeout` - 60s timeout
11. `TestWebSocketClient_ActivityTimeout` - 15s warning
12. `TestWebSocketClient_ManualRetry` - User-initiated retry
13. `TestWebSocketClient_Stop` - User-initiated stop
14. `TestWebSocketClient_ResultsAccumulation` - Partial results

```typescript
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { StravaImportWebSocket } from './strava-websocket';
import { MockWebSocket, setupMockWebSocket } from '../test-utils/mock-websocket';
import { mockProgressMessage } from '../test-utils/strava-mocks';

describe('StravaImportWebSocket', () => {
  let mockWs: MockWebSocket;
  let client: StravaImportWebSocket;
  let callbacks: any;

  beforeEach(() => {
    setupMockWebSocket();
    callbacks = {
      onStateChange: vi.fn(),
      onProgress: vi.fn(),
      onComplete: vi.fn(),
      onDone: vi.fn(),
      onError: vi.fn()
    };
  });

  afterEach(() => {
    client?.close();
    vi.clearAllTimers();
  });

  it('should connect and send import request', async () => {
    client = new StravaImportWebSocket(callbacks);

    await client.connect('test@example.com', [
      { event_id: '123', overrides: {} }
    ]);

    const ws = client['ws'];
    expect(ws).toBeDefined();
    expect(ws.send).toHaveBeenCalledWith(
      expect.stringContaining('"email":"test@example.com"')
    );
  });

  it('should handle progress messages', async () => {
    client = new StravaImportWebSocket(callbacks);
    await client.connect('test@example.com', [{ event_id: '123', overrides: {} }]);

    const ws = client['ws'] as MockWebSocket;
    ws.simulateMessage(mockProgressMessage('progress'));

    expect(callbacks.onProgress).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'progress',
        event_id: '789012',
        step: 'fetching'
      })
    );
  });

  it('should reconnect on connection loss', async () => {
    vi.useFakeTimers();
    client = new StravaImportWebSocket(callbacks);
    await client.connect('test@example.com', [{ event_id: '123', overrides: {} }]);

    const firstWs = client['ws'] as MockWebSocket;
    firstWs.simulateClose();

    await vi.advanceTimersByTimeAsync(2000); // Wait for reconnect

    expect(callbacks.onStateChange).toHaveBeenCalledWith('connecting');
    expect(client.getReconnectAttempts()).toEqual({ current: 1, max: 3 });
  });

  it('should stop after max reconnect attempts', async () => {
    vi.useFakeTimers();
    client = new StravaImportWebSocket(callbacks);
    await client.connect('test@example.com', [{ event_id: '123', overrides: {} }]);

    // Simulate 3 disconnects
    for (let i = 0; i < 3; i++) {
      const ws = client['ws'] as MockWebSocket;
      ws.simulateClose();
      await vi.advanceTimersByTimeAsync(10000);
    }

    expect(client.getState()).toBe('error');
    expect(callbacks.onError).toHaveBeenCalledWith(
      expect.stringContaining('Max reconnection attempts')
    );
  });

  it('should accumulate partial results', async () => {
    client = new StravaImportWebSocket(callbacks);
    await client.connect('test@example.com', [
      { event_id: '123', overrides: {} },
      { event_id: '456', overrides: {} }
    ]);

    const ws = client['ws'] as MockWebSocket;

    // First event completes
    ws.simulateMessage(mockProgressMessage('complete'));

    const results = client.getCompletedResults();
    expect(results).toHaveLength(1);
    expect(results[0].event_id).toBe('789012');
  });
});
```

**Total Core Tests:** 3 files, ~1,020 lines

---

### 4.2 Component Tests

#### 4.2.1 EmailInput Component

**File:** `frontends/form/src/lib/components/strava/EmailInput.test.ts` (NEW, ~200 lines)

**Tests:**
1. `TestEmailInput_Renders` - Component renders
2. `TestEmailInput_ValidEmail` - Accepts valid email
3. `TestEmailInput_InvalidEmail` - Shows error for invalid
4. `TestEmailInput_EmptyEmail` - Shows error for empty
5. `TestEmailInput_WhitespaceTrims` - Trims whitespace
6. `TestEmailInput_GroupCodeSelector` - Shows group codes
7. `TestEmailInput_Submit` - Calls onSubmit callback

```typescript
import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import EmailInput from './EmailInput.svelte';

describe('EmailInput', () => {
  it('should render input and submit button', () => {
    render(EmailInput, { props: { onSubmit: vi.fn() } });

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /continue/i })).toBeInTheDocument();
  });

  it('should validate valid email', async () => {
    const onSubmit = vi.fn();
    render(EmailInput, { props: { onSubmit } });

    const input = screen.getByLabelText(/email/i);
    await fireEvent.input(input, { target: { value: 'test@example.com' } });
    await fireEvent.click(screen.getByRole('button', { name: /continue/i }));

    expect(onSubmit).toHaveBeenCalledWith({
      email: 'test@example.com',
      groupCode: expect.any(String)
    });
  });

  it('should show error for invalid email', async () => {
    render(EmailInput, { props: { onSubmit: vi.fn() } });

    const input = screen.getByLabelText(/email/i);
    await fireEvent.input(input, { target: { value: 'invalid-email' } });
    await fireEvent.click(screen.getByRole('button', { name: /continue/i }));

    expect(screen.getByText(/valid email/i)).toBeInTheDocument();
  });

  it('should trim whitespace from email', async () => {
    const onSubmit = vi.fn();
    render(EmailInput, { props: { onSubmit } });

    const input = screen.getByLabelText(/email/i);
    await fireEvent.input(input, { target: { value: '  test@example.com  ' } });
    await fireEvent.click(screen.getByRole('button', { name: /continue/i }));

    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ email: 'test@example.com' })
    );
  });
});
```

#### 4.2.2 ImportProgress Component

**File:** `frontends/form/src/lib/components/strava/ImportProgress.test.ts` (NEW, ~250 lines)

**Tests:**
1. `TestImportProgress_RendersSingleEvent` - Shows single event
2. `TestImportProgress_RendersMultipleEvents` - Shows multiple
3. `TestImportProgress_CalculatesPercentage` - Correct percentage
4. `TestImportProgress_ShowsStepProgress` - Step indicators
5. `TestImportProgress_ShowsReconnecting` - Reconnection UI
6. `TestImportProgress_ShowsHeartbeat` - Heartbeat indicator
7. `TestImportProgress_ShowsRetryButton` - After failure
8. `TestImportProgress_DisabledDuringImport` - Button disabled

```typescript
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import ImportProgress from './ImportProgress.svelte';

describe('ImportProgress', () => {
  it('should calculate progress percentage correctly', () => {
    render(ImportProgress, {
      props: {
        events: [
          { id: '1', title: 'Event 1', status: 'complete', step: 'database' },
          { id: '2', title: 'Event 2', status: 'in_progress', step: 'coordinates' },
          { id: '3', title: 'Event 3', status: 'pending', step: null }
        ],
        isConnected: true,
        reconnectAttempts: { current: 0, max: 3 }
      }
    });

    // 1 complete (100%) + 1 at coordinates (50%) + 1 pending (0%) = 150/300 = 50%
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '50');
  });

  it('should show reconnecting state', () => {
    render(ImportProgress, {
      props: {
        events: [],
        isConnected: false,
        reconnectAttempts: { current: 2, max: 3 }
      }
    });

    expect(screen.getByText(/reconnecting/i)).toBeInTheDocument();
    expect(screen.getByText(/attempt 2 of 3/i)).toBeInTheDocument();
  });

  it('should show step progress for each event', () => {
    render(ImportProgress, {
      props: {
        events: [
          { id: '1', title: 'Event 1', status: 'in_progress', step: 'route' }
        ],
        isConnected: true,
        reconnectAttempts: { current: 0, max: 3 }
      }
    });

    expect(screen.getByText(/fetching/i)).toBeInTheDocument();
    expect(screen.getByText(/coordinates/i)).toBeInTheDocument();
    expect(screen.getByText(/route/i)).toBeInTheDocument();
    expect(screen.getByText(/database/i)).toBeInTheDocument();
  });
});
```

#### 4.2.3 EventCard Component (Lower Priority)

**File:** `frontends/form/src/lib/components/strava/EventCard.test.ts` (NEW, ~150 lines)

Basic rendering and interaction tests.

**Total Component Tests:** 3 files, ~600 lines

---

### 4.3 Frontend Testing Summary

| Category | Files | Lines | Priority |
|----------|-------|-------|----------|
| Test Infrastructure | 4 | ~180 | Critical |
| Test Utilities | 3 | ~330 | Critical |
| Core Unit Tests | 3 | ~1,020 | High |
| Component Tests | 3 | ~600 | Medium |
| **Total** | **13** | **~2,130** | |

**Frontend Test Commands:**
```bash
cd frontends/form

# Run tests
npm run test

# Watch mode
npm run test:watch

# Coverage report
npm run test:coverage

# UI mode
npm run test:ui

# Expected: >80% coverage for Strava code
```

---

## PHASE 5: Documentation Updates (2-3 hours)

### 5.1 Root README

**File:** `/Users/joseduarte/personal/cyclescene/README.md`

**Current:** No mention of Strava integration

**Add Section (after line 155):**

```markdown
### Strava Integration

Users can import group rides directly from Strava clubs they admin/own:
- **OAuth Flow**: Secure Strava authentication
- **Club Filtering**: Only shows clubs where user is admin/owner
- **Event Import**: Import upcoming group events with routes
- **Real-time Progress**: WebSocket-based progress tracking
- **Magic Links**: Email edit links for imported events

See [Strava Documentation](docs/strava/README.md) for setup instructions.
```

**Location:** After "Data Flow" section, before "Privacy and Analytics"

---

### 5.2 Backend README

**File:** `/Users/joseduarte/personal/cyclescene/backend/README.md`

**Current:** No API endpoint documentation for Strava

**Add Section (after line 172):**

```markdown
## Strava API Endpoints

### Authentication

**`GET /strava/auth/initiate`**
- Initiates Strava OAuth flow
- Query params: `city` (optional, defaults to 'pdx')
- Returns: `{ url: string }` - OAuth authorization URL

**`GET /strava/auth/callback`**
- OAuth callback endpoint
- Query params: `code`, `state`
- Sets session cookie
- Redirects to frontend

**`POST /strava/auth/logout`**
- Clears Strava session
- Returns: `{ message: "Logged out successfully" }`

**`GET /strava/auth/session`**
- Checks current session status
- Returns: `{ valid: boolean, athlete_id?: string }`

### Clubs & Events

**`GET /strava/clubs/admin`**
- Gets clubs where user is admin/owner
- Requires: Valid session cookie
- Returns: `{ clubs: StravaClub[] }`
- Filters: Cycling clubs matching city

**`GET /strava/clubs/:clubId/events`**
- Gets upcoming events for a club
- Requires: Valid session cookie
- Returns: `{ events: StravaGroupEvent[] }`

### Import

**`WS /strava/import`**
- WebSocket endpoint for importing events
- Requires: Valid session cookie
- Request: `ImportRequest` (email, events array)
- Messages: `ProgressMessage` (heartbeat, progress, complete, done, error)

### Rate Limiting

All endpoints respect Strava's rate limits:
- 15-minute limit: 100 requests
- Daily limit: 1,000 requests
- 429 responses include `retry_after_seconds`

### Environment Variables

```bash
STRAVA_CLIENT_ID=your_client_id
STRAVA_CLIENT_SECRET=your_client_secret
STRAVA_REDIRECT_URI=https://api.cyclescene.cc/strava/auth/callback
STRAVA_DEBUG=false  # Enable debug logging
```

See [Strava Documentation](../docs/strava/README.md) for detailed setup.
```

**Location:** After "API Documentation" section (line 172)

---

### 5.3 Frontend Form README

**File:** `/Users/joseduarte/personal/cyclescene/frontends/form/README.md`

**Current:** No mention of Strava import

**Add Section (after line 56):**

```markdown
## Strava Import Feature

Import group rides directly from Strava clubs you admin or own.

### How It Works

1. **Connect**: Click "Import from Strava" button
2. **Authenticate**: OAuth popup to authorize Strava access
3. **Select**: Choose clubs and events to import
4. **Import**: Real-time progress tracking via WebSocket
5. **Edit**: Receive email with edit links for all imported events

### Features

- **Filtered Clubs**: Only shows clubs where you're admin/owner
- **City Filtering**: Automatically filters to Portland (PDX) or Salt Lake City (SLC) clubs
- **Route Support**: Imports route distance and elevation
- **Batch Import**: Import multiple events simultaneously
- **Progress Tracking**: Real-time 4-step pipeline (fetch → coordinates → route → database)
- **Error Recovery**: Automatic reconnection with partial results preserved
- **Rate Limit Handling**: Graceful handling of Strava API limits

### UI Components

- `StravaImportButton.svelte` - Entry point button
- `StravaImport.svelte` - Main orchestrator (5-step flow)
- `EmailInput.svelte` - Email form with group code selector
- `ClubList.svelte` - Expandable club list
- `EventCard.svelte` - Event display with customization
- `ImportProgress.svelte` - Real-time progress visualization
- `ImportResults.svelte` - Final results with edit links

### Testing

```bash
# Run Strava-specific tests
npm run test -- strava

# Watch mode
npm run test:watch

# Coverage
npm run test:coverage
```

### Debug Mode

Enable debug logging:
```bash
PUBLIC_STRAVA_DEBUG=true npm run dev
```

See [Strava Documentation](../../docs/strava/README.md) for developer setup.
```

**Location:** After "Styling" section, before "Troubleshooting"

---

### 5.4 Create Strava OAuth Learnings Doc

**File:** `/Users/joseduarte/personal/cyclescene/docs/strava/STRAVA_OAUTH_LEARNINGS.md` (NEW)

```markdown
# Strava OAuth & Group Events API - Learnings & Gotchas

**Created:** 2026-02-03
**Last Updated:** 2026-02-03

This document captures key learnings, gotchas, and non-obvious behaviors discovered while implementing Strava OAuth and Group Events import.

---

## OAuth Implementation

### Popup vs Redirect Flow

**Decision:** Used popup window for better UX

**Why:**
- Keeps user in-app (no full redirect)
- Maintains form state
- Modern browsers allow popups for user-initiated actions

**Gotchas:**
- Must detect popup blockers
- Need polling fallback if `postMessage` fails
- 5-minute timeout to prevent hanging

**Code Pattern:**
```typescript
const popup = window.open(authUrl, '_blank', 'width=600,height=800');
window.addEventListener('message', (event) => {
  if (event.data.type === 'strava-auth-success') {
    // Handle success
  }
});
```

### Session Management

**Decision:** Server-side sessions with HTTP-only cookies

**Why:**
- More secure than localStorage
- No XSS exposure
- Automatic expiry (15 minutes)

**Gotchas:**
- Must use `credentials: 'include'` in fetch calls
- CORS must allow credentials
- Session cleanup required (in-memory store with goroutine)

---

## Group Events API

### Undocumented Endpoint

**Key Discovery:** `/api/v3/clubs/{id}/group_events` is NOT in official docs

**Source:** Found via network inspection in Strava web app

**Behavior:**
- Returns 404 for clubs without events (not documented)
- Requires read scope (not read_all)
- No pagination (returns all upcoming events)

**Gotcha:** Must handle 404 as "no events" (not an error)

### Admin Detection

**Challenge:** API doesn't provide direct "is_admin" field on clubs list endpoint

**Solution:** Must call `/api/v3/clubs/{id}` individually for each club

**Fields:**
- `membership` - "member" or "pending"
- `admin` - Boolean (true if admin)
- `owner` - Boolean (true if owner)

**Optimization:** Only fetch details for cycling clubs in target city

### Event Location Data

**Inconsistency:** Events may have:
1. `address` only (string like "Portland, OR")
2. `address` + lat/lng (decimal degrees)
3. Neither (null/missing)

**Solution:** Fallback chain:
1. Use event lat/lng if present
2. Geocode `address` if present
3. Use club city as fallback

**Gotcha:** Some events have `address: ""` (empty string, not null) - must check `HasLocation()` method

### Route Data

**Structure:** Events with routes have `route_id`

**Gotcha:** Must make separate call to `/api/v3/routes/{id}` for full route data

**Returned:**
- `distance` (meters, not miles)
- `elevation_gain` (meters)
- `map.summary_polyline` (encoded polyline)

**Not Returned:**
- Turn-by-turn directions
- Waypoints
- Segment details

### Timezone Handling

**Challenge:** Event times are in club's local timezone, but API returns UTC strings

**Fields:**
- `upcoming_occurrences` - Array of ISO 8601 strings (UTC)
- No timezone field on event object

**Solution:** Parse as UTC, display in user's local time

**Gotcha:** Strava web app shows times in club timezone, but API doesn't provide timezone data

---

## Rate Limiting

### Dual Header System

**Discovery:** Strava uses TWO sets of rate limit headers:

1. **15-minute window:**
   - `X-Ratelimit-Limit`: 100
   - `X-Ratelimit-Usage`: Current usage
   - `X-Ratelimit-Next`: Unix timestamp

2. **Daily window:**
   - `X-Readratelimit-Limit`: 1000
   - `X-Readratelimit-Usage`: Current usage

**Gotcha:** Header names differ (`Ratelimit` vs `Readratelimit`)

**Best Practice:** Monitor both, prioritize 15-minute limit

### 429 Response Handling

**Response Body:**
```json
{
  "message": "Rate Limit Exceeded",
  "errors": [
    {
      "resource": "Application",
      "field": "rate limit",
      "code": "exceeded"
    }
  ]
}
```

**Headers:**
- `Retry-After` - Seconds until reset (NOT standard format)

**Implementation:**
- Parse `Retry-After` header
- Return `RateLimitError` with `retry_after_seconds`
- Display countdown timer in UI

---

## WebSocket Import

### Heartbeat Necessity

**Why:** WebSocket connections can silently die

**Implementation:**
- Server sends heartbeat every 30s
- Client expects message every 60s
- 15s warning before timeout

**Gotcha:** Cloud Run can terminate idle WebSockets after 60s (must send heartbeat)

### Reconnection Strategy

**Pattern:** Exponential backoff with max attempts

**Details:**
- Max 3 attempts
- Delays: 2s, 4s, 8s
- Preserve partial results
- Allow manual retry

**Gotcha:** Must accumulate results before reconnect (send to client before disconnect)

### Message Ordering

**Not Guaranteed:** WebSocket messages may arrive out of order

**Solution:** Use `event_id` field to match messages to events

**Progress Steps:** Always in order: fetching → coordinates → route → database

---

## JavaScript Precision Issues

### Event ID Conversion

**Problem:** Strava event IDs are int64 (up to 2^63-1)

**JavaScript limit:** 2^53-1 (Number.MAX_SAFE_INTEGER)

**Solution:** Convert to strings server-side before JSON serialization

**Code:**
```go
type EventResponse struct {
    ID string `json:"id"` // NOT int64
}
```

**Gotcha:** Must convert back to int64 for API calls

---

## City Filtering

### Club City Matching

**Fields Available:**
- `city` - String (e.g., "Portland")
- `state` - String (e.g., "OR")
- `country` - String (e.g., "USA")

**Challenge:** No standardized city names

**Solution:** Case-insensitive partial matching

**Examples:**
- "Portland" matches "portland" and "Portland, OR"
- "Salt Lake City" matches "Salt Lake City" and "SLC"

**Configuration:**
```go
var CityConfig = map[string]CityInfo{
    "pdx": {
        Name: "Portland",
        Aliases: []string{"Portland", "PDX"},
        State: "OR",
    },
    "slc": {
        Name: "Salt Lake City",
        Aliases: []string{"Salt Lake City", "SLC", "Salt Lake"},
        State: "UT",
    },
}
```

---

## Testing Challenges

### OAuth Testing

**Challenge:** Can't easily mock OAuth popup flow in tests

**Solution:**
- Unit test OAuth initiation (URL generation)
- Unit test callback handling (token exchange)
- Skip integration tests for full popup flow

### WebSocket Testing

**Challenge:** Hard to test real WebSocket connections

**Solution:**
- Mock WebSocket class in tests
- Test message handling independently
- Test reconnection logic with fake timers

**Library:** `MockWebSocket` test utility

### Rate Limit Testing

**Challenge:** Hard to trigger real rate limits

**Solution:**
- Mock API responses with 429 status
- Test header parsing independently
- Test retry logic with fake delays

---

## Production Considerations

### Environment Variables

**Required:**
- `STRAVA_CLIENT_ID` - OAuth client ID
- `STRAVA_CLIENT_SECRET` - OAuth secret
- `STRAVA_REDIRECT_URI` - Full callback URL

**Optional:**
- `STRAVA_DEBUG=true` - Enable debug logging

**Security:** Never log access tokens or client secrets

### Monitoring

**Key Metrics:**
- API call count (track usage)
- Rate limit usage (warn at 80%)
- Session cleanup (prevent memory leaks)
- WebSocket connection count

**Logged to Database:**
```sql
CREATE TABLE strava_api_calls (
    endpoint TEXT,
    status_code INT,
    rate_limit_remaining INT,
    response_time_ms INT,
    timestamp DATETIME
);
```

### Error Handling

**User-Facing Errors:**
- Rate limit exceeded → Show retry countdown
- Session expired → Prompt re-authentication
- Network error → Offer retry button
- Duplicate event → Show existing edit link

**Never Show:**
- Stack traces
- API tokens
- Internal error details

---

## Future Improvements

### Potential Enhancements

1. **Token Refresh:** Currently sessions expire after 15min. Could implement refresh tokens.

2. **Webhook Support:** Strava offers webhooks for event updates. Could auto-sync changes.

3. **Batch Route Fetching:** Currently fetch routes one-by-one. Could parallelize.

4. **City Auto-Detection:** Could use IP geolocation to default city.

5. **Caching:** Could cache club/event data for 10-15 minutes to reduce API calls.

### Known Limitations

1. **No Event Editing:** Once imported, events must be edited in CycleScene (not synced back to Strava)

2. **No Recurring Events:** Strava supports recurring events, but we only import next occurrence

3. **No RSVP Data:** Can't see who RSVP'd to Strava events

4. **No Photo Import:** Event photos stay on Strava

---

## References

- [Strava API Documentation](https://developers.strava.com/docs/reference/)
- [OAuth 2.0 Spec](https://oauth.net/2/)
- [WebSocket Protocol](https://tools.ietf.org/html/rfc6455)
- [JavaScript Number Precision](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MAX_SAFE_INTEGER)

---

**Contributors:** Add your learnings here as you discover new behaviors!
```

**Estimated:** ~600 lines

---

### 5.5 Documentation Summary

| File | Type | Action | Est. Lines |
|------|------|--------|-----------|
| README.md (root) | Update | Add Strava section | +10 |
| backend/README.md | Update | Add API endpoints | +60 |
| frontends/form/README.md | Update | Add Strava feature docs | +50 |
| STRAVA_OAUTH_LEARNINGS.md | Create | New learnings doc | +600 |
| **Total** | | | **+720 lines** |

---

## PHASE 6: End-to-End Manual Testing (2-3 hours)

### 6.1 Manual Test Checklist

**Test Environment:**
- Local development setup
- Real Strava account
- Multiple clubs (admin of at least 1)

**Tests:**

#### OAuth Flow
- [ ] Click "Import from Strava" button
- [ ] Popup opens with Strava OAuth screen
- [ ] Authorize access
- [ ] Popup closes automatically
- [ ] Session cookie set
- [ ] Redirected to email input

#### Club & Event Loading
- [ ] Only admin/owner clubs displayed
- [ ] Clubs filtered by city (Portland/SLC)
- [ ] Cycling clubs only (no running clubs)
- [ ] Click club to expand
- [ ] Events load (lazy loading)
- [ ] Only upcoming events shown
- [ ] Event details correct (title, date, location)

#### Import Single Event
- [ ] Enter valid email
- [ ] Select 1 event
- [ ] Click import
- [ ] WebSocket connects
- [ ] Progress bar updates
- [ ] All 4 steps complete (fetching → coordinates → route → database)
- [ ] Success message shown
- [ ] Edit link displayed
- [ ] Email received with edit link
- [ ] Edit link works

#### Import Multiple Events
- [ ] Select 3+ events from different clubs
- [ ] Click import
- [ ] All events show progress
- [ ] Progress percentage updates correctly
- [ ] All events complete
- [ ] Results summary correct (X successful, Y failed)
- [ ] All edit links work
- [ ] Email contains all event titles

#### Error Scenarios
- [ ] Rate limit (if possible) - retry countdown shown
- [ ] Session expired - prompt to re-authenticate
- [ ] Network error - retry button works
- [ ] Duplicate event - existing link shown
- [ ] Invalid email - validation error
- [ ] Popup blocked - fallback message

#### Reconnection
- [ ] Start import
- [ ] Kill WebSocket connection (browser DevTools)
- [ ] Reconnecting message shown
- [ ] Connection restored
- [ ] Import continues
- [ ] Partial results preserved

#### Mobile Testing
- [ ] Test on iPhone Safari
- [ ] Test on Android Chrome
- [ ] OAuth popup works on mobile
- [ ] UI responsive
- [ ] Progress updates work

#### Accessibility
- [ ] Keyboard navigation works
- [ ] Screen reader announces progress
- [ ] Focus management correct
- [ ] ARIA labels present
- [ ] Error messages announced

### 6.2 Performance Testing

- [ ] User in 10+ clubs → loads in <5s
- [ ] Import 10 events → completes in <2min
- [ ] Memory usage stable (no leaks)
- [ ] WebSocket cleanup on unmount

### 6.3 Browser Compatibility

- [ ] Chrome/Edge (Chromium)
- [ ] Firefox
- [ ] Safari (macOS)
- [ ] Safari (iOS)
- [ ] Chrome (Android)

---

## SUCCESS CRITERIA

### Backend
- ✅ All Go tests pass: `go test ./... -v`
- ✅ Test coverage >90% for Strava code
- ✅ No build errors: `go build ./...`
- ✅ monitoring.go has test coverage
- ✅ handler/websocket tests complete

### Frontend
- ✅ Testing framework installed and configured
- ✅ All TypeScript tests pass: `npm run test`
- ✅ Test coverage >80% for Strava code
- ✅ No TypeScript errors: `npm run check`
- ✅ Build succeeds: `npm run build`

### Documentation
- ✅ Root README mentions Strava
- ✅ Backend README documents API endpoints
- ✅ Frontend README documents import feature
- ✅ Learnings doc created with gotchas
- ✅ Environment variables documented

### Manual Testing
- ✅ Full OAuth flow works end-to-end
- ✅ Single event import succeeds
- ✅ Multi-event import succeeds
- ✅ Rate limit handling works
- ✅ Reconnection works
- ✅ Mobile browsers work
- ✅ Accessibility verified

---

## ESTIMATED TIMELINE

| Phase | Description | Time |
|-------|-------------|------|
| **Phase 1** | Fix go.mod bug | 5 minutes |
| **Phase 2** | Backend tests (480 lines) | 4-6 hours |
| **Phase 3** | Frontend infrastructure | 2-3 hours |
| **Phase 4** | Frontend tests (2,130 lines) | 4-6 hours |
| **Phase 5** | Documentation (720 lines) | 2-3 hours |
| **Phase 6** | Manual E2E testing | 2-3 hours |
| **Total** | | **14-22 hours** |

---

## IMPLEMENTATION ORDER

1. **CRITICAL:** Fix go.mod (blocking)
2. **Backend:** Complete tests (monitoring → handler → websocket)
3. **Frontend:** Setup infrastructure (dependencies → config → utilities)
4. **Frontend:** Core tests (types → API → WebSocket)
5. **Frontend:** Component tests (EmailInput → ImportProgress)
6. **Documentation:** Update READMEs + create learnings doc
7. **Manual Testing:** E2E validation

---

## VALIDATION COMMANDS

```bash
# Fix go.mod
cd backend
sed -i '' 's/go 1.24.0/go 1.24/' go.mod

# Backend tests
go test ./internal/strava/... -v -cover
go test ./internal/api/strava/... -v -cover
go test ./... -v  # All tests

# Frontend setup
cd ../frontends/form
pnpm add -D vitest @vitest/ui @testing-library/svelte jsdom

# Frontend tests
npm run test
npm run test:coverage

# Build verification
cd ../../backend
go build ./...

cd ../frontends/form
npm run check
npm run build
```

---

## FILES TO CREATE (Summary)

### Backend (1 file)
- `backend/internal/strava/monitoring_test.go` (~200 lines)

### Frontend (13 files)
- `frontends/form/vitest.config.ts` (~30 lines)
- `frontends/form/src/setupTests.ts` (~20 lines)
- `frontends/form/src/lib/test-utils/strava-mocks.ts` (~150 lines)
- `frontends/form/src/lib/test-utils/mock-websocket.ts` (~100 lines)
- `frontends/form/src/lib/test-utils/fetch-mocks.ts` (~80 lines)
- `frontends/form/src/lib/types/strava.test.ts` (~120 lines)
- `frontends/form/src/lib/api/strava.test.ts` (~400 lines)
- `frontends/form/src/lib/utils/strava-websocket.test.ts` (~500 lines)
- `frontends/form/src/lib/components/strava/EmailInput.test.ts` (~200 lines)
- `frontends/form/src/lib/components/strava/ImportProgress.test.ts` (~250 lines)
- `frontends/form/src/lib/components/strava/EventCard.test.ts` (~150 lines)
- `frontends/form/src/lib/components/strava/ClubList.test.ts` (~100 lines)
- `frontends/form/src/lib/components/strava/StravaImport.test.ts` (~200 lines)

### Documentation (1 file)
- `docs/strava/STRAVA_OAUTH_LEARNINGS.md` (~600 lines)

### FILES TO MODIFY (Summary)

### Backend (2 files)
- `backend/go.mod` (fix version)
- `backend/internal/api/strava/handler_test.go` (+100 lines)
- `backend/internal/api/strava/websocket_test.go` (+180 lines)

### Frontend (1 file)
- `frontends/form/package.json` (add test scripts + dependencies)

### Documentation (3 files)
- `README.md` (+10 lines)
- `backend/README.md` (+60 lines)
- `frontends/form/README.md` (+50 lines)

---

## TOTAL LINES OF CODE

- **Backend:** ~480 lines (tests)
- **Frontend:** ~2,130 lines (infrastructure + tests)
- **Documentation:** ~720 lines
- **TOTAL:** ~3,330 lines

---

## NEXT STEPS

1. Review this detailed plan with stakeholder
2. Get approval on scope and priorities
3. Begin implementation (Phase 1 → Phase 6)
4. Mark Milestone 6 complete when all success criteria met

**Ready to begin implementation!** 🚀
