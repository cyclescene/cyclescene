-- Restructure ride_groups table: change PK from code to id (UUID)
-- SQLite doesn't support direct PK changes, so we use temp table approach

-- Create new ride_groups table with id UUID as PK
CREATE TABLE ride_groups_new (
    id TEXT PRIMARY KEY,                    -- UUID primary key
    code TEXT UNIQUE NOT NULL,              -- Keep code as unique constraint
    name TEXT NOT NULL,
    description TEXT,
    city TEXT,
    icon_url TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    web_url TEXT,
    edit_token TEXT UNIQUE,
    created_at TEXT NOT NULL,
    public_id TEXT UNIQUE,                  -- NEW: URL-friendly slug for sprite key
    marker TEXT                             -- NEW: Sprite metadata key
);

-- Copy data from old table, generating UUIDs for id
INSERT INTO ride_groups_new (id, code, name, description, city, icon_url, is_active, web_url, edit_token, created_at, public_id, marker)
SELECT
    lower(hex(randomblob(16))), -- Generate UUID-like random ID
    code,
    name,
    description,
    city,
    icon_url,
    is_active,
    web_url,
    edit_token,
    created_at,
    NULL,  -- public_id will be set when markers are uploaded
    NULL   -- marker will be set when markers are uploaded
FROM ride_groups;

-- Drop old table
DROP TABLE ride_groups;

-- Rename new table to ride_groups
ALTER TABLE ride_groups_new RENAME TO ride_groups;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_groups_code ON ride_groups (code);
CREATE INDEX IF NOT EXISTS idx_groups_edit_token ON ride_groups (edit_token);
CREATE INDEX IF NOT EXISTS idx_groups_public_id ON ride_groups (public_id);
