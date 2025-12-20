-- Write your down sql migration here
-- 1. Start a transaction for atomicity
PRAGMA foreign_keys = OFF; -- Temporarily disable foreign key checks

BEGIN TRANSACTION;

-- 2. Create a temporary table with the *original* schema (excluding latitude and longitude)
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


    FOREIGN KEY (group_code)
        REFERENCES ride_groups(code)
        ON DELETE SET NULL
);

-- 3. Copy data from the old table to the new table
INSERT INTO events_temp SELECT 
    id, title, tinytitle, description, image_url, audience, ride_length, area, date_type,
    venue_name, address, location_details, ending_location, is_loop_ride, city,
    organizer_name, organizer_email, organizer_phone, web_url, web_name, newsflash,
    hide_email, hide_phone, hide_contact_name, group_code, edit_token, is_published,
    is_featured, moderation_notes, moderated_at, created_at, updated_at
FROM events;

-- 4. Drop the original table
DROP TABLE events;

-- 5. Rename the temporary table back to the original name
ALTER TABLE events_temp RENAME TO events;

-- 6. Re-create the indices
CREATE INDEX IF NOT EXISTS idx_published ON events (is_published);
CREATE INDEX IF NOT EXISTS idx_group_code ON events (group_code);
-- (You would also re-create the optional idx_coordinates index if it were included in the 'up' migration)

-- 7. Commit the transaction
COMMIT;

PRAGMA foreign_keys = ON; -- Re-enable foreign key checks
