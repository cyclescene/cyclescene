package main

import (
	"database/sql"
	"fmt"
	"os"
)

// ConnectToDB connects to the main ride event database
func ConnectToDB() (*sql.DB, error) {
	dbURL := os.Getenv("TURSO_DB_URL")
	authToken := os.Getenv("TURSO_DB_RW_TOKEN")

	fullURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	return sql.Open("libsql", fullURL)
}

// ConnectToMonitoringDB connects to the separate monitoring/analytics database
func ConnectToMonitoringDB() (*sql.DB, error) {
	dbURL := os.Getenv("TURSO_MONITORING_DB_URL")
	authToken := os.Getenv("TURSO_MONITORING_DB_RW_TOKEN")

	fullURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	return sql.Open("libsql", fullURL)
}
