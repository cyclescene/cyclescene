package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"

	// "github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var apiHandler http.Handler
var db *sql.DB

func init() {
	var err error
	// Used while in development
	// err = godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("failed to read environment variables: %v", err)
	// }
	db, err = ConnectToDB()
	if err != nil {
		log.Fatalf("unable to connect to TursoDB: %v", err)
	}
	apiHandler = NewRideAPIRouter(db)

}

func main() {

	slog.Info("API Gateway started", "listening_on", ":8080")
	err := http.ListenAndServe(":8080", apiHandler)
	if err != nil {
		slog.Error("unable to start server", "error", err)
		log.Fatalf("FATAL: unable to start server: %v", err)
	}
}
