package strava

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/spacesedan/cyclescene/backend/internal/api/ride"
	"github.com/spacesedan/cyclescene/backend/internal/routes"
	"github.com/spacesedan/cyclescene/backend/internal/scraper"
)

// Service provides business logic for Strava OAuth and event import
type Service struct {
	client         *Client
	sessionStore   *SessionStore
	monitoringRepo *MonitoringRepository
	routeRepo      *routes.Repository
	callbackURL    string
	debug          bool
}

// NewService creates a new Strava service
func NewService(
	client *Client,
	sessionStore *SessionStore,
	monitoringRepo *MonitoringRepository,
	callbackURL string,
) *Service {
	return &Service{
		client:         client,
		sessionStore:   sessionStore,
		monitoringRepo: monitoringRepo,
		callbackURL:    callbackURL,
		debug:          os.Getenv("STRAVA_DEBUG") == "true",
	}
}

// InitiateOAuth generates a state token and returns the Strava authorization URL
// cityCode is stored with the state for retrieval during callback
func (s *Service) InitiateOAuth(ctx context.Context, cityCode string) (string, error) {
	// Validate city code
	if _, ok := cityMappings[cityCode]; !ok && cityCode != "" {
		return "", fmt.Errorf("unsupported city code: %s", cityCode)
	}

	// Generate state token with city context
	state, err := s.sessionStore.GenerateStateWithCity(cityCode)
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Build Strava authorization URL
	authURL := fmt.Sprintf(
		"https://www.strava.com/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=read,activity:read&state=%s",
		s.client.clientID,
		s.callbackURL,
		state,
	)

	if s.debug {
		slog.Debug("OAuth flow initiated",
			"state", state[:8]+"...",
			"city_code", cityCode,
			"callback_url", s.callbackURL,
		)
	}

	return authURL, nil
}

// HandleOAuthCallback exchanges the authorization code for tokens and creates a session
// Returns the session ID for the frontend to store
func (s *Service) HandleOAuthCallback(ctx context.Context, code, state string) (string, error) {
	// Validate state and get city code
	cityCode, valid := s.sessionStore.ValidateStateAndGetCity(state)
	if !valid {
		return "", fmt.Errorf("invalid or expired state token")
	}

	// Exchange code for token
	tokenResp, metrics, err := s.client.ExchangeToken(ctx, code)
	s.logAPICall(metrics, 0)
	if err != nil {
		return "", fmt.Errorf("token exchange failed: %w", err)
	}

	// Create session with city code
	session := &Session{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Unix(tokenResp.ExpiresAt, 0),
		AthleteID:    tokenResp.Athlete.ID,
		AthleteName:  fmt.Sprintf("%s %s", tokenResp.Athlete.FirstName, tokenResp.Athlete.LastName),
		CityCode:     cityCode,
	}

	sessionID, err := s.sessionStore.CreateSession(session)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	if s.debug {
		slog.Debug("OAuth callback processed",
			"athlete_id", tokenResp.Athlete.ID,
			"athlete_name", session.AthleteName,
			"session_id", sessionID[:8]+"...",
			"city_code", cityCode,
			"expires_at", session.ExpiresAt,
		)
	}

	return sessionID, nil
}

