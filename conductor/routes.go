package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
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
	err := Error{
		Error: error,
	}
	json, e := json.Marshal(err)
	if e != nil {
		panic(e)
	}
	return json
}

// GetServiceByID : gets specific service from database
// func GetServiceByID(db *sql.DB, id string) (Services, error) {
// 	var services []Service
// 	rows, queryErr := db.Query("SELECT hostname, id FROM service WHERE id = $1", id)
// 	if queryErr != nil {
// 		return nil, queryErr
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var service Service
// 		if e := rows.Scan(&service.Hostname, &service.ID); e != nil {
// 			return nil, e
// 		}
// 		services = append(services, service)
// 	}
// 	return services, nil
// }

// GetServiceByIDHandler : handles retrival of service from API request
// func GetServiceByIDHandler(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		params := mux.Vars(r)
// 		id := params["id"]

// 		services, err := GetServiceByID(db, id)
// 		if len(services) < 1 {
// 			json, err := json.Marshal(services)
// 			if err != nil {
// 				panic(err)
// 			}
// 			w.Header().Set("Content-Type", "application/json")
// 			w.WriteHeader(http.StatusOK)
// 			w.Write(json)
// 			return
// 		}

// 		service := &Service{
// 			ID:       services[0].ID,
// 			Hostname: services[0].Hostname,
// 		}

// 		json, err := json.Marshal(service)
// 		if err != nil {
// 			panic(err)
// 		}

// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(json)
// 	}
// }

// PostGuestHandler : Handles creation of guests on compute nodes
// func PostGuestHandler(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// TODO: Create a block volume
// 		// TODO: Inject the OS image into the volume
// 		// TODO: Expose the volume over ISCSI
// 		// TODO: Mount the ISCSI target on the compute node
// 		// TODO: Create the virtual machine
// 		// TODO: Cloudinit configuration stuff
// 	}
// }

// GetServiceByHostnameHandler : handles retrival of service from API request
func GetServiceByHostnameHandler(db *gorm.DB) http.HandlerFunc {
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

// PostServiceHandler : handles creation of service
func PostServiceHandler(db *gorm.DB) http.HandlerFunc {
	type requestBody struct {
		Hostname      string    `json:"hostname"`
		ServiceTypeID uuid.UUID `json:"service_type_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var serviceType *ServiceType
		db.Where(&ServiceType{Name: "block"}).First(&serviceType)
		if serviceType == nil {
			json := HandleError("invalid_service_type")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write(json)
			return
		}

		var data requestBody
		_ = json.NewDecoder(r.Body).Decode(&data)

		service := Service{Hostname: data.Hostname, ServiceTypeID: serviceType.ID}
		db.Create(&service)

		json, err := json.Marshal(service)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

// GetServiceTypeByNameHandler : handles retrival of servicetype from API request
func GetServiceTypeByNameHandler(db *gorm.DB) http.HandlerFunc {
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
