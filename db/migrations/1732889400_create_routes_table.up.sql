CREATE TABLE routes (
  id TEXT PRIMARY KEY,
  source TEXT NOT NULL,
  source_id TEXT NOT NULL,
  source_url TEXT NOT NULL,
  geojson TEXT NOT NULL,
  distance_km REAL,
  distance_mi REAL,
  created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
  UNIQUE(source, source_id)
);

CREATE INDEX idx_routes_source_id ON routes (source, source_id);

ALTER TABLE events ADD COLUMN route_id TEXT REFERENCES routes(id) ON DELETE SET NULL;

CREATE INDEX idx_events_route_id ON events (route_id);

ALTER TABLE shift2bikes_events ADD COLUMN route_id TEXT REFERENCES routes(id) ON DELETE SET NULL;

CREATE INDEX idx_shift2bikes_events_route_id ON shift2bikes_events (route_id);