// GetAdminClubs fetches clubs where the user is an admin or owner
// Filters by city and cycling sport type to reduce API calls (Decision #4)
func (s *Service) GetAdminClubs(ctx context.Context, sessionID string) ([]ClubDetail, error) {
	session, ok := s.sessionStore.GetSession(sessionID)
	if !ok {
		return nil, ErrUnauthorized
	}

	// Fetch all clubs user belongs to
	clubs, metrics, err := s.client.GetAthleteClubs(ctx, session.AccessToken)
	metrics.AthleteID = session.AthleteID
	s.logAPICall(metrics, session.AthleteID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clubs: %w", err)
	}

	if s.debug {
		slog.Debug("Fetched athlete clubs",
			"athlete_id", session.AthleteID,
			"total_clubs", len(clubs),
			"city_code", session.CityCode,
		)
	}

	// Filter clubs by city and sport type BEFORE checking admin status
	// This reduces API calls significantly (83% reduction per Decision #4)
	var candidateClubs []Club
	for _, club := range clubs {
		// Skip non-cycling clubs
		if !club.IsCyclingClub() {
			if s.debug {
				slog.Debug("Skipping non-cycling club",
					"club_id", club.ID,
					"club_name", club.Name,
					"sport_type", club.SportType,
				)
			}
			continue
		}

		// Skip clubs not in target city (if city code is set)
		if session.CityCode != "" && !club.MatchesCity(session.CityCode) {
			if s.debug {
				slog.Debug("Skipping club outside target city",
					"club_id", club.ID,
					"club_name", club.Name,
					"club_city", club.City,
					"target_city", session.CityCode,
				)
			}
			continue
		}

		candidateClubs = append(candidateClubs, club)
	}

	if s.debug {
		slog.Debug("Filtered clubs for admin check",
			"total_clubs", len(clubs),
			"candidate_clubs", len(candidateClubs),
			"city_code", session.CityCode,
		)
	}

	// Now check admin status for remaining clubs
	var adminClubs []ClubDetail
	for _, club := range candidateClubs {
		detail, metrics, err := s.client.GetClubDetails(ctx, session.AccessToken, club.ID)
		metrics.AthleteID = session.AthleteID
		s.logAPICall(metrics, session.AthleteID)

		if err != nil {
			if s.debug {
				slog.Debug("Failed to fetch club details",
					"club_id", club.ID,
					"club_name", club.Name,
					"error", err,
				)
			}
			continue
		}

		if detail.IsAdminOrOwner() {
			adminClubs = append(adminClubs, *detail)
			if s.debug {
				slog.Debug("Found admin club",
					"club_id", detail.ID,
					"club_name", detail.Name,
					"admin", detail.Admin,
					"owner", detail.Owner,
				)
			}
		}
	}

	if s.debug {
		slog.Debug("Admin clubs filtered",
			"candidate_clubs", len(candidateClubs),
			"admin_clubs", len(adminClubs),
		)
	}

	return adminClubs, nil
}

// GetClubEvents fetches group events for a specific club
// Only call this after verifying user is admin of the club
func (s *Service) GetClubEvents(ctx context.Context, sessionID string, clubID int64) ([]GroupEvent, error) {
	session, ok := s.sessionStore.GetSession(sessionID)
	if !ok {
		return nil, ErrUnauthorized
	}

	events, metrics, err := s.client.GetClubEvents(ctx, session.AccessToken, clubID)
	metrics.AthleteID = session.AthleteID
	s.logAPICall(metrics, session.AthleteID)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch club events: %w", err)
	}

	if s.debug {
		slog.Debug("Fetched club events",
			"club_id", clubID,
			"events_count", len(events),
		)
	}

	return events, nil
}

// ConvertEventToSubmission converts a Strava event to a CycleScene submission
// Handles geocoding fallback if event doesn't have coordinates
func (s *Service) ConvertEventToSubmission(ctx context.Context, sessionID string, event *GroupEvent, organizerEmail string) (*ride.Submission, float64, float64, error) {
	session, ok := s.sessionStore.GetSession(sessionID)
	if !ok {
		return nil, 0, 0, ErrUnauthorized
	}

	cityCode := session.CityCode
	if cityCode == "" {
		return nil, 0, 0, fmt.Errorf("city code not set in session")
	}

	// Get coordinates - prefer event coordinates, fallback to geocoding
	var lat, lng float64

	if event.HasLocation() {
		lat = event.GetLatitude()
		lng = event.GetLongitude()
		if s.debug {
			slog.Debug("Using event coordinates",
				"event_id", event.ID,
				"lat", lat,
				"lng", lng,
			)
		}
	} else if event.Address != "" {
		// Geocode the address
		var err error
		lat, lng, err = scraper.GeocodeQuery(event.Address, cityCode)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("unable to import: no location information. Please add address or starting location in Strava and retry: %w", err)
		}
		if s.debug {
			slog.Debug("Geocoded event address",
				"event_id", event.ID,
				"address", event.Address,
				"lat", lat,
				"lng", lng,
			)
		}
	} else {
		return nil, 0, 0, fmt.Errorf("unable to import: no location information. Please add address or starting location in Strava and retry")
	}

	// Convert to submission
	submission, err := event.ToSubmission(cityCode, organizerEmail, lat, lng)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to convert event: %w", err)
	}

	return submission, lat, lng, nil
}

