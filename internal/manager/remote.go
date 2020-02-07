package manager

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/raft/raftpb"
	"google.golang.org/grpc"
)

// RemoteServer : manager remote requests
type RemoteServer struct {
	Node *cluster.Node
}

// Init : adds nodes to the raft
func (s *RemoteServer) Init(ctx context.Context, in *api.ManagerRemoteInitReq) (*api.ManagerRemoteInitRes, error) {
	if in.Addr == "" {
		return nil, errors.New("invalid_raft_addr")
	}

	if s.Node.RaftNode == nil {
		return nil, errors.New("invalid_raft_state")
	}

	_, ok := s.Node.RaftState.Lookup("members")

	if ok {
		return nil, errors.New("invalid_raft_members_state")
	}

	member, _, err := cluster.GetMemberByAddr(in.Addr, s.Node.RaftNode.Members)

	if err != nil {
		return nil, errors.New("invalid_raft_member") // does not exist in initd cluster peers
	}

	res := &api.ManagerRemoteInitRes{
		MemberId: member.ID,
		Members:  s.Node.RaftNode.Members,
	}

	return res, nil
}

// Join : adds nodes to the raft
func (s *RemoteServer) Join(ctx context.Context, in *api.ManagerRemoteJoinReq) (*api.ManagerRemoteJoinRes, error) {
	if s.Node.RaftNode == nil {
		return nil, errors.New("invalid_raft_state")
	}

	var members []*api.Member

	m, ok := s.Node.RaftState.Lookup("members")

	if !ok {
		members = s.Node.RaftNode.Members
	} else {
		err := json.Unmarshal([]byte(m), &members)

		if err != nil {
			return nil, err
		}
	}

	_, _, err := cluster.GetMemberByAddr(in.Addr, members)

	if err == nil {
		return nil, errors.New("invalid_raft_member")
	}

	// need a uniq ident for node ids so we will use the
	// raft commit index as it is also uniq and an index
	commitIndex := s.Node.RaftNode.Node.Status().Commit

	memberID := uint64(commitIndex)

	cc := raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  memberID,
		Context: []byte(in.Addr),
	}

	s.Node.RaftNode.ConfChangeC <- cc

	log.Printf("New member proposed to cluster %d", memberID)

	res := &api.ManagerRemoteJoinRes{
		MemberId: memberID,
		Members:  members,
	}

	members = append(members, &api.Member{ID: memberID, Addr: in.Addr})

	json, err := json.Marshal(members)

	if err != nil {
		return nil, err
	}

	s.Node.RaftState.Propose("members", string(json))

	return res, nil
}

// Remove : adds nodes to the raft
func (s *RemoteServer) Remove(ctx context.Context, in *api.ManagerRemoteRemoveReq) (*api.ManagerRemoteRemoveRes, error) {
	if s.Node.RaftNode == nil {
		return nil, errors.New("invalid_raft_state")
	}

	var members []*api.Member

	m, ok := s.Node.RaftState.Lookup("members")

	if !ok {
		members = s.Node.RaftNode.Members
	} else {
		err := json.Unmarshal([]byte(m), &members)

		if err != nil {
			return nil, err
		}
	}

	member, memberIndex, err := cluster.GetMemberByID(in.MemberId, members)

	if err != nil {
		return nil, err
	}

	cc := raftpb.ConfChange{
		Type:   raftpb.ConfChangeRemoveNode,
		NodeID: member.ID,
	}

	s.Node.RaftNode.ConfChangeC <- cc

	members = append(members[:*memberIndex], members[*memberIndex+1:]...)

	json, err := json.Marshal(members)

	if err != nil {
		return nil, err
	}

	s.Node.RaftState.Propose("members", string(json))

	res := &api.ManagerRemoteRemoveRes{}

	log.Printf("Removed member from cluster %d", in.MemberId)

	return res, nil
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

	logrus.Info("Started manager remote gRPC tcp server.")

	api.RegisterManagerRemoteServer(s, server)

	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve")
	}
}
