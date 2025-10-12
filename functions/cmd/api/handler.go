package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"slices"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spacesedan/cyclescene/functions/internal/api/auth"
	"github.com/spacesedan/cyclescene/functions/internal/api/group"
	"github.com/spacesedan/cyclescene/functions/internal/api/ride"
)

var allowedDomains = []string{
	"https://cyclescene.cc",
	"https://www.cyclescene.cc",
	"https://form.cyclescene.cc",
	"https://pdx.cyclescene.cc",
	"https://slc.cyclescene.cc",
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
			AllowedMethods:   []string{http.MethodGet, http.MethodPut, http.MethodPost},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-BFF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}
	} else {
		corsOptions = cors.Options{
			AllowOriginFunc:  isAllowedOrigin,
			AllowedMethods:   []string{http.MethodGet, http.MethodPut, http.MethodPost},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-BFF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}
	}

	r := chi.NewMux()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(corsOptions))

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)

	rideRepo := ride.NewRepository(db)
	rideService := ride.NewService(rideRepo)
	rideHandler := ride.NewHandler(rideService)

	groupRepo := group.NewRepository(db)
	groupService := group.NewService(groupRepo)
	groupHandler := group.NewHandler(groupService)

	r.Route("/v1", func(r chi.Router) {
		// auth handlers -- /tokens
		authHandler.RegisterRoutes(r)

		// ride handlers scraped and user submitted -- /rides
		rideHandler.RegisterRoutes(r)

		// group handlers
		groupHandler.RegisterRoutes(r)
	})

	return r
}
