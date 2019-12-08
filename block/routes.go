package main

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse : struct for describing a pv error
type ErrorResponse struct {
	Message string `json:"message"`
}

// RequestBody : describes HTTP post request body
type RequestBody struct {
	Device string `json:"device"`
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
