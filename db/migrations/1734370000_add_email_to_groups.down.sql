-- Rollback: Remove email column from ride_groups table
ALTER TABLE ride_groups DROP COLUMN IF EXISTS email;

-- Drop index
DROP INDEX IF EXISTS idx_groups_email;
