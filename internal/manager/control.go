package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/gossip"
	"github.com/google/uuid"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// controlServer : manager remote requests
type controlServer struct {
	Manager    *Manager
	Memberlist *memberlist.Memberlist
}

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

func GetServiceByID(services []*api.Service, id uuid.UUID) *api.Service {
	for _, service := range services {
		if service.Id == id.String() {
			return service
		}
	}

	return nil
}

func (s *controlServer) Init(ctx context.Context, in *api.ManagerControlInitRequest) (*api.ManagerControlInitResponse, error) {
	etcd, err := NewEtcdClient(s.Manager.Flags.EtcdEndpoints)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	services, err := s.Manager.GetServices()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	for _, service := range services {
		if service.Addr == s.Manager.Flags.ListenServiceAddr.String() {
			st := status.New(codes.AlreadyExists, codes.AlreadyExists.String())

			return nil, st.Err()
		}
	}

	serviceID := uuid.New()

	service := &api.Service{
		Addr: s.Manager.Flags.ListenServiceAddr.String(),
		Id:   serviceID.String(),
		Type: api.ServiceType_MANAGER.String(),
	}

	serviceKey := fmt.Sprintf("/service/%s", service.Id)

	serviceJSON, err := json.Marshal(service)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	_, putErr := etcd.Put(ctx, serviceKey, string(serviceJSON))

	if putErr != nil {
		st := status.New(codes.Internal, putErr.Error())

		return nil, st.Err()
	}

	gossipID := uuid.New()

	gossipMember := &gossip.Member{
		Id:          gossipID.String(),
		ServiceAddr: service.Addr,
		ServiceId:   service.Id,
		ServiceType: service.Type,
	}

	memberlist, err := gossip.NewMemberList(gossipMember, s.Manager.Flags.ListenGossipAddr.Port)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if len(services) > 0 {
		remoteService := services[0]

		conn, err := grpc.Dial(remoteService.Addr, grpc.WithInsecure())

		if err != nil {
			log.Fatal(err)
		}

		defer conn.Close()

		c := api.NewManagerRemoteClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		defer cancel()

		opts := &api.ManagerRemoteGossipRequest{
			ServiceId: service.Id,
		}

		gossipRes, err := c.Gossip(ctx, opts)

		logrus.Print(gossipRes.GossipAddr)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		_, joinErr := memberlist.Join([]string{gossipRes.GossipAddr})

		if joinErr != nil {
			st := status.New(codes.Internal, joinErr.Error())

			return nil, st.Err()
		}
	}

	s.Memberlist = memberlist

	res := &api.ManagerControlInitResponse{
		Id: service.Id,
	}

	return res, nil

}

func (s *controlServer) Remove(ctx context.Context, in *api.ManagerControlRemoveRequest) (*api.ManagerControlRemoveResponse, error) {
	etcd, err := NewEtcdClient(s.Manager.Flags.EtcdEndpoints)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	serviceID, err := uuid.Parse(in.Id)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	services, err := s.Manager.GetServices()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	service := GetServiceByID(services, serviceID)

	if service == nil {
		st := status.New(codes.NotFound, "service_not_found")

		return nil, st.Err()
	}

	serviceKey := fmt.Sprintf("/service/%s", service.Id)

	_, delErr := etcd.Delete(ctx, serviceKey)

	if delErr != nil {
		st := status.New(codes.Internal, delErr.Error())

		return nil, st.Err()
	}

	res := &api.ManagerControlRemoveResponse{}

	return res, nil
}
