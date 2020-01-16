package block

import (
	"log"
	"net"

	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"google.golang.org/grpc"
)

// Node : block node
type Node struct {
	Key   *config.Key
	Raft  *cluster.RaftNode
	State *cluster.RaftState
}

// NewNode : creates a new manager struct
func NewNode(f *config.Flags) (*Node, error) {
	k, err := config.GetKey(f.ConfigDir)

	if err != nil {
		return nil, err
	}

	if f.JoinAddr != nil {
		log.Print("hello")
	}

	// node, store := cluster.NewRaft(configDir)

	m := &Node{
		Key: k,
		// Raft:  node,
		// State: store,
	}

	return m, nil
}

// Start : starts Raft memebership server
func Start(f *config.Flags, n *Node) {
	lis, err := net.Listen("tcp", f.ListenRemoteAPI.String())

	if err != nil {
		log.Fatal("Failed to listen")
	}

	s := grpc.NewServer()

	// server := &blockServer{
	// 	Node: n,
	// }

	// api.RegisterBlockServer(s, server)

	log.Print("Started block gRPC endpoints.")

	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve")
	}
}
