package utils

import "github.com/hashicorp/consul/api"

// NewConsulClient : creates a new Consul client
func NewConsulClient(address string) (*api.Client, error) {
	config := api.DefaultConfig()

	config.Address = address

	client, err := api.NewClient(config)

	if err != nil {
		return nil, err
	}

	return client, nil
}
