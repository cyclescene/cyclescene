-- Write your down sql migration here
DROP TABLE events;
DROP INDEX idx_published ON events;
DROP INDEX idx_group_code ON events;
