-- Drop admin_api_keys table and indexes
DROP INDEX IF EXISTS idx_admin_api_keys_revoked;
DROP INDEX IF EXISTS idx_admin_api_keys_api_key;
DROP TABLE IF EXISTS admin_api_keys;