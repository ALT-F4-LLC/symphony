package utils

import (
	"fmt"

	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
)

// KvKey : creates a key-value key representation
func KvKey(resourceID uuid.UUID, resourceType api.ResourceType) string {
	return fmt.Sprintf("%s/%s", resourceType.String(), resourceID.String())
}

// NewConsulClient : creates a new Consul client
func NewConsulClient(address string) (*consul.Client, error) {
	config := consul.DefaultConfig()

	config.Address = address

	client, err := consul.NewClient(config)

	if err != nil {
		return nil, err
	}

	return client, nil
}
