-- Rollback migration: Remove Strava integration tables

DROP TABLE IF EXISTS strava_event_metadata;
DROP TABLE IF EXISTS strava_connections;
