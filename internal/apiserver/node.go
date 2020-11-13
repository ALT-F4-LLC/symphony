package apiserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/erkrnt/symphony/internal/pkg/state"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

// APIServer : defines apiserver service
type APIServer struct {
	Flags     *Flags
	Node      *config.Node
	Resources *Resources
}

const (
	EtcdDialTimeout    = 5 * time.Second
	GrpcContextTimeout = 5 * time.Second
)

// New : creates a new apiserver instance
func New() (*APIServer, error) {
	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	if flags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	key, err := config.GetKey(flags.ConfigDir)

	if err != nil {
		return nil, err
	}

	serviceType := api.ServiceType_APISERVER

	serf, err := cluster.NewSerf(flags.ListenAddr, key.ServiceID, serviceType)

	if err != nil {
		return nil, err
	}

	resources := &Resources{
		EtcdEndpoints: flags.EtcdEndpoints,
	}

	node := &config.Node{
		ErrorC: make(chan string),
		Key:    key,
		Serf:   serf,
	}

	apiserver := &APIServer{
		Flags:     flags,
		Node:      node,
		Resources: resources,
	}

	return apiserver, nil
}

// Start : handles start of apiserver service
func (s *APIServer) Start() {
	if s.Node.Key.ClusterID != nil && s.Node.Key.JoinAddr != nil && s.Node.Key.ServiceID != nil {
		restartErr := s.restart(s.Node.Key.JoinAddr)

		if restartErr != nil {
			s.Node.ErrorC <- restartErr.Error()
		}
	}

	go s.listenRemote()
}

func (s *APIServer) listenRemote() {
	listen, err := net.Listen("tcp", s.Flags.ListenAddr.String())

	if err != nil {
		s.Node.ErrorC <- err.Error()
	}

	server := grpc.NewServer()

	apiserver := &GRPCServerAPIServer{
		APIServer: s,
	}

	api.RegisterAPIServerServer(server, apiserver)

	logrus.Info("Started apiserver gRPC server.")

	serveErr := server.Serve(listen)

	if serveErr != nil {
		s.Node.ErrorC <- serveErr.Error()
	}
}

func (s *APIServer) removeResource(name string, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), EtcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: EtcdDialTimeout,
		Endpoints:   s.Flags.EtcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return err
	}

	defer client.Close()

	lv, err := s.Resources.getLogicalVolumeByID(id)

	if err != nil {
		return err
	}

	etcdKey := fmt.Sprintf("/logicalvolume/%s", lv.ID)

	_, delRes := client.Delete(ctx, etcdKey)

	if delRes != nil {
		return delRes
	}

	return nil
}

func (s *APIServer) restart(joinAddr *net.TCPAddr) error {
	cluster, err := s.Resources.getClusterByID(*s.Node.Key.ClusterID)

	if err != nil {
		return err
	}

	service, err := s.Resources.getServiceByID(*s.Node.Key.ServiceID)

	if err != nil {
		return err
	}

	if cluster != nil && joinAddr != nil && service != nil {
		serfError := s.Node.RestartSerf(s.Flags.ListenAddr, service.Type)

		if serfError != nil {
			return serfError
		}

		conn, err := grpc.Dial(joinAddr.String(), grpc.WithInsecure())

		if err != nil {
			return err
		}

		defer conn.Close()

		ctx, cancel := context.WithTimeout(context.Background(), GrpcContextTimeout)

		defer cancel()

		r := api.NewAPIServerClient(conn)

		opts := &api.RequestServiceJoin{
			ClusterID: cluster.ID,
			ServiceID: service.ID,
		}

		res, err := r.ServiceJoin(ctx, opts)

		if err != nil {
			return err
		}

		_, serfErr := s.Node.Serf.Join([]string{res.PeerAddr}, true)

		if serfErr != nil {
			return serfErr
		}
	}

	return nil
}
