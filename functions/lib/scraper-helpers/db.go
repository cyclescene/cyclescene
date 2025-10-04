package scraperhelpers

import (
	"database/sql"
)

func createTables(db *sql.DB) error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS rides (
        composite_event_id TEXT PRIMARY KEY,
        id TEXT NOT NULL,
        address TEXT,
        audience TEXT NOT NULL,
        cancelled INTEGER NOT NULL,
        date TEXT NOT NULL,
        details TEXT,
        endtime TEXT,
        email TEXT,
        eventduration INTEGER,
        image TEXT,
        lat REAL,
        lon REAL,
        locdetails TEXT,
        locend TEXT,
        loopride INTEGER NOT NULL,
        newsflash TEXT,
        organizer TEXT,
        safetyplan INTEGER NOT NULL,
        shareable TEXT,
        starttime TEXT NOT NULL,
        timedetails TEXT,
        title TEXT NOT NULL,
        venue TEXT,
        webname TEXT,
        weburl TEXT,
        source_data TEXT NOT NULL
    );`

	createGeocodeCacheTableSQL := `
    CREATE TABLE IF NOT EXISTS geocode_cache (
    address_key TEXT PRIMARY KEY,
    lat REAL NOT NULL,
    lon REAL NOT NULL,
    last_updated TEXT NOT NULL
    );`

	if _, err := db.Exec(createTableSQL); err != nil {
		return err
	}

	if _, err := db.Exec(createGeocodeCacheTableSQL); err != nil {
		return err
	}

	return nil
}
