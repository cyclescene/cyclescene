package synclog

import (
	"database/sql"
	"log/slog"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// LogSync inserts a new sync log entry into the database
func (r *Repository) LogSync(log *SyncLog) error {
	query := `
		INSERT INTO sync_logs (client_id, sync_type, status, error_msg, ride_count, duration, timestamp, city_code, os)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(
		query,
		log.ClientID,
		log.SyncType,
		log.Status,
		log.ErrorMsg,
		log.RideCount,
		log.Duration,
		log.Timestamp,
		log.CityCode,
		log.OS,
	)

	if err != nil {
		slog.Error("Failed to log sync event", "error", err, "client_id", log.ClientID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("Failed to get last insert ID", "error", err)
		return err
	}

	log.ID = id
	return nil
}

// GetRecentLogs retrieves recent sync logs across all clients
func (r *Repository) GetRecentLogs(limit int) ([]SyncLog, error) {
	query := `
		SELECT id, client_id, sync_type, status, error_msg, ride_count, duration, timestamp, city_code, os
		FROM sync_logs
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		slog.Error("Failed to query recent logs", "error", err)
		return nil, err
	}
	defer rows.Close()

	var logs []SyncLog
	for rows.Next() {
		var log SyncLog
		err := rows.Scan(
			&log.ID,
			&log.ClientID,
			&log.SyncType,
			&log.Status,
			&log.ErrorMsg,
			&log.RideCount,
			&log.Duration,
			&log.Timestamp,
			&log.CityCode,
			&log.OS,
		)
		if err != nil {
			slog.Error("Failed to scan log row", "error", err)
			return nil, err
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		slog.Error("Error iterating log rows", "error", err)
		return nil, err
	}

	return logs, nil
}

// GetClientLogs retrieves sync logs for a specific client
func (r *Repository) GetClientLogs(clientID string, limit int) ([]SyncLog, error) {
	query := `
		SELECT id, client_id, sync_type, status, error_msg, ride_count, duration, timestamp, city_code, os
		FROM sync_logs
		WHERE client_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, clientID, limit)
	if err != nil {
		slog.Error("Failed to query client logs", "error", err, "client_id", clientID)
		return nil, err
	}
	defer rows.Close()

	var logs []SyncLog
	for rows.Next() {
		var log SyncLog
		err := rows.Scan(
			&log.ID,
			&log.ClientID,
			&log.SyncType,
			&log.Status,
			&log.ErrorMsg,
			&log.RideCount,
			&log.Duration,
			&log.Timestamp,
			&log.CityCode,
			&log.OS,
		)
		if err != nil {
			slog.Error("Failed to scan log row", "error", err)
			return nil, err
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		slog.Error("Error iterating log rows", "error", err)
		return nil, err
	}

	return logs, nil
}

// GetSyncStats returns aggregate statistics for syncs
func (r *Repository) GetSyncStats(since time.Time) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_syncs,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as successful_syncs,
			SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as failed_syncs,
			AVG(duration) as avg_duration,
			COUNT(DISTINCT client_id) as unique_clients
		FROM sync_logs
		WHERE timestamp >= ?
	`

	var (
		totalSyncs      int
		successfulSyncs int
		failedSyncs     int
		avgDuration     sql.NullInt64
		uniqueClients   int
	)

	err := r.db.QueryRow(query, since).Scan(
		&totalSyncs,
		&successfulSyncs,
		&failedSyncs,
		&avgDuration,
		&uniqueClients,
	)

	if err != nil {
		slog.Error("Failed to query sync stats", "error", err)
		return nil, err
	}

	stats := map[string]interface{}{
		"total_syncs":       totalSyncs,
		"successful_syncs":  successfulSyncs,
		"failed_syncs":      failedSyncs,
		"unique_clients":    uniqueClients,
	}

	if avgDuration.Valid {
		stats["avg_duration_ms"] = avgDuration.Int64
	}

	return stats, nil
}
