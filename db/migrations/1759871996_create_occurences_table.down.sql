-- Write your down sql migration here
DROP TABLE event_occurrences;
DROP INDEX idx_occurrence_datetime ON event_occurrences;
DROP INDEX idx_occurrence_event_id ON event_occurrences;
