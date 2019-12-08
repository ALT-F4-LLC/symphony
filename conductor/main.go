package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	port = ":36837"
)

func main() {
	flags := GetFlags()

	db, err := GetDatabase(flags)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	r := mux.NewRouter()

	r.Path("/service").Queries("hostname", "{hostname}").HandlerFunc(GetServiceByHostnameHandler(db)).Methods("GET")
	r.Path("/service").HandlerFunc(PostServiceHandler(db)).Methods("POST")
	r.Path("/service/{id}").HandlerFunc(GetServiceByIDHandler(db)).Methods("GET")

	r.Path("/servicetype").Queries("name", "{name}").HandlerFunc(GetServiceTypeByNameHandler(db)).Methods("GET")

	r.Path("/physicalvolume").Queries("device", "{device}").Queries("service_id", "{service_id}").HandlerFunc(GetPhysicalVolumeByDeviceHandler(db)).Methods("GET")
	r.Path("/physicalvolume").HandlerFunc(PostPhysicalVolumeHandler(db)).Methods("POST")
	r.Path("/physicalvolume/{id}").HandlerFunc(DeletePhysicalVolumeHandler(db)).Methods("DELETE")

	log.Printf("Started conductor on 0.0.0.0" + port)

	log.Fatal(http.ListenAndServe(port, r))
}
