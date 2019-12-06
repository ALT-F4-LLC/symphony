package main

import (
	"log"
	"net/http"

	"github.com/erkrnt/symphony/services"
	"github.com/gorilla/mux"
)

const (
	port = ":50051"
)

func main() {
	// Get command line flags
	flags := GetFlags()

	// Get service configuration from FS
	config, configErr := GetConfig(flags.Config)
	if configErr != nil {
		panic(configErr)
	}

	// Handle service discovery
	service, serviceErr := services.GetService(config.Conductor.Hostname, config.Hostname, "block")
	if serviceErr != nil {
		panic(serviceErr)
	}

	// Handle database initialization
	db, dbErr := GetDatabase(flags)
	if dbErr != nil {
		panic(dbErr)
	}

	r := mux.NewRouter()
	r.Path("/pv").Queries("device", "{device}").HandlerFunc(GetPvByDeviceHandler(db, service)).Methods("GET")

	// Log successful listen
	log.Printf("Started block \"%s\" on 0.0.0.0%s", service.ID, port)

	// Logs the error if ListenAndServe fails.
	log.Fatal(http.ListenAndServe(port, r))
}
