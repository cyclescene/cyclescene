package strava

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// SyncService orchestrates background sync of Strava events
type SyncService struct {
	db             *sql.DB
	client         *Client
	connRepo       *ConnectionRepository
	metadataRepo   *EventMetadataRepository
	monitoringRepo *MonitoringRepository
	config         *SyncConfig
}

// NewSyncService creates a new sync service
func NewSyncService(
	db *sql.DB,
	client *Client,
	connRepo *ConnectionRepository,
	metadataRepo *EventMetadataRepository,
	monitoringRepo *MonitoringRepository,
	config *SyncConfig,
) *SyncService {
	return &SyncService{
		db:             db,
		client:         client,
		connRepo:       connRepo,
		metadataRepo:   metadataRepo,
		monitoringRepo: monitoringRepo,
		config:         config,
	}
}

// Run executes the background sync process
func (s *SyncService) Run(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{
		StartedAt: time.Now(),
	}

	slog.Info("sync_started",
		"max_connections", s.config.MaxConnectionsPerRun,
		"max_requests_15min", s.config.MaxRequestsPer15Min,
		"continue_on_error", s.config.ContinueOnError,
	)

	// Fetch connections that need syncing (filtered by last_synced_at)
	connections, err := s.connRepo.GetConnectionsForSync(ctx, s.config.MaxConnectionsPerRun)
	if err != nil {
		slog.Error("failed_to_get_connections_for_sync", "error", err)
		return nil, err
	}

	result.TotalConnections = len(connections)
	slog.Info("connections_to_sync", "count", len(connections))

	// Process each connection
	for i, conn := range connections {
		// Check context cancellation
		if ctx.Err() != nil {
			slog.Warn("sync_cancelled", "reason", ctx.Err())
			break
		}

		// Check rate limits before processing
		if s.shouldStopDueToRateLimits(result) {
			slog.Warn("stopping_sync_rate_limit",
				"api_requests_used", result.APIRequestsUsed,
				"rate_limit_15min", result.RateLimitUsage15Min,
				"connections_processed", result.ProcessedConnections,
			)
			result.StoppedDueToRateLimit = true
			break
		}

		// Log progress every 10 connections
		if i > 0 && i%10 == 0 {
			slog.Info("sync_progress",
				"processed", i,
				"total", len(connections),
				"api_requests", result.APIRequestsUsed,
			)
		}

		// Sync this athlete
		athleteResult := s.syncAthlete(ctx, conn)
		result.ProcessedConnections++

		// Aggregate stats
		result.APIRequestsUsed += athleteResult.APIRequestsUsed
		result.EventsRefreshed += athleteResult.EventsRefreshed
		result.EventsDeleted += athleteResult.EventsDeleted

		// Track rate limit usage
		if athleteResult.RateLimitUsage15Min > result.RateLimitUsage15Min {
			result.RateLimitUsage15Min = athleteResult.RateLimitUsage15Min
		}
		if athleteResult.RateLimitUsageDaily > result.RateLimitUsageDaily {
			result.RateLimitUsageDaily = athleteResult.RateLimitUsageDaily
		}

		// Handle errors
		if athleteResult.Error != nil {
			result.FailedConnections++
			result.Errors = append(result.Errors, *athleteResult.Error)

			slog.Warn("athlete_sync_failed",
				"athlete_id", conn.AthleteID,
				"error_type", athleteResult.Error.ErrorType,
				"message", athleteResult.Error.Message,
			)

			if !s.config.ContinueOnError {
				slog.Error("stopping_sync_due_to_error")
				break
			}
		} else {
			result.SuccessfulConnections++
			slog.Debug("athlete_sync_success",
				"athlete_id", conn.AthleteID,
				"events_refreshed", athleteResult.EventsRefreshed,
				"events_deleted", athleteResult.EventsDeleted,
			)
		}
	}

	result.CompletedAt = time.Now()

	slog.Info("sync_completed",
		"duration_ms", result.Duration().Milliseconds(),
		"connections_processed", result.ProcessedConnections,
		"connections_successful", result.SuccessfulConnections,
		"connections_failed", result.FailedConnections,
		"events_refreshed", result.EventsRefreshed,
		"events_deleted", result.EventsDeleted,
		"api_requests_used", result.APIRequestsUsed,
		"stopped_rate_limit", result.StoppedDueToRateLimit,
	)

	return result, nil
}

