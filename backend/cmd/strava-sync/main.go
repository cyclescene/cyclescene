package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spacesedan/cyclescene/backend/internal/alerts"
	"github.com/spacesedan/cyclescene/backend/internal/strava"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// Strava Background Sync Service
// Runs every 3 days at 2am PST to refresh Strava event data
// See docs/strava/BACKGROUND_SYNC_SERVICE.md for design details

const (
	// Default timeout for the entire sync run
	DefaultSyncTimeout = 25 * time.Minute

	// TestEncryptionKey is a 32-byte key for testing (DO NOT use in production!)
	// Base64 encoded: 32 bytes of zeros
	TestEncryptionKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
)

func main() {
	// Load environment variables
	if os.Getenv("APP_ENV") == "dev" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: failed to load .env file: %v", err)
		}
	}

	// Configure logging
	configureLogging()

	slog.Info("strava_sync_starting",
		"version", "1.0.0",
		"app_env", os.Getenv("APP_ENV"),
	)

	// Initialize alerting
	notifier := alerts.NewNotifier()
	if notifier.IsConfigured() {
		slog.Info("alerting_configured")
	} else {
		slog.Debug("alerting_not_configured")
	}

	// Run sync with timeout
	ctx, cancel := context.WithTimeout(context.Background(), DefaultSyncTimeout)
	defer cancel()

	result, err := runSync(ctx)

	// Check for critical failures and send alerts
	var shouldAlert bool
	var alertTitle, alertMessage string

	if err != nil {
		shouldAlert = true
		alertTitle = "Strava Sync Job Failed"
		alertMessage = fmt.Sprintf("Sync job encountered critical error: %v", err)
		slog.Error("sync_failed", "error", err)
	} else if result.ProcessedConnections > 0 && result.SuccessfulConnections == 0 {
		shouldAlert = true
		alertTitle = "Zero Athletes Synced"
		alertMessage = fmt.Sprintf("Job ran but synced 0/%d connections successfully. All connections failed.", result.ProcessedConnections)
		slog.Error("all_connections_failed",
			"processed", result.ProcessedConnections,
			"failed", result.FailedConnections,
		)
	}

	// Send alert if needed (don't let alerting failure affect exit code)
	if shouldAlert {
		alertCtx, alertCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer alertCancel()
		if alertErr := notifier.SendCriticalAlert(alertCtx, alertTitle, alertMessage); alertErr != nil {
			slog.Error("failed_to_send_alert", "error", alertErr)
		}
		os.Exit(1)
	}

	slog.Info("strava_sync_finished",
		"successful", result.SuccessfulConnections,
		"failed", result.FailedConnections,
		"events_refreshed", result.EventsRefreshed,
		"events_deleted", result.EventsDeleted,
	)
}

func runSync(ctx context.Context) (*strava.SyncResult, error) {
	// Connect to main database
	db, err := connectToDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to main database: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping main database: %w", err)
	}
	slog.Info("connected_to_main_db")

	// Connect to monitoring database
	monitoringDB, err := connectToMonitoringDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to monitoring database: %w", err)
	}
	defer monitoringDB.Close()

	if err := monitoringDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping monitoring database: %w", err)
	}
	slog.Info("connected_to_monitoring_db")

	// Initialize encryption
	encryption, err := strava.NewTokenEncryption()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryption: %w", err)
	}
	slog.Debug("encryption_initialized")

	// Initialize Strava client
	stravaConfig := &strava.Config{
		ClientID:     os.Getenv("STRAVA_CLIENT_ID"),
		ClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
		Debug:        os.Getenv("STRAVA_DEBUG") == "true",
	}

	if stravaConfig.ClientID == "" || stravaConfig.ClientSecret == "" {
		return nil, fmt.Errorf("STRAVA_CLIENT_ID and STRAVA_CLIENT_SECRET must be set")
	}

	client := strava.NewClient(stravaConfig)

	// Initialize repositories
	connRepo := strava.NewConnectionRepository(db, encryption)
	metadataRepo := strava.NewEventMetadataRepository(db)
	monitoringRepo := strava.NewMonitoringRepository(monitoringDB, stravaConfig.Debug)

	// Initialize sync config
	syncConfig := strava.NewSyncConfigFromEnv()
	slog.Info("sync_config_loaded",
		"max_connections", syncConfig.MaxConnectionsPerRun,
		"max_requests_15min", syncConfig.MaxRequestsPer15Min,
		"max_requests_day", syncConfig.MaxRequestsPerDay,
		"continue_on_error", syncConfig.ContinueOnError,
	)

	// Initialize and run sync service
	syncService := strava.NewSyncService(
		db,
		client,
		connRepo,
		metadataRepo,
		monitoringRepo,
		syncConfig,
	)

	return syncService.Run(ctx)
}

func configureLogging() {
	level := slog.LevelInfo
	if os.Getenv("STRAVA_DEBUG") == "true" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Use JSON handler for production (easier to parse in Cloud Logging)
	var handler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func connectToDB() (*sql.DB, error) {
	dbURL := os.Getenv("TURSO_DB_URL")
	authToken := os.Getenv("TURSO_DB_RW_TOKEN")

	if dbURL == "" {
		return nil, fmt.Errorf("TURSO_DB_URL environment variable not set")
	}

	fullURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)
	return sql.Open("libsql", fullURL)
}

func connectToMonitoringDB() (*sql.DB, error) {
	dbURL := os.Getenv("TURSO_MONITORING_DB_URL")
	authToken := os.Getenv("TURSO_MONITORING_DB_RW_TOKEN")

	if dbURL == "" {
		return nil, fmt.Errorf("TURSO_MONITORING_DB_URL environment variable not set")
	}

	fullURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)
	return sql.Open("libsql", fullURL)
}
