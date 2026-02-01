-- Create strava_api_logs table to track Strava API calls for monitoring and rate limit tracking

CREATE TABLE IF NOT EXISTS strava_api_logs (
    -- Primary key
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    -- Request details
    endpoint TEXT NOT NULL,              -- '/athlete/clubs', '/clubs/123/group_events'
    method TEXT NOT NULL,                -- 'GET', 'POST'
    status_code INTEGER NOT NULL,        -- 200, 401, 429, etc.
    response_time_ms INTEGER,            -- Latency tracking (how long Strava took)

    -- General rate limit tracking (from X-Ratelimit-* headers)
    rate_limit_15min_usage INTEGER,      -- Current usage in 15min window (e.g., 7)
    rate_limit_15min_limit INTEGER,      -- Max allowed in 15min (e.g., 200)
    rate_limit_daily_usage INTEGER,      -- Current usage in 24hr window (e.g., 7)
    rate_limit_daily_limit INTEGER,      -- Max allowed in 24hr (e.g., 2000)

    -- Read-only rate limit tracking (from X-Readratelimit-* headers)
    -- IMPORTANT: This is the ACTUAL limiting factor for our GET-heavy operations!
    read_limit_15min_usage INTEGER,      -- Current read usage in 15min window (e.g., 7)
    read_limit_15min_limit INTEGER,      -- Max read allowed in 15min (100)
    read_limit_daily_usage INTEGER,      -- Current read usage in 24hr window (e.g., 7)
    read_limit_daily_limit INTEGER,      -- Max read allowed in 24hr (1000)

    -- Response data (privacy-safe)
    message TEXT,                        -- 'ok' for 200, error details for failures
    clubs_count INTEGER,                 -- Number of clubs returned (for /athlete/clubs)
    events_count INTEGER,                -- Number of events returned (for /group_events)

    -- User context (privacy-safe identifiers only)
    athlete_id INTEGER                   -- Strava athlete ID (public identifier, not sensitive)
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_strava_status ON strava_api_logs(status_code);
CREATE INDEX IF NOT EXISTS idx_strava_endpoint ON strava_api_logs(endpoint);
CREATE INDEX IF NOT EXISTS idx_strava_created_at ON strava_api_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_strava_athlete ON strava_api_logs(athlete_id);
CREATE INDEX IF NOT EXISTS idx_strava_rate_limit ON strava_api_logs(rate_limit_15min_usage, rate_limit_15min_limit);
CREATE INDEX IF NOT EXISTS idx_strava_read_limit ON strava_api_logs(read_limit_15min_usage, read_limit_15min_limit);
