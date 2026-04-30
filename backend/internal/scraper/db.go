package scraper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const shift2BikesRideSource = "Shift2Bikes"

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
			_ = tx.Rollback()
			panic(r)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`
        INSERT INTO geocode_cache (location_key, lat, lng, city, last_updated)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(location_key) DO UPDATE SET
						city=excluded.city,
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
			loc.City,
			now,
		)
		if err != nil {
			slog.Error("Failed to upsert single location in batch", "loc", loc, "error", err.Error())
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
			_ = tx.Rollback()
			panic(r)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = bulkUpsertRideDataTx(tx, rideData); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk transaction: %w", err)
	}

	return nil
}

func SyncShift2BikesRideData(db *sql.DB, cityCode string, rideData []Shift2BikeEvent) error {
	today, err := todayForCity(cityCode)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin sync transaction: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	storedUpcomingIDs, err := getStoredUpcomingShift2BikesEventIDs(tx, cityCode, today)
	if err != nil {
		return err
	}

	scrapedUpcomingIDs := make(map[string]struct{}, len(rideData))
	for i := range rideData {
		if rideData[i].Date < today {
			continue
		}
		id := rideData[i].CompositeID()
		if id == "" {
			slog.Warn("scraped Shift2Bikes event has no stable ID; skipping diff tracking", "title", rideData[i].Title, "date", rideData[i].Date)
			continue
		}
		scrapedUpcomingIDs[id] = struct{}{}
	}
	if len(scrapedUpcomingIDs) == 0 && len(storedUpcomingIDs) > 0 {
		return fmt.Errorf("current Shift2Bikes scrape returned no upcoming event IDs; refusing to delete %d stored upcoming events", len(storedUpcomingIDs))
	}

	var idsToDelete []string
	for id := range storedUpcomingIDs {
		if _, found := scrapedUpcomingIDs[id]; !found {
			idsToDelete = append(idsToDelete, id)
		}
	}

	if len(idsToDelete) > 0 {
		var deletedCount int64
		deletedCount, err = deleteShift2BikesEvents(tx, cityCode, idsToDelete)
		if err != nil {
			return err
		}
		slog.Info("deleted stored Shift2Bikes events missing from current scrape", "city", cityCode, "count", deletedCount)
	}

	if err = bulkUpsertRideDataTx(tx, rideData); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit sync transaction: %w", err)
	}

	newCount := 0
	for id := range scrapedUpcomingIDs {
		if _, found := storedUpcomingIDs[id]; !found {
			newCount++
		}
	}

	slog.Info("synced Shift2Bikes ride data",
		"city", cityCode,
		"scraped_upcoming_count", len(scrapedUpcomingIDs),
		"stored_upcoming_count", len(storedUpcomingIDs),
		"new_count", newCount,
		"deleted_count", len(idsToDelete),
	)

	return nil
}

func getStoredUpcomingShift2BikesEventIDs(tx *sql.Tx, cityCode string, today string) (map[string]struct{}, error) {
	rows, err := tx.Query(`
		SELECT composite_event_id
		FROM shift2bikes_events
		WHERE citycode = ?
			AND ridesource = ?
			AND date >= ?
	`, cityCode, shift2BikesRideSource, today)
	if err != nil {
		return nil, fmt.Errorf("failed to query stored upcoming Shift2Bikes events: %w", err)
	}
	defer rows.Close()

	ids := make(map[string]struct{})
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan stored Shift2Bikes event ID: %w", err)
		}
		ids[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate stored Shift2Bikes event IDs: %w", err)
	}

	return ids, nil
}

func deleteShift2BikesEvents(tx *sql.Tx, cityCode string, eventIDs []string) (int64, error) {
	stmt, err := tx.Prepare(`
		DELETE FROM shift2bikes_events
		WHERE citycode = ?
			AND ridesource = ?
			AND composite_event_id = ?
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare Shift2Bikes delete statement: %w", err)
	}
	defer stmt.Close()

	var deletedCount int64
	for _, id := range eventIDs {
		result, err := stmt.Exec(cityCode, shift2BikesRideSource, id)
		if err != nil {
			return deletedCount, fmt.Errorf("failed to delete Shift2Bikes event %s: %w", id, err)
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return deletedCount, fmt.Errorf("failed to read delete result for Shift2Bikes event %s: %w", id, err)
		}
		deletedCount += rowsAffected
	}

	return deletedCount, nil
}

func bulkUpsertRideDataTx(tx *sql.Tx, rideData []Shift2BikeEvent) error {
	stmt, err := tx.Prepare(`
        INSERT INTO shift2bikes_events (
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
						citycode,
						ridesource,
						source_data,
						route_id
        )
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
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
						citycode=excluded.citycode,
						ridesource=excluded.ridesource,
            source_data=excluded.source_data,
            route_id=excluded.route_id;
        `)
	if err != nil {
		return fmt.Errorf("failed to prepare ride data upsert statement: %v", err)
	}
	defer stmt.Close()

	for i := range rideData {
		ride := rideData[i]

		compositeKey := ride.CompositeID()
		if compositeKey == "" {
			slog.Warn("skipping Shift2Bikes event with no stable ID", "title", ride.Title, "date", ride.Date)
			continue
		}

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

		var routeID interface{}
		if ride.RouteID != "" {
			routeID = ride.RouteID
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
			ride.CityCode,
			ride.SourcedFrom,
			string(sourceData),
			routeID,
		)
		if err != nil {
			slog.Error("Failed to upsert single ride in batch", "key", compositeKey, "error", err.Error())
			return fmt.Errorf("failed to execute batch upsert for key %s: %w", compositeKey, err)
		}
	}

	return nil
}

func todayForCity(cityCode string) (string, error) {
	tzName := "America/Los_Angeles"
	if cityCode == "slc" {
		tzName = "America/Denver"
	}

	tz, err := time.LoadLocation(tzName)
	if err != nil {
		return "", fmt.Errorf("failed to load timezone for city %s: %w", cityCode, err)
	}

	return time.Now().In(tz).Format("2006-01-02"), nil
}
