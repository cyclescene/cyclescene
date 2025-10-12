-- Write your up sql migration here
CREATE TABLE IF NOT EXISTS geocode_cache (
  location_key TEXT PRIMARY KEY,
  lat REAL NOT NULL,
  lng REAL NOT NULL,
  city TEXT NOT NULL,
  last_updated TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_geocode_key ON geocode_cache (location_key);
CREATE INDEX IF NOT EXISTS idx_geocode_city ON geocode_cache (city);
