package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/erkrnt/symphony/internal/pkg/state"
	"github.com/google/uuid"
	"github.com/hashicorp/serf/serf"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Manager : manager node
type Manager struct {
	Node config.Node

	flags          *flags
	watchers       map[string]Watcher
	watchersReady  []string
	watchersReadyC chan struct{}
}

// Watcher : struct for watching resources
type Watcher struct {
	AssignedC   clientv3.WatchChan
	UnassignedC clientv3.WatchChan
}

var (
	etcdDialTimeout    = 5 * time.Second
	etcdWatchResources = []string{"LogicalVolume", "PhysicalVolume", "VolumeGroup"}
	grpcContextTimeout = 5 * time.Second
)

// Start : handles start of manager service
func (m *Manager) Start() {
	if m.Node.Key.ClusterID != nil && m.Node.Key.ServiceID != nil {
		restartErr := m.restart()

		if restartErr != nil {
			m.Node.ErrorC <- restartErr.Error()
		}
	}

	for k, w := range m.watchers {
		go m.startWatcherForKey(k, w)
	}

	<-m.watchersReadyC

	go m.listenControl()

	go m.listenRemote()
}

func (m *Manager) startWatcherForKey(key string, watcher Watcher) {
	config := clientv3.Config{
		DialTimeout: etcdDialTimeout,
		Endpoints:   m.flags.etcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		m.Node.ErrorC <- err.Error()
	}

	defer client.Close()

	assignedKey := fmt.Sprintf("Assigned/%s", key)

	unassignedKey := fmt.Sprintf("Unassigned/%s", key)

	watcher.AssignedC = client.Watch(context.Background(), assignedKey, clientv3.WithPrefix())

	watcher.UnassignedC = client.Watch(context.Background(), unassignedKey, clientv3.WithPrefix())

	/*  for wresp := range watcher.AssignedC { */
	// for _, ev := range wresp.Events {
	// // TODO : handle returned data
	// fmt.Printf("AssignedC: %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
	// }
	// }

	// for wresp := range watcher.UnassignedC {
	// for _, ev := range wresp.Events {
	// // TODO : handle returned data
	// fmt.Printf("UnassignedC: %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
	// }
	// }

	// NewLogicalVolume -> "logicalvolume:unassigned/*" -> Manager (watcher update) -> Begin scheduling

	// watch "logicalvolume:assigned/*" === lvAssignedC
	// watch "logicalvolume:unassigned/*" === lvUnassignedC

	m.watchersReady = append(m.watchersReady, key)

	if len(m.watchersReady) == len(etcdWatchResources) {
		m.watchersReadyC <- struct{}{}

		logrus.Info("Resource watchers registered...")
	}
}

func (m *Manager) listenControl() {
	socketPath := fmt.Sprintf("%s/control.sock", m.flags.configDir)

	if err := os.RemoveAll(socketPath); err != nil {
		m.Node.ErrorC <- err.Error()
	}

	listen, err := net.Listen("unix", socketPath)

	if err != nil {
		m.Node.ErrorC <- err.Error()
	}

	server := grpc.NewServer()

	endpoints := &endpoints{
		manager: m,
	}

	control := &control{
		endpoints: endpoints,
		manager:   m,
	}

	api.RegisterManagerServer(server, control)

	logrus.Info("Started control gRPC socket server.")

	serveErr := server.Serve(listen)

	if serveErr != nil {
		m.Node.ErrorC <- serveErr.Error()
	}
}

func (m *Manager) listenRemote() {
	listen, err := net.Listen("tcp", m.flags.listenAddr.String())

	if err != nil {
		m.Node.ErrorC <- err.Error()
	}

	endpoints := &endpoints{
		manager: m,
	}

	remote := &remote{
		endpoints: endpoints,
		manager:   m,
	}

	// TODO : start the peers server with authentication

	server := grpc.NewServer()

	api.RegisterManagerServer(server, remote)

	logrus.Info("Started manager gRPC tcp server.")

	serveErr := server.Serve(listen)

	if serveErr != nil {
		m.Node.ErrorC <- serveErr.Error()
	}
}

func (m *Manager) getCluster() (*api.Cluster, error) {
	results, err := m.getStateByKey("/cluster")

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var cluster *api.Cluster

	for _, ev := range results.Kvs {
		err := json.Unmarshal(ev.Value, &cluster)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}
	}

	return cluster, nil
}

