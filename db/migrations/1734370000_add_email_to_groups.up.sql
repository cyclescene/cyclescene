-- Add email column to ride_groups table
ALTER TABLE ride_groups ADD COLUMN email TEXT;

-- Create index on email for faster lookups if needed
CREATE INDEX IF NOT EXISTS idx_groups_email on ride_groups (email);
