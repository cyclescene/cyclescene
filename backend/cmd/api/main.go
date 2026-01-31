package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

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

// findAvailablePort attempts to find an available port starting from the given port
func findAvailablePort(startPort int) (int, error) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		port := startPort + i
		addr := fmt.Sprintf(":%d", port)

		// Try to listen on the port
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			// Port is available, close the listener and return the port
			listener.Close()
			return port, nil
		}

		slog.Warn("Port in use, trying next port", "port", port, "attempt", i+1)
	}

	return 0, fmt.Errorf("could not find available port after %d attempts starting from %d", maxAttempts, startPort)
}

func main() {
	// Get port from environment or use default
	defaultPort := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			defaultPort = p
		}
	}

	// Find an available port
	port, err := findAvailablePort(defaultPort)
	if err != nil {
		slog.Error("unable to find available port", "error", err)
		log.Fatalf("FATAL: %v", err)
	}

	addr := fmt.Sprintf(":%d", port)
	if port != defaultPort {
		slog.Warn("Configured port was in use, using alternative port", "configured_port", defaultPort, "actual_port", port)
	}

	slog.Info("API Gateway started", "listening_on", addr, "eventarc_enabled", true)
	err = http.ListenAndServe(addr, apiHandler)
	if err != nil {
		slog.Error("unable to start server", "error", err)
		log.Fatalf("FATAL: unable to start server: %v", err)
	}
	defer db.Close()
}
