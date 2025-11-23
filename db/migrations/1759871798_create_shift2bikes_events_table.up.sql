-- Write your up sql migration here
CREATE TABLE IF NOT EXISTS shift2bikes_events (
  composite_event_id TEXT PRIMARY KEY,

  id TEXT NOT NULL,
  title TEXT NOT NULL,
  lat REAL NOT NULL,
  lng REAL NOT NULL,
  address TEXT NOT NULL,
  audience TEXT NOT NULL,
  cancelled INTEGER NOT NULL,
  date TEXT NOT NULL,
  starttime TEXT NOT NULL,
  safetyplan INTEGER NOT NULL,
  details TEXT NOT NULL,
  venue TEXT NOT NULL,
  organizer TEXT NOT NULL,
  loopride INTEGER NOT NULL,
  shareable TEXT NOT NULL,


  endtime TEXT,
  email TEXT,
  eventduration INTEGER,
  image TEXT,
  locdetails TEXT,
  locend TEXT,
  newsflash TEXT,
  timedetails TEXT,
  webname TEXT,
  weburl TEXT,

  citycode TEXT NOT NULL,
  ridesource TEXT NOT NULL,
  source_data TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_citycode ON shift2bikes_events (citycode);
CREATE INDEX IF NOT EXISTS idx_date ON shift2bikes_events (date);
