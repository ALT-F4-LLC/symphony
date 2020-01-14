package manager

import (
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// RaftMembershipServer : struct used for grpc service
type RaftMembershipServer struct {
	Member *cluster.Member
}

// StartRaftMembershipServer : starts Raft memebership server
func StartRaftMembershipServer(addr *net.TCPAddr, node *cluster.Member) {
	lis, err := net.Listen("tcp", addr.String())

	if err != nil {
		logrus.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	server := &RaftMembershipServer{
		Member: node,
	}

	api.RegisterRaftMembershipServer(s, server)

	logFields := logrus.Fields{"listen-remote-api": addr.String()}

	logrus.WithFields(logFields).Info("Started manager gRPC endpoints.")

	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("Failed to serve: %v", err)
	}
}
