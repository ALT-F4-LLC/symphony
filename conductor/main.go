package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	port = ":36837"
)

func main() {
	// Get service configuration from FS
	config, configErr := GetServiceConfig(os.Getenv("CONDUCTOR_CONFIG_FILE"))
	if configErr != nil {
		panic(configErr)
	}

	// Setup connection to Postgres database
	connStr := fmt.Sprintf("dbname=%s host=%s port=%d password=%s sslmode=disable user=%s", config.Postgres.DBName, config.Postgres.Host, config.Postgres.Port, config.Postgres.Password, config.Postgres.User)
	db, connErr := sql.Open("postgres", connStr)
	if connErr != nil {
		panic(connErr)
	}

	// Endpoints that handle cluster initialization
	r := mux.NewRouter()

	// Resources: service
	r.Path("/service").Queries("hostname", "{hostname}").HandlerFunc(GetServiceByHostnameHandler(db)).Methods("GET")
	r.Path("/service").HandlerFunc(PostServiceHandler(db)).Methods("POST")
	r.Path("/service/{id}").HandlerFunc(GetServiceByIDHandler(db)).Methods("GET")

	// Resources: servicetype
	r.Path("/servicetype").Queries("name", "{name}").HandlerFunc(GetServiceTypeByNameHandler(db)).Methods("GET")

	// Log successful listen
	log.Printf("Started conductor listening on 0.0.0.0" + port)

	// Logs the error if ListenAndServe fails.
	log.Fatal(http.ListenAndServe(port, r))
}