// syncAthlete syncs all events for a single athlete
func (s *SyncService) syncAthlete(ctx context.Context, conn *Connection) *AthleteSync {
	result := NewAthleteSync(conn.AthleteID, conn.CityCode)

	slog.Debug("sync_athlete_start",
		"athlete_id", conn.AthleteID,
		"city_code", conn.CityCode,
	)

	// 1. Refresh the access token
	tokenResp, metrics, err := s.client.RefreshToken(ctx, conn.RefreshToken)
	result.APIRequestsUsed++

	// Log API call to monitoring
	if metrics != nil {
		metrics.AthleteID = conn.AthleteID
		s.monitoringRepo.LogAPICall(metrics)
		s.monitoringRepo.CheckRateLimitWarning(metrics)
		result.RateLimitUsage15Min = metrics.ReadLimit15minUsage
		result.RateLimitUsageDaily = metrics.ReadLimitDailyUsage
	}

	if err != nil {
		errorType := s.classifyError(err)
		result.SetError(errorType, err.Error())
		slog.Warn("token_refresh_failed",
			"athlete_id", conn.AthleteID,
			"error_type", errorType,
			"error", err,
		)
		return result
	}

	accessToken := tokenResp.AccessToken
	slog.Debug("token_refreshed", "athlete_id", conn.AthleteID)

	// 2. Fetch athlete's clubs
	clubs, metrics, err := s.client.GetAthleteClubs(ctx, accessToken)
	result.APIRequestsUsed++

	if metrics != nil {
		metrics.AthleteID = conn.AthleteID
		s.monitoringRepo.LogAPICall(metrics)
		s.monitoringRepo.CheckRateLimitWarning(metrics)
		result.RateLimitUsage15Min = metrics.ReadLimit15minUsage
		result.RateLimitUsageDaily = metrics.ReadLimitDailyUsage
	}

	if err != nil {
		errorType := s.classifyError(err)
		result.SetError(errorType, err.Error())
		slog.Warn("fetch_clubs_failed",
			"athlete_id", conn.AthleteID,
			"error_type", errorType,
			"error", err,
		)
		return result
	}

	result.ClubsFound = len(clubs)
	slog.Debug("clubs_fetched",
		"athlete_id", conn.AthleteID,
		"clubs_count", len(clubs),
	)

	// 3. Filter to admin clubs (cycling, matching city, admin/owner)
	adminClubs := s.filterAdminClubs(ctx, accessToken, clubs, conn.CityCode, result)
	result.AdminClubsFound = len(adminClubs)

	if len(adminClubs) == 0 {
		slog.Debug("no_admin_clubs",
			"athlete_id", conn.AthleteID,
			"city_code", conn.CityCode,
		)
		// Not an error - athlete just doesn't have any qualifying clubs
		// Update last_synced_at to avoid re-processing
		if err := s.connRepo.UpdateLastSynced(ctx, conn.AthleteID); err != nil {
			slog.Warn("failed_to_update_last_synced", "athlete_id", conn.AthleteID, "error", err)
		}
		return result
	}

	slog.Debug("admin_clubs_found",
		"athlete_id", conn.AthleteID,
		"admin_clubs_count", len(adminClubs),
	)

	// 4. Sync events for each admin club
	for _, club := range adminClubs {
		s.syncClubEvents(ctx, accessToken, conn, club, result)

		// Check if we hit an error during club sync
		if result.Error != nil {
			return result
		}
	}

	// 5. Update last_synced_at on success
	if err := s.connRepo.UpdateLastSynced(ctx, conn.AthleteID); err != nil {
		slog.Warn("failed_to_update_last_synced", "athlete_id", conn.AthleteID, "error", err)
	}

	slog.Debug("sync_athlete_complete",
		"athlete_id", conn.AthleteID,
		"events_refreshed", result.EventsRefreshed,
		"events_deleted", result.EventsDeleted,
		"api_requests", result.APIRequestsUsed,
	)

	return result
}

