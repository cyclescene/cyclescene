-- Write your down sql migration here
DROP TABLE geocode_cache;
DROP INDEX idx_geocode_key;
CREATE INDEX IF EXISTS idx_geocode_city;
