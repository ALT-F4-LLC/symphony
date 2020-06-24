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

	flags      *flags
	key        *config.Key
	memberlist *memberlist.Memberlist
}

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

	go m.listenControl()

	go m.listenRemote()
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

	control := &controlServer{
		manager: m,
	}

	api.RegisterManagerControlServer(server, control)

	logrus.Info("Started manager control gRPC socket server.")

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

	remote := &remoteServer{
		manager: m,
	}

	server := grpc.NewServer()

	api.RegisterManagerRemoteServer(server, remote)

	logrus.Info("Started manager remote gRPC tcp server.")

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

func (m *Manager) getStateByKey(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   m.flags.etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	results, err := etcd.Get(ctx, key, opts...)

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
					managers = append(services, s)
				}

			}

			for _, manager := range managers {

				if manager.ID != service.ID {
					joinAddr, err := net.ResolveTCPAddr("tcp", manager.Addr)

					if err != nil {
						return err
					}

					conn, err := grpc.Dial(joinAddr.String(), grpc.WithInsecure())

					if err != nil {
						logrus.Debug("Restart join failed to remote peer... skipping.")
						continue
					}

					defer conn.Close()

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

					defer cancel()

					r := api.NewManagerRemoteClient(conn)

					opts := &api.ManagerRemoteJoinRequest{
						ClusterID: cluster.ID,
						ServiceID: service.ID,
					}

					join, err := r.Join(ctx, opts)

					if err != nil {
						logrus.Debug("Restart join failed to remote peer... skipping.")
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
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   m.flags.etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return err
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	clusterJSON, err := json.Marshal(cluster)

	if err != nil {
		return err
	}

	_, putErr := etcd.Put(ctx, "/cluster", string(clusterJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (m *Manager) saveLogicalVolume(lv *api.LogicalVolume) error {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   m.flags.etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return err
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	lvJSON, err := json.Marshal(lv)

	if err != nil {
		return err
	}

	lvKey := fmt.Sprintf("/logicalvolume/%s", lv.ID)

	_, putErr := etcd.Put(ctx, lvKey, string(lvJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (m *Manager) savePhysicalVolume(pv *api.PhysicalVolume) error {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   m.flags.etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return err
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	pvJSON, err := json.Marshal(pv)

	if err != nil {
		return err
	}

	volumeKey := fmt.Sprintf("/physicalvolume/%s", pv.ID)

	_, putErr := etcd.Put(ctx, volumeKey, string(pvJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (m *Manager) saveVolumeGroup(vg *api.VolumeGroup) error {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   m.flags.etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return err
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	vgJSON, err := json.Marshal(vg)

	if err != nil {
		return err
	}

	volumeKey := fmt.Sprintf("/volumegroup/%s", vg.ID)

	_, putErr := etcd.Put(ctx, volumeKey, string(vgJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (m *Manager) saveService(service *api.Service) error {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   m.flags.etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return err
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	serviceKey := fmt.Sprintf("/service/%s", service.ID)

	serviceJSON, err := json.Marshal(service)

	if err != nil {
		return err
	}

	_, putErr := etcd.Put(ctx, serviceKey, string(serviceJSON))

	if putErr != nil {
		return err
	}

	return nil
}

// New : creates a new manager
func New() (*Manager, error) {
	errorC := make(chan string)

	key := &config.Key{}

	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	if flags.verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	manager := &Manager{
		ErrorC: errorC,

		flags: flags,
		key:   key,
	}

	return manager, nil
}
