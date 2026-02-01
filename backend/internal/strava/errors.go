package strava

import (
	"errors"
	"fmt"
)

// Sentinel errors for Strava API operations
var (
	ErrUnauthorized      = errors.New("strava: unauthorized or expired token")
	ErrRateLimitExceeded = errors.New("strava: rate limit exceeded")
	ErrNotFound          = errors.New("strava: resource not found")
	ErrForbidden         = errors.New("strava: access forbidden")
	ErrServerError       = errors.New("strava: server error")
	ErrInvalidResponse   = errors.New("strava: invalid response from API")
)

// APIError wraps Strava API errors with details
type APIError struct {
	StatusCode int
	Message    string
	Endpoint   string
	Err        error
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Endpoint != "" {
		return fmt.Sprintf("strava api error %d on %s: %s", e.StatusCode, e.Endpoint, e.Message)
	}
	return fmt.Sprintf("strava api error %d: %s", e.StatusCode, e.Message)
}

// Unwrap returns the underlying error for errors.Is/As support
func (e *APIError) Unwrap() error {
	return e.Err
}

// NewAPIError creates a new APIError with the given details
func NewAPIError(statusCode int, message, endpoint string, err error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Endpoint:   endpoint,
		Err:        err,
	}
}

// IsUnauthorized checks if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsRateLimitExceeded checks if the error is a rate limit error
func IsRateLimitExceeded(err error) bool {
	return errors.Is(err, ErrRateLimitExceeded)
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
