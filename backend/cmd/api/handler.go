package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	chi "github.com/go-chi/chi/v5"
	chimi "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spacesedan/cyclescene/backend/internal/api/auth"
	"github.com/spacesedan/cyclescene/backend/internal/api/clienterror"
	"github.com/spacesedan/cyclescene/backend/internal/api/events"
	"github.com/spacesedan/cyclescene/backend/internal/api/group"
	"github.com/spacesedan/cyclescene/backend/internal/api/magiclink"
	apimi "github.com/spacesedan/cyclescene/backend/internal/api/middleware"
	"github.com/spacesedan/cyclescene/backend/internal/api/ride"
	routesapi "github.com/spacesedan/cyclescene/backend/internal/api/routes"
	"github.com/spacesedan/cyclescene/backend/internal/api/storage"
	"github.com/spacesedan/cyclescene/backend/internal/api/synclog"
	"github.com/spacesedan/cyclescene/backend/internal/config"
)

func NewRideAPIRouter(db *sql.DB, monitoringDB *sql.DB) http.Handler {
	slog.Info("Initializing CycleScene API router!")

	// Load CORS config from centralized environment config
	corsOptions := config.CORSConfig()
	slog.Info("Loading CORS config",
		"env", config.GetEnvironment(),
		"origins", config.GetAllowedOrigins(),
	)

	r := chi.NewMux()
	r.Use(chimi.Logger)
	r.Use(chimi.Recoverer)
	r.Use(cors.Handler(corsOptions))

	// Health check endpoint for load balancers and monitoring
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"healthy","version":"1.0.2"}`)); err != nil {
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
	// Configure group service with magic link support if available
	var groupService *group.Service
	if magicLinkService != nil {
		// Get the group edit link base URL from environment or use default
		groupEditLinkBaseURL := os.Getenv("GROUP_EDIT_LINK_BASE_URL")
		if groupEditLinkBaseURL == "" {
			groupEditLinkBaseURL = "https://form.cyclescene.cc/group/edit" // Default for production
		}
		groupService = group.NewServiceWithMagicLink(groupRepo, magicLinkService, groupEditLinkBaseURL)
	} else {
		groupService = group.NewService(groupRepo)
	}
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

	// Routes handler
	routesRepo := routesapi.NewRepository(db)
	routesHandler := routesapi.NewHandler(routesRepo)

	// Sync logs handler (uses monitoring DB for analytics isolation)
	synclogRepo := synclog.NewRepository(monitoringDB)
	synclogHandler := synclog.NewHandler(synclogRepo)

	// Client errors handler (uses monitoring DB for analytics isolation)
	clientErrorRepo := clienterror.NewRepository(monitoringDB)
	clientErrorHandler := clienterror.NewHandler(clientErrorRepo)

	r.Route("/v1", func(r chi.Router) {
		// auth handlers -- /tokens
		authHandler.RegisterRoutes(r)

		// storage handlers -- /storage (signed URLs for uploads)
		storageHandler.RegisterRoutes(r)

		// routes handlers -- /routes
		routesHandler.RegisterRoutes(r)

		// sync logs handlers -- /sync-logs
		synclogHandler.RegisterRoutes(r)

		r.Post("/admin/sync/shift2bikes", rideHandler.TriggerShift2BikesSync)

		// client error handler -- POST /client-errors only
		r.Post("/client-errors", clientErrorHandler.LogError)

		// ride handlers scraped and user submitted -- /rides
		r.Route("/rides", func(r chi.Router) {
			// Apply rate limiting only to submission endpoint
			r.Post("/submit", submissionRateLimiter.Middleware(http.HandlerFunc(rideHandler.SubmitRide)).ServeHTTP)
			// Register other routes without rate limiting
			r.Get("/edit/{token}", rideHandler.GetRideByEditToken)
			r.Put("/edit/{token}", rideHandler.UpdateRide)
			r.Patch("/edit/{token}/details", rideHandler.UpdateEventDetails)
			r.Patch("/edit/{token}/occurrences/{occurrenceId}", rideHandler.UpdateOccurrence)
			r.Get("/admin/pending", rideHandler.GetPendingRides)
			r.Patch("/admin/{id}/publish", rideHandler.PublishRide)
			r.Get("/search", rideHandler.SearchRides)
			r.Get("/upcoming", rideHandler.GetUpcomingRides)
			r.Get("/past", rideHandler.GetPastRides)
			r.Get("/ics", rideHandler.GenerateICS)
		})

		// group handlers
		groupHandler.RegisterRoutes(r)
	})

	return r
}
