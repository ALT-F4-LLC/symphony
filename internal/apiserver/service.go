package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/service"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
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
	results, err := apiserver.getStateByKey("/Cluster", clientv3.WithPrefix())

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	clusters := make([]*api.Cluster, 0)

	for _, ev := range results.Kvs {
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
	clusterKey := fmt.Sprintf("/Cluster/%s", clusterID.String())

	results, err := apiserver.getStateByKey(clusterKey)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var cluster *api.Cluster

	for i, ev := range results.Kvs {

		if i == 0 {
			err := json.Unmarshal(ev.Value, &cluster)

			if err != nil {
				st := status.New(codes.Internal, err.Error())

				return nil, st.Err()
			}
		}
	}

	return cluster, nil
}

func (apiserver *APIServer) getLogicalVolumes() ([]*api.LogicalVolume, error) {
	results, err := apiserver.getStateByKey("/LogicalVolume", clientv3.WithPrefix())

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	lvs := make([]*api.LogicalVolume, 0)

	for _, ev := range results.Kvs {
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
	etcdKey := fmt.Sprintf("/LogicalVolume/%s", id.String())

	results, err := apiserver.getStateByKey(etcdKey)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var lv *api.LogicalVolume

	for _, ev := range results.Kvs {
		var v *api.LogicalVolume

		err := json.Unmarshal(ev.Value, &v)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		if v.ID == id.String() {
			lv = v
		}
	}

	return lv, nil
}

func (apiserver *APIServer) getPhysicalVolumes() ([]*api.PhysicalVolume, error) {
	results, err := apiserver.getStateByKey("/PhysicalVolume", clientv3.WithPrefix())

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	pvs := make([]*api.PhysicalVolume, 0)

	for _, ev := range results.Kvs {
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
	etcdKey := fmt.Sprintf("/physicalvolume/%s", id.String())

	results, err := apiserver.getStateByKey(etcdKey)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var volume *api.PhysicalVolume

	for _, ev := range results.Kvs {
		var s *api.PhysicalVolume

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		if s.ID == id.String() {
			volume = s
		}
	}

	return volume, nil
}

func (apiserver *APIServer) getServices() ([]*api.Service, error) {
	results, err := apiserver.getStateByKey("/Service", clientv3.WithPrefix())

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	services := make([]*api.Service, 0)

	for _, ev := range results.Kvs {
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
	etcdKey := fmt.Sprintf("/Service/%s", id.String())

	results, err := apiserver.getStateByKey(etcdKey)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var service *api.Service

	for _, ev := range results.Kvs {
		var s *api.Service

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		if s.ID == id.String() {
			service = s
		}
	}

	return service, nil
}

func (apiserver *APIServer) getStateByKey(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), service.DialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: service.DialTimeout,
		Endpoints:   apiserver.Flags.EtcdEndpoints,
	}

	client, err := clientv3.New(config)

	if err != nil {
		return nil, err
	}

	defer client.Close()

	results, err := client.Get(ctx, key, opts...)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (apiserver *APIServer) getVolumeGroups() ([]*api.VolumeGroup, error) {
	results, err := apiserver.getStateByKey("/VolumeGroup", clientv3.WithPrefix())

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	vgs := make([]*api.VolumeGroup, 0)

	for _, ev := range results.Kvs {
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
	etcdKey := fmt.Sprintf("/VolumeGroup/%s", id.String())

	results, err := apiserver.getStateByKey(etcdKey)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var volumeGroup *api.VolumeGroup

	for _, ev := range results.Kvs {
		var vg *api.VolumeGroup

		err := json.Unmarshal(ev.Value, &vg)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		if vg.ID == id.String() {
			volumeGroup = vg
		}
	}

	return volumeGroup, nil
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
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   apiserver.Flags.EtcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return st.Err()
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	_, res := etcd.Delete(ctx, key)

	if res != nil {
		st := status.New(codes.Internal, res.Error())

		return st.Err()
	}

	return nil
}

func (apiserver *APIServer) saveCluster(cluster *api.Cluster) error {
	ctx, cancel := context.WithTimeout(context.Background(), service.DialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: service.DialTimeout,
		Endpoints:   apiserver.Flags.EtcdEndpoints,
	}

	client, err := clientv3.New(config)

	if err != nil {
		return err
	}

	defer client.Close()

	clusterJSON, err := json.Marshal(cluster)

	if err != nil {
		return err
	}

	clusterKey := fmt.Sprintf("/Cluster/%s", cluster.ID)

	_, putErr := client.Put(ctx, clusterKey, string(clusterJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (apiserver *APIServer) saveLogicalVolume(lv *api.LogicalVolume) error {
	ctx, cancel := context.WithTimeout(context.Background(), service.DialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: service.DialTimeout,
		Endpoints:   apiserver.Flags.EtcdEndpoints,
	}

	client, err := clientv3.New(config)

	if err != nil {
		return err
	}

	defer client.Close()

	lvJSON, err := json.Marshal(lv)

	if err != nil {
		return err
	}

	lvKey := fmt.Sprintf("/LogicalVolume/%s", lv.ID)

	_, putErr := client.Put(ctx, lvKey, string(lvJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (apiserver *APIServer) savePhysicalVolume(pv *api.PhysicalVolume) error {
	ctx, cancel := context.WithTimeout(context.Background(), service.DialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: service.DialTimeout,
		Endpoints:   apiserver.Flags.EtcdEndpoints,
	}

	client, err := clientv3.New(config)

	if err != nil {
		return err
	}

	defer client.Close()

	pvJSON, err := json.Marshal(pv)

	if err != nil {
		return err
	}

	volumeKey := fmt.Sprintf("/PhysicalVolume/%s", pv.ID)

	_, putErr := client.Put(ctx, volumeKey, string(pvJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (apiserver *APIServer) saveVolumeGroup(vg *api.VolumeGroup) error {
	ctx, cancel := context.WithTimeout(context.Background(), service.DialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: service.DialTimeout,
		Endpoints:   apiserver.Flags.EtcdEndpoints,
	}

	client, err := clientv3.New(config)

	if err != nil {
		return err
	}

	defer client.Close()

	vgJSON, err := json.Marshal(vg)

	if err != nil {
		return err
	}

	volumeKey := fmt.Sprintf("/VolumeGroup/%s", vg.ID)

	_, putErr := client.Put(ctx, volumeKey, string(vgJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (apiserver *APIServer) saveService(s *api.Service) error {
	ctx, cancel := context.WithTimeout(context.Background(), service.DialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: service.DialTimeout,
		Endpoints:   apiserver.Flags.EtcdEndpoints,
	}

	client, err := clientv3.New(config)

	if err != nil {
		return err
	}

	defer client.Close()

	serviceKey := fmt.Sprintf("/Service/%s", s.ID)

	serviceJSON, err := json.Marshal(s)

	if err != nil {
		return err
	}

	_, putErr := client.Put(ctx, serviceKey, string(serviceJSON))

	if putErr != nil {
		return err
	}

	return nil
}
