package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"time"

	chi "github.com/go-chi/chi/v5"
	chimi "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spacesedan/cyclescene/functions/internal/api/auth"
	"github.com/spacesedan/cyclescene/functions/internal/api/events"
	"github.com/spacesedan/cyclescene/functions/internal/api/group"
	"github.com/spacesedan/cyclescene/functions/internal/api/magiclink"
	apimi "github.com/spacesedan/cyclescene/functions/internal/api/middleware"
	"github.com/spacesedan/cyclescene/functions/internal/api/ride"
	"github.com/spacesedan/cyclescene/functions/internal/api/storage"
)

var allowedDomains = []string{
	"https://cyclescene.cc",
	"https://www.cyclescene.cc",
	"https://form.cyclescene.cc",
	"https://pdx.cyclescene.cc",
	"https://slc.cyclescene.cc",
	"http://localhost:5173",
	"http://localhost:5174",
	"http://localhost:5175",
}

func isAllowedOrigin(_ *http.Request, origin string) bool {
	return slices.Contains(allowedDomains, origin)
}

func NewRideAPIRouter(db *sql.DB) http.Handler {
	var corsOptions cors.Options

	if os.Getenv("APP_ENV") == "dev" {
		slog.Info("loading cors with dev options")
		corsOptions = cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodPatch, http.MethodOptions},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-BFF-Token", "X-Admin-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}
	} else {
		corsOptions = cors.Options{
			AllowOriginFunc:  isAllowedOrigin,
			AllowedMethods:   []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodPatch, http.MethodOptions},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-BFF-Token", "X-Admin-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}
	}

	r := chi.NewMux()
	r.Use(chimi.Logger)
	r.Use(chimi.Recoverer)
	r.Use(cors.Handler(corsOptions))

	// Health check endpoint for load balancers and monitoring
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"healthy","version":"1.0.0"}`)); err != nil {
			slog.Error("Failed to write health check response", "error", err)
		}
	})

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)

	// Initialize Resend client for email sending
	resendAPIKey := os.Getenv("RESEND_API_KEY")
	if resendAPIKey == "" {
		slog.Warn("RESEND_API_KEY not configured - magic link emails will not be sent")
	}

	// Magic link service (email sending)
	var magicLinkService *magiclink.Service
	if resendAPIKey != "" {
		magicLinkService = magiclink.NewService(resendAPIKey)
	}

	// Create Eventarc client for triggering image optimization
	eventarcClient := events.NewEventarcClient()

	rideRepo := ride.NewRepository(db)
	// Configure ride service with magic link support if available
	var rideService *ride.Service
	if magicLinkService != nil {
		// Get the edit link base URL from environment or use default
		editLinkBaseURL := os.Getenv("EDIT_LINK_BASE_URL")
		if editLinkBaseURL == "" {
			editLinkBaseURL = "https://form.cyclescene.cc/rides/edit" // Default for production
		}
		rideService = ride.NewServiceWithMagicLink(rideRepo, magicLinkService, editLinkBaseURL)
	} else {
		rideService = ride.NewService(rideRepo)
	}
	rideHandler := ride.NewHandler(rideService, authService, eventarcClient)

	groupRepo := group.NewRepository(db)
	groupService := group.NewService(groupRepo)
	groupHandler := group.NewHandler(groupService, eventarcClient)

	// Storage handler for signed URLs (image uploads)
	storageService, err := storage.NewService()
	if err != nil {
		slog.Error("Failed to initialize storage service", "error", err)
		// Don't fail startup, but log the error
	}
	storageHandler := storage.NewHandler(storageService)

	// Rate limiter: 10 submissions per minute per IP
	submissionRateLimiter := apimi.NewRateLimiter(10, time.Minute)

	r.Route("/v1", func(r chi.Router) {
		// auth handlers -- /tokens
		authHandler.RegisterRoutes(r)

		// storage handlers -- /storage (signed URLs for uploads)
		storageHandler.RegisterRoutes(r)

		// ride handlers scraped and user submitted -- /rides
		r.Route("/rides", func(r chi.Router) {
			// Apply rate limiting only to submission endpoint
			r.Post("/submit", submissionRateLimiter.Middleware(http.HandlerFunc(rideHandler.SubmitRide)).ServeHTTP)
			// Register other routes without rate limiting
			r.Get("/edit/{token}", rideHandler.GetRideByEditToken)
			r.Put("/edit/{token}", rideHandler.UpdateRide)
			r.Patch("/edit/{token}/occurrences/{occurrenceId}", rideHandler.UpdateOccurrence)
			r.Get("/admin/pending", rideHandler.GetPendingRides)
			r.Patch("/admin/{id}/publish", rideHandler.PublishRide)
			r.Get("/upcoming", rideHandler.GetUpcomingRides)
			r.Get("/past", rideHandler.GetPastRides)
			r.Get("/ics", rideHandler.GenerateICS)
		})

		// group handlers
		groupHandler.RegisterRoutes(r)
	})

	return r
}
