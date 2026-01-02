package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var apiHandler http.Handler
var db *sql.DB
var monitoringDB *sql.DB

func init() {
	var err error
	if os.Getenv("APP_ENV") == "dev" {
		// Used while in development
		err = godotenv.Load()
		if err != nil {
			log.Fatalf("failed to read environment variables: %v", err)
		}
	}

	// Connect to main database
	db, err = ConnectToDB()
	if err != nil {
		log.Fatalf("unable to connect to TursoDB: %v", err)
	}

	if err := db.Ping(); err != nil {
		slog.Error("failed to connect to TursoDB", "error", err)
		log.Fatalf("failed to connect to TursoDB")
	}
	slog.Info("Connected to main Turso database")

	// Connect to monitoring database
	monitoringDB, err = ConnectToMonitoringDB()
	if err != nil {
		log.Fatalf("unable to connect to Turso monitoring DB: %v", err)
	}

	if err := monitoringDB.Ping(); err != nil {
		log.Fatalf("failed to connect to Turso monitoring DB")
	}
	slog.Info("Connected to monitoring Turso database")

	apiHandler = NewRideAPIRouter(db, monitoringDB)

}

func main() {
	slog.Info("API Gateway started", "listening_on", ":8080", "eventarc_enabled", true)
	err := http.ListenAndServe(":8080", apiHandler)
	if err != nil {
		slog.Error("unable to start server", "error", err)
		log.Fatalf("FATAL: unable to start server: %v", err)
	}
	defer db.Close()
}
