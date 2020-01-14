package manager

import (
	"net"

	"github.com/erkrnt/symphony/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RaftMembershipServer : struct used for grpc service
type RaftMembershipServer struct {
	Manager *Manager
}

// StartRaftMembershipServer : starts Raft memebership server
func StartRaftMembershipServer(m *Manager) {
	lis, err := net.Listen("tcp", m.Flags.ListenRemoteAPI.String())

	if err != nil {
		m.Logger.Fatal("Failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()

	server := &RaftMembershipServer{
		Manager: m,
	}

	api.RegisterRaftMembershipServer(s, server)

	m.Logger.Info(
		"Started manager gRPC endpoints.",
		zap.String("listen-remote-api", m.Flags.ListenRemoteAPI.String()),
	)

	if err := s.Serve(lis); err != nil {
		m.Logger.Fatal("Failed to serve", zap.Error(err))
	}
}
