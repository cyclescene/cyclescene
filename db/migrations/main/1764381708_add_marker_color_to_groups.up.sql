-- Add marker_color column to ride_groups table
-- Stores hex color for group marker teardrop (#RRGGBB format)

ALTER TABLE ride_groups ADD COLUMN marker_color TEXT DEFAULT '#3B82F6';

-- Create index for faster queries if needed
CREATE INDEX IF NOT EXISTS idx_groups_marker_color ON ride_groups (marker_color);
