-- Create admin_api_keys table for API key authentication
CREATE TABLE IF NOT EXISTS admin_api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    api_key TEXT UNIQUE NOT NULL,
    admin_name TEXT,
    created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    revoked_at TEXT,
    last_used_at TEXT
);

-- Create indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_admin_api_keys_api_key ON admin_api_keys(api_key);
CREATE INDEX IF NOT EXISTS idx_admin_api_keys_revoked ON admin_api_keys(revoked_at);