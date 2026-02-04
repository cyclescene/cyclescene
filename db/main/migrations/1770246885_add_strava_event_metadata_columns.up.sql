-- Migration: Add missing columns to strava_event_metadata for background sync tracking
-- These columns support Phase 1: Core Sync Infrastructure

-- Add imported_at to track when the event was first imported
ALTER TABLE strava_event_metadata
ADD COLUMN imported_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'));

-- Add refresh_count to track how many times the event has been synced
ALTER TABLE strava_event_metadata
ADD COLUMN refresh_count INTEGER NOT NULL DEFAULT 0;

-- Create index for efficient querying of stale events (not refreshed in 7+ days)
CREATE INDEX IF NOT EXISTS idx_strava_event_metadata_imported_at
ON strava_event_metadata (imported_at);
