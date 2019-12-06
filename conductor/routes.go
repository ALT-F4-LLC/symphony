package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// Error : struct for handling errors
type Error struct {
	Error string
}

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

// GetServiceByIDHandler : handles retrival of service from API request
func GetServiceByIDHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		params := mux.Vars(r)
		id := params["id"]

		service, err := GetServiceByID(db, id)
		if err != nil {
			json := HandleError(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(json)
			return
		}

		data := make([]Service, 0)
		if service != nil {
			data = append(data, *service)
		}
		json, err := json.Marshal(data)
		if err != nil {
			json := HandleError(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(json)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

// GetServiceByHostnameHandler : handles retrival of service from API request
func GetServiceByHostnameHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		hostname := r.FormValue("hostname")
		service, err := GetServiceByHostname(db, hostname)
		if err != nil {
			json := HandleError(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(json)
			return
		}

		data := make([]Service, 0)
		if service != nil {
			data = append(data, *service)
		}

		json, err := json.Marshal(data)
		if err != nil {
			json := HandleError(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(json)
			return
		}

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

		// Decode JSON
		body := requestBody{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		// Lookup servicetype to confirm exists
		serviceType, err := GetServiceTypeByID(db, body.ServiceTypeID)
		if err != nil {
			json := HandleError(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(json)
			return
		}
		if serviceType == nil {
			json := HandleError("invalid_service_type")
			w.WriteHeader(http.StatusNotFound)
			w.Write(json)
			return
		}

		// Create our new service
		service := Service{Hostname: body.Hostname, ServiceTypeID: serviceType.ID}
		if err := db.Create(&service).Error; err != nil {
			json := HandleError(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(json)
			return
		}

		// Turn into JSON
		json, err := json.Marshal(service)
		if err != nil {
			json := HandleError(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(json)
			return
		}

		// Handle response
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

// GetServiceTypeByNameHandler : handles retrival of servicetype from API request
func GetServiceTypeByNameHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		name := r.FormValue("name")
		servicetype, err := GetServiceTypeByName(db, name)
		if err != nil {
			json := HandleError(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(json)
			return
		}

		data := make([]ServiceType, 0)
		if servicetype != nil {
			data = append(data, *servicetype)
		}

		json, err := json.Marshal(data)
		if err != nil {
			json := HandleError(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(json)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}
