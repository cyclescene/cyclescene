-- Write your up sql migration here
ALTER TABLE events
ADD COLUMN latitude REAL;

ALTER TABLE events
ADD COLUMN longitude REAL;
