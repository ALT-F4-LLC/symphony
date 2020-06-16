package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type remoteServer struct {
	Manager *Manager
}

func (s *remoteServer) Gossip(ctx context.Context, in *api.ManagerRemoteGossipRequest) (*api.ManagerRemoteGossipResponse, error) {
	services, err := s.Manager.GetServices()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(in.ServiceId)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	service := GetServiceByID(services, serviceID)

	if service == nil {
		st := status.New(codes.NotFound, "invalid_service_id")

		return nil, st.Err()
	}

	if service.Type != api.ServiceType_MANAGER.String() {
		st := status.New(codes.PermissionDenied, "invalid_service_type")

		return nil, st.Err()
	}

	res := &api.ManagerRemoteGossipResponse{
		GossipAddr: s.Manager.Flags.ListenGossipAddr.String(),
	}

	return res, nil
}

func (s *remoteServer) Init(ctx context.Context, in *api.ManagerRemoteInitRequest) (*api.ManagerRemoteInitResponse, error) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   s.Manager.Flags.EtcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	services, err := etcd.Get(ctx, "/service", clientv3.WithPrefix())

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	for _, ev := range services.Kvs {
		var srvc api.Service

		err := json.Unmarshal(ev.Value, &srvc)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		if srvc.Addr == in.Addr {
			st := status.New(codes.AlreadyExists, codes.AlreadyExists.String())

			return nil, st.Err()
		}
	}

	serviceID := uuid.New()

	service := api.Service{
		Addr: in.Addr,
		Id:   serviceID.String(),
		Type: in.Type.String(),
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

	res := &api.ManagerRemoteInitResponse{
		GossipAddr: s.Manager.Flags.ListenGossipAddr.String(),
		Id:         service.Id,
	}

	return res, nil
}
