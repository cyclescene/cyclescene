package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// BaseURL is the Strava API base URL
	BaseURL = "https://www.strava.com/api/v3"

	// OAuthTokenURL is the Strava OAuth token endpoint
	OAuthTokenURL = "https://www.strava.com/oauth/token"

	// DefaultTimeout is the default HTTP request timeout
	DefaultTimeout = 30 * time.Second
)

// Client is the Strava API client
type Client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	baseURL      string
	debug        bool
}

// NewClient creates a new Strava API client
func NewClient(config *Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		baseURL:      BaseURL,
		debug:        config.Debug,
	}
}

// ExchangeToken exchanges an authorization code for access and refresh tokens
func (c *Client) ExchangeToken(ctx context.Context, code string) (*TokenResponse, *APICallMetrics, error) {
	start := time.Now()
	endpoint := "/oauth/token"

	if c.debug {
		slog.Debug("Exchanging OAuth code for token",
			"endpoint", endpoint,
		)
	}

	// Build form data
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", OAuthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange token: %w", err)
	}
	defer resp.Body.Close()

	metrics := c.buildMetrics(endpoint, "POST", resp, start)

	if resp.StatusCode != http.StatusOK {
		apiErr := c.handleErrorResponse(resp, endpoint)
		metrics.Message = apiErr.Message
		return nil, metrics, apiErr
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		metrics.Message = "failed to decode response"
		return nil, metrics, fmt.Errorf("failed to decode token response: %w", err)
	}

	metrics.Message = "ok"
	metrics.AthleteID = tokenResp.Athlete.ID

	if c.debug {
		slog.Debug("OAuth token exchange successful",
			"athlete_id", tokenResp.Athlete.ID,
			"expires_in", tokenResp.ExpiresIn,
			"token_type", tokenResp.TokenType,
			// NEVER log access_token or refresh_token
		)
	}

	return &tokenResp, metrics, nil
}

// RefreshToken refreshes an expired access token
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, *APICallMetrics, error) {
	start := time.Now()
	endpoint := "/oauth/token"

	if c.debug {
		slog.Debug("Refreshing OAuth token",
			"endpoint", endpoint,
		)
	}

	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, "POST", OAuthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	metrics := c.buildMetrics(endpoint, "POST", resp, start)

	if resp.StatusCode != http.StatusOK {
		apiErr := c.handleErrorResponse(resp, endpoint)
		metrics.Message = apiErr.Message
		return nil, metrics, apiErr
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		metrics.Message = "failed to decode response"
		return nil, metrics, fmt.Errorf("failed to decode token response: %w", err)
	}

	metrics.Message = "ok"
	metrics.AthleteID = tokenResp.Athlete.ID

	if c.debug {
		slog.Debug("OAuth token refresh successful",
			"athlete_id", tokenResp.Athlete.ID,
			"expires_in", tokenResp.ExpiresIn,
		)
	}

	return &tokenResp, metrics, nil
}

// GetAthleteClubs fetches all clubs the authenticated athlete belongs to
func (c *Client) GetAthleteClubs(ctx context.Context, accessToken string) ([]Club, *APICallMetrics, error) {
	start := time.Now()
	endpoint := "/athlete/clubs"

	if c.debug {
		slog.Debug("Fetching athlete clubs",
			"endpoint", endpoint,
		)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch clubs: %w", err)
	}
	defer resp.Body.Close()

	metrics := c.buildMetrics(endpoint, "GET", resp, start)

	if resp.StatusCode != http.StatusOK {
		apiErr := c.handleErrorResponse(resp, endpoint)
		metrics.Message = apiErr.Message
		return nil, metrics, apiErr
	}

	var clubs []Club
	if err := json.NewDecoder(resp.Body).Decode(&clubs); err != nil {
		metrics.Message = "failed to decode response"
		return nil, metrics, fmt.Errorf("failed to decode clubs response: %w", err)
	}

	metrics.Message = "ok"
	metrics.ClubsCount = len(clubs)

	if c.debug {
		slog.Debug("Fetched athlete clubs",
			"endpoint", endpoint,
			"clubs_count", len(clubs),
			"read_limit_usage", metrics.ReadLimit15minUsage,
			"read_limit_limit", metrics.ReadLimit15minLimit,
		)
	}

	return clubs, metrics, nil
}

