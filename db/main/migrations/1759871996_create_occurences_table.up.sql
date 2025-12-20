-- Write your up sql migration here
CREATE TABLE IF NOT EXISTS event_occurrences (
    id INTEGER PRIMARY KEY,
    
    /* Foreign Key to the Master Event */
    event_id INTEGER NOT NULL,
    
    /* Time & Date Fields (Three fields for UX/DB) */
    start_date TEXT NOT NULL,       -- e.g., '2025-10-07'
    start_time TEXT NOT NULL,       -- e.g., '18:30:00'
    start_datetime TEXT NOT NULL,   -- e.g., '2025-10-07 18:30:00' (The Sort Field)

    event_duration_minutes INTEGER,
    event_time_details TEXT, -- Notes specific to this date (e.g., 'Delayed due to rain')
    is_cancelled INTEGER NOT NULL DEFAULT 0, -- 1 if this single date is cancelled
    
    FOREIGN KEY (event_id) 
        REFERENCES events(id) 
        ON DELETE CASCADE
);

-- Essential Index for fast chronological queries
CREATE INDEX IF NOT EXISTS idx_occurrence_datetime ON event_occurrences (start_datetime);
CREATE INDEX IF NOT EXISTS idx_occurrence_event_id ON event_occurrences (event_id);
