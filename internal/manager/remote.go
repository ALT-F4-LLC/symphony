package manager

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"go.etcd.io/etcd/raft/raftpb"
)

// RemoteServer : manager remote requests
type RemoteServer struct {
	Node *Node
}

// ManagerRemoteInitialize : adds nodes to the raft
func (s *RemoteServer) ManagerRemoteInitialize(ctx context.Context, in *api.ManagerRemoteInitializeRequest) (*api.ManagerRemoteInitializeResponse, error) {
	if in.Addr == "" {
		return nil, errors.New("invalid_raft_addr")
	}

	if s.Node.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	_, ok := s.Node.State.Lookup("members")

	if ok {
		return nil, errors.New("invalid_raft_members_state")
	}

	member, _, err := cluster.GetMemberByAddr(in.Addr, s.Node.Raft.Members)

	if err != nil {
		return nil, errors.New("invalid_raft_member") // does not exist in initialized cluster peers
	}

	res := &api.ManagerRemoteInitializeResponse{
		MemberId: member.ID,
		Members:  s.Node.Raft.Members,
	}

	return res, nil
}

// ManagerRemoteJoin : adds nodes to the raft
func (s *RemoteServer) ManagerRemoteJoin(ctx context.Context, in *api.ManagerRemoteJoinRequest) (*api.ManagerRemoteJoinResponse, error) {
	if s.Node.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	var members []*api.Member

	m, ok := s.Node.State.Lookup("members")

	if !ok {
		members = s.Node.Raft.Members
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
	commitIndex := s.Node.Raft.Node.Status().Commit

	memberID := uint64(commitIndex)

	cc := raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  memberID,
		Context: []byte(in.Addr),
	}

	s.Node.Raft.ConfChangeC <- cc

	log.Printf("New member proposed to cluster %d", memberID)

	res := &api.ManagerRemoteJoinResponse{
		MemberId: memberID,
		Members:  members,
	}

	members = append(members, &api.Member{ID: memberID, Addr: in.Addr})

	json, err := json.Marshal(members)

	if err != nil {
		return nil, err
	}

	s.Node.State.Propose("members", string(json))

	return res, nil
}

// ManagerRemoteRemove : adds nodes to the raft
func (s *RemoteServer) ManagerRemoteRemove(ctx context.Context, in *api.ManagerRemoteRemoveRequest) (*api.ManagerRemoteRemoveResponse, error) {
	if s.Node.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	var members []*api.Member

	m, ok := s.Node.State.Lookup("members")

	if !ok {
		members = s.Node.Raft.Members
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

	s.Node.Raft.ConfChangeC <- cc

	members = append(members[:*memberIndex], members[*memberIndex+1:]...)

	json, err := json.Marshal(members)

	if err != nil {
		return nil, err
	}

	s.Node.State.Propose("members", string(json))

	res := &api.ManagerRemoteRemoveResponse{}

	log.Printf("Removed member from cluster %d", in.MemberId)

	return res, nil
}