// GetSession returns the session for the given ID (for handlers that need session data)
func (s *Service) GetSession(sessionID string) (*Session, bool) {
	return s.sessionStore.GetSession(sessionID)
}

// DeleteSession removes a session (for logout)
func (s *Service) DeleteSession(sessionID string) {
	s.sessionStore.DeleteSession(sessionID)
}

// SetRouteRepository sets the route repository for route processing
func (s *Service) SetRouteRepository(repo *routes.Repository) {
	s.routeRepo = repo
}

// ProcessRoute fetches a Strava route and stores it in the database
// Returns the route ID for linking to events
func (s *Service) ProcessRoute(ctx context.Context, sessionID string, routeID int64, cityCode string) (*string, error) {
	session, ok := s.sessionStore.GetSession(sessionID)
	if !ok {
		return nil, ErrUnauthorized
	}

	if s.routeRepo == nil {
		return nil, fmt.Errorf("route repository not configured")
	}

	// Create fetcher with user's access token (per-request since each user has their own token)
	fetcher := routes.NewRouteFetcher(http.DefaultClient, session.AccessToken, "", "")

	// Construct Strava route export URL
	routeURL := fmt.Sprintf("https://www.strava.com/api/v3/routes/%d/export_gpx", routeID)

	if s.debug {
		slog.Debug("Fetching Strava route",
			"route_id", routeID,
			"athlete_id", session.AthleteID,
			"city_code", cityCode,
		)
	}

	// Fetch and convert route to GeoJSON
	feature, err := fetcher.FetchAndConvert(routeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch route: %w", err)
	}

	// Extract distance from feature properties (routes package sets this)
	distanceKm := 0.0
	distanceMi := 0.0
	if dist, ok := feature.Properties["distance"].(float64); ok {
		distanceKm = dist / 1000.0 // meters to km
		distanceMi = distanceKm * 0.621371
	}

	// Store route with deduplication (CreateRoute handles existing routes)
	routeIDStr, err := s.routeRepo.CreateRoute(
		ctx,
		"strava",
		fmt.Sprintf("%d", routeID),
		routeURL,
		cityCode,
		feature,
		distanceKm,
		distanceMi,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to store route: %w", err)
	}

	if s.debug {
		slog.Debug("Route stored successfully",
			"route_id", routeIDStr,
			"strava_route_id", routeID,
			"distance_km", distanceKm,
			"city_code", cityCode,
		)
	}

	return &routeIDStr, nil
}

// logAPICall logs API metrics to the monitoring database
func (s *Service) logAPICall(metrics *APICallMetrics, athleteID int64) {
	if metrics == nil {
		return
	}

	if athleteID != 0 {
		metrics.AthleteID = athleteID
	}

	// Log to monitoring DB (if configured)
	if s.monitoringRepo != nil {
		if err := s.monitoringRepo.LogAPICall(metrics); err != nil {
			slog.Error("Failed to log API call to monitoring DB",
				"error", err,
				"endpoint", metrics.Endpoint,
			)
		}

		// Check rate limit warnings
		s.monitoringRepo.CheckRateLimitWarning(metrics)
	}
}
