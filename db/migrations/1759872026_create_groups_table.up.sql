-- Write your up sql migration here
CREATE TABLE IF NOT EXISTS ride_groups (
    code TEXT PRIMARY KEY, -- The 4-character identifier
    name TEXT NOT NULL,
    description TEXT,
    city TEXT,
    icon_url TEXT,
    is_active INTEGER NOT NULL DEFAULT 1, -- soft delete for groups
    web_url TEXT,
    edit_token TEXT UNIQUE,
    created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

CREATE INDEX IF NOT EXISTS idx_groups_edit_token on ride_groups (edit_token);
