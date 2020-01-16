package manager

import (
	"context"

	"github.com/erkrnt/symphony/api"
)

type managerServer struct {
	Node *Node
}

func nodeIsMember(addr string, peers []string) bool {
	for _, a := range peers {
		if a == addr {
			return true
		}
	}

	return false
}

func (s *managerServer) Join(ctx context.Context, in *api.JoinRequest) (*api.JoinResponse, error) {
	// TODO: Take in the server address and lookup in kv

	// isMember := nodeIsMember(in.Addr, s.Member.Raft.)

	// logrus.Debug(isMember)

	// if !ok {
	// 	return nil, errors.New("store lookup failure")
	// }

	return &api.JoinResponse{}, nil
}

func (s *managerServer) Leave(ctx context.Context, in *api.LeaveRequest) (*api.LeaveResponse, error) {
	return &api.LeaveResponse{}, nil
}
