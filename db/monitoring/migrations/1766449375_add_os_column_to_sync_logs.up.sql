-- Add os column to sync_logs table to track which platform/OS is syncing
ALTER TABLE sync_logs ADD COLUMN os TEXT DEFAULT 'Unknown';
