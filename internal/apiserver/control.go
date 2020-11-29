package apiserver

import (
	"context"
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServerControl : control server struct
type GRPCServerControl struct {
	APIServer *APIServer
}

// ServiceNew : handles creation of service from control
func (s GRPCServerControl) ServiceNew(ctx context.Context, in *api.RequestServiceNew) (*api.ResponseServiceNew, error) {
	if in.APIServerAddr == "" {
		st := status.New(codes.InvalidArgument, "invalid_apiserver_addr")

		return nil, st.Err()
	}

	options := NewServiceOptions{
		ClusterID:   in.ClusterID,
		ServiceType: in.ServiceType,
	}

	newService, err := s.APIServer.Resources.newService(options)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	clusterID, err := uuid.Parse(newService.ClusterID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(newService.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	key := &service.Key{
		ClusterID: &clusterID,
		ServiceID: &serviceID,
	}

	configDir := s.APIServer.Flags.ConfigDir

	saveErr := s.APIServer.Service.Key.Save(configDir, key)

	if saveErr != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	peerAddr, err := net.ResolveTCPAddr("tcp", in.APIServerAddr)

	if err != nil {
		return nil, err
	}

	restartErr := s.APIServer.restart(peerAddr)

	if restartErr != nil {
		st := status.New(codes.Internal, restartErr.Error())

		return nil, st.Err()
	}

	res := &api.ResponseServiceNew{
		ClusterID: newService.ClusterID,
		ServiceID: newService.ID,
	}

	return res, nil
}
