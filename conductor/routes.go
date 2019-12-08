package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erkrnt/symphony/schemas"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// ErrorResponse : struct for describing a pv error
type ErrorResponse struct {
	Message string `json:"message"`
}

// HandleErrorResponse : translates error to json responses
func HandleErrorResponse(w http.ResponseWriter, err error) {
	res := &ErrorResponse{Message: err.Error()}
	json, err := json.Marshal(res)
	if err != nil {
		http.Error(w, "invalid_json", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(json)
}

// HandleResponse : translates response to json
func HandleResponse(w http.ResponseWriter, data interface{}) {
	json, err := json.Marshal(data)
	if err != nil {
		HandleErrorResponse(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

// GetServiceByIDHandler : handles retrival of service from API request
func GetServiceByIDHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		params := mux.Vars(r)
		serviceID := params["id"]

		id, err := uuid.Parse(serviceID)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		service, err := GetServiceByID(db, id)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		data := make([]schemas.Service, 0)
		if service != nil {
			data = append(data, *service)
		}

		json, err := json.Marshal(data)
		if err != nil {
			HandleErrorResponse(w, err)
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
			HandleErrorResponse(w, err)
			return
		}

		data := make([]schemas.Service, 0)
		if service != nil {
			data = append(data, *service)
		}

		json, err := json.Marshal(data)
		if err != nil {
			HandleErrorResponse(w, err)
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

		body := requestBody{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		serviceType, err := GetServiceTypeByID(db, body.ServiceTypeID)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}
		if serviceType == nil {
			HandleErrorResponse(w, errors.New("invalid_service_type"))
			return
		}

		service := schemas.Service{Hostname: body.Hostname, ServiceTypeID: serviceType.ID}
		if err := db.Create(&service).Error; err != nil {
			HandleErrorResponse(w, err)
			return
		}

		json, err := json.Marshal(service)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

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
			HandleErrorResponse(w, err)
			return
		}

		data := make([]schemas.ServiceType, 0)
		if servicetype != nil {
			data = append(data, *servicetype)
		}

		HandleResponse(w, data)
	}
}

// GetPhysicalVolumeByDeviceHandler : handles HTTP request for getting pvs
func GetPhysicalVolumeByDeviceHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		device := r.FormValue("device")
		id := r.FormValue("service_id")
		json := make([]schemas.PhysicalVolume, 0)

		serviceID, err := uuid.Parse(id)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		pv, err := GetPhysicalVolumeByDeviceAndServiceID(db, device, serviceID)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}
		if pv == nil {
			HandleResponse(w, json)
			return
		}

		metadata, err := GetPhysicalVolumeMetadataByDeviceOnHost(db, device, serviceID)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		pv.Metadata = metadata
		json = append(json, *pv)

		HandleResponse(w, json)
	}
}

// PostPhysicalVolumeHandler : handles creation of physical volume on a host
func PostPhysicalVolumeHandler(db *gorm.DB) http.HandlerFunc {
	type requestBody struct {
		Device    string    `json:"device"`
		ServiceID uuid.UUID `json:"service_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body := requestBody{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		existing, err := GetPhysicalVolumeByDeviceAndServiceID(db, body.Device, body.ServiceID)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}
		if existing != nil {
			HandleErrorResponse(w, errors.New("physical_volume_exists"))
			return
		}

		service, err := GetServiceByID(db, body.ServiceID)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}
		if service == nil {
			HandleErrorResponse(w, errors.New("invalid_service_id"))
			return
		}

		serviceType, err := GetServiceTypeByID(db, service.ServiceTypeID)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}
		if serviceType == nil {
			HandleErrorResponse(w, errors.New("invalid_service_type_id"))
			return
		}
		if serviceType.Name != "block" {
			HandleErrorResponse(w, errors.New("invalid_service_type"))
			return
		}

		metadata, err := CreatePhysicalVolumeOnHost(service.Hostname, body.Device)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		pv := schemas.PhysicalVolume{Device: body.Device, ServiceID: service.ID}
		if err := db.Create(&pv).Error; err != nil {
			HandleErrorResponse(w, err)
			return
		}

		pv.Metadata = metadata

		HandleResponse(w, pv)
	}
}

// DeletePhysicalVolumeHandler : handles creation of physical volume on a host
func DeletePhysicalVolumeHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		params := mux.Vars(r)
		deviceID := params["id"]

		id, err := uuid.Parse(deviceID)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		pv, err := GetPhysicalVolumeByDeviceID(db, id)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}
		if pv == nil {
			HandleErrorResponse(w, errors.New("invalid_physical_volume"))
			return
		}

		service, err := GetServiceByID(db, pv.ServiceID)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}
		if service == nil {
			HandleErrorResponse(w, errors.New("invalid_service_id"))
			return
		}

		delErr := DeletePhysicalVolumeOnHost(service.Hostname, pv.Device)
		if delErr != nil {
			HandleErrorResponse(w, delErr)
			return
		}

		dbErr := DeletePhysicalVolumeByDeviceID(db, pv.ID)
		if dbErr != nil {
			HandleErrorResponse(w, dbErr)
			return
		}

		json := make([]schemas.PhysicalVolume, 0)
		HandleResponse(w, json)
	}
}
