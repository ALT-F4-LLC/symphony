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

	// Handle service discovery
	service, serviceErr := services.GetService(flags.Conductor, flags.Hostname, "block")
	if serviceErr != nil {
		panic(serviceErr)
	}

	r := mux.NewRouter()
	r.Path("/physicalvolume").Queries("device", "{device}").HandlerFunc(GetPhysicalVolumeByDeviceHandler()).Methods("GET")
	r.Path("/physicalvolume").HandlerFunc(PostPhysicalVolumeHandler()).Methods("POST")
	r.Path("/physicalvolume").HandlerFunc(DeletePhysicalVolumeHandler()).Methods("DELETE")

	// Log successful listen
	log.Printf("Started block \"%s\" on 0.0.0.0%s", service.ID, port)

	// Logs the error if ListenAndServe fails.
	log.Fatal(http.ListenAndServe(port, r))
}
