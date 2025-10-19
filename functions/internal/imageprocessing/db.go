package imageprocessing

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

// DBConnector handles database connections
type DBConnector struct {
	db *sql.DB
}

func NewDBConnector() (*DBConnector, error) {
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_DB_RW_TOKEN") == "" {
		log.Fatal("FATAL: Turso env variables not set properly")
	}

	dbURL := os.Getenv("TURSO_DB_URL")
	authToken := os.Getenv("TURSO_DB_RW_TOKEN")

	fullURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	db, err := sql.Open("libsql", fullURL)
	if err != nil {
		return nil, err
	}

	return &DBConnector{db: db}, nil
}

// UpdateImageURL updates the image_url field for a ride or group
func (d *DBConnector) UpdateImageURL(entityType, entityID, imageURL string) error {
	var table string

	// Determine which table to update based on entityType
	if entityType == "ride" {
		table = "rides"
	} else if entityType == "group" {
		table = "groups"
	} else {
		return fmt.Errorf("invalid entityType: %s", entityType)
	}

	query := fmt.Sprintf("UPDATE %s SET image_url = ? WHERE id = ?", table)

	result, err := d.db.Exec(query, imageURL, entityID)
	if err != nil {
		return fmt.Errorf("failed to update %s: %v", table, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s with id %s not found", table, entityID)
	}

	return nil
}

// Close closes the database connection
func (d *DBConnector) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// Ping checks if the database is reachable
func (d *DBConnector) Ping() error {
	return d.db.Ping()
}
