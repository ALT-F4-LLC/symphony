package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erkrnt/symphony/services"
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
	w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

// GetPvByDeviceHandler : handles HTTP request for getting pvs
func GetPvByDeviceHandler(db *gorm.DB, service *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		device := r.FormValue("device")
		json := make([]Pv, 0)

		pv, err := GetPvByDevice(db, device, service)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}
		if pv != nil {
			json = append(json, *pv)
		}

		pvd, pvdErr := pvExists(device)
		if pvdErr != nil {
			HandleErrorResponse(w, pvdErr)
			return
		}
		if pvd == nil {
			HandleErrorResponse(w, errors.New("invalid_pv_device"))
			return
		}

		HandleResponse(w, json)
	}
}
