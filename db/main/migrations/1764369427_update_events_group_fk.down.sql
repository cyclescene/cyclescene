-- Rollback: restore events table to use only group_code FK

-- Create old events table structure
CREATE TABLE events_old (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    tinytitle TEXT,
    description TEXT NOT NULL,
    image_url TEXT,
    audience TEXT,
    ride_length TEXT,
    area TEXT,
    date_type TEXT,
    venue_name TEXT,
    address TEXT,
    location_details TEXT,
    ending_location TEXT,
    is_loop_ride INTEGER NOT NULL DEFAULT 0,
    city TEXT NOT NULL,
    organizer_name TEXT,
    organizer_email TEXT,
    organizer_phone TEXT,
    web_url TEXT,
    web_name TEXT,
    newsflash TEXT,
    hide_email INTEGER NOT NULL DEFAULT 0,
    hide_phone INTEGER NOT NULL DEFAULT 0,
    hide_contact_name INTEGER NOT NULL DEFAULT 0,
    group_code TEXT,
    edit_token TEXT UNIQUE,
    is_published INTEGER NOT NULL DEFAULT 0,
    is_featured INTEGER NOT NULL DEFAULT 0,
    moderation_notes TEXT,
    moderated_at TEXT,
    created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    updated_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    latitude REAL,
    longitude REAL,
    image_uuid TEXT,
    FOREIGN KEY (group_code) REFERENCES ride_groups(code) ON DELETE SET NULL
);

-- Copy data back (dropping group_id)
INSERT INTO events_old (
    id, title, tinytitle, description, image_url, audience, ride_length, area, date_type,
    venue_name, address, location_details, ending_location, is_loop_ride, city,
    organizer_name, organizer_email, organizer_phone, web_url, web_name, newsflash,
    hide_email, hide_phone, hide_contact_name, group_code, edit_token,
    is_published, is_featured, moderation_notes, moderated_at, created_at, updated_at,
    latitude, longitude, image_uuid
)
SELECT
    id, title, tinytitle, description, image_url, audience, ride_length, area, date_type,
    venue_name, address, location_details, ending_location, is_loop_ride, city,
    organizer_name, organizer_email, organizer_phone, web_url, web_name, newsflash,
    hide_email, hide_phone, hide_contact_name, group_code, edit_token,
    is_published, is_featured, moderation_notes, moderated_at, created_at, updated_at,
    latitude, longitude, image_uuid
FROM events;

-- Drop new table
DROP TABLE events;

-- Rename old table
ALTER TABLE events_old RENAME TO events;

-- Recreate original indexes
CREATE INDEX IF NOT EXISTS idx_published ON events (is_published);
CREATE INDEX IF NOT EXISTS idx_group_code ON events (group_code);
