-- Write your up sql migration here
CREATE TABLE IF NOT EXISTS groups (
    code TEXT PRIMARY KEY, -- The 4-character identifier
    name TEXT NOT NULL,
    description TEXT,
    city TEXT,
    icon_url TEXT,
    web_url TEXT,
    created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);
