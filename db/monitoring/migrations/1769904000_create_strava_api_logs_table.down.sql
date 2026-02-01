-- Drop strava_api_logs table and its indexes

DROP INDEX IF EXISTS idx_strava_read_limit;
DROP INDEX IF EXISTS idx_strava_rate_limit;
DROP INDEX IF EXISTS idx_strava_athlete;
DROP INDEX IF EXISTS idx_strava_created_at;
DROP INDEX IF EXISTS idx_strava_endpoint;
DROP INDEX IF EXISTS idx_strava_status;
DROP TABLE IF EXISTS strava_api_logs;
