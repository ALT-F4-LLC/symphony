package manager

import (
	"context"
	"log"
	"net"

	"github.com/erkrnt/symphony/api"
	"google.golang.org/grpc"
)

// RaftMembershipServer : struct used for grpc service
type RaftMembershipServer struct {
	Manager *Manager
}

// NodeIsMember : checks if a specific node is a peer in raft
func NodeIsMember(addr string, peers []string) bool {
	for _, a := range peers {
		if a == addr {
			return true
		}
	}

	return false
}

// StartRaftMembershipServer : starts Raft memebership server
func StartRaftMembershipServer(m *Manager) {
	lis, err := net.Listen("tcp", m.Flags.ListenRemoteAPI.String())

	if err != nil {
		log.Fatal("Failed to listen")
	}

	s := grpc.NewServer()

	server := &RaftMembershipServer{
		Manager: m,
	}

	api.RegisterRaftMembershipServer(s, server)

	log.Print("Started manager gRPC endpoints.")

	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve")
	}
}

// Join : handles joining servers to raft members
func (s *RaftMembershipServer) Join(ctx context.Context, in *api.JoinRequest) (*api.JoinResponse, error) {
	// TODO: Take in the server address and lookup in kv

	// isMember := NodeIsMember(in.Addr, s.Member.Raft.)

	// logrus.Debug(isMember)

	// if !ok {
	// 	return nil, errors.New("store lookup failure")
	// }

	return &api.JoinResponse{}, nil
}

// Leave : handles removing servers from raft members
func (s *RaftMembershipServer) Leave(ctx context.Context, in *api.LeaveRequest) (*api.LeaveResponse, error) {
	return &api.LeaveResponse{}, nil
}
