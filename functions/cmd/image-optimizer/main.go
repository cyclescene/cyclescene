package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/spacesedan/cyclescene/functions/internal/imageprocessing"
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

func main() {
	slog.Info("Image Optimizer started", "listening_on", ":8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		slog.Error("unable to start server", "error", err)
		log.Fatalf("FATAL: unable to start server: %v", err)
	}
	defer dbConnector.Close()
}
