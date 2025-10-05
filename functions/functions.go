package functions

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/spacesedan/cyclescene/functions/api"
)

var apiHandler http.Handler
var db *sql.DB

func init() {
	var err error
	db, err = api.ConnectToDB()
	if err != nil {
		log.Fatal("unable to connect to TursoDB")
	}
	apiHandler = api.NewRideAPIRouter(db)

	functions.HTTP("RideApi", serveRideApi)
}

func serveRideApi(w http.ResponseWriter, r *http.Request) {
	apiHandler.ServeHTTP(w, r)
}
