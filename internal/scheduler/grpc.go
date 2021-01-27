package scheduler

import (
	"context"

	"github.com/erkrnt/symphony/api"
	"github.com/sirupsen/logrus"
)

// GRPCServerScheduler : Scheduler GRPC endpoints
type GRPCServerScheduler struct {
	Scheduler *Scheduler
}

func (s *GRPCServerScheduler) ResourceStatus(ctx context.Context, in *api.RequestResourceStatus) (*api.ResponseStatus, error) {
	logrus.Debug(in)

	res := &api.ResponseStatus{
		SUCCESS: true,
	}

	return res, nil
}
