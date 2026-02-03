// Package config provides centralized environment-specific configuration
// for CORS, cookies, and other settings that differ between dev and prod.
package config

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/cors"
)

// Environment represents the application environment
type Environment string

const (
	EnvDev  Environment = "dev"
	EnvProd Environment = "prod"
)

// GetEnvironment returns the current environment based on APP_ENV
func GetEnvironment() Environment {
	if os.Getenv("APP_ENV") == "dev" {
		return EnvDev
	}
	return EnvProd
}

// IsDev returns true if running in development mode
func IsDev() bool {
	return GetEnvironment() == EnvDev
}

// IsProd returns true if running in production mode
func IsProd() bool {
	return GetEnvironment() == EnvProd
}

// =============================================================================
// CORS Configuration
// =============================================================================

// DevOrigins are the allowed origins for development
var DevOrigins = []string{
	"http://localhost:5173",
	"http://localhost:5174",
	"http://localhost:5175",
	"http://localhost:5176",
	"http://localhost:5177",
	"http://localhost:5178",
	"http://localhost:5179",
	"http://localhost:5180",
	"https://dev.nadhatter.com",
}

// ProdOrigins are the allowed origins for production
var ProdOrigins = []string{
	"https://cyclescene.cc",
	"https://form.cyclescene.cc",
	"https://pdx.cyclescene.cc",
	"https://slc.cyclescene.cc",
	"https://dashboard.cyclescene.cc",
}

// CORSConfig returns the CORS options for the current environment
func CORSConfig() cors.Options {
	var origins []string
	var allowOriginFunc func(r *http.Request, origin string) bool

	if IsDev() {
		origins = DevOrigins
		// In dev mode, also allow any *.nadhatter.com subdomain
		allowOriginFunc = func(r *http.Request, origin string) bool {
			// Check if it's in the explicit list
			for _, allowed := range DevOrigins {
				if origin == allowed {
					return true
				}
			}
			// Allow any *.nadhatter.com subdomain in dev
			if strings.HasSuffix(origin, ".nadhatter.com") {
				return true
			}
			return false
		}
	} else {
		origins = ProdOrigins
	}

	return cors.Options{
		AllowedOrigins:   origins,
		AllowOriginFunc:  allowOriginFunc,
		AllowedMethods:   []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodPatch, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-BFF-Token", "X-Admin-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}
}

// GetAllowedOrigins returns the allowed origins for the current environment
func GetAllowedOrigins() []string {
	if IsDev() {
		return DevOrigins
	}
	return ProdOrigins
}

// =============================================================================
// Cookie Configuration
// =============================================================================

// CookieConfig holds cookie settings for the current environment
type CookieConfig struct {
	HttpOnly bool
	Secure   bool
	SameSite http.SameSite
}

// GetCookieConfig returns cookie settings for the current environment
func GetCookieConfig() CookieConfig {
	if IsDev() {
		return CookieConfig{
			HttpOnly: false, // Allow JS to read for WebSocket auth in dev
			Secure:   false, // No HTTPS in local dev
			SameSite: http.SameSiteLaxMode,
		}
	}
	return CookieConfig{
		HttpOnly: true,  // Protect from XSS
		Secure:   true,  // HTTPS only
		SameSite: http.SameSiteLaxMode,
	}
}

// NewSessionCookie creates a new session cookie with environment-appropriate settings
func NewSessionCookie(name, value string, maxAge int) *http.Cookie {
	cfg := GetCookieConfig()
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: cfg.HttpOnly,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	}
}

// ClearSessionCookie creates a cookie that clears the session
// (MaxAge: -1 tells the browser to delete it)
func ClearSessionCookie(name string) *http.Cookie {
	return NewSessionCookie(name, "", -1)
}
