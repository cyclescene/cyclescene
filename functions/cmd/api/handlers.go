package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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
			AllowedMethods:   []string{http.MethodGet},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}
	} else {
		corsOptions = cors.Options{
			AllowOriginFunc:  isAllowedOrigin,
			AllowedMethods:   []string{http.MethodGet},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}
	}

	r := chi.NewMux()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(corsOptions))

	r.Route("/v1/rides", func(r chi.Router) {
		r.Get("/upcoming", MakeRidesHandler(db, getUpcomingRides))
		r.Get("/past", MakeRidesHandler(db, getPastRides))
		r.Get("/ics", GenerateICSHandler(db, getRide))

	})

	// register the BFFRoutes
	addBFFRoutes(r, db)
	return r
}

type ridesFetcher func(db *sql.DB, cityCode string) ([]RideFromDB, error)

func MakeRidesHandler(db *sql.DB, fetcher ridesFetcher) http.HandlerFunc {
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

type rideFetcher func(db *sql.DB, cityCode string, id string) ([]RideFromDB, error)

func GenerateICSHandler(db *sql.DB, fetcher rideFetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rideId := r.URL.Query().Get("id")
		cityCode := r.URL.Query().Get("city")
		if cityCode == "" {
			cityCode = "pdx"
		}
		storedRides, err := fetcher(db, cityCode, rideId)
		if err != nil {
			slog.Error("Database query failed", "error", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		slog.Info("Stored Rides Get!", "num of rides", len(storedRides))

		ride := storedRides[0].ToRide()

		startTimeStr := fmt.Sprintf("%s %s", ride.Date, ride.StartTime)
		start, err := time.Parse("2006-01-02 15:04:05", startTimeStr)
		if err != nil {
			http.Error(w, "Invalid start time", http.StatusInternalServerError)
			return
		}

		end := start.Add(2 * time.Hour)
		if ride.EndTime != "" {
			if endTime, err := time.Parse("2006-01-02 15:04:05", fmt.Sprintf("%s %s", ride.Date, ride.EndTime)); err == nil {
				end = endTime
			}
		}

		formatICS := func(t time.Time) string {
			return t.UTC().Format("20060102T150405Z")
		}

		// Clean newlines and commas
		desc := strings.ReplaceAll(ride.Details, "\n", "\\n")
		desc = strings.ReplaceAll(desc, ",", "\\,")

		icsContent := strings.Join([]string{
			"BEGIN:VCALENDAR",
			"VERSION:2.0",
			"PRODID:-//CycleScene//EN",
			"BEGIN:VEVENT",
			fmt.Sprintf("UID:%s@cyclescene.com", ride.ID),
			fmt.Sprintf("DTSTAMP:%s", formatICS(time.Now())),
			fmt.Sprintf("DTSTART:%s", formatICS(start)),
			fmt.Sprintf("DTEND:%s", formatICS(end)),
			fmt.Sprintf("SUMMARY:%s", ride.Title),
			fmt.Sprintf("LOCATION:%s", ride.Venue),
			fmt.Sprintf("DESCRIPTION:%s\\nURL:%s", desc, ride.Shareable),
			"STATUS:CONFIRMED",
			"END:VEVENT",
			"END:VCALENDAR",
			"", // ensure trailing CRLF
		}, "\r\n")

		filename := url.QueryEscape(ride.Title)
		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.ics\"", filename))
		w.Header().Set("Cache-Control", "public, max-age=3600")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(icsContent))
	}

}
