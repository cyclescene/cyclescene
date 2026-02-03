-- Rollback migration: Remove source tracking columns from events table
PRAGMA foreign_keys = OFF;

BEGIN TRANSACTION;

-- 1. Drop the deduplication index first
DROP INDEX IF EXISTS idx_events_source_dedup;

-- 2. Create a temporary table with the schema before this migration
CREATE TABLE events_temp (
    id INTEGER PRIMARY KEY,

    /* Core Content & Recurrence Type */
    title TEXT NOT NULL,
    tinytitle TEXT,
    description TEXT NOT NULL,
    image_url TEXT,
    audience TEXT,
    ride_length TEXT,
    area TEXT,
    date_type TEXT,

    /* Location Fields (Constant per Series) */
    venue_name TEXT,
    address TEXT,
    location_details TEXT,
    ending_location TEXT,
    is_loop_ride INTEGER NOT NULL DEFAULT 0,
    city TEXT NOT NULL,

    /* Contact Fields & Privacy Flags */
    organizer_name TEXT,
    organizer_email TEXT,
    organizer_phone TEXT,
    web_url TEXT,
    web_name TEXT,
    newsflash TEXT,
    hide_email INTEGER NOT NULL DEFAULT 0,
    hide_phone INTEGER NOT NULL DEFAULT 0,
    hide_contact_name INTEGER NOT NULL DEFAULT 0,

    /* Group Identifier */
    group_code TEXT,

    /* Magic Link Editing System (Permanent Key) */
    edit_token TEXT UNIQUE,

    /* System/Moderation Fields */
    is_published INTEGER NOT NULL DEFAULT 0,
    is_featured INTEGER NOT NULL DEFAULT 0,
    moderation_notes TEXT,
    moderated_at TEXT,

    /* Audit Fields (Turso/SQLite) */
    created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    updated_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),

    /* Location coordinates (from alter_001) */
    latitude REAL,
    longitude REAL,

    /* Image optimization (from alter_002) */
    image_uuid TEXT,

    FOREIGN KEY (group_code)
        REFERENCES ride_groups(code)
        ON DELETE SET NULL
);

-- 3. Copy data from the old table to the new table (excluding source and source_id)
INSERT INTO events_temp SELECT
    id, title, tinytitle, description, image_url, audience, ride_length, area, date_type,
    venue_name, address, location_details, ending_location, is_loop_ride, city,
    organizer_name, organizer_email, organizer_phone, web_url, web_name, newsflash,
    hide_email, hide_phone, hide_contact_name, group_code, edit_token, is_published,
    is_featured, moderation_notes, moderated_at, created_at, updated_at,
    latitude, longitude, image_uuid
FROM events;

-- 4. Drop the original table
DROP TABLE events;

-- 5. Rename the temporary table back to the original name
ALTER TABLE events_temp RENAME TO events;

-- 6. Re-create the existing indices
CREATE INDEX IF NOT EXISTS idx_published ON events (is_published);
CREATE INDEX IF NOT EXISTS idx_group_code ON events (group_code);
CREATE INDEX IF NOT EXISTS idx_events_image_uuid ON events (image_uuid);

COMMIT;

PRAGMA foreign_keys = ON;
