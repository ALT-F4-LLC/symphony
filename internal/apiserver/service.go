package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// APIServer : defines apiserver service
type APIServer struct {
	Consul *consul.Client
	ErrorC chan error
	Flags  *Flags
}

type scheduleResourceStatusOptions struct {
	serviceID      string
	resourceID     string
	resourceStatus api.ResourceStatus
	resourceType   api.ResourceType
}

func GetService(agentService *consul.AgentService) (*api.Service, error) {
	var serviceType api.ServiceType

	switch agentService.Meta["ServiceType"] {
	case "BLOCK":
		serviceType = api.ServiceType_BLOCK
	}

	if serviceType == api.ServiceType_UNKNOWN_SERVICE_TYPE {
		return nil, errors.New("invalid_service_type")
	}

	service := &api.Service{
		ID:   agentService.ID,
		Type: serviceType,
	}

	return service, nil
}

func GetServices(agentServices map[string]*consul.AgentService) ([]*api.Service, error) {
	services := make([]*api.Service, 0)

	for _, ev := range agentServices {
		s, err := GetService(ev)

		if err != nil {
			return nil, err
		}

		services = append(services, s)
	}

	return services, nil
}

// New : creates a new apiserver instance
func New() (*APIServer, error) {
	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	if flags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	consulClient, err := utils.NewConsulClient(flags.ConsulAddr)

	if err != nil {
		logrus.Fatal(err)
	}

	errorC := make(chan error)

	apiserver := &APIServer{
		Consul: consulClient,
		ErrorC: errorC,
		Flags:  flags,
	}

	return apiserver, nil
}

// Start : handles start of apiserver service
func (apiserver *APIServer) Start() {
	go apiserver.listenGRPC()
}

func (apiserver *APIServer) getAgentServices() (map[string]*consul.AgentService, error) {
	agent := apiserver.Consul.Agent()

	services, err := agent.Services()

	if err != nil {
		return nil, err
	}

	return services, nil
}

func (apiserver *APIServer) getAgentServiceByID(id uuid.UUID) (*consul.AgentService, error) {
	agent := apiserver.Consul.Agent()

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

func (apiserver *APIServer) getAgentServicesByType(serviceType api.ServiceType) (map[string]*consul.AgentService, error) {
	agent := apiserver.Consul.Agent()

	filter := fmt.Sprintf("Meta.ServiceType==%s", serviceType.String())

	services, err := agent.ServicesWithFilter(filter)

	if err != nil {
		return nil, err
	}

	return services, nil
}

func (apiserver *APIServer) getLogicalVolumes() ([]*api.LogicalVolume, error) {
	results, err := apiserver.getResources("logicalvolume/")

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	lvs := make([]*api.LogicalVolume, 0)

	for _, ev := range results {
		var lv *api.LogicalVolume

		err := json.Unmarshal(ev.Value, &lv)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		lvs = append(lvs, lv)
	}

	return lvs, nil
}

func (apiserver *APIServer) getLogicalVolumeByID(id uuid.UUID) (*api.LogicalVolume, error) {
	key := fmt.Sprintf("logicalvolume/%s", id.String())

	result, err := apiserver.getResource(key)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var lv *api.LogicalVolume

	jsonErr := json.Unmarshal(result.Value, &lv)

	if jsonErr != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	return lv, nil
}

func (apiserver *APIServer) getPhysicalVolumes() ([]*api.PhysicalVolume, error) {
	results, err := apiserver.getResources("physicalvolume/")

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	pvs := make([]*api.PhysicalVolume, 0)

	for _, ev := range results {
		var s *api.PhysicalVolume

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		pvs = append(pvs, s)
	}

	return pvs, nil
}

func (apiserver *APIServer) getPhysicalVolumeByID(id uuid.UUID) (*api.PhysicalVolume, error) {
	key := fmt.Sprintf("physicalvolume/%s", id.String())

	result, err := apiserver.getResource(key)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var pv *api.PhysicalVolume

	jsonErr := json.Unmarshal(result.Value, &pv)

	if jsonErr != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	return pv, nil
}

func (apiserver *APIServer) getResource(key string) (*consul.KVPair, error) {
	kv := apiserver.Consul.KV()

	results, _, err := kv.Get(key, nil)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (apiserver *APIServer) getResources(key string) (consul.KVPairs, error) {
	kv := apiserver.Consul.KV()

	results, _, err := kv.List(key, nil)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (apiserver *APIServer) getVolumeGroups() ([]*api.VolumeGroup, error) {
	results, err := apiserver.getResources("volumegroup/")

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	vgs := make([]*api.VolumeGroup, 0)

	for _, ev := range results {
		var s *api.VolumeGroup

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		vgs = append(vgs, s)
	}

	return vgs, nil
}

func (apiserver *APIServer) getVolumeGroupByID(id uuid.UUID) (*api.VolumeGroup, error) {
	key := fmt.Sprintf("volumegroup/%s", id.String())

	result, err := apiserver.getResource(key)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var vg *api.VolumeGroup

	jsonErr := json.Unmarshal(result.Value, &vg)

	if jsonErr != nil {
		st := status.New(codes.Internal, jsonErr.Error())

		return nil, st.Err()
	}

	return vg, nil
}

func (apiserver *APIServer) listenGRPC() {
	listen, err := net.Listen("tcp", apiserver.Flags.ListenAddr.String())

	if err != nil {
		apiserver.ErrorC <- err

		return
	}

	grpcServer := grpc.NewServer()

	apiserverServer := &GRPCServerAPIServer{
		APIServer: apiserver,
	}

	api.RegisterAPIServerServer(grpcServer, apiserverServer)

	logrus.Info("Started TCP server")

	serveErr := grpcServer.Serve(listen)

	if serveErr != nil {
		apiserver.ErrorC <- serveErr

		return
	}
}

func (apiserver *APIServer) removeResource(key string) error {
	kv := apiserver.Consul.KV()

	_, resErr := kv.Delete(key, nil)

	if resErr != nil {
		st := status.New(codes.Internal, resErr.Error())

		return st.Err()
	}

	return nil
}

func (apiserver *APIServer) saveLogicalVolume(lv *api.LogicalVolume) error {
	kv := apiserver.Consul.KV()

	lvKey := fmt.Sprintf("logicalvolume/%s", lv.ID)

	lvValue, err := json.Marshal(lv)

	if err != nil {
		return err
	}

	kvPair := &consul.KVPair{
		Key:   lvKey,
		Value: lvValue,
	}

	_, putErr := kv.Put(kvPair, nil)

	if putErr != nil {
		return err
	}

	return nil
}

func (apiserver *APIServer) savePhysicalVolume(pv *api.PhysicalVolume) error {
	kv := apiserver.Consul.KV()

	pvKey := fmt.Sprintf("physicalvolume/%s", pv.ID)

	pvValue, err := json.Marshal(pv)

	if err != nil {
		return err
	}

	kvPair := &consul.KVPair{
		Key:   pvKey,
		Value: pvValue,
	}

	_, putErr := kv.Put(kvPair, nil)

	if putErr != nil {
		return err
	}

	return nil
}

func (apiserver *APIServer) saveVolumeGroup(vg *api.VolumeGroup) error {
	kv := apiserver.Consul.KV()

	vgKey := fmt.Sprintf("volumegroup/%s", vg.ID)

	vgValue, err := json.Marshal(vg)

	if err != nil {
		return err
	}

	kvPair := &consul.KVPair{
		Key:   vgKey,
		Value: vgValue,
	}

	_, putErr := kv.Put(kvPair, nil)

	if putErr != nil {
		return err
	}

	return nil
}
