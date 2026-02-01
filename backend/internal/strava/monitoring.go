package strava

import (
	"database/sql"
	"log/slog"
)

// APICallMetrics contains metrics from a Strava API call for monitoring
type APICallMetrics struct {
	// Request details
	Endpoint       string
	Method         string
	StatusCode     int
	ResponseTimeMs int

	// General rate limit tracking (from X-Ratelimit-* headers)
	// IMPORTANT: Headers are case-sensitive! Use "X-Ratelimit-*" NOT "X-RateLimit-*"
	RateLimit15minUsage int // Parsed from "X-Ratelimit-Usage: 7,7" -> 7
	RateLimit15minLimit int // Parsed from "X-Ratelimit-Limit: 200,2000" -> 200
	RateLimitDailyUsage int // Parsed from "X-Ratelimit-Usage: 7,7" -> 7
	RateLimitDailyLimit int // Parsed from "X-Ratelimit-Limit: 200,2000" -> 2000

	// Read-only rate limit tracking (from X-Readratelimit-* headers)
	// CRITICAL: This is the ACTUAL constraint for our GET-heavy operations!
	ReadLimit15minUsage int // Parsed from "X-Readratelimit-Usage: 7,7" -> 7
	ReadLimit15minLimit int // Parsed from "X-Readratelimit-Limit: 100,1000" -> 100
	ReadLimitDailyUsage int // Parsed from "X-Readratelimit-Usage: 7,7" -> 7
	ReadLimitDailyLimit int // Parsed from "X-Readratelimit-Limit: 100,1000" -> 1000

	// Response data (privacy-safe)
	Message     string
	ClubsCount  int   // For /athlete/clubs endpoint
	EventsCount int   // For /group_events endpoint
	AthleteID   int64 // From token response or context
}

// MonitoringRepository handles logging API calls to the monitoring database
type MonitoringRepository struct {
	db    *sql.DB
	debug bool
}

// NewMonitoringRepository creates a new monitoring repository
func NewMonitoringRepository(db *sql.DB, debug bool) *MonitoringRepository {
	return &MonitoringRepository{
		db:    db,
		debug: debug,
	}
}

// LogAPICall logs a Strava API call to the monitoring database
func (r *MonitoringRepository) LogAPICall(metrics *APICallMetrics) error {
	if r.db == nil {
		if r.debug {
			slog.Debug("Monitoring DB not configured, skipping API call log",
				"endpoint", metrics.Endpoint,
				"status_code", metrics.StatusCode,
			)
		}
		return nil
	}

	query := `
		INSERT INTO strava_api_logs (
			endpoint,
			method,
			status_code,
			response_time_ms,
			rate_limit_15min_usage,
			rate_limit_15min_limit,
			rate_limit_daily_usage,
			rate_limit_daily_limit,
			read_limit_15min_usage,
			read_limit_15min_limit,
			read_limit_daily_usage,
			read_limit_daily_limit,
			message,
			clubs_count,
			events_count,
			athlete_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		metrics.Endpoint,
		metrics.Method,
		metrics.StatusCode,
		metrics.ResponseTimeMs,
		metrics.RateLimit15minUsage,
		metrics.RateLimit15minLimit,
		metrics.RateLimitDailyUsage,
		metrics.RateLimitDailyLimit,
		metrics.ReadLimit15minUsage,
		metrics.ReadLimit15minLimit,
		metrics.ReadLimitDailyUsage,
		metrics.ReadLimitDailyLimit,
		metrics.Message,
		metrics.ClubsCount,
		metrics.EventsCount,
		metrics.AthleteID,
	)

	if err != nil {
		slog.Error("Failed to log Strava API call to monitoring DB",
			"error", err,
			"endpoint", metrics.Endpoint,
			"status_code", metrics.StatusCode,
		)
		return err
	}

	if r.debug {
		slog.Debug("Logged Strava API call",
			"endpoint", metrics.Endpoint,
			"status_code", metrics.StatusCode,
			"response_time_ms", metrics.ResponseTimeMs,
			"read_limit_usage", metrics.ReadLimit15minUsage,
			"read_limit_limit", metrics.ReadLimit15minLimit,
		)
	}

	return nil
}

// CheckRateLimitWarning logs a warning if rate limits are approaching threshold
func (r *MonitoringRepository) CheckRateLimitWarning(metrics *APICallMetrics) {
	// Warn at 80% of read limit (the actual constraint for our operations)
	if metrics.ReadLimit15minLimit > 0 {
		usagePercent := float64(metrics.ReadLimit15minUsage) / float64(metrics.ReadLimit15minLimit) * 100
		if usagePercent >= 80 {
			slog.Warn("Strava read rate limit approaching threshold",
				"usage", metrics.ReadLimit15minUsage,
				"limit", metrics.ReadLimit15minLimit,
				"percent", usagePercent,
			)
		}
	}

	// Also check daily limits
	if metrics.ReadLimitDailyLimit > 0 {
		usagePercent := float64(metrics.ReadLimitDailyUsage) / float64(metrics.ReadLimitDailyLimit) * 100
		if usagePercent >= 80 {
			slog.Warn("Strava daily read rate limit approaching threshold",
				"usage", metrics.ReadLimitDailyUsage,
				"limit", metrics.ReadLimitDailyLimit,
				"percent", usagePercent,
			)
		}
	}
}
