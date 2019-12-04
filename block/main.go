package main

import (
	"log"
	"net/http"
	"os"

	"github.com/erkrnt/symphony/services"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	port = ":50051"
)

func main() {
	// Get command line flags
	flags := GetFlags()

	// Get service configuration from FS
	config, configErr := GetConfig(os.Getenv("BLOCK_CONFIG_FILE"))
	if configErr != nil {
		panic(configErr)
	}

	// Handle service discovery
	service, err := services.GetService(config)
	if err != nil {
		panic(err)
	}

	// Handle database initialization
	db, err := LoadDB(flags)
	if err != nil {
		panic(err)
	}

	log.Println(db)

	r := mux.NewRouter()
	// r.Path("/pv").Queries("device", "{device}").HandlerFunc(GetPvsByDeviceHandler(db, service)).Methods("GET")

	// Log successful listen
	log.Printf("Started block service with identifier: %s", service.ID)

	// Logs the error if ListenAndServe fails.
	log.Fatal(http.ListenAndServe(port, r))
}
