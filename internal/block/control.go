package block

import (
	"context"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServerControl : control service struct
type GRPCServerControl struct {
	Block *Block
}

// ServiceNew : handles creating a new service definition
func (s *GRPCServerControl) ServiceNew(ctx context.Context, in *api.RequestServiceNew) (*api.ResponseServiceNew, error) {
	if s.Block.Key.ServiceID != nil {
		st := status.New(codes.AlreadyExists, "service_already_exists")

		return nil, st.Err()
	}

	apiserverAddr := s.Block.Flags.APIServerAddr

	if apiserverAddr == nil {
		st := status.New(codes.InvalidArgument, "invalid_apiserver_addr")

		return nil, st.Err()
	}

	conn := service.NewClientConnTCP(apiserverAddr.String())

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), service.ContextTimeout)

	defer cancel()

	c := api.NewAPIServerClient(conn)

	healthServiceAddr := s.Block.Flags.HealthServiceAddr.String()

	newServiceOptions := &api.RequestNewService{
		HealthServiceAddr: healthServiceAddr,
		ServiceType:       api.ServiceType_BLOCK,
	}

	newService, err := c.NewService(ctx, newServiceOptions)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(newService.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	saveKeyOptions := service.SaveKeyOptions{
		ConfigDir: s.Block.Flags.ConfigDir,
		ServiceID: serviceID,
	}

	saveErr := s.Block.Key.Save(saveKeyOptions)

	if saveErr != nil {
		st := status.New(codes.Internal, saveErr.Error())

		return nil, st.Err()
	}

	res := &api.ResponseServiceNew{
		ServiceID: newService.ID,
	}

	go s.Block.listenGRPC()

	return res, nil
}
