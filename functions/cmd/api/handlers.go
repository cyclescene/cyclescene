package main

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRideAPIRouter(db *sql.DB) http.Handler {
	r := chi.NewMux()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/v1/rides", func(r chi.Router) {
		r.Get("/upcoming", MakeRidesHandler(db, getUpcomingRides))
		r.Get("/past", MakeRidesHandler(db, getPastRides))

	})

	return r
}

type rideFetcher func(db *sql.DB, cityCode string) ([]RideFromDB, error)

func MakeRidesHandler(db *sql.DB, fetcher rideFetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cityCode := r.URL.Query().Get("city")
		if cityCode == "" {
			cityCode = "pdx"
		}
		storedRides, err := fetcher(db, cityCode)
		if err != nil {
			slog.Error("Database query failed", "error", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		slog.Info("Stored Rides Get!", "num of rides", len(storedRides))

		var rides []Ride
		for i := range storedRides {
			rdb := storedRides[i]
			rides = append(rides, rdb.ToRide())
		}

		if rides == nil {
			rides = []Ride{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(rides); err != nil {
			slog.Error("Failed to encode rides to JSON", "error", err.Error())

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
