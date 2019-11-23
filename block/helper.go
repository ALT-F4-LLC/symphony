package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

// Service : struct for service in postgres
type Service struct {
	ID       uuid.UUID
	Hostname string
}

// ServiceConfig : service configuration
type ServiceConfig struct {
	Conductor struct {
		Hostname string
	}
	Hostname string
	Postgres struct {
		DBName   string
		Host     string
		Password string
		Port     int
		User     string
	}
}

// ServiceType : struct for service in postgres
type ServiceType struct {
	ID   uuid.UUID
	Name string
}

// GetServiceByHostname : requests service id from conductor
func GetServiceByHostname(config *ServiceConfig) (*Service, error) {
	baseURL := fmt.Sprintf("http://%s/service", config.Conductor.Hostname)
	getURL := fmt.Sprintf("%s?hostname=%s", baseURL, config.Hostname)
	serviceRes, err := http.Get(getURL)
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

// GetServiceTypeIDByName : requests servicetype id from conductor
func GetServiceTypeIDByName(config *ServiceConfig) (*uuid.UUID, error) {
	getURL := fmt.Sprintf("http://%s/servicetype?name=%s", config.Conductor.Hostname, "block")
	serviceTypesRes, err := http.Get(getURL)
	if err != nil {
		return nil, err
	}
	defer serviceTypesRes.Body.Close()
	body, err := ioutil.ReadAll(serviceTypesRes.Body)
	if err != nil {
		return nil, err
	}
	servicetypes := []ServiceType{}
	if err := json.Unmarshal(body, &servicetypes); err != nil {
		return nil, err
	}
	if len(servicetypes) < 1 {
		return nil, errors.New("invalid_servicetype")
	}
	return &servicetypes[0].ID, nil
}

// CreateService : creates a service entry in conductor
func CreateService(config *ServiceConfig) (*Service, error) {
	serviceTypeID, err := GetServiceTypeIDByName(config)
	if err != nil {
		return nil, err
	}
	data := fmt.Sprintf(`{"hostname":"%s","servicetype_id":"%s"}`, config.Hostname, serviceTypeID)
	baseURL := fmt.Sprintf("http://%s/service", config.Conductor.Hostname)
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

// HandleInit : handles initialization and clustering of instance
func HandleInit(config *ServiceConfig) (*Service, error) {
	service, err := GetServiceByHostname(config)
	if err != nil {
		return nil, err
	}
	if service == nil {
		srvc, err := CreateService(config)
		if err != nil {
			return nil, err
		}
		return srvc, nil
	}
	return service, nil
}

// GetServiceConfig : gets service configuration data
func GetServiceConfig(path string) (*ServiceConfig, error) {
	config := ServiceConfig{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	yamlErr := yaml.Unmarshal(data, &config)
	if yamlErr != nil {
		return nil, yamlErr
	}
	return &config, nil
}
