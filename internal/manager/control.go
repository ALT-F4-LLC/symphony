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

// controlServer : manager remote requests
type controlServer struct {
	Manager *Manager
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

func GetServiceByID(services []*api.Service, id string) *api.Service {
	for _, service := range services {
		if service.Id == id {
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
		if service.Addr == s.Manager.Flags.ListenAddr.String() {
			st := status.New(codes.AlreadyExists, codes.AlreadyExists.String())

			return nil, st.Err()
		}
	}

	serviceID := uuid.New()

	service := api.Service{
		Addr: s.Manager.Flags.ListenAddr.String(),
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

	service := GetServiceByID(services, serviceID.String())

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
