package main

import (
	"net/http"

	"github.com/erkrnt/symphony/services"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	port = ":50051"
)

func main() {
	flags := GetFlags()
	if flags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	service, serviceErr := services.GetService(flags.Conductor, flags.Hostname, "block")
	if serviceErr != nil {
		panic(serviceErr)
	}

	r := mux.NewRouter()

	r.Path("/physicalvolume").Queries("device", "{device}").Handler(services.RegisterHandler(GetPhysicalVolumeByDeviceHandler())).Methods("GET")
	r.Path("/physicalvolume").Handler(services.RegisterHandler(PostPhysicalVolumeHandler())).Methods("POST")
	r.Path("/physicalvolume").Handler(services.RegisterHandler(DeletePhysicalVolumeHandler())).Methods("DELETE")

	r.Path("/volumegroup").Handler(services.RegisterHandler(PostVolumeGroupHandler())).Methods("POST")
	r.Path("/volumegroup/{id}").Handler(services.RegisterHandler(GetVolumeGroupByIDHandler())).Methods("GET")
	r.Path("/volumegroup/{id}").Handler(services.RegisterHandler(DeleteVolumeGroupHandler())).Methods("DELETE")

	logrus.WithFields(logrus.Fields{"port": port, "service_id": service.ID}).Info("Started block service.")

	logrus.Fatal(http.ListenAndServe(port, r))
}
