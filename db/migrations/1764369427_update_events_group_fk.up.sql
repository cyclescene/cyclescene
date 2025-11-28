-- Update events table to use group_id (UUID FK) instead of group_code

-- Create new events table with group_id FK
CREATE TABLE events_new (
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
    group_code TEXT,                           -- Keep for backwards compatibility
    group_id TEXT,                             -- NEW: UUID FK to ride_groups(id)
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
    FOREIGN KEY (group_id) REFERENCES ride_groups(id) ON DELETE SET NULL
);

-- Copy data, joining on group_code to get the new id
INSERT INTO events_new (
    id, title, tinytitle, description, image_url, audience, ride_length, area, date_type,
    venue_name, address, location_details, ending_location, is_loop_ride, city,
    organizer_name, organizer_email, organizer_phone, web_url, web_name, newsflash,
    hide_email, hide_phone, hide_contact_name, group_code, group_id, edit_token,
    is_published, is_featured, moderation_notes, moderated_at, created_at, updated_at,
    latitude, longitude, image_uuid
)
SELECT
    e.id, e.title, e.tinytitle, e.description, e.image_url, e.audience, e.ride_length, e.area, e.date_type,
    e.venue_name, e.address, e.location_details, e.ending_location, e.is_loop_ride, e.city,
    e.organizer_name, e.organizer_email, e.organizer_phone, e.web_url, e.web_name, e.newsflash,
    e.hide_email, e.hide_phone, e.hide_contact_name, e.group_code,
    rg.id,  -- Get the new UUID id from ride_groups
    e.edit_token, e.is_published, e.is_featured, e.moderation_notes, e.moderated_at,
    e.created_at, e.updated_at, e.latitude, e.longitude, e.image_uuid
FROM events e
LEFT JOIN ride_groups rg ON e.group_code = rg.code;

-- Drop old table
DROP TABLE events;

-- Rename new table
ALTER TABLE events_new RENAME TO events;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_published ON events (is_published);
CREATE INDEX IF NOT EXISTS idx_group_code ON events (group_code);
CREATE INDEX IF NOT EXISTS idx_group_id ON events (group_id);
