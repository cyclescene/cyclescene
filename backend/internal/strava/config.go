package strava

import "os"

// Default callback path for OAuth flow
const DefaultCallbackPath = "/v1/strava/auth/callback"

// Config holds the Strava API configuration
type Config struct {
	ClientID     string
	ClientSecret string
	CallbackPath string // Path only, not full URL (e.g., "/v1/strava/auth/callback")
	Debug        bool
}

// LoadConfig loads Strava configuration from environment variables
func LoadConfig() *Config {
	callbackPath := os.Getenv("STRAVA_CALLBACK_PATH")
	if callbackPath == "" {
		callbackPath = DefaultCallbackPath
	}

	return &Config{
		ClientID:     os.Getenv("STRAVA_CLIENT_ID"),
		ClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
		CallbackPath: callbackPath,
		Debug:        os.Getenv("STRAVA_DEBUG") == "true",
	}
}

// IsConfigured returns true if the minimum required config is set
func (c *Config) IsConfigured() bool {
	return c.ClientID != "" && c.ClientSecret != ""
}

// Validate checks if the configuration is valid for production use
func (c *Config) Validate() error {
	if c.ClientID == "" {
		return NewAPIError(0, "STRAVA_CLIENT_ID is required", "", nil)
	}
	if c.ClientSecret == "" {
		return NewAPIError(0, "STRAVA_CLIENT_SECRET is required", "", nil)
	}
	return nil
}
