package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/erkrnt/symphony/internal/pkg/gossip"
	"github.com/erkrnt/symphony/internal/pkg/state"
	"github.com/google/uuid"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Manager : manager node
type Manager struct {
	ErrorC chan string

	flags          *flags
	key            *config.Key
	memberlist     *memberlist.Memberlist
	watchers       []string
	watchersReady  []string
	watchersReadyC chan struct{}
}

var (
	ETCD_DIAL_TIMEOUT    = 5 * time.Second
	ETCD_WATCH_KEYS      = []string{"logicalvolume:assigned/*", "logicalvolume:unassigned/*"}
	GRPC_CONTEXT_TIMEOUT = 5 * time.Second
)

// Start : handles start of manager service
func (m *Manager) Start() {
	key, err := m.key.Get(m.flags.configDir)

	if err != nil {
		m.ErrorC <- err.Error()
	}

	if key.ClusterID != nil && key.ServiceID != nil {
		restartErr := m.restart(key)

		if restartErr != nil {
			m.ErrorC <- restartErr.Error()
		}
	}

	for _, key := range m.watchers {
		go m.startWatcherForKey(key)
	}

	<-m.watchersReadyC

	go m.listenControl()

	go m.listenRemote()
}

func (m *Manager) startWatcherForKey(key string) {
	// TODO : setup watch for state in etcd

	// NewLogicalVolume -> "logicalvolume:unassigned/*" -> Manager (watcher update) -> Begin scheduling

	// watch "logicalvolume:assigned/*" === lvAssignedC
	// watch "logicalvolume:unassigned/*" === lvUnassignedC

	m.watchersReady = append(m.watchersReady, key)

	if len(m.watchersReady) == len(m.watchers) {
		m.watchersReadyC <- struct{}{}
	}
}

func (m *Manager) listenControl() {
	socketPath := fmt.Sprintf("%s/control.sock", m.flags.configDir)

	if err := os.RemoveAll(socketPath); err != nil {
		m.ErrorC <- err.Error()
	}

	listen, err := net.Listen("unix", socketPath)

	if err != nil {
		m.ErrorC <- err.Error()
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
		m.ErrorC <- serveErr.Error()
	}
}

func (m *Manager) listenRemote() {
	listen, err := net.Listen("tcp", m.flags.listenServiceAddr.String())

	if err != nil {
		m.ErrorC <- err.Error()
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
		m.ErrorC <- serveErr.Error()
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

func (m *Manager) getMemberFromService(service *api.Service) (*gossip.Member, error) {
	var member *gossip.Member

	members := m.memberlist.Members()

	for _, m := range members {
		var metadata *gossip.Member

		if err := json.Unmarshal(m.Meta, &metadata); err != nil {
			st := status.New(codes.InvalidArgument, err.Error())

			return nil, st.Err()
		}

		if service.ID == metadata.ServiceID {
			member = metadata
		}
	}

	return member, nil
}

func (m *Manager) getStateByKey(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ETCD_DIAL_TIMEOUT)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: ETCD_DIAL_TIMEOUT,
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

	isLocalInit := serviceAddr == m.flags.listenServiceAddr.String()

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

	gossipAddr := m.flags.listenGossipAddr

	res := &api.ManagerServiceInitResponse{
		ClusterID:  cluster.ID,
		Endpoints:  endpoints,
		GossipAddr: gossipAddr.String(),
		ServiceID:  service.ID,
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

	grpcCtx, grpcCancel := context.WithTimeout(context.Background(), GRPC_CONTEXT_TIMEOUT)

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

	clientCtx, clientCancel := context.WithTimeout(context.Background(), ETCD_DIAL_TIMEOUT)

	defer clientCancel()

	config := clientv3.Config{
		DialTimeout: ETCD_DIAL_TIMEOUT,
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

func (m *Manager) restart(key *config.Key) error {
	cluster, err := m.getCluster()

	if err != nil {
		return err
	}

	if cluster != nil {
		service, err := m.getServiceByID(*key.ServiceID)

		if err != nil {
			return err
		}

		managerServiceType := api.ServiceType_MANAGER.String()

		if service != nil && service.Type == managerServiceType {
			gossipMember := &gossip.Member{
				ServiceAddr: m.flags.listenServiceAddr.String(),
				ServiceID:   service.ID,
				ServiceType: service.Type,
			}

			listenGossipAddr := m.flags.listenGossipAddr

			memberlist, err := gossip.NewMemberList(gossipMember, listenGossipAddr.Port)

			if err != nil {
				return err
			}

			m.memberlist = memberlist

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

					ctx, cancel := context.WithTimeout(context.Background(), GRPC_CONTEXT_TIMEOUT)

					defer cancel()

					r := api.NewManagerClient(conn)

					opts := &api.ManagerServiceJoinRequest{
						ClusterID: cluster.ID,
						ServiceID: service.ID,
					}

					join, err := r.ServiceJoin(ctx, opts)

					if err != nil {
						fields := logrus.Fields{
							"error": err.Error(),
						}

						logrus.WithFields(fields).Debug("Restart join failed... waiting for next attempt.")

						time.Sleep(15 * time.Second)

						continue
					}

					_, joinErr := memberlist.Join([]string{join.GossipAddr})

					if joinErr != nil {
						return err
					}

					return nil
				}
			}
		}
	}

	return nil
}

func (m *Manager) saveCluster(cluster *api.Cluster) error {
	ctx, cancel := context.WithTimeout(context.Background(), ETCD_DIAL_TIMEOUT)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: ETCD_DIAL_TIMEOUT,
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
	ctx, cancel := context.WithTimeout(context.Background(), ETCD_DIAL_TIMEOUT)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: ETCD_DIAL_TIMEOUT,
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
	ctx, cancel := context.WithTimeout(context.Background(), ETCD_DIAL_TIMEOUT)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: ETCD_DIAL_TIMEOUT,
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
	ctx, cancel := context.WithTimeout(context.Background(), ETCD_DIAL_TIMEOUT)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: ETCD_DIAL_TIMEOUT,
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
	ctx, cancel := context.WithTimeout(context.Background(), ETCD_DIAL_TIMEOUT)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: ETCD_DIAL_TIMEOUT,
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

	key := &config.Key{}

	manager := &Manager{
		ErrorC: make(chan string),

		flags:          flags,
		key:            key,
		watchers:       ETCD_WATCH_KEYS,
		watchersReady:  make([]string, 0),
		watchersReadyC: make(chan struct{}),
	}

	return manager, nil
}
