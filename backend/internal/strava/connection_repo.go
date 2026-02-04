package strava

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ConnectionRepository manages persistent Strava connections
type ConnectionRepository struct {
	db         *sql.DB
	encryption *TokenEncryption
}

// NewConnectionRepository creates a new connection repository
func NewConnectionRepository(db *sql.DB, encryption *TokenEncryption) *ConnectionRepository {
	return &ConnectionRepository{
		db:         db,
		encryption: encryption,
	}
}

// Connection represents a stored Strava connection
type Connection struct {
	AthleteID    int64
	RefreshToken string // Decrypted
	CityCode     string
	LastSyncedAt *time.Time
	CreatedAt    time.Time
}

// SaveConnection stores or updates a Strava connection with encrypted refresh token
func (r *ConnectionRepository) SaveConnection(ctx context.Context, athleteID int64, refreshToken, cityCode string) error {
	// Encrypt refresh token
	encryptedToken, nonce, err := r.encryption.Encrypt(refreshToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt refresh token: %w", err)
	}

	// Upsert connection (INSERT or REPLACE on conflict)
	query := `
		INSERT INTO strava_connections (
			athlete_id,
			refresh_token_encrypted,
			encryption_nonce,
			city_code,
			created_at
		) VALUES (?, ?, ?, ?, STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
		ON CONFLICT(athlete_id) DO UPDATE SET
			refresh_token_encrypted = excluded.refresh_token_encrypted,
			encryption_nonce = excluded.encryption_nonce,
			city_code = excluded.city_code
	`

	_, err = r.db.ExecContext(ctx, query, athleteID, encryptedToken, nonce, cityCode)
	if err != nil {
		return fmt.Errorf("failed to save connection: %w", err)
	}

	return nil
}

// GetConnection retrieves a connection and decrypts the refresh token
func (r *ConnectionRepository) GetConnection(ctx context.Context, athleteID int64) (*Connection, error) {
	query := `
		SELECT
			athlete_id,
			refresh_token_encrypted,
			encryption_nonce,
			city_code,
			last_synced_at,
			created_at
		FROM strava_connections
		WHERE athlete_id = ?
	`

	var conn Connection
	var encryptedToken, nonce []byte
	var lastSyncedStr sql.NullString

	err := r.db.QueryRowContext(ctx, query, athleteID).Scan(
		&conn.AthleteID,
		&encryptedToken,
		&nonce,
		&conn.CityCode,
		&lastSyncedStr,
		&conn.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Connection not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query connection: %w", err)
	}

	// Decrypt refresh token
	refreshToken, err := r.encryption.Decrypt(encryptedToken, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
	}
	conn.RefreshToken = refreshToken

	// Parse last_synced_at if present
	if lastSyncedStr.Valid && lastSyncedStr.String != "" {
		t, err := time.Parse("2006-01-02 15:04:05.000", lastSyncedStr.String)
		if err == nil {
			conn.LastSyncedAt = &t
		}
	}

	return &conn, nil
}

// UpdateLastSynced updates the last_synced_at timestamp for a connection
func (r *ConnectionRepository) UpdateLastSynced(ctx context.Context, athleteID int64) error {
	query := `
		UPDATE strava_connections
		SET last_synced_at = STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')
		WHERE athlete_id = ?
	`

	_, err := r.db.ExecContext(ctx, query, athleteID)
	if err != nil {
		return fmt.Errorf("failed to update last_synced_at: %w", err)
	}

	return nil
}

// DeleteConnection removes a connection (for deauthorization)
func (r *ConnectionRepository) DeleteConnection(ctx context.Context, athleteID int64) error {
	query := `DELETE FROM strava_connections WHERE athlete_id = ?`

	_, err := r.db.ExecContext(ctx, query, athleteID)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	return nil
}

// ListConnections returns all connections (for background sync)
func (r *ConnectionRepository) ListConnections(ctx context.Context) ([]*Connection, error) {
	query := `
		SELECT
			athlete_id,
			refresh_token_encrypted,
			encryption_nonce,
			city_code,
			last_synced_at,
			created_at
		FROM strava_connections
		ORDER BY last_synced_at ASC NULLS FIRST
	`

	return r.queryConnections(ctx, query)
}

// GetConnectionsForSync returns connections that need syncing
// (last_synced_at is NULL or older than 3 days)
func (r *ConnectionRepository) GetConnectionsForSync(ctx context.Context, limit int) ([]*Connection, error) {
	query := `
		SELECT
			athlete_id,
			refresh_token_encrypted,
			encryption_nonce,
			city_code,
			last_synced_at,
			created_at
		FROM strava_connections
		WHERE last_synced_at IS NULL OR last_synced_at < datetime('now', '-3 days')
		ORDER BY last_synced_at ASC NULLS FIRST
		LIMIT ?
	`

	return r.queryConnectionsWithArgs(ctx, query, limit)
}

// queryConnections executes a query and returns connections (no args)
func (r *ConnectionRepository) queryConnections(ctx context.Context, query string) ([]*Connection, error) {
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query connections: %w", err)
	}
	defer rows.Close()

	return r.scanConnections(rows)
}

// queryConnectionsWithArgs executes a query with args and returns connections
func (r *ConnectionRepository) queryConnectionsWithArgs(ctx context.Context, query string, args ...interface{}) ([]*Connection, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query connections: %w", err)
	}
	defer rows.Close()

	return r.scanConnections(rows)
}

// scanConnections scans rows into Connection structs
func (r *ConnectionRepository) scanConnections(rows *sql.Rows) ([]*Connection, error) {
	var connections []*Connection
	for rows.Next() {
		var conn Connection
		var encryptedToken, nonce []byte
		var lastSyncedStr sql.NullString

		err := rows.Scan(
			&conn.AthleteID,
			&encryptedToken,
			&nonce,
			&conn.CityCode,
			&lastSyncedStr,
			&conn.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}

		// Decrypt refresh token
		refreshToken, err := r.encryption.Decrypt(encryptedToken, nonce)
		if err != nil {
			// Log error but continue with other connections
			continue
		}
		conn.RefreshToken = refreshToken

		// Parse last_synced_at if present
		if lastSyncedStr.Valid && lastSyncedStr.String != "" {
			t, err := time.Parse("2006-01-02 15:04:05.000", lastSyncedStr.String)
			if err == nil {
				conn.LastSyncedAt = &t
			}
		}

		connections = append(connections, &conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating connections: %w", err)
	}

	return connections, nil
}

