package main

import (
	_ "github.com/lib/pq"
)

const (
	port = ":36837"
)

func main() {
	// Get command line flags
	flags := GetFlags()

	// Load database here
	db, err := LoadDB(flags)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Endpoints that handle cluster initialization
	// r := mux.NewRouter()

	// Resources: service
	// r.Path("/service").Queries("hostname", "{hostname}").HandlerFunc(GetServiceByHostnameHandler(db)).Methods("GET")
	// r.Path("/service").HandlerFunc(PostServiceHandler(db)).Methods("POST")
	// r.Path("/service/{id}").HandlerFunc(GetServiceByIDHandler(db)).Methods("GET")

	// Resources: servicetype
	// r.Path("/servicetype").Queries("name", "{name}").HandlerFunc(GetServiceTypeByNameHandler(db)).Methods("GET")

	// Log successful listen
	// log.Printf("Started conductor listening on 0.0.0.0" + port)

	// Logs the error if ListenAndServe fails.
	// log.Fatal(http.ListenAndServe(port, r))
}
