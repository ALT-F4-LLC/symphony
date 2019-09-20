package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	port = ":22878"
)

func main() {
	connStr := "user=postgres dbname=postgres password=b774e74299739cc265a03bc9eecf12b241d3e5c1a2011b9a9d sslmode=disable"
	db, connErr := sql.Open("postgres", connStr)
	if connErr != nil {
		log.Fatal(connErr)
	}

	r := mux.NewRouter()
	r.HandleFunc("/image", GetImageHandler(db)).Methods("GET")
	r.HandleFunc("/image", PostImageHandler(db)).Methods("POST")
	r.HandleFunc("/image/{id}", GetImageByIDHandler(db)).Methods("GET")

	// Log successful listen
	log.Printf("Started image service")

	// Logs the error if ListenAndServe fails.
	log.Fatal(http.ListenAndServe(port, r))
}
