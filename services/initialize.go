package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
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
func getServiceTypeByName(config *Config) (*serviceType, error) {
	url := fmt.Sprintf("http://%s/servicetype?name=%s", config.Conductor.Hostname, "block")
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
func newService(config *Config) (*Service, error) {
	serviceType, err := getServiceTypeByName(config)
	if err != nil {
		return nil, err
	}
	data := fmt.Sprintf(`{"hostname":"%s","servicetype_id":"%s"}`, config.Hostname, serviceType.id)
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

// getServiceByHostname : requests service id from conductor
func getServiceByHostname(config *Config) (*Service, error) {
	base := fmt.Sprintf("http://%s/service", config.Conductor.Hostname)
	url := fmt.Sprintf("%s?hostname=%s", base, config.Hostname)
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
func GetService(config *Config) (*Service, error) {
	service, err := getServiceByHostname(config)
	if err != nil {
		return nil, err
	}
	if service == nil {
		srvc, err := newService(config)
		if err != nil {
			return nil, err
		}
		return srvc, nil
	}
	return service, nil
}

// GetDatabase : get database connection for service
func GetDatabase(debug bool) (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		return nil, err
	}
	if debug == true {
		db.LogMode(true)
	}
	return db, nil
}
