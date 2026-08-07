CREATE TABLE google_calendar_sync (
    calendar_id TEXT NOT NULL,
    google_event_id TEXT NOT NULL,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    google_updated_at TEXT,
    last_seen_at TEXT NOT NULL,
    PRIMARY KEY (calendar_id, google_event_id),
    UNIQUE (event_id)
);

CREATE INDEX idx_google_calendar_sync_last_seen_at
ON google_calendar_sync (last_seen_at);
