package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/spacesedan/cyclescene/backend/internal/imageprocessing"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var dbConnector *imageprocessing.DBConnector
var router *chi.Mux

func init() {
	var err error
	if os.Getenv("APP_ENV") == "dev" {
		err = godotenv.Load()
		if err != nil {
			log.Fatalf("failed to read environment variables: %v", err)
		}
	}

	// Set up logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})))

	slog.Info("Starting image optimizer service")

	// Connect to database
	dbConnector, err = imageprocessing.NewDBConnector()
	if err != nil {
		slog.Error("unable to connect to database", "error", err)
		log.Fatalf("unable to connect to database: %v", err)
	}

	if err := dbConnector.Ping(); err != nil {
		slog.Error("failed to ping database", "error", err)
		log.Fatalf("failed to ping database: %v", err)
	}

	slog.Info("Connected to database")

	// Set up router
	router = chi.NewRouter()
	setupRoutes(router, dbConnector)
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

	// Image optimizer service with proper storage and Eventarc permissions
	slog.Info("Image Optimizer started", "listening_on", addr, "storage_permissions_configured", true)
	err = http.ListenAndServe(addr, router)
	if err != nil {
		slog.Error("unable to start server", "error", err)
		log.Fatalf("FATAL: unable to start server: %v", err)
	}
	defer dbConnector.Close()
}
