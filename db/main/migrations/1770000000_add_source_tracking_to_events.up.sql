-- Add source tracking columns to events table for external import deduplication
-- source: identifies where the event came from
-- source_id: external identifier for deduplication

ALTER TABLE events ADD COLUMN source TEXT;

ALTER TABLE events ADD COLUMN source_id TEXT;

-- Create partial unique index for deduplication
-- Only applies when both fields are set (manual submissions can have NULL values)
CREATE UNIQUE INDEX IF NOT EXISTS idx_events_source_dedup
ON events (source, source_id)
WHERE source IS NOT NULL AND source_id IS NOT NULL;
