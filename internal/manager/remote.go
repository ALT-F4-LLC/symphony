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
	manager *Manager
}

func (s *remoteServer) Init(ctx context.Context, in *api.ManagerRemoteInitRequest) (*api.ManagerRemoteInitResponse, error) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   s.manager.flags.etcdEndpoints,
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

		if srvc.Addr == in.ServiceAddr {
			st := status.New(codes.AlreadyExists, codes.AlreadyExists.String())

			return nil, st.Err()
		}
	}

	serviceID := uuid.New()

	service := api.Service{
		Addr: in.ServiceAddr,
		ID:   serviceID.String(),
		Type: in.ServiceType.String(),
	}

	serviceKey := fmt.Sprintf("/service/%s", service.ID)

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

	gossipAddr := s.manager.flags.listenGossipAddr

	res := &api.ManagerRemoteInitResponse{
		GossipAddr: gossipAddr.String(),
		ServiceID:  service.ID,
	}

	return res, nil
}

func (s *remoteServer) Leave(ctx context.Context, in *api.ManagerRemoteLeaveRequest) (*api.ManagerRemoteLeaveResponse, error) {
	res := &api.ManagerRemoteLeaveResponse{}

	return res, nil
}
