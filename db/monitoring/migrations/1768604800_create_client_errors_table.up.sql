-- Create client_errors table to track errors from PWA clients

CREATE TABLE IF NOT EXISTS client_errors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id TEXT NOT NULL,
    error_type TEXT NOT NULL,
    error_msg TEXT NOT NULL,
    stack_trace TEXT,
    component TEXT,
    action TEXT,
    url TEXT,
    user_agent TEXT,
    os TEXT DEFAULT 'Unknown',
    city_code TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_client_errors_client_id ON client_errors (client_id);
CREATE INDEX IF NOT EXISTS idx_client_errors_timestamp ON client_errors (timestamp);
CREATE INDEX IF NOT EXISTS idx_client_errors_type ON client_errors (error_type);
CREATE INDEX IF NOT EXISTS idx_client_errors_component ON client_errors (component);
