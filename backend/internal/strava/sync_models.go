package strava

import (
	"time"
)

// SyncResult contains the aggregated results from a sync run
type SyncResult struct {
	// Timing
	StartedAt   time.Time
	CompletedAt time.Time

	// Connection stats
	TotalConnections      int // Total connections found
	ProcessedConnections  int // Connections we attempted to sync
	SuccessfulConnections int // Connections synced without error
	FailedConnections     int // Connections that failed

	// Event stats
	EventsRefreshed int // Events with updated refresh timestamp
	EventsDeleted   int // Stale events removed from database

	// API usage
	APIRequestsUsed      int // Total API requests made this run
	RateLimitUsage15Min  int // Last seen 15-minute rate limit usage
	RateLimitUsageDaily  int // Last seen daily rate limit usage
	StoppedDueToRateLimit bool // True if sync stopped early due to rate limits

	// Errors (not individual athlete errors, but service-level issues)
	Errors []SyncError
}

// Duration returns the total duration of the sync run
func (r *SyncResult) Duration() time.Duration {
	if r.CompletedAt.IsZero() {
		return time.Since(r.StartedAt)
	}
	return r.CompletedAt.Sub(r.StartedAt)
}

// SyncError represents an error that occurred during sync
type SyncError struct {
	AthleteID int64
	ErrorType string // "token_revoked", "rate_limit", "api_error", "db_error", "decrypt_error"
	Message   string
	Timestamp time.Time
}

// AthleteSync contains the results from syncing a single athlete's events
type AthleteSync struct {
	AthleteID int64
	CityCode  string

	// Stats
	ClubsFound      int // Total clubs athlete belongs to
	AdminClubsFound int // Clubs where athlete is admin/owner
	EventsRefreshed int // Events refreshed for this athlete
	EventsDeleted   int // Events deleted for this athlete
	APIRequestsUsed int // API requests used for this athlete

	// Rate limit info (from last API call)
	RateLimitUsage15Min int
	RateLimitUsageDaily int

	// Error (nil if successful)
	Error *SyncError
}

// EventComparison holds the results of comparing Strava events to stored events
type EventComparison struct {
	ToRefresh []int64 // Event IDs that exist on both Strava and DB (refresh timestamp)
	ToDelete  []int64 // Event IDs in DB but not on Strava (delete)
}

// NewAthleteSync creates a new AthleteSync result for the given athlete
func NewAthleteSync(athleteID int64, cityCode string) *AthleteSync {
	return &AthleteSync{
		AthleteID: athleteID,
		CityCode:  cityCode,
	}
}

// SetError sets an error on the athlete sync result
func (a *AthleteSync) SetError(errorType, message string) {
	a.Error = &SyncError{
		AthleteID: a.AthleteID,
		ErrorType: errorType,
		Message:   message,
		Timestamp: time.Now(),
	}
}
