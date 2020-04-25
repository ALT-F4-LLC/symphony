package manager

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/raft/raftpb"
	"google.golang.org/grpc"
)

type remoteServer struct {
	Manager *Manager
}

// Join : adds nodes to the raft
func (s *remoteServer) Join(ctx context.Context, in *api.ManagerRemoteJoinRequest) (*api.ManagerRemoteJoinResponse, error) {
	if s.Manager.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	members, err := s.Manager.GetRaftMembers()

	if err != nil {
		return nil, err
	}

	_, member := GetRaftMemberByAddr(members, in.Addr)

	if member != nil {
		return nil, errors.New("invalid_raft_member")
	}

	// need a uniq ident for node ids so we will use the
	// raft commit index as it is also uniq and an index
	commitIndex := s.Manager.Raft.Node.Status().Commit

	added := &api.RaftMember{
		Addr: in.Addr,
		Id:   uint64(commitIndex),
	}

	cc := raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  added.Id,
		Context: []byte(in.Addr),
	}

	s.Manager.Raft.ConfChangeC <- cc

	log.Printf("New member proposed to cluster %d", added.Id)

	members = append(members, added)

	json, err := json.Marshal(members)

	if err != nil {
		return nil, err
	}

	s.Manager.State.Propose("members", string(json))

	res := &api.ManagerRemoteJoinResponse{
		RaftId:      added.Id,
		RaftMembers: members,
	}

	return res, nil
}

// StartRemoteServer : starts Raft membership server
func StartRemoteServer(m *Manager) {
	lis, err := net.Listen("tcp", m.Flags.ListenRemoteAddr.String())

	if err != nil {
		log.Fatal("Failed to listen")
	}

	server := grpc.NewServer()

	remote := &remoteServer{
		Manager: m,
	}

	logrus.Info("Started manager remote gRPC tcp server.")

	api.RegisterManagerRemoteServer(server, remote)

	if err := server.Serve(lis); err != nil {
		log.Fatal("Failed to serve")
	}
}
