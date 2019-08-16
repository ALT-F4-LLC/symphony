package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "user=postgres dbname=postgres password=b774e74299739cc265a03bc9eecf12b241d3e5c1a2011b9a9d sslmode=disable"
	db, connErr := sql.Open("postgres", connStr)
	if connErr != nil {
		log.Fatal(connErr)
	}
	images, err := GetImages(db)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(images)
}
