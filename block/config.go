package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config : service configuration
type Config struct {
	Conductor struct {
		Hostname string
	}
	Hostname string
}

// GetConfig : gets service configuration data
func GetConfig(path string) (*Config, error) {
	if path == "" {
		path = "./config.yml"
	}
	config := Config{}
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
