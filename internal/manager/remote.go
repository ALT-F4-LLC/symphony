package manager

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/raft/raftpb"
	"google.golang.org/grpc"
)

type remoteServer struct {
	Node *cluster.Node
}

// Join : adds nodes to the raft
func (s *remoteServer) Join(ctx context.Context, in *api.ManagerRemoteJoinRequest) (*api.ManagerRemoteJoinResponse, error) {
	if s.Node.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	members, err := s.Node.GetRaftMembers()

	if err != nil {
		return nil, err
	}

	_, member := cluster.GetRaftMemberByAddr(members, in.Addr)

	if member != nil {
		return nil, errors.New("invalid_raft_member")
	}

	// need a uniq ident for node ids so we will use the
	// raft commit index as it is also uniq and an index
	commitIndex := s.Node.Raft.Node.Status().Commit

	added := &api.RaftMember{
		Addr: in.Addr,
		Id:   uint64(commitIndex),
	}

	cc := raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  added.Id,
		Context: []byte(in.Addr),
	}

	s.Node.Raft.ConfChangeC <- cc

	log.Printf("New member proposed to cluster %d", added.Id)

	members = append(members, added)

	addedGossipID := uuid.New().String()

	json, err := json.Marshal(members)

	if err != nil {
		return nil, err
	}

	s.Node.State.Propose("members", string(json))

	res := &api.ManagerRemoteJoinResponse{
		GossipAddr:  s.Node.Flags.ListenGossipAddr.String(),
		GossipId:    addedGossipID,
		RaftId:      added.Id,
		RaftMembers: members,
	}

	return res, nil
}

// StartRemoteServer : starts Raft membership server
func StartRemoteServer(node *cluster.Node) {
	lis, err := net.Listen("tcp", node.Flags.ListenRemoteAddr.String())

	if err != nil {
		log.Fatal("Failed to listen")
	}

	server := grpc.NewServer()

	remote := &remoteServer{
		Node: node,
	}

	logrus.Info("Started manager remote gRPC tcp server.")

	api.RegisterManagerRemoteServer(server, remote)

	if err := server.Serve(lis); err != nil {
		log.Fatal("Failed to serve")
	}
}
