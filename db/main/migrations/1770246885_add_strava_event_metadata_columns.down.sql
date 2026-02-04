-- Rollback: Remove strava_event_metadata columns

DROP INDEX IF EXISTS idx_strava_event_metadata_imported_at;

-- Note: SQLite doesn't support DROP COLUMN in older versions
-- If you need to rollback, you'll need to recreate the table
-- For now, this is a placeholder
-- ALTER TABLE strava_event_metadata DROP COLUMN refresh_count;
-- ALTER TABLE strava_event_metadata DROP COLUMN imported_at;
