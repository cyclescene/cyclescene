-- Rollback: Remove marker_color column from ride_groups table

DROP INDEX IF EXISTS idx_groups_marker_color;

ALTER TABLE ride_groups DROP COLUMN marker_color;
