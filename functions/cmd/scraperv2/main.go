package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/spacesedan/shift2bikes-scraper/lib/scraper-helpers"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	// Load Envs (dev)
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Check for ENV variables
	// DB Vars
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_DB_RW_TOKEN") == "" {
		log.Fatal("FATAL: Turso env variable not set properly")
	}

	// // GOOGLE Vars
	// if os.Getenv("GOOGLE_GEOCODING_API_KEY") == "" {
	// 	log.Fatal("FATAL: GOOGLE_GEOCODING_API_KEY not properly set")
	// }

	// set up logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})))

	// connect to DB(Turso)
	dbURL := os.Getenv("TURSO_DB_URL")
	authToken := os.Getenv("TURSO_DB_RW_TOKEN")

	fullURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	db, err := sql.Open("libsql", fullURL)
	if err != nil {
		log.Fatalf("failed to open Turso DB connection: %v", err)
	}
	defer db.Close()

	////// READY TO START /////////////////////////

	// get upcoming Rides
	shift2BikesEvents, err := scraperhelpers.GetRides()
	if err != nil {
		fmt.Println("POOP")
	}

	for _, event := range shift2BikesEvents.Events {
		fmt.Printf("Address: %s Venue: %s Location Details: %s\n", event.Address, event.Venue, event.Locdetails)
	}
}
