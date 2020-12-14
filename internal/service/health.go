package service

import (
	"context"

	"github.com/erkrnt/symphony/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServerHealth struct{}

func (s *GRPCServerHealth) Check(ctx context.Context, in *api.HealthCheckRequest) (*api.HealthCheckResponse, error) {
	checkStatus := api.HealthCheckResponse_SERVING

	res := &api.HealthCheckResponse{
		Status: checkStatus,
	}

	return res, nil
}

func (s *GRPCServerHealth) Watch(in *api.HealthCheckRequest, stream api.Health_WatchServer) error {
	checkStatus := api.HealthCheckResponse_SERVING

	res := &api.HealthCheckResponse{
		Status: checkStatus,
	}

	streamErr := stream.Send(res)

	if streamErr != nil {
		st := status.New(codes.Internal, streamErr.Error())

		return st.Err()
	}

	return nil
}
