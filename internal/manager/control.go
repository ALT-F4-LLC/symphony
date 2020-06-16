package manager

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/erkrnt/symphony/internal/pkg/gossip"
	"github.com/google/uuid"
	"github.com/hashicorp/memberlist"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// controlServer : manager remote requests
type controlServer struct {
	manager    *Manager
	memberlist *memberlist.Memberlist
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

	if in.Addr != "" {
		defaultInitAddr = in.Addr
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	opts := &api.ManagerRemoteInitRequest{
		ServiceAddr: s.manager.flags.listenServiceAddr.String(),
		ServiceType: api.ServiceType_MANAGER,
	}

	init, err := r.Init(ctx, opts)

	if err != nil {
		return nil, err
	}

	serviceID, err := uuid.Parse(init.ServiceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	key := &config.Key{
		ServiceID: serviceID,
	}

	saveErr := key.Save(s.manager.flags.configPath)

	if saveErr != nil {
		st := status.New(codes.Internal, saveErr.Error())

		return nil, st.Err()
	}

	gossipID := uuid.New()

	gossipMember := &gossip.Member{
		ServiceAddr: opts.ServiceAddr,
		ServiceID:   serviceID.String(),
		ServiceType: opts.ServiceType.String(),
	}

	listenGossipAddr := s.manager.flags.listenGossipAddr

	memberlist, err := gossip.NewMemberList(gossipID, gossipMember, listenGossipAddr.Port)

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

	s.memberlist = memberlist

	res := &api.ManagerControlInitResponse{
		ServiceID: serviceID.String(),
	}

	return res, nil
}

func (s *controlServer) Leave(ctx context.Context, in *api.ManagerControlLeaveRequest) (*api.ManagerControlLeaveResponse, error) {
	res := &api.ManagerControlLeaveResponse{}

	return res, nil
}

func (s *controlServer) Remove(ctx context.Context, in *api.ManagerControlRemoveRequest) (*api.ManagerControlRemoveResponse, error) {
	etcd, err := NewEtcdClient(s.manager.flags.etcdEndpoints)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	serviceID, err := uuid.Parse(in.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	services, err := s.manager.GetServices()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	service := GetServiceByID(services, serviceID)

	if service == nil {
		st := status.New(codes.NotFound, "service_not_found")

		return nil, st.Err()
	}

	serviceKey := fmt.Sprintf("/service/%s", service.ID)

	_, delErr := etcd.Delete(ctx, serviceKey)

	if delErr != nil {
		st := status.New(codes.Internal, delErr.Error())

		return nil, st.Err()
	}

	res := &api.ManagerControlRemoveResponse{}

	return res, nil
}
