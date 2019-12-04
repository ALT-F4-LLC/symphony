package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Error : struct for handling errors
type Error struct {
	Error string
}

// ServiceConfig : service configuration
type ServiceConfig struct {
	Postgres struct {
		DBName   string
		Host     string
		Password string
		Port     int
		User     string
	}
}

// Services : struct for multiple services
type Services []Service

// ServiceTypes : struct for multiple servicetypes
type ServiceTypes []ServiceType

// HandleError : handles response
func HandleError(error string) []byte {
	err := &Error{
		Error: error,
	}
	json, e := json.Marshal(err)
	if e != nil {
		panic(e)
	}
	return json
}

// GetServiceByID : gets specific service from database
func GetServiceByID(db *sql.DB, id string) (Services, error) {
	var services []Service
	rows, queryErr := db.Query("SELECT hostname, id FROM service WHERE id = $1", id)
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()
	for rows.Next() {
		var service Service
		if e := rows.Scan(&service.Hostname, &service.ID); e != nil {
			return nil, e
		}
		services = append(services, service)
	}
	return services, nil
}

// GetServiceByHostnameHandler : handles retrival of service from API request
func GetServiceByHostnameHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostname := r.FormValue("hostname")
		services, err := GetServiceByHostname(db, hostname)
		if err != nil {
			panic(err)
		}
		json, err := json.Marshal(services)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

// GetServiceByIDHandler : handles retrival of service from API request
func GetServiceByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]

		services, err := GetServiceByID(db, id)
		if len(services) < 1 {
			json, err := json.Marshal(services)
			if err != nil {
				panic(err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(json)
			return
		}

		service := &Service{
			ID:       services[0].ID,
			Hostname: services[0].Hostname,
		}

		json, err := json.Marshal(service)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

// GetServiceTypeByName : gets specific service from database
func GetServiceTypeByName(db *sql.DB, name string) (ServiceTypes, error) {
	servicetypes := make(ServiceTypes, 0)
	rows, queryErr := db.Query("SELECT id, name FROM servicetype WHERE name = $1", name)
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()
	for rows.Next() {
		var servicetype ServiceType
		if e := rows.Scan(&servicetype.ID, &servicetype.Name); e != nil {
			return nil, e
		}
		servicetypes = append(servicetypes, servicetype)
	}
	return servicetypes, nil
}

// PostServiceHandler : handles creation of service
func PostServiceHandler(db *sql.DB) http.HandlerFunc {
	type requestBody struct {
		Hostname      string    `json:"hostname"`
		ServiceTypeID uuid.UUID `json:"servicetype_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// JSON data from service
		var data requestBody
		_ = json.NewDecoder(r.Body).Decode(&data)

		// Generate a new service in database
		var service Service
		id := uuid.New()
		sql := "INSERT INTO service(id, hostname, servicetype_id) VALUES ($1, $2, $3) RETURNING hostname, id, servicetype_id"
		err := db.QueryRow(sql, id, data.Hostname, data.ServiceTypeID).Scan(&service.Hostname, &service.ID, &service.ServiceTypeID)
		if err != nil {
			panic(err)
		}
		json, err := json.Marshal(service)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

// PostGuestHandler : Handles creation of guests on compute nodes
func PostGuestHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Create a block volume
		// TODO: Inject the OS image into the volume
		// TODO: Expose the volume over ISCSI
		// TODO: Mount the ISCSI target on the compute node
		// TODO: Create the virtual machine
		// TODO: Cloudinit configuration stuff
	}
}

// GetServiceTypeByNameHandler : handles retrival of servicetype from API request
func GetServiceTypeByNameHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		servicetypes, err := GetServiceTypeByName(db, name)
		if err != nil {
			panic(err)
		}
		json, err := json.Marshal(servicetypes)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}