// GetClubDetails fetches detailed information about a specific club
// This is the KEY method for admin detection - returns admin and owner flags
func (c *Client) GetClubDetails(ctx context.Context, accessToken string, clubID int64) (*ClubDetail, *APICallMetrics, error) {
	start := time.Now()
	endpoint := fmt.Sprintf("/clubs/%d", clubID)

	if c.debug {
		slog.Debug("Fetching club details",
			"endpoint", endpoint,
			"club_id", clubID,
		)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch club details: %w", err)
	}
	defer resp.Body.Close()

	metrics := c.buildMetrics(endpoint, "GET", resp, start)

	if resp.StatusCode != http.StatusOK {
		apiErr := c.handleErrorResponse(resp, endpoint)
		metrics.Message = apiErr.Message
		return nil, metrics, apiErr
	}

	var clubDetail ClubDetail
	if err := json.NewDecoder(resp.Body).Decode(&clubDetail); err != nil {
		metrics.Message = "failed to decode response"
		return nil, metrics, fmt.Errorf("failed to decode club details response: %w", err)
	}

	metrics.Message = "ok"

	if c.debug {
		slog.Debug("Fetched club details",
			"endpoint", endpoint,
			"club_id", clubID,
			"club_name", clubDetail.Name,
			"admin", clubDetail.Admin,
			"owner", clubDetail.Owner,
		)
	}

	return &clubDetail, metrics, nil
}

// GetClubEvents fetches group events for a specific club
// NOTE: This uses an UNDOCUMENTED endpoint discovered through testing
func (c *Client) GetClubEvents(ctx context.Context, accessToken string, clubID int64) ([]GroupEvent, *APICallMetrics, error) {
	start := time.Now()
	endpoint := fmt.Sprintf("/clubs/%d/group_events", clubID)

	if c.debug {
		slog.Debug("Fetching club events",
			"endpoint", endpoint,
			"club_id", clubID,
		)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch club events: %w", err)
	}
	defer resp.Body.Close()

	metrics := c.buildMetrics(endpoint, "GET", resp, start)

	if resp.StatusCode == http.StatusNotFound {
		// Special handling for undocumented endpoint
		// This could mean the endpoint was removed or the club has no events
		metrics.Message = "ALERT: /group_events endpoint returned 404 - may have been removed by Strava or club has no events"
		if c.debug {
			slog.Warn("Club events endpoint returned 404",
				"endpoint", endpoint,
				"club_id", clubID,
			)
		}
		// Return empty array, not an error - club might just have no events
		return []GroupEvent{}, metrics, nil
	}

	if resp.StatusCode != http.StatusOK {
		apiErr := c.handleErrorResponse(resp, endpoint)
		metrics.Message = apiErr.Message
		return nil, metrics, apiErr
	}

	var events []GroupEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		metrics.Message = "failed to decode response"
		return nil, metrics, fmt.Errorf("failed to decode events response: %w", err)
	}

	metrics.Message = "ok"
	metrics.EventsCount = len(events)

	if c.debug {
		slog.Debug("Fetched club events",
			"endpoint", endpoint,
			"club_id", clubID,
			"events_count", len(events),
		)
	}

	return events, metrics, nil
}

// GetRoute fetches route metadata (NOT the GPX file)
// For GPX export, use the routes.RouteFetcher
func (c *Client) GetRoute(ctx context.Context, accessToken string, routeID int64) (*Route, *APICallMetrics, error) {
	start := time.Now()
	endpoint := fmt.Sprintf("/routes/%d", routeID)

	if c.debug {
		slog.Debug("Fetching route metadata",
			"endpoint", endpoint,
			"route_id", routeID,
		)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch route: %w", err)
	}
	defer resp.Body.Close()

	metrics := c.buildMetrics(endpoint, "GET", resp, start)

	if resp.StatusCode != http.StatusOK {
		apiErr := c.handleErrorResponse(resp, endpoint)
		metrics.Message = apiErr.Message
		return nil, metrics, apiErr
	}

	var route Route
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		metrics.Message = "failed to decode response"
		return nil, metrics, fmt.Errorf("failed to decode route response: %w", err)
	}

	metrics.Message = "ok"

	if c.debug {
		slog.Debug("Fetched route metadata",
			"endpoint", endpoint,
			"route_id", routeID,
			"route_name", route.Name,
			"distance_m", route.Distance,
		)
	}

	return &route, metrics, nil
}

