package manager

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

// Manager : defines manager service
type Manager struct {
	ErrorC chan error
	State  *StateMachine

	consul *consul.Client
	flags  *flags
}

// New : creates a new apiserver instance
func New() (*Manager, error) {
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

	manager := &Manager{
		ErrorC: make(chan error),
		State:  newState(),

		consul: consulClient,
		flags:  flags,
	}

	return manager, nil
}

// Start : handles start of apiserver service
func (m *Manager) Start() {
	listen, err := net.Listen("tcp", m.flags.ServiceAddr.String())

	if err != nil {
		m.ErrorC <- err

		return
	}

	grpcServer := grpc.NewServer()

	managerServer := &grpcServerManager{
		manager: m,
	}

	api.RegisterManagerServer(grpcServer, managerServer)

	logrus.Info("Started TCP server")

	serveErr := grpcServer.Serve(listen)

	if serveErr != nil {
		m.ErrorC <- serveErr

		return
	}
}

func (m *Manager) agentServices() (map[string]*consul.AgentService, error) {
	agent := m.consul.Agent()

	services, err := agent.Services()

	if err != nil {
		return nil, err
	}

	return services, nil
}

func (m *Manager) agentServicesByType(serviceType api.ServiceType) (map[string]*consul.AgentService, error) {
	agent := m.consul.Agent()

	filter := fmt.Sprintf("Meta.ServiceType==%s", serviceType.String())

	services, err := agent.ServicesWithFilter(filter)

	if err != nil {
		return nil, err
	}

	return services, nil
}

func (m *Manager) agentServiceByID(id uuid.UUID) (*consul.AgentService, error) {
	agent := m.consul.Agent()

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

func (m *Manager) agentServiceHealth(agentServiceID string) (string, error) {
	agent := m.consul.Agent()

	status, info, err := agent.AgentHealthServiceByID(agentServiceID)

	if status == "critical" && info == nil {
		return status, errors.New("invalid_service_id")
	}

	if err != nil {
		return status, err
	}

	return status, nil
}

func (m *Manager) deleteResource(key string) error {
	kv := m.consul.KV()

	_, resErr := kv.Delete(key, nil)

	if resErr != nil {
		st := status.New(codes.Internal, resErr.Error())

		return st.Err()
	}

	return nil
}

func (m *Manager) logicalVolumes() ([]*api.LogicalVolume, error) {
	results, err := m.resources("logicalvolume/")

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
func (m *Manager) logicalVolumeByID(id uuid.UUID) (*api.LogicalVolume, error) {
	key := fmt.Sprintf("logicalvolume/%s", id.String())

	result, err := m.resource(key)

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

func (m *Manager) physicalVolumes() ([]*api.PhysicalVolume, error) {
	results, err := m.resources("physicalvolume/")

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

func (m *Manager) physicalVolumeByID(id uuid.UUID) (*api.PhysicalVolume, error) {
	key := fmt.Sprintf("physicalvolume/%s", id.String())

	result, err := m.resource(key)

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

func (m *Manager) resources(key string) (consul.KVPairs, error) {
	kv := m.consul.KV()

	results, _, err := kv.List(key, nil)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (m *Manager) resource(key string) (*consul.KVPair, error) {
	kv := m.consul.KV()

	results, _, err := kv.Get(key, nil)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (m *Manager) saveLogicalVolume(lv *api.LogicalVolume) error {
	kv := m.consul.KV()

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

func (m *Manager) savePhysicalVolume(pv *api.PhysicalVolume) error {
	kv := m.consul.KV()

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

func (m *Manager) saveVolumeGroup(vg *api.VolumeGroup) error {
	kv := m.consul.KV()

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

func (m *Manager) service(agentService *consul.AgentService) (*api.Service, error) {
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

func (m *Manager) services(agentServices map[string]*consul.AgentService) ([]*api.Service, error) {
	services := make([]*api.Service, 0)

	for _, ev := range agentServices {
		s, err := m.service(ev)
		if err != nil {
			return nil, err
		}

		services = append(services, s)
	}

	return services, nil
}

func (m *Manager) volumeGroups() ([]*api.VolumeGroup, error) {
	results, err := m.resources("volumegroup/")

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

func (m *Manager) volumeGroupByID(id uuid.UUID) (*api.VolumeGroup, error) {
	key := fmt.Sprintf("volumegroup/%s", id.String())

	result, err := m.resource(key)

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
