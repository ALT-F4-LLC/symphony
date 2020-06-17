package manager

import (
	"context"
	"net"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/erkrnt/symphony/internal/pkg/gossip"
	"github.com/google/uuid"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// controlServer : manager remote requests
type controlServer struct {
	manager *Manager
}

// NewEtcdClient : creates new etcd client
func NewEtcdClient(endpoints []string) (*clientv3.Client, error) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	return etcd, nil
}

// GetServiceByID : gets service from services
func GetServiceByID(services []*api.Service, id uuid.UUID) *api.Service {
	for _, service := range services {
		if service.ID == id.String() {
			return service
		}
	}

	return nil
}

func (s *controlServer) Init(ctx context.Context, in *api.ManagerControlInitRequest) (*api.ManagerControlInitResponse, error) {
	defaultInitAddr := s.manager.flags.listenServiceAddr.String()

	if in.ServiceAddr != "" {
		defaultInitAddr = in.ServiceAddr
	}

	initAddr, err := net.ResolveTCPAddr("tcp", defaultInitAddr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(initAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	r := api.NewManagerRemoteClient(conn)

	opts := &api.ManagerRemoteInitRequest{
		ServiceAddr: s.manager.flags.listenServiceAddr.String(),
		ServiceType: api.ServiceType_MANAGER,
	}

	init, err := r.Init(ctx, opts)

	if err != nil {
		return nil, err
	}

	clusterID, err := uuid.Parse(init.ClusterID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(init.ServiceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	key := &config.Key{
		ClusterID: &clusterID,
		Endpoints: init.Endpoints,
		ServiceID: &serviceID,
	}

	saveErr := key.Save(s.manager.flags.configDir)

	if saveErr != nil {
		st := status.New(codes.Internal, saveErr.Error())

		return nil, st.Err()
	}

	gossipMember := &gossip.Member{
		ServiceAddr: opts.ServiceAddr,
		ServiceID:   serviceID.String(),
		ServiceType: opts.ServiceType.String(),
	}

	listenGossipAddr := s.manager.flags.listenGossipAddr

	memberlist, err := gossip.NewMemberList(gossipMember, listenGossipAddr.Port)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if init.GossipAddr != listenGossipAddr.String() {
		_, joinErr := memberlist.Join([]string{init.GossipAddr})

		if joinErr != nil {
			st := status.New(codes.Internal, joinErr.Error())

			return nil, st.Err()
		}
	}

	s.manager.memberlist = memberlist

	res := &api.ManagerControlInitResponse{
		ServiceID: serviceID.String(),
	}

	return res, nil
}

func (s *controlServer) Remove(ctx context.Context, in *api.ManagerControlRemoveRequest) (*api.SuccessStatusResponse, error) {
	serviceID, err := uuid.Parse(in.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	removeAddr, err := net.ResolveTCPAddr("tcp", s.manager.flags.listenServiceAddr.String())

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(removeAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	remote := api.NewManagerRemoteClient(conn)

	opts := &api.ManagerRemoteRemoveRequest{
		ServiceID: serviceID.String(),
	}

	res, err := remote.Remove(ctx, opts)

	if err != nil {
		return nil, err
	}

	return res, nil
}