// buildMetrics creates APICallMetrics from an HTTP response
func (c *Client) buildMetrics(endpoint, method string, resp *http.Response, start time.Time) *APICallMetrics {
	metrics := &APICallMetrics{
		Endpoint:       endpoint,
		Method:         method,
		StatusCode:     resp.StatusCode,
		ResponseTimeMs: int(time.Since(start).Milliseconds()),
	}

	// Parse general rate limit headers
	// IMPORTANT: Headers use "X-Ratelimit-*" (lowercase 'l'), NOT "X-RateLimit-*"
	if limitHeader := resp.Header.Get("X-Ratelimit-Limit"); limitHeader != "" {
		metrics.RateLimit15minLimit, metrics.RateLimitDailyLimit = parseRateLimitHeader(limitHeader)
	}
	if usageHeader := resp.Header.Get("X-Ratelimit-Usage"); usageHeader != "" {
		metrics.RateLimit15minUsage, metrics.RateLimitDailyUsage = parseRateLimitHeader(usageHeader)
	}

	// Parse read-only rate limit headers (the ACTUAL constraint for GET requests!)
	if readLimitHeader := resp.Header.Get("X-Readratelimit-Limit"); readLimitHeader != "" {
		metrics.ReadLimit15minLimit, metrics.ReadLimitDailyLimit = parseRateLimitHeader(readLimitHeader)
	}
	if readUsageHeader := resp.Header.Get("X-Readratelimit-Usage"); readUsageHeader != "" {
		metrics.ReadLimit15minUsage, metrics.ReadLimitDailyUsage = parseRateLimitHeader(readUsageHeader)
	}

	if c.debug {
		slog.Debug("Rate limit info",
			"endpoint", endpoint,
			"general_limit", resp.Header.Get("X-Ratelimit-Limit"),
			"general_usage", resp.Header.Get("X-Ratelimit-Usage"),
			"read_limit", resp.Header.Get("X-Readratelimit-Limit"),
			"read_usage", resp.Header.Get("X-Readratelimit-Usage"),
		)
	}

	return metrics
}

// parseRateLimitHeader parses a rate limit header like "200,2000" into (200, 2000)
func parseRateLimitHeader(header string) (int, int) {
	parts := strings.Split(header, ",")
	if len(parts) != 2 {
		return 0, 0
	}
	first, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	second, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	return first, second
}

// handleErrorResponse processes an error response from the Strava API
func (c *Client) handleErrorResponse(resp *http.Response, endpoint string) *APIError {
	body, _ := io.ReadAll(resp.Body)

	var message string
	var stravaErr struct {
		Message string `json:"message"`
		Errors  []struct {
			Resource string `json:"resource"`
			Field    string `json:"field"`
			Code     string `json:"code"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &stravaErr); err == nil && stravaErr.Message != "" {
		message = stravaErr.Message
		if len(stravaErr.Errors) > 0 {
			message = fmt.Sprintf("%s: %s %s (%s)",
				stravaErr.Message,
				stravaErr.Errors[0].Resource,
				stravaErr.Errors[0].Field,
				stravaErr.Errors[0].Code,
			)
		}
	} else {
		message = string(body)
	}

	if c.debug {
		slog.Debug("Strava API error",
			"endpoint", endpoint,
			"status_code", resp.StatusCode,
			"message", message,
		)
	}

	var baseErr error
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		baseErr = ErrUnauthorized
		if message == "" {
			message = "Unauthorized or expired token"
		}
	case http.StatusForbidden:
		baseErr = ErrForbidden
		if message == "" {
			message = "Access forbidden"
		}
	case http.StatusNotFound:
		baseErr = ErrNotFound
		if message == "" {
			message = "Resource not found"
		}
	case http.StatusTooManyRequests:
		baseErr = ErrRateLimitExceeded
		if message == "" {
			message = "Rate limit exceeded"
		}
	default:
		if resp.StatusCode >= 500 {
			baseErr = ErrServerError
			if message == "" {
				message = "Server error"
			}
		} else {
			baseErr = ErrInvalidResponse
			if message == "" {
				message = fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
			}
		}
	}

	return NewAPIError(resp.StatusCode, message, endpoint, baseErr)
}

// SetBaseURL allows overriding the base URL (for testing)
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}
