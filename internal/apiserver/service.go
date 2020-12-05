package apiserver

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/service"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// APIServer : defines apiserver service
type APIServer struct {
	ErrorC chan string
	Flags  *Flags
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

	apiserver := &APIServer{
		ErrorC: make(chan string),
		Flags:  flags,
	}

	return apiserver, nil
}

// Start : handles start of apiserver service
func (apiserver *APIServer) Start() {
	listen, err := net.Listen("tcp", apiserver.Flags.ListenAddr.String())

	if err != nil {
		apiserver.ErrorC <- err.Error()
	}

	grpcServer := grpc.NewServer()

	apiserverServer := &GRPCServerAPIServer{
		APIServer: apiserver,
	}

	api.RegisterAPIServerServer(grpcServer, apiserverServer)

	logrus.Info("Started apiserver grpc server.")

	serveErr := grpcServer.Serve(listen)

	if serveErr != nil {
		apiserver.ErrorC <- serveErr.Error()
	}
}

func (apiserver *APIServer) getClusters() ([]*api.Cluster, error) {
	results, err := apiserver.getResources("cluster/")

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	clusters := make([]*api.Cluster, 0)

	for _, ev := range results {
		var c *api.Cluster

		err := json.Unmarshal(ev.Value, &c)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		clusters = append(clusters, c)
	}

	return clusters, nil
}

func (apiserver *APIServer) getClusterByID(clusterID uuid.UUID) (*api.Cluster, error) {
	clusterKey := fmt.Sprintf("cluster/%s", clusterID.String())

	result, err := apiserver.getResource(clusterKey)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var cluster *api.Cluster

	jsonErr := json.Unmarshal(result.Value, &cluster)

	if jsonErr != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	return cluster, nil
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

func (apiserver *APIServer) getServices() ([]*api.Service, error) {
	results, err := apiserver.getResources("service/")

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	services := make([]*api.Service, 0)

	for _, ev := range results {
		var s *api.Service

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		services = append(services, s)
	}

	return services, nil
}

func (apiserver *APIServer) getServiceByID(id uuid.UUID) (*api.Service, error) {
	key := fmt.Sprintf("service/%s", id.String())

	result, err := apiserver.getResource(key)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var s *api.Service

	jsonErr := json.Unmarshal(result.Value, &s)

	if jsonErr != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}
	return s, nil
}

func (apiserver *APIServer) getResource(key string) (*consul.KVPair, error) {
	client, err := service.NewConsulClient(apiserver.Flags.ConsulAddr)

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

func (apiserver *APIServer) getResources(key string) (consul.KVPairs, error) {
	client, err := service.NewConsulClient(apiserver.Flags.ConsulAddr)

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

func (apiserver *APIServer) newService(clusterID uuid.UUID, serviceType api.ServiceType) (*api.Service, error) {
	cluster, err := apiserver.getClusterByID(clusterID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if cluster == nil {
		st := status.New(codes.NotFound, "invalid_cluster")

		return nil, st.Err()
	}

	serviceID := uuid.New()

	service := &api.Service{
		ClusterID: cluster.ID,
		ID:        serviceID.String(),
		Status:    api.ResourceStatus_CREATE_COMPLETED,
		Type:      serviceType,
	}

	serviceSaveErr := apiserver.saveService(service)

	if serviceSaveErr != nil {
		st := status.New(codes.Internal, serviceSaveErr.Error())

		return nil, st.Err()
	}

	return service, nil
}

func (apiserver *APIServer) removeResource(key string) error {
	client, err := service.NewConsulClient(apiserver.Flags.ConsulAddr)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return st.Err()
	}

	kv := client.KV()

	_, resErr := kv.Delete(key, nil)

	if resErr != nil {
		st := status.New(codes.Internal, resErr.Error())

		return st.Err()
	}

	return nil
}

func (apiserver *APIServer) saveCluster(cluster *api.Cluster) error {
	client, err := service.NewConsulClient(apiserver.Flags.ConsulAddr)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return st.Err()
	}

	kv := client.KV()

	clusterValue, err := json.Marshal(cluster)

	if err != nil {
		return err
	}

	clusterKey := fmt.Sprintf("cluster/%s", cluster.ID)

	kvPair := &consul.KVPair{
		Key:   clusterKey,
		Value: clusterValue,
	}

	_, putErr := kv.Put(kvPair, nil)

	if putErr != nil {
		return err
	}

	return nil
}

func (apiserver *APIServer) saveLogicalVolume(lv *api.LogicalVolume) error {
	client, err := service.NewConsulClient(apiserver.Flags.ConsulAddr)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return st.Err()
	}

	kv := client.KV()

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
	client, err := service.NewConsulClient(apiserver.Flags.ConsulAddr)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return st.Err()
	}

	kv := client.KV()

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
	client, err := service.NewConsulClient(apiserver.Flags.ConsulAddr)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return st.Err()
	}

	kv := client.KV()

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

func (apiserver *APIServer) saveService(s *api.Service) error {
	client, err := service.NewConsulClient(apiserver.Flags.ConsulAddr)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return st.Err()
	}

	kv := client.KV()

	serviceKey := fmt.Sprintf("service/%s", s.ID)

	serviceValue, err := json.Marshal(s)

	if err != nil {
		return err
	}

	kvPair := &consul.KVPair{
		Key:   serviceKey,
		Value: serviceValue,
	}
	_, putErr := kv.Put(kvPair, nil)

	if putErr != nil {
		return err
	}

	return nil
}
