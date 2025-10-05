package scraperhelpers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"golang.org/x/exp/slog"
	"strings"
	"time"
)

func CreateTables(db *sql.DB) error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS rides (
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
	
				ridesource TEXT NOT NULL,
        source_data TEXT NOT NULL
    );`

	createGeocodeCacheTableSQL := `
    CREATE TABLE IF NOT EXISTS geocode_cache (
  	location_key TEXT PRIMARY KEY,
    lat REAL NOT NULL,
    lng REAL NOT NULL,
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

func GetGeocodeCache(db *sql.DB) (map[string]GeoCodeCached, error) {
	rows, err := db.Query("SELECT location_key, lat, lng FROM geocode_cache")
	if err != nil {
		slog.Error("Something went wrong when calling for geocode cache", "error", err.Error())
		return nil, fmt.Errorf("failed to query geocode cache: %w", err)

	}
	defer rows.Close()

	cachedAddresses := make(map[string]GeoCodeCached)

	for rows.Next() {
		var key string
		var lat, lng float64
		if err := rows.Scan(&key, &lat, &lng); err != nil {
			slog.Error("Unalble to scan row", "error", err.Error())
			continue
		}
		cachedAddresses[strings.ToLower(key)] = GeoCodeCached{Latitude: lat, Longitude: lng}
	}
	if err = rows.Err(); err != nil {
		slog.Error("Something went wrong while iterating though stored rows", "error", err.Error())
		return nil, fmt.Errorf("row iteration error: %w", err)

	}
	slog.Info("Cached geocoded get successfull", "cached_amount", len(cachedAddresses))

	return cachedAddresses, nil
}

func BulkUpsertGeocodeData(db *sql.DB, locations []Location) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin bulk transaction: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`
        INSERT INTO geocode_cache (location_key, lat, lng, last_updated)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(location_key) DO UPDATE SET
            lat=excluded.lat,
            lng=excluded.lng,
            last_updated=excluded.last_updated;
        `)
	if err != nil {
		return fmt.Errorf("failed to prepare geocode cache upsert statement: %v", err)
	}
	defer stmt.Close()

	now := time.Now().Format(time.RFC3339)

	for i := range locations {
		loc := locations[i]

		_, err := stmt.Exec(
			strings.ToLower(loc.Query),
			loc.Latitude,
			loc.Longitude,
			now,
		)
		if err != nil {
			slog.Error("Failed to upsert single location in batch", "query", loc.Query, "error", err.Error())
			return fmt.Errorf("failed to execute batch upsert for key %s: %w", loc.Query, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk transaction: %w", err)
	}

	return nil
}

func BulkUpsertRideData(db *sql.DB, rideData []Shift2BikeEvent) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin bulk transaction: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`
        INSERT INTO rides (
            composite_event_id,
            id,
						title,
						lat,
						lng,
						address,
						audience,
						cancelled,
						date,
						starttime,
						safetyplan,
						details,
						venue,
						organizer,
						loopride,
						shareable,
						endtime,
						email,
						eventduration,
						image,
						locdetails,
						locend,
						newsflash,
						timedetails,
						webname,
						weburl,
						ridesource,
						source_data
        )
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
        ON CONFLICT(composite_event_id) DO UPDATE SET
            id=excluded.id,
            address=excluded.address,
            audience=excluded.audience,
            cancelled=excluded.cancelled,
            date=excluded.date,
            details=excluded.details,
            endtime=excluded.endtime,
            email=excluded.email,
            eventduration=excluded.eventduration,
            image=excluded.image,
            lat=excluded.lat,
            lng=excluded.lng,
            locdetails=excluded.locdetails,
            locend=excluded.locend,
            loopride=excluded.loopride,
            newsflash=excluded.newsflash,
            organizer=excluded.organizer,
            safetyplan=excluded.safetyplan,
            shareable=excluded.shareable,
            starttime=excluded.starttime,
            timedetails=excluded.timedetails,
            title=excluded.title,
            venue=excluded.venue,
            webname=excluded.webname,
            weburl=excluded.weburl,
						ridesource=excluded.ridesource,
            source_data=excluded.source_data;
        `)
	if err != nil {
		return fmt.Errorf("failed to prepare ride data upsert statement: %v", err)
	}
	defer stmt.Close()

	for i := range rideData {
		ride := rideData[i]

		compositeKey := fmt.Sprintf("%s_%s", ride.ID, ride.Date)

		sourceData, marshalErr := json.Marshal(ride)
		if marshalErr != nil {
			sourceData = []byte("{}")
		}

		isCancelled := 0
		if ride.Cancelled {
			isCancelled = 1
		}

		isLoopRide := 0
		if ride.Loopride {
			isLoopRide = 1
		}

		isSafetyPlan := 0
		if ride.Safetyplan {
			isSafetyPlan = 1
		}
		_, err = stmt.Exec(
			compositeKey,
			ride.ID,
			ride.Title,
			ride.Location.Latitude,
			ride.Location.Longitude,
			ride.Address,
			ride.Audience,
			isCancelled,
			ride.Date,
			ride.Time,
			isSafetyPlan,
			ride.Details,
			ride.Venue,
			ride.Organizer,
			isLoopRide,
			ride.Shareable,
			ride.Endtime,
			ride.Email,
			ride.Eventduration,
			ride.Image,
			ride.Locdetails,
			ride.Locend,
			ride.Newsflash,
			ride.Timedetails,
			ride.Webname,
			ride.Weburl,
			ride.SourcedFrom,
			string(sourceData),
		)
		if err != nil {
			slog.Error("Failed to upsert single location in batch", "key", compositeKey, "error", err.Error())
			return fmt.Errorf("failed to execute batch upsert for key %s: %w", compositeKey, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk transaction: %w", err)
	}

	return nil
}
