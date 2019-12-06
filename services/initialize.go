package services // import "github.com/erkrnt/symphony/services"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
)

// Service : struct for service in postgres
type Service struct {
	ID       uuid.UUID
	Hostname string
}

// ServiceType : struct for service in postgres
type ServiceType struct {
	ID   uuid.UUID
	Name string
}

// getServiceByHostname : requests service id from conductor
func getServiceByHostname(conductorHostname string, hostname string) (*Service, error) {
	base := fmt.Sprintf("http://%s/service", conductorHostname)
	url := fmt.Sprintf("%s?hostname=%s", base, hostname)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	services := make([]Service, 0)
	if err := json.Unmarshal(body, &services); err != nil {
		return nil, err
	}
	if len(services) < 1 {
		return nil, nil
	}
	return &services[0], nil
}

// getServiceTypeByName : requests servicetype from conductor
func getServiceTypeByName(conductorHostname string, serviceTypeName string) (*ServiceType, error) {
	url := fmt.Sprintf("http://%s/servicetype?name=%s", conductorHostname, serviceTypeName)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	serviceTypes := make([]ServiceType, 0)
	if err := json.Unmarshal(body, &serviceTypes); err != nil {
		return nil, err
	}
	if len(serviceTypes) < 1 {
		return nil, errors.New("invalid_service_type")
	}
	return &serviceTypes[0], nil
}

// newService : creates a service entry in conductor
func newService(conductorHostname string, hostname string, serviceTypeName string) (*Service, error) {
	serviceType, err := getServiceTypeByName(conductorHostname, serviceTypeName)
	if err != nil {
		return nil, err
	}
	data := fmt.Sprintf(`{"hostname":"%s","service_type_id":"%s"}`, hostname, serviceType.ID)
	url := fmt.Sprintf("http://%s/service", conductorHostname)
	res, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	service := Service{}
	if err := json.Unmarshal(body, &service); err != nil {
		return nil, err
	}
	return &service, nil
}

// GetService : handles initialization and clustering of instance
func GetService(conductorHostname string, hostname string, serviceTypeName string) (*Service, error) {
	service, err := getServiceByHostname(conductorHostname, hostname)
	if err != nil {
		return nil, err
	}
	if service == nil {
		service, err := newService(conductorHostname, hostname, serviceTypeName)
		if err != nil {
			return nil, err
		}
		return service, nil
	}
	return service, nil
}