// filterAdminClubs filters clubs to only those where the athlete is admin/owner
// and that match the city and are cycling clubs
func (s *SyncService) filterAdminClubs(
	ctx context.Context,
	accessToken string,
	clubs []Club,
	cityCode string,
	result *AthleteSync,
) []*ClubDetail {
	var adminClubs []*ClubDetail

	for _, club := range clubs {
		// 1. Check if it's a cycling club
		if !club.IsCyclingClub() {
			slog.Debug("club_not_cycling",
				"club_id", club.ID,
				"club_name", club.Name,
				"sport_type", club.SportType,
			)
			continue
		}

		// 2. Check if it matches the athlete's city
		if !club.MatchesCity(cityCode) {
			slog.Debug("club_city_mismatch",
				"club_id", club.ID,
				"club_name", club.Name,
				"club_city", club.City,
				"athlete_city", cityCode,
			)
			continue
		}

		// 3. Fetch club details to check admin status
		// This requires an API call per club
		clubDetail, metrics, err := s.client.GetClubDetails(ctx, accessToken, club.ID)
		result.APIRequestsUsed++

		if metrics != nil {
			metrics.AthleteID = result.AthleteID
			s.monitoringRepo.LogAPICall(metrics)
			s.monitoringRepo.CheckRateLimitWarning(metrics)
			result.RateLimitUsage15Min = metrics.ReadLimit15minUsage
			result.RateLimitUsageDaily = metrics.ReadLimitDailyUsage
		}

		if err != nil {
			slog.Warn("failed_to_get_club_details",
				"club_id", club.ID,
				"error", err,
			)
			// Continue to next club, don't fail entire sync
			continue
		}

		// 4. Check if athlete is admin or owner
		if !clubDetail.IsAdminOrOwner() {
			slog.Debug("club_not_admin",
				"club_id", club.ID,
				"club_name", club.Name,
				"admin", clubDetail.Admin,
				"owner", clubDetail.Owner,
			)
			continue
		}

		slog.Debug("admin_club_found",
			"club_id", club.ID,
			"club_name", club.Name,
			"admin", clubDetail.Admin,
			"owner", clubDetail.Owner,
		)

		adminClubs = append(adminClubs, clubDetail)
	}

	return adminClubs
}

// syncClubEvents syncs events for a single club
func (s *SyncService) syncClubEvents(
	ctx context.Context,
	accessToken string,
	conn *Connection,
	club *ClubDetail,
	result *AthleteSync,
) {
	slog.Debug("sync_club_events_start",
		"athlete_id", conn.AthleteID,
		"club_id", club.ID,
		"club_name", club.Name,
	)

	// 1. Fetch club events from Strava
	stravaEvents, metrics, err := s.client.GetClubEvents(ctx, accessToken, club.ID)
	result.APIRequestsUsed++

	if metrics != nil {
		metrics.AthleteID = conn.AthleteID
		s.monitoringRepo.LogAPICall(metrics)
		s.monitoringRepo.CheckRateLimitWarning(metrics)
		result.RateLimitUsage15Min = metrics.ReadLimit15minUsage
		result.RateLimitUsageDaily = metrics.ReadLimitDailyUsage
	}

	if err != nil {
		errorType := s.classifyError(err)
		slog.Warn("fetch_club_events_failed",
			"club_id", club.ID,
			"error_type", errorType,
			"error", err,
		)
		// Don't fail entire athlete sync, just skip this club
		return
	}

	// 2. Filter to upcoming events only
	upcomingEvents := FilterUpcomingEvents(stravaEvents)
	slog.Debug("club_events_fetched",
		"club_id", club.ID,
		"total_events", len(stravaEvents),
		"upcoming_events", len(upcomingEvents),
	)

	// 3. Get stored events for this athlete
	storedMetadata, err := s.metadataRepo.GetUpcomingStravaEventsByAthlete(ctx, conn.AthleteID)
	if err != nil {
		slog.Warn("failed_to_get_stored_events",
			"athlete_id", conn.AthleteID,
			"error", err,
		)
		return
	}

	// 4. Filter stored metadata to this club only
	clubMetadata := filterMetadataByClub(storedMetadata, club.ID)
	slog.Debug("stored_events_for_club",
		"club_id", club.ID,
		"total_stored", len(storedMetadata),
		"club_stored", len(clubMetadata),
	)

	// 5. Compare Strava events with stored events
	comparison := s.compareEvents(upcomingEvents, clubMetadata)
	slog.Debug("event_comparison",
		"club_id", club.ID,
		"to_refresh", len(comparison.ToRefresh),
		"to_delete", len(comparison.ToDelete),
	)

	// 6. Process updates (refresh timestamps, delete stale events)
	if err := s.processEventUpdates(ctx, comparison, result); err != nil {
		slog.Warn("process_event_updates_failed",
			"club_id", club.ID,
			"error", err,
		)
	}

	slog.Debug("sync_club_events_complete",
		"club_id", club.ID,
		"refreshed", len(comparison.ToRefresh),
		"deleted", len(comparison.ToDelete),
	)
}

