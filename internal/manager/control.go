package manager

import (
	"context"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type control struct {
	*endpoints

	manager *Manager
}

func (c *control) ServiceInit(ctx context.Context, in *api.ManagerServiceInitRequest) (*api.ManagerServiceInitResponse, error) {
	serviceAddr := c.manager.flags.listenServiceAddr.String()

	serviceType := api.ServiceType_MANAGER

	res, err := c.manager.newService(serviceAddr, serviceType)

	if err != nil {
		return nil, err
	}

	clusterID, err := uuid.Parse(res.ClusterID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(res.ServiceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	key := &config.Key{
		ClusterID: &clusterID,
		ServiceID: &serviceID,
	}

	saveErr := key.Save(c.manager.flags.configDir)

	if saveErr != nil {
		st := status.New(codes.Internal, saveErr.Error())

		return nil, st.Err()
	}

	restartErr := c.manager.restart(key)

	if restartErr != nil {
		st := status.New(codes.Internal, restartErr.Error())

		return nil, st.Err()
	}

	return res, nil
}
