CREATE TABLE IF NOT EXISTS strava_connections (
    athlete_id INTEGER PRIMARY KEY,
    refresh_token_encrypted BLOB NOT NULL,
    encryption_nonce BLOB NOT NULL,
    city_code TEXT NOT NULL,
    last_synced_at TEXT,
    created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

CREATE INDEX IF NOT EXISTS idx_strava_connections_city ON strava_connections (city_code);
CREATE INDEX IF NOT EXISTS idx_strava_connections_last_synced ON strava_connections (last_synced_at);

CREATE TABLE IF NOT EXISTS strava_event_metadata (
    event_id INTEGER PRIMARY KEY,
    strava_event_id INTEGER NOT NULL,
    strava_club_id INTEGER NOT NULL,
    imported_by_athlete_id INTEGER NOT NULL,
    last_refreshed_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    imported_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    refresh_count INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE,
    FOREIGN KEY (imported_by_athlete_id) REFERENCES strava_connections (athlete_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_strava_event_metadata_strava_ids ON strava_event_metadata (strava_club_id, strava_event_id);
CREATE INDEX IF NOT EXISTS idx_strava_event_metadata_athlete ON strava_event_metadata (imported_by_athlete_id);
CREATE INDEX IF NOT EXISTS idx_strava_event_metadata_refreshed ON strava_event_metadata (last_refreshed_at);
CREATE INDEX IF NOT EXISTS idx_strava_event_metadata_imported_at ON strava_event_metadata (imported_at);
