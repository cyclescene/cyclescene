package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {

	// used in development
	if os.Getenv("APP_ENV") == "dev" {
		_ = godotenv.Load()
	}
	// DB Vars
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_DB_RW_TOKEN") == "" {
		log.Fatal("FATAL: Turso env variable not set properly")
	}

	// GOOGLE Vars
	if os.Getenv("GOOGLE_GEOCODING_API_KEY") == "" {
		log.Fatal("FATAL: GOOGLE_GEOCODING_API_KEY not properly set")
	}
	//
	// // set up logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})))
	//
	// connect to DB(Turso)
	dbURL := os.Getenv("TURSO_DB_URL")
	authToken := os.Getenv("TURSO_DB_RW_TOKEN")

	fullURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	db, err := sql.Open("libsql", fullURL)
	if err != nil {
		log.Fatalf("failed to open Turso DB connection: %v", err)
	}
	defer db.Close()

	startTime := time.Now()
	nowStr := time.Now().Format(time.RFC3339)

	query := `
		DELETE FROM submission_tokens
		WHERE expires_at < ?;`

	if _, err = db.Exec(query, nowStr); err != nil {
		slog.Error("Failed to remove expired submission_tokens", "time", nowStr, "error", err)
		log.Fatalf("failed to removed expired submission tokens: %v", err)
	}

	endTime := time.Since(time.Now())
	slog.Info("Cleared out all expired tokens!", "start_time", startTime, "end_time", endTime)

}
