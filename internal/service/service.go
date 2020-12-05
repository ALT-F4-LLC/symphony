package service

import "github.com/hashicorp/consul/api"

// Service : service member
type Service struct {
	ErrorC chan string
	Key    *Key
}

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
