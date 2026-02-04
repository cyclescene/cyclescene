-- Migration: Add minimal Strava integration tables
-- Privacy-first: Only store refresh tokens and athlete IDs, no personal data

-- Table 1: Strava Connections (just enough to re-authenticate)
CREATE TABLE IF NOT EXISTS strava_connections (
    athlete_id INTEGER PRIMARY KEY,              -- Strava athlete ID (unique identifier)
    refresh_token_encrypted BLOB NOT NULL,       -- Encrypted refresh token (for getting access tokens)
    encryption_nonce BLOB NOT NULL,              -- Nonce for AES-GCM decryption
    city_code TEXT NOT NULL,                     -- City context (pdx, slc)
    last_synced_at TEXT,                         -- Last successful background sync (for monitoring)
    created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

CREATE INDEX IF NOT EXISTS idx_strava_connections_city ON strava_connections (city_code);
CREATE INDEX IF NOT EXISTS idx_strava_connections_last_synced ON strava_connections (last_synced_at);


-- Table 2: Strava Event Metadata (links events to Strava for sync and display)
CREATE TABLE IF NOT EXISTS strava_event_metadata (
    event_id INTEGER PRIMARY KEY,                -- References events.id
    strava_event_id INTEGER NOT NULL,            -- Strava's event ID (for deduplication and sync checks)
    strava_club_id INTEGER NOT NULL,             -- Strava's club ID (for "View on Strava" links)
    imported_by_athlete_id INTEGER NOT NULL,     -- Which connection imported this (for sync ownership)
    last_refreshed_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')), -- 7-day compliance tracking

    FOREIGN KEY (event_id)
        REFERENCES events (id)
        ON DELETE CASCADE,
    FOREIGN KEY (imported_by_athlete_id)
        REFERENCES strava_connections (athlete_id)
        ON DELETE CASCADE  -- If connection deleted, delete their imported events
);

CREATE INDEX IF NOT EXISTS idx_strava_event_metadata_strava_ids ON strava_event_metadata (strava_club_id, strava_event_id);
CREATE INDEX IF NOT EXISTS idx_strava_event_metadata_athlete ON strava_event_metadata (imported_by_athlete_id);
CREATE INDEX IF NOT EXISTS idx_strava_event_metadata_refreshed ON strava_event_metadata (last_refreshed_at);
