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
		r.Get("/ics", MakeRideHandler(db, getRide))

	})

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

func MakeRideHandler(db *sql.DB, fetcher rideFetcher) http.HandlerFunc {
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

func GenerateICSHandler(db *sql.DB, fetcher rideFetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cityCode string
		rideId := r.URL.Query().Get("id")
		// fmt.Println(rideId)

		if r.URL.Query().Get("city") == "" {
			cityCode = "pdx"
		} else {
			cityCode = r.URL.Query().Get("city")
		}

		fmt.Printf("ID: %s\nCity: %s", rideId, cityCode)

		storedRides, err := getRide(db, rideId, cityCode)
		if err != nil {
			slog.Error("Database query failed", "error", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		slog.Info("Ride Found!", "id", rideId, "cityCode", cityCode)

		ride := storedRides[0].ToRide()

		// --- 1. DATE/TIME PREPARATION (CRITICAL) ---
		// Combine date and time and parse it into a Go time.Time object
		startTimeStr := fmt.Sprintf("%s %s", ride.Date, ride.StartTime)
		start, err := time.Parse("2006-01-02 15:04:05", startTimeStr) // Adjust layout to match your DB time format
		if err != nil {
			slog.Error("Failed to parse ride start time", "time", startTimeStr, "error", err)
			http.Error(w, "Invalid ride time data", http.StatusInternalServerError)
			return
		}

		// Set End Time (Default to 2 hours if missing)
		end := start.Add(2 * time.Hour)
		if ride.EndTime != "" {
			// If end time is present, parse it instead
			endTimeStr := fmt.Sprintf("%s %s", ride.Date, ride.EndTime)
			end, err = time.Parse("2006-01-02 15:04:05", endTimeStr)
			if err != nil {
				// Log error but use the default 2-hour end time
				slog.Warn("Failed to parse ride end time, defaulting to 2 hours", "error", err)
				end = start.Add(2 * time.Hour)
			}
		}

		// Helper function to format Go time.Time to ICS format (YYYYMMDDTHHMMSSZ)
		// ICS requires the time in UTC/Zulu format (the .UTC().Format part is critical)
		formatICS := func(t time.Time) string {
			return t.UTC().Format("20060102T150405Z")
		}

		// --- 2. ICS CONTENT GENERATION (MANUAL STRING BUILD) ---

		// Escape necessary characters for ICS (newline to \n, comma to \,)
		detailsEscaped := strings.ReplaceAll(ride.Details, "\n", "\\n")
		detailsEscaped = strings.ReplaceAll(detailsEscaped, ",", "\\,")

		// Build the ICS file content
		icsContent := fmt.Sprintf(`BEGIN:VCALENDAR
		 		VERSION:2.0
		 		PRODID:-//CycleScene//EN
		 		BEGIN:VEVENT
		 		UID:%s@cyclescene.com
		 		DTSTAMP:%s
		 		DTSTART:%s
		 		DTEND:%s
		 		SUMMARY:%s
		 		LOCATION:%s
		 		DESCRIPTION:%s\nURL:%s
		 		END:VEVENT
		 		END:VCALENDAR`,
			ride.ID,
			formatICS(time.Now()),
			formatICS(start),
			formatICS(end),
			ride.Title,
			ride.Venue,
			detailsEscaped,
			ride.Shareable,
		)

		// --- 3. SERVE THE FILE ---

		// Set the Content-Disposition header to force the browser to download the file
		filename := url.QueryEscape(ride.Title) // URL-encode the filename
		w.Header().Set("Content-Disposition", fmt.Sprintf("filename=\"%s.ics\"", filename))
		// w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(icsContent)))

		// Write the content to the HTTP response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(icsContent))
	}

}
