ALTER TABLE events DROP COLUMN route_id;

ALTER TABLE shift2bikes_events DROP COLUMN route_id;

DROP INDEX idx_events_route_id;
DROP INDEX idx_shift2bikes_events_route_id;
DROP INDEX idx_routes_source_id;

DROP TABLE routes;
