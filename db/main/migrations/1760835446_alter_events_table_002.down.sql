-- Rollback migration: Remove image_uuid and image_url columns from events and ride_groups tables
PRAGMA foreign_keys = OFF; -- Temporarily disable foreign key checks

-- ==== ROLLBACK EVENTS TABLE ====
BEGIN TRANSACTION;

-- 1. Create a temporary table with the original schema (excluding image_uuid)
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
    latitude REAL,
    longitude REAL,

    FOREIGN KEY (group_code)
        REFERENCES ride_groups(code)
        ON DELETE SET NULL
);

-- 2. Copy data from the old table to the new table (excluding image_uuid)
INSERT INTO events_temp SELECT
    id, title, tinytitle, description, image_url, audience, ride_length, area, date_type,
    venue_name, address, location_details, ending_location, is_loop_ride, city,
    organizer_name, organizer_email, organizer_phone, web_url, web_name, newsflash,
    hide_email, hide_phone, hide_contact_name, group_code, edit_token, is_published,
    is_featured, moderation_notes, moderated_at, created_at, updated_at, latitude, longitude
FROM events;

-- 3. Drop the original table
DROP TABLE events;

-- 4. Rename the temporary table back to the original name
ALTER TABLE events_temp RENAME TO events;

-- 5. Re-create the indices
CREATE INDEX IF NOT EXISTS idx_published ON events (is_published);
CREATE INDEX IF NOT EXISTS idx_group_code ON events (group_code);

-- 6. Commit the transaction
COMMIT;

-- ==== ROLLBACK RIDE_GROUPS TABLE ====
BEGIN TRANSACTION;

-- 1. Create a temporary table with the original schema (excluding image_uuid and image_url)
CREATE TABLE ride_groups_temp (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    city TEXT,
    icon_url TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    web_url TEXT,
    edit_token TEXT UNIQUE,
    created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

-- 2. Copy data from the old table to the new table (excluding image_uuid and image_url)
INSERT INTO ride_groups_temp SELECT
    code, name, description, city, icon_url, is_active, web_url, edit_token, created_at
FROM ride_groups;

-- 3. Drop the original table
DROP TABLE ride_groups;

-- 4. Rename the temporary table back to the original name
ALTER TABLE ride_groups_temp RENAME TO ride_groups;

-- 5. Re-create the indices
CREATE INDEX IF NOT EXISTS idx_groups_edit_token on ride_groups (edit_token);

-- 6. Commit the transaction
COMMIT;

PRAGMA foreign_keys = ON; -- Re-enable foreign key checks