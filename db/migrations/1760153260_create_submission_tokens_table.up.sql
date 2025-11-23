-- Write your up sql migration here
CREATE TABLE IF NOT EXISTS submission_tokens (
  token TEXT PRIMARY KEY,
  city TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
  expires_at TEXT NOT NULL,
  used INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_token_used ON submission_tokens (used);
CREATE INDEX IF NOT EXISTS idx_token_expiry ON submission_tokens (expires_at);

