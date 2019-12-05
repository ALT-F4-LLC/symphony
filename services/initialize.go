package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
)

// serviceType : struct for service in postgres
type serviceType struct {
	id   uuid.UUID
	name string
}

// Service : struct for service in postgres
type Service struct {
	ID       uuid.UUID
	Hostname string
}

// getServiceTypeByName : requests servicetype from conductor
func getServiceTypeByName(conductorHostname string) (*serviceType, error) {
	url := fmt.Sprintf("http://%s/servicetype?name=%s", conductorHostname, "block")
	serviceTypesRes, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer serviceTypesRes.Body.Close()
	body, err := ioutil.ReadAll(serviceTypesRes.Body)
	if err != nil {
		return nil, err
	}
	serviceTypes := []serviceType{}
	if err := json.Unmarshal(body, &serviceTypes); err != nil {
		return nil, err
	}
	if len(serviceTypes) < 1 {
		return nil, errors.New("invalid_servicetype")
	}
	return &serviceTypes[0], nil
}

// newService : creates a service entry in conductor
func newService(conductorHostname string, hostname string) (*Service, error) {
	serviceType, err := getServiceTypeByName(conductorHostname)
	if err != nil {
		return nil, err
	}
	data := fmt.Sprintf(`{"hostname":"%s","servicetype_id":"%s"}`, hostname, serviceType.id)
	baseURL := fmt.Sprintf("http://%s/service", conductorHostname)
	serviceRes, err := http.Post(baseURL, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	defer serviceRes.Body.Close()
	body, err := ioutil.ReadAll(serviceRes.Body)
	if err != nil {
		return nil, err
	}
	service := Service{}
	if err := json.Unmarshal(body, &service); err != nil {
		return nil, err
	}
	return &service, nil
}

// getServiceByHostname : requests service id from conductor
func getServiceByHostname(conductorHostname string, hostname string) (*Service, error) {
	base := fmt.Sprintf("http://%s/service", conductorHostname)
	url := fmt.Sprintf("%s?hostname=%s", base, hostname)
	serviceRes, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer serviceRes.Body.Close()
	body, err := ioutil.ReadAll(serviceRes.Body)
	if err != nil {
		return nil, err
	}
	services := []Service{}
	if err := json.Unmarshal(body, &services); err != nil {
		return nil, err
	}
	if len(services) < 1 {
		return nil, nil
	}
	return &services[0], nil
}

// GetService : handles initialization and clustering of instance
func GetService(conductorHostname string, hostname string) (*Service, error) {
	service, err := getServiceByHostname(conductorHostname, hostname)
	if err != nil {
		return nil, err
	}
	if service == nil {
		srvc, err := newService(conductorHostname, hostname)
		if err != nil {
			return nil, err
		}
		return srvc, nil
	}
	return service, nil
}