// compareEvents compares Strava events with stored event metadata
func (s *SyncService) compareEvents(
	stravaEvents []GroupEvent,
	storedMetadata []*EventMetadata,
) *EventComparison {
	result := &EventComparison{}

	// Build map of Strava event IDs for O(1) lookup
	stravaEventMap := make(map[int64]bool)
	for _, event := range stravaEvents {
		stravaEventMap[event.ID] = true
	}

	// Check each stored event
	for _, meta := range storedMetadata {
		if stravaEventMap[meta.StravaEventID] {
			// Event exists on Strava - refresh timestamp
			result.ToRefresh = append(result.ToRefresh, meta.EventID)
		} else {
			// Event no longer on Strava - delete
			result.ToDelete = append(result.ToDelete, meta.EventID)
		}
	}

	return result
}

// processEventUpdates refreshes and deletes events based on comparison
func (s *SyncService) processEventUpdates(
	ctx context.Context,
	comparison *EventComparison,
	result *AthleteSync,
) error {
	// 1. Refresh events (update last_refreshed_at and increment refresh_count)
	for _, eventID := range comparison.ToRefresh {
		if err := s.metadataRepo.UpdateLastRefreshed(ctx, eventID); err != nil {
			slog.Warn("failed_to_refresh_event",
				"event_id", eventID,
				"error", err,
			)
			// Continue with other events
			continue
		}
		result.EventsRefreshed++
		slog.Debug("event_refreshed", "event_id", eventID)
	}

	// 2. Delete stale events (no longer on Strava)
	for _, eventID := range comparison.ToDelete {
		if err := s.deleteEvent(ctx, eventID); err != nil {
			slog.Warn("failed_to_delete_event",
				"event_id", eventID,
				"error", err,
			)
			// Continue with other events
			continue
		}
		result.EventsDeleted++
		slog.Info("event_deleted",
			"event_id", eventID,
			"reason", "no_longer_on_strava",
		)
	}

	return nil
}

// deleteEvent removes a stale event from the database
func (s *SyncService) deleteEvent(ctx context.Context, eventID int64) error {
	// Delete from events table - CASCADE will delete from strava_event_metadata
	query := `DELETE FROM events WHERE id = ? AND source = 'strava'`
	result, err := s.db.ExecContext(ctx, query, eventID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Event may have been detached (source changed to 'cyclescene')
		slog.Debug("event_not_deleted_may_be_detached", "event_id", eventID)
	}

	return nil
}

// shouldStopDueToRateLimits checks if we should stop syncing due to rate limits
func (s *SyncService) shouldStopDueToRateLimits(result *SyncResult) bool {
	// Check our own request counter
	if result.APIRequestsUsed >= s.config.MaxRequestsPer15Min {
		slog.Warn("approaching_15min_rate_limit",
			"requests_used", result.APIRequestsUsed,
			"limit", s.config.MaxRequestsPer15Min,
		)
		return true
	}

	// Check Strava's reported usage (from headers)
	if result.RateLimitUsage15Min >= 90 {
		slog.Warn("strava_15min_rate_limit_high",
			"usage", result.RateLimitUsage15Min,
		)
		return true
	}

	// Check daily limit
	if result.RateLimitUsageDaily >= s.config.MaxRequestsPerDay {
		slog.Warn("approaching_daily_rate_limit",
			"usage", result.RateLimitUsageDaily,
			"limit", s.config.MaxRequestsPerDay,
		)
		return true
	}

	return false
}

// classifyError classifies an error into a sync error type
func (s *SyncService) classifyError(err error) string {
	if IsUnauthorized(err) {
		return "token_revoked"
	}
	if IsRateLimitExceeded(err) {
		return "rate_limit"
	}
	if IsNotFound(err) {
		return "not_found"
	}
	return "api_error"
}

// filterMetadataByClub filters event metadata to only events from a specific club
func filterMetadataByClub(metadata []*EventMetadata, clubID int64) []*EventMetadata {
	var filtered []*EventMetadata
	for _, meta := range metadata {
		if meta.StravaClubID == clubID {
			filtered = append(filtered, meta)
		}
	}
	return filtered
}