func (m *Manager) getLogicalVolumes() ([]*api.LogicalVolume, error) {
	results, err := m.getStateByKey("/logicalvolume", clientv3.WithPrefix())

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

func (m *Manager) getPhysicalVolumes() ([]*api.PhysicalVolume, error) {
	results, err := m.getStateByKey("/physicalvolume", clientv3.WithPrefix())

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

func (m *Manager) getVolumeGroups() ([]*api.VolumeGroup, error) {
	results, err := m.getStateByKey("/volumegroup", clientv3.WithPrefix())

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

func (m *Manager) getServices() ([]*api.Service, error) {
	results, err := m.getStateByKey("/service", clientv3.WithPrefix())

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

func (m *Manager) getLogicalVolumeByID(id uuid.UUID) (*api.LogicalVolume, error) {
	etcdKey := fmt.Sprintf("/logicalvolume/%s", id.String())

	results, err := m.getStateByKey(etcdKey)

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

func (m *Manager) getPhysicalVolumeByID(id uuid.UUID) (*api.PhysicalVolume, error) {
	etcdKey := fmt.Sprintf("/physicalvolume/%s", id.String())

	results, err := m.getStateByKey(etcdKey)

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

func (m *Manager) getServiceByID(id uuid.UUID) (*api.Service, error) {
	etcdKey := fmt.Sprintf("/service/%s", id.String())

	results, err := m.getStateByKey(etcdKey)

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

func (m *Manager) getMemberFromService(service *api.Service) (*serf.Member, error) {
	var member *serf.Member

	members := m.Node.Serf.Members()

	for _, m := range members {
		if m.Tags["ServiceID"] == service.ID {
			member = &m
		}
	}

	return member, nil
}

func (m *Manager) getStateByKey(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), etcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: etcdDialTimeout,
		Endpoints:   m.flags.etcdEndpoints,
	}

	client, err := state.NewClient(config)

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

func (m *Manager) getVolumeGroupByID(id uuid.UUID) (*api.VolumeGroup, error) {
	etcdKey := fmt.Sprintf("/volumegroup/%s", id.String())

	results, err := m.getStateByKey(etcdKey)

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

func (m *Manager) newService(serviceAddr string, serviceType api.ServiceType) (*api.ManagerServiceInitResponse, error) {
	cluster, err := m.getCluster()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	isLocalInit := serviceAddr == m.flags.listenAddr.String()

	if cluster == nil && !isLocalInit {
		st := status.New(codes.InvalidArgument, "cluster_not_initialized")

		return nil, st.Err()
	}

	if cluster == nil && isLocalInit {
		clusterID := uuid.New()

		cluster = &api.Cluster{
			ID: clusterID.String(),
		}

		err := m.saveCluster(cluster)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}
	}

	services, err := m.getServices()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	for _, service := range services {
		if service.Addr == serviceAddr {
			st := status.New(codes.AlreadyExists, codes.AlreadyExists.String())

			return nil, st.Err()
		}
	}

	serviceID := uuid.New()

	service := &api.Service{
		Addr: serviceAddr,
		ID:   serviceID.String(),
		Type: serviceType.String(),
	}

	saveErr := m.saveService(service)

	if saveErr != nil {
		st := status.New(codes.Internal, saveErr.Error())

		return nil, st.Err()
	}

	endpoints := make([]string, 0)

	for _, service := range services {
		if service.Type == api.ServiceType_MANAGER.String() {
			endpoints = append(endpoints, service.Addr)
		}
	}

	res := &api.ManagerServiceInitResponse{
		ClusterID: cluster.ID,
		Endpoints: endpoints,
		ServiceID: service.ID,
	}

	return res, nil
}

func (m *Manager) removeService(id string) (*api.SuccessStatusResponse, error) {
	serviceID, err := uuid.Parse(id)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	cluster, err := m.getCluster()

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	if cluster == nil {
		st := status.New(codes.NotFound, "cluster_not_initialized")

		return nil, st.Err()
	}

	service, err := m.getServiceByID(serviceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	leaveAddr, err := net.ResolveTCPAddr("tcp", service.Addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(leaveAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	grpcCtx, grpcCancel := context.WithTimeout(context.Background(), grpcContextTimeout)

	defer grpcCancel()

	if service.Type == api.ServiceType_BLOCK.String() {
		peer := api.NewBlockClient(conn)

		opts := &api.BlockServiceLeaveRequest{
			ServiceID: service.ID,
		}

		_, err := peer.ServiceLeave(grpcCtx, opts)

		if err != nil {
			return nil, err
		}
	}

	if service.Type == api.ServiceType_MANAGER.String() {
		peer := api.NewManagerClient(conn)

		opts := &api.ManagerServiceLeaveRequest{
			ServiceID: service.ID,
		}

		_, err := peer.ServiceLeave(grpcCtx, opts)

		if err != nil {
			return nil, err
		}
	}

	serviceKey := fmt.Sprintf("/service/%s", service.ID)

	clientCtx, clientCancel := context.WithTimeout(context.Background(), etcdDialTimeout)

	defer clientCancel()

	config := clientv3.Config{
		DialTimeout: etcdDialTimeout,
		Endpoints:   m.flags.etcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return nil, err
	}

	defer client.Close()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	defer client.Close()

	_, delErr := client.Delete(clientCtx, serviceKey)

	if delErr != nil {
		st := status.New(codes.Internal, delErr.Error())

		return nil, st.Err()
	}

	res := &api.SuccessStatusResponse{Success: true}

	return res, nil
}

func (m *Manager) restart() error {
	cluster, err := m.getCluster()

	if err != nil {
		return err
	}

	if cluster != nil {
		service, err := m.getServiceByID(*m.Node.Key.ServiceID)

		if err != nil {
			return err
		}

		managerServiceType := api.ServiceType_MANAGER.String()

		if service != nil && service.Type == managerServiceType {
			restartSerfErr := m.Node.RestartSerf(m.flags.listenAddr)

			if restartSerfErr != nil {
				return restartSerfErr
			}

			managers := make([]*api.Service, 0)

			services, err := m.getServices()

			if err != nil {
				return err
			}

			for _, s := range services {

				if s.Type == managerServiceType {
					managers = append(managers, s)
				}

			}

			for _, manager := range managers {

				if manager.ID != service.ID {
					joinAddr, err := net.ResolveTCPAddr("tcp", manager.Addr)

					if err != nil {
						return err
					}

					conn, err := grpc.Dial(joinAddr.String(), grpc.WithInsecure())

					defer conn.Close()

					if err != nil {
						fields := logrus.Fields{
							"error": err.Error(),
						}

						logrus.WithFields(fields).Debug("Restart dial failed... waiting for next attempt.")

						time.Sleep(15 * time.Second)

						continue
					}

					ctx, cancel := context.WithTimeout(context.Background(), grpcContextTimeout)

					defer cancel()

					r := api.NewManagerClient(conn)

					opts := &api.ManagerServiceJoinRequest{
						ClusterID: cluster.ID,
						ServiceID: service.ID,
					}

					joinRes, err := r.ServiceJoin(ctx, opts)

					if err != nil {
						fields := logrus.Fields{
							"error": err.Error(),
						}

						logrus.WithFields(fields).Debug("Restart join failed... waiting for next attempt.")

						time.Sleep(15 * time.Second)

						continue
					}

					_, serfErr := m.Node.Serf.Join([]string{joinRes.SerfAddress}, true)

					if serfErr != nil {
						return serfErr
					}

					return nil
				}
			}
		}
	}

	return nil
}

func (m *Manager) saveCluster(cluster *api.Cluster) error {
	ctx, cancel := context.WithTimeout(context.Background(), etcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: etcdDialTimeout,
		Endpoints:   m.flags.etcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return err
	}

	defer client.Close()

	clusterJSON, err := json.Marshal(cluster)

	if err != nil {
		return err
	}

	_, putErr := client.Put(ctx, "/cluster", string(clusterJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (m *Manager) saveLogicalVolume(lv *api.LogicalVolume) error {
	ctx, cancel := context.WithTimeout(context.Background(), etcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: etcdDialTimeout,
		Endpoints:   m.flags.etcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return err
	}

	defer client.Close()

	lvJSON, err := json.Marshal(lv)

	if err != nil {
		return err
	}

	lvKey := fmt.Sprintf("/logicalvolume/%s", lv.ID)

	_, putErr := client.Put(ctx, lvKey, string(lvJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (m *Manager) savePhysicalVolume(pv *api.PhysicalVolume) error {
	ctx, cancel := context.WithTimeout(context.Background(), etcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: etcdDialTimeout,
		Endpoints:   m.flags.etcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return err
	}

	defer client.Close()

	pvJSON, err := json.Marshal(pv)

	if err != nil {
		return err
	}

	volumeKey := fmt.Sprintf("/physicalvolume/%s", pv.ID)

	_, putErr := client.Put(ctx, volumeKey, string(pvJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (m *Manager) saveVolumeGroup(vg *api.VolumeGroup) error {
	ctx, cancel := context.WithTimeout(context.Background(), etcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: etcdDialTimeout,
		Endpoints:   m.flags.etcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return err
	}

	defer client.Close()

	vgJSON, err := json.Marshal(vg)

	if err != nil {
		return err
	}

	volumeKey := fmt.Sprintf("/volumegroup/%s", vg.ID)

	_, putErr := client.Put(ctx, volumeKey, string(vgJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (m *Manager) saveService(service *api.Service) error {
	ctx, cancel := context.WithTimeout(context.Background(), etcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: etcdDialTimeout,
		Endpoints:   m.flags.etcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return err
	}

	defer client.Close()

	serviceKey := fmt.Sprintf("/service/%s", service.ID)

	serviceJSON, err := json.Marshal(service)

	if err != nil {
		return err
	}

	_, putErr := client.Put(ctx, serviceKey, string(serviceJSON))

	if putErr != nil {
		return err
	}

	return nil
}

// New : creates a new manager
func New() (*Manager, error) {
	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	if flags.verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	key, err := config.NewKey(flags.configDir)

	if err != nil {
		return nil, err
	}

	serfInstance, err := cluster.NewSerf(flags.listenAddr, key.ServiceID, api.ServiceType_MANAGER)

	if err != nil {
		return nil, err
	}

	node := config.Node{
		ErrorC: make(chan string),
		Key:    key,
		Serf:   serfInstance,
	}

	watchers := make(map[string]Watcher)

	for _, w := range etcdWatchResources {
		watchers[w] = Watcher{}
	}

	manager := &Manager{
		Node: node,

		flags:          flags,
		watchers:       watchers,
		watchersReady:  make([]string, 0),
		watchersReadyC: make(chan struct{}),
	}

	return manager, nil
}
