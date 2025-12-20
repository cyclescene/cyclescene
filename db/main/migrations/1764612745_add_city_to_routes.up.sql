ALTER TABLE routes ADD COLUMN city TEXT DEFAULT 'pdx';

CREATE INDEX idx_routes_city ON routes (city);
