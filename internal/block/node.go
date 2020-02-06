package block

import (
	"log"
	"net"

	"github.com/erkrnt/symphony/api"
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
func NewNode(flags *config.Flags) (*Node, error) {
	key, err := config.GetKey(flags.ConfigDir)

	if err != nil {
		return nil, err
	}

	var peers []string

	if flags.JoinAddr != nil {
		p, err := cluster.JoinRaft(flags, key)

		if err != nil {
			return nil, err
		}

		peers = p
	}

	node, store := cluster.NewRaft(flags, key, peers)

	m := &Node{
		Key:   key,
		Raft:  node,
		State: store,
	}

	return m, nil
}

// Start : starts Raft memebership server
func Start(flags *config.Flags, node *Node) {
	lis, err := net.Listen("tcp", flags.ListenRemoteAddr.String())

	if err != nil {
		log.Fatal("Failed to listen")
	}

	s := grpc.NewServer()

	server := &blockServer{
		Node: node,
	}

	api.RegisterBlockServer(s, server)

	log.Print("Started block gRPC endpoints.")

	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve")
	}
}
