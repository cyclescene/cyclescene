package admin

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/spacesedan/cyclescene/backend/internal/admin"
)

// SyncHandler handles admin endpoints for the Strava sync service
type SyncHandler struct {
	monitoringDB *sql.DB
	projectID    string
	region       string
}

// NewSyncHandler creates a new admin sync handler
func NewSyncHandler(monitoringDB *sql.DB) *SyncHandler {
	projectID := os.Getenv("GCP_PROJECT")
	if projectID == "" {
		projectID = "leaguefindr"
	}

	region := os.Getenv("GCP_REGION")
	if region == "" {
		region = "us-west1"
	}

	return &SyncHandler{
		monitoringDB: monitoringDB,
		projectID:    projectID,
		region:       region,
	}
}

// TriggerSyncResponse is the response for triggering a sync
type TriggerSyncResponse struct {
	Status    string `json:"status"`
	Execution string `json:"execution,omitempty"`
	Message   string `json:"message,omitempty"`
}

// SyncStatusResponse is the response for sync status
type SyncStatusResponse struct {
	LastSyncTime          *time.Time `json:"last_sync_time,omitempty"`
	TotalAPICallsLast7d   int        `json:"total_api_calls_7d"`
	SuccessfulCalls       int        `json:"successful_calls"`
	FailedCalls           int        `json:"failed_calls"`
	RateLimitUsage15m     int        `json:"rate_limit_usage_15m"`
	RateLimitUsageDaily   int        `json:"rate_limit_usage_daily"`
	UniqueAthletesLast7d  int        `json:"unique_athletes_7d"`
	RecentErrors          []APIError `json:"recent_errors,omitempty"`
}

// APIError represents a recent API error
type APIError struct {
	Timestamp  time.Time `json:"timestamp"`
	Endpoint   string    `json:"endpoint"`
	StatusCode int       `json:"status_code"`
	AthleteID  int64     `json:"athlete_id,omitempty"`
}

// TriggerSync handles POST /admin/sync/trigger
func (h *SyncHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create Cloud Run client
	runClient, err := admin.NewRunClient(ctx, h.projectID, h.region)
	if err != nil {
		slog.Error("failed_to_create_run_client", "error", err)
		h.writeJSON(w, http.StatusInternalServerError, TriggerSyncResponse{
			Status:  "error",
			Message: "Failed to create Cloud Run client",
		})
		return
	}
	defer runClient.Close()

	// Trigger the sync job
	executionName, err := runClient.TriggerStravaSync(ctx)
	if err != nil {
		slog.Error("failed_to_trigger_sync", "error", err)
		h.writeJSON(w, http.StatusInternalServerError, TriggerSyncResponse{
			Status:  "error",
			Message: "Failed to trigger sync job",
		})
		return
	}

	slog.Info("sync_triggered_manually",
		"execution", executionName,
		"remote_addr", r.RemoteAddr,
	)

	h.writeJSON(w, http.StatusOK, TriggerSyncResponse{
		Status:    "triggered",
		Execution: executionName,
	})
}

// GetStatus handles GET /admin/sync/status
func (h *SyncHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := SyncStatusResponse{}

	// Query for last sync time and stats
	if h.monitoringDB != nil {
		// Get total API calls in last 7 days
		row := h.monitoringDB.QueryRowContext(ctx, `
			SELECT
				COUNT(*) as total,
				SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END) as success,
				SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as failed,
				COUNT(DISTINCT athlete_id) as unique_athletes
			FROM strava_api_logs
			WHERE created_at > datetime('now', '-7 days')
		`)
		if err := row.Scan(&status.TotalAPICallsLast7d, &status.SuccessfulCalls, &status.FailedCalls, &status.UniqueAthletesLast7d); err != nil {
			slog.Debug("failed_to_query_api_stats", "error", err)
		}

		// Get current rate limit usage
		row = h.monitoringDB.QueryRowContext(ctx, `
			SELECT
				COALESCE(MAX(read_limit_15min_usage), 0) as usage_15m,
				COALESCE(MAX(read_limit_daily_usage), 0) as usage_daily
			FROM strava_api_logs
			WHERE created_at > datetime('now', '-15 minutes')
		`)
		if err := row.Scan(&status.RateLimitUsage15m, &status.RateLimitUsageDaily); err != nil {
			slog.Debug("failed_to_query_rate_limits", "error", err)
		}

		// Get last sync time
		row = h.monitoringDB.QueryRowContext(ctx, `
			SELECT created_at
			FROM strava_api_logs
			ORDER BY created_at DESC
			LIMIT 1
		`)
		var lastSyncStr string
		if err := row.Scan(&lastSyncStr); err == nil {
			if t, err := time.Parse("2006-01-02 15:04:05", lastSyncStr); err == nil {
				status.LastSyncTime = &t
			}
		}

		// Get recent errors
		rows, err := h.monitoringDB.QueryContext(ctx, `
			SELECT created_at, endpoint, status_code, athlete_id
			FROM strava_api_logs
			WHERE status_code >= 400
			AND created_at > datetime('now', '-7 days')
			ORDER BY created_at DESC
			LIMIT 10
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var apiErr APIError
				var timestampStr string
				var athleteID sql.NullInt64
				if err := rows.Scan(&timestampStr, &apiErr.Endpoint, &apiErr.StatusCode, &athleteID); err == nil {
					if t, err := time.Parse("2006-01-02 15:04:05", timestampStr); err == nil {
						apiErr.Timestamp = t
					}
					if athleteID.Valid {
						apiErr.AthleteID = athleteID.Int64
					}
					status.RecentErrors = append(status.RecentErrors, apiErr)
				}
			}
		}
	}

	h.writeJSON(w, http.StatusOK, status)
}

// writeJSON writes a JSON response
func (h *SyncHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed_to_write_json_response", "error", err)
	}
}
