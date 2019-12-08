package main

import (
	"encoding/json"
	"net/http"

	"github.com/erkrnt/symphony/schemas"
)

type requestBody struct {
	Device string `json:"device"`
}

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

// GetPhysicalVolumeByDeviceHandler : handles HTTP request for getting pvs
func GetPhysicalVolumeByDeviceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		device := r.FormValue("device")
		json := make([]schemas.PhysicalVolumeMetadata, 0)

		pv, pvErr := pvExists(device)
		if pvErr != nil {
			HandleErrorResponse(w, pvErr)
			return
		}
		if pv != nil {
			json = append(json, *pv)
		}

		HandleResponse(w, json)
	}
}

// PostPhysicalVolumeHandler : creates a physical volume on host
func PostPhysicalVolumeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body := requestBody{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		pv, err := pvCreate(body.Device)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		HandleResponse(w, pv)
	}
}

// DeletePhysicalVolumeHandler : delete a physical volume on host
func DeletePhysicalVolumeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body := requestBody{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		err := pvRemove(body.Device)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		json := make([]schemas.PhysicalVolumeMetadata, 0)
		HandleResponse(w, json)
	}
}
