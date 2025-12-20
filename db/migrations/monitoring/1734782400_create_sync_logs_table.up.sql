-- Create sync_logs table to track background sync events from clients

CREATE TABLE IF NOT EXISTS sync_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id TEXT NOT NULL,
    sync_type TEXT NOT NULL,
    status TEXT NOT NULL,
    error_msg TEXT,
    ride_count INTEGER DEFAULT 0,
    duration INTEGER DEFAULT 0,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    city_code TEXT
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_sync_logs_client_id ON sync_logs (client_id);
CREATE INDEX IF NOT EXISTS idx_sync_logs_timestamp ON sync_logs (timestamp);
CREATE INDEX IF NOT EXISTS idx_sync_logs_status ON sync_logs (status);
CREATE INDEX IF NOT EXISTS idx_sync_logs_city_code ON sync_logs (city_code);
