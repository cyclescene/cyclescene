-- Rollback: restore ride_groups table to code as PK

-- Create old ride_groups table with code as PK
CREATE TABLE ride_groups_old (
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

-- Copy data back (dropping new columns)
INSERT INTO ride_groups_old (code, name, description, city, icon_url, is_active, web_url, edit_token, created_at)
SELECT code, name, description, city, icon_url, is_active, web_url, edit_token, created_at
FROM ride_groups;

-- Drop new table
DROP TABLE ride_groups;

-- Rename old table back to ride_groups
ALTER TABLE ride_groups_old RENAME TO ride_groups;

-- Recreate original indexes
CREATE INDEX IF NOT EXISTS idx_groups_edit_token ON ride_groups (edit_token);
