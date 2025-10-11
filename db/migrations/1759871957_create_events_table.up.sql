-- Write your up sql migration here
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY,
    
    /* Core Content & Recurrence Type */
    title TEXT NOT NULL,
    tinytitle TEXT,
    description TEXT NOT NULL,
    image_url TEXT,
    audience TEXT, -- G=General, F=Family, A=Adults Only, E=Experienced
    ride_length TEXT,
    area TEXT,
    date_type TEXT, -- S=Single, R=Recurring, O=One-Off, etc.

    /* Location Fields (Constant per Series) */
    venue_name TEXT,
    address TEXT,
    location_details TEXT,
    ending_location TEXT,
    is_loop_ride INTEGER NOT NULL DEFAULT 0, -- 0=FALSE, 1=TRUE
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
    is_published INTEGER NOT NULL DEFAULT 0, -- CRITICAL: 0=Pending Review, 1=Published
    is_featured INTEGER NOT NULL DEFAULT 0,
    moderation_notes TEXT,
    moderated_at TEXT,
    
    /* Audit Fields (Turso/SQLite) */
    created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    updated_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),


    FOREIGN KEY (group_code)
        REFERENCES groups(code)
        ON DELETE SET NULL
);

-- Index for fast moderation and public display queries
CREATE INDEX IF NOT EXISTS idx_published ON events (is_published);
CREATE INDEX IF NOT EXISTS idx_group_code ON events (group_code);
