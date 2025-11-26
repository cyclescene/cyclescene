-- Add newsflash column to event_occurrences table
-- This allows each occurrence to have its own newsflash/alert message
-- Different dates of the same recurring event can have different newsflashes

ALTER TABLE event_occurrences
ADD COLUMN newsflash TEXT;
