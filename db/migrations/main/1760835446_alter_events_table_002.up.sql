-- Add image_uuid column to events table (rides) for tracking uploaded images
ALTER TABLE events
ADD COLUMN image_uuid TEXT;

-- Add image_uuid and image_url columns to ride_groups table (groups) for tracking uploaded images
ALTER TABLE ride_groups
ADD COLUMN image_uuid TEXT;

ALTER TABLE ride_groups
ADD COLUMN image_url TEXT;

-- Create indexes for image_uuid lookups during optimization
CREATE INDEX IF NOT EXISTS idx_events_image_uuid ON events (image_uuid);
CREATE INDEX IF NOT EXISTS idx_ride_groups_image_uuid ON ride_groups (image_uuid);