package manager

import (
	"context"

	"github.com/erkrnt/symphony/api"
)

type remote struct {
	*endpoints

	manager *Manager
}

func (r *remote) ServiceInit(ctx context.Context, in *api.ManagerServiceInitRequest) (*api.ManagerServiceInitResponse, error) {
	res, err := r.manager.newService(in.ServiceAddr, in.ServiceType)

	if err != nil {
		return nil, err
	}

	return res, nil
}
