package manager

import (
	"errors"
	"fmt"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func agentServices(consulAddr string) (map[string]*consul.AgentService, error) {
	client, err := utils.NewConsulClient(consulAddr)

	if err != nil {
		return nil, err
	}

	agent := client.Agent()

	services, err := agent.Services()

	if err != nil {
		return nil, err
	}

	return services, nil
}

func agentServicesByType(consulAddr string, serviceType api.ServiceType) (map[string]*consul.AgentService, error) {
	client, err := utils.NewConsulClient(consulAddr)

	if err != nil {
		return nil, err
	}

	agent := client.Agent()

	filter := fmt.Sprintf("Meta.ServiceType==%s", serviceType.String())

	services, err := agent.ServicesWithFilter(filter)

	if err != nil {
		return nil, err
	}

	return services, nil
}

func agentServiceByID(consulAddr string, id uuid.UUID) (*consul.AgentService, error) {
	client, err := utils.NewConsulClient(consulAddr)

	if err != nil {
		return nil, err
	}

	agent := client.Agent()

	serviceOpts := &consul.QueryOptions{}

	service, _, err := agent.Service(id.String(), serviceOpts)

	if err != nil {
		return nil, err
	}

	if service == nil {
		return nil, errors.New("invalid_service_id")
	}

	return service, nil
}

func agentServiceHealth(consulAddr string, agentServiceID string) (string, error) {
	client, err := utils.NewConsulClient(consulAddr)

	if err != nil {
		return "critical", err
	}

	agent := client.Agent()

	status, info, err := agent.AgentHealthServiceByID(agentServiceID)

	if status == "critical" && info == nil {
		return status, errors.New("invalid_service_id")
	}

	if err != nil {
		return status, err
	}

	return status, nil
}

func deleteKvPair(consulAddr string, key string) error {
	client, err := utils.NewConsulClient(consulAddr)

	if err != nil {
		return err
	}

	kv := client.KV()

	_, resErr := kv.Delete(key, nil)

	if resErr != nil {
		st := status.New(codes.Internal, resErr.Error())

		return st.Err()
	}

	return nil
}

func kvPair(consulAddr string, key string) (*consul.KVPair, error) {
	client, err := utils.NewConsulClient(consulAddr)

	if err != nil {
		return nil, err
	}

	kv := client.KV()

	results, _, err := kv.Get(key, nil)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func kvPairs(consulAddr string, key string) (consul.KVPairs, error) {
	client, err := utils.NewConsulClient(consulAddr)

	if err != nil {
		return nil, err
	}

	kv := client.KV()

	results, _, err := kv.List(key, nil)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func putKVPair(consulAddr string, key string, value []byte) error {
	client, err := utils.NewConsulClient(consulAddr)

	if err != nil {
		return err
	}

	kv := client.KV()

	kvPair := &consul.KVPair{
		Key:   key,
		Value: value,
	}

	_, putErr := kv.Put(kvPair, nil)

	if putErr != nil {
		return putErr
	}

	return nil
}
