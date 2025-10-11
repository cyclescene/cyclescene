-- Write your down sql migration here
DROP TABLE event_occurrences;
DROP INDEX IF EXISTS idx_occurrence_datetime;
DROP INDEX IF EXISTS idx_occurrence_event_id;
