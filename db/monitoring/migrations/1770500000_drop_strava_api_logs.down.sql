CREATE TABLE IF NOT EXISTS strava_api_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER,
    rate_limit_15min_usage INTEGER,
    rate_limit_15min_limit INTEGER,
    rate_limit_daily_usage INTEGER,
    rate_limit_daily_limit INTEGER,
    read_limit_15min_usage INTEGER,
    read_limit_15min_limit INTEGER,
    read_limit_daily_usage INTEGER,
    read_limit_daily_limit INTEGER,
    message TEXT,
    clubs_count INTEGER,
    events_count INTEGER,
    athlete_id INTEGER
);

CREATE INDEX IF NOT EXISTS idx_strava_status ON strava_api_logs(status_code);
CREATE INDEX IF NOT EXISTS idx_strava_endpoint ON strava_api_logs(endpoint);
CREATE INDEX IF NOT EXISTS idx_strava_created_at ON strava_api_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_strava_athlete ON strava_api_logs(athlete_id);
CREATE INDEX IF NOT EXISTS idx_strava_rate_limit ON strava_api_logs(rate_limit_15min_usage, rate_limit_15min_limit);
CREATE INDEX IF NOT EXISTS idx_strava_read_limit ON strava_api_logs(read_limit_15min_usage, read_limit_15min_limit);
