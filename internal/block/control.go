package block

import (
	"context"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServerControl struct {
	Block *Block
}

func (s *GRPCServerControl) ServiceNew(ctx context.Context, in *api.RequestServiceNew) (*api.ResponseServiceNew, error) {
	if in.APIServerAddr == "" {
		st := status.New(codes.InvalidArgument, "invalid_apiserver_addr")

		return nil, st.Err()
	}

	conn := service.NewClientConnTCP(in.APIServerAddr)

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), service.ContextTimeout)

	defer cancel()

	c := api.NewAPIServerClient(conn)

	healthServiceAddr := s.Block.Flags.HealthServiceAddr.String()

	options := &api.RequestNewService{
		HealthServiceAddr: healthServiceAddr,
		ServiceType:       api.ServiceType_BLOCK,
	}

	// TODO: replace _ with service response

	_, err := c.NewService(ctx, options)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	res := &api.ResponseServiceNew{}

	return res, nil
}
