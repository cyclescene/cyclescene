package strava

import (
	"os"
	"strconv"
)

// SyncConfig holds configuration for the background sync service
type SyncConfig struct {
	// Connection limits
	MaxConnectionsPerRun int // Max connections to process per run (default: 100)

	// Rate limit thresholds (conservative to avoid hitting Strava limits)
	MaxRequestsPer15Min int // Stop if we've made this many requests (default: 90, Strava limit: 100)
	MaxRequestsPerDay   int // Stop if approaching daily limit (default: 900, Strava limit: 1000)

	// Behavior
	ContinueOnError bool // Continue processing other athletes if one fails (default: true)
	Debug           bool // Enable debug logging
	Force           bool // Force sync all connections, ignoring last_synced_at (default: false, for testing)
}

// DefaultSyncConfig returns a SyncConfig with sensible defaults
func DefaultSyncConfig() *SyncConfig {
	return &SyncConfig{
		MaxConnectionsPerRun: 100,
		MaxRequestsPer15Min:  90,  // Conservative: 90% of Strava's 100/15min limit
		MaxRequestsPerDay:    900, // Conservative: 90% of Strava's 1000/day limit
		ContinueOnError:      true,
		Debug:                false,
	}
}

// NewSyncConfigFromEnv creates a SyncConfig from environment variables
// Falls back to defaults if env vars are not set
func NewSyncConfigFromEnv() *SyncConfig {
	config := DefaultSyncConfig()

	if v := os.Getenv("SYNC_MAX_CONNECTIONS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.MaxConnectionsPerRun = n
		}
	}

	if v := os.Getenv("SYNC_MAX_REQUESTS_15MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.MaxRequestsPer15Min = n
		}
	}

	if v := os.Getenv("SYNC_MAX_REQUESTS_DAY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.MaxRequestsPerDay = n
		}
	}

	if os.Getenv("SYNC_CONTINUE_ON_ERROR") == "false" {
		config.ContinueOnError = false
	}

	if os.Getenv("STRAVA_DEBUG") == "true" {
		config.Debug = true
	}

	if os.Getenv("SYNC_FORCE") == "true" {
		config.Force = true
	}

	return config
}
