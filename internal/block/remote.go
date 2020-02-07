package block

import (
	"log"
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// RemoteServer : block remote requests
type RemoteServer struct {
	Node *cluster.Node
}

// StartRemoteServer : starts Raft memebership server
func StartRemoteServer(node *cluster.Node) {
	lis, err := net.Listen("tcp", node.Flags.ListenRemoteAddr.String())

	if err != nil {
		log.Fatal("Failed to listen")
	}

	s := grpc.NewServer()

	server := &RemoteServer{
		Node: node,
	}

	logrus.Info("Started block remote gRPC tcp server.")

	api.RegisterBlockRemoteServer(s, server)

	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve")
	}
}
