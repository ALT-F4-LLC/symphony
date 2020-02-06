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
	if s.Node.Raft == nil {
		return nil, errors.New("invalid_raft_cluster")
	}

	if in.Addr == "" {
		return nil, errors.New("invalid_raft_addr")
	}

	membersLength := len(s.Node.Raft.Peers)

	if membersLength > 3 {
		return nil, errors.New("invalid_raft_state") // cluster has already been Initializeialized
	}

	memberIndex, err := cluster.GetMemberIndex(in.Addr, s.Node.Raft.Peers)

	if membersLength <= 3 && err != nil {
		return nil, errors.New("invalid_raft_member") // does not exist in Initialize cluster peers
	}

	res := &api.ManagerRemoteInitializeResponse{
		NodeId: uint64(*memberIndex + 1),
		Peers:  s.Node.Raft.Peers,
	}

	return res, nil
}

// ManagerRemoteJoin : adds nodes to the raft
func (s *RemoteServer) ManagerRemoteJoin(ctx context.Context, in *api.ManagerRemoteJoinRequest) (*api.ManagerRemoteJoinResponse, error) {
	if s.Node.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	var peers []string

	p, ok := s.Node.State.Lookup("raft:peers")

	if !ok {
		peers = s.Node.Raft.Peers
	} else {
		err := json.Unmarshal([]byte(p), &peers)

		if err != nil {
			return nil, err
		}
	}

	_, err := cluster.GetMemberIndex(in.Addr, peers)

	if err == nil {
		return nil, errors.New("invalid_raft_member")
	}

	nodeID := uint64(len(peers) + 1)

	cc := raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  nodeID,
		Context: []byte(in.Addr),
	}

	s.Node.Raft.ConfChangeC <- cc

	log.Printf("New member proposed to cluster %d", nodeID)

	res := &api.ManagerRemoteJoinResponse{
		NodeId: nodeID,
		Peers:  peers,
	}

	peers = append(peers, in.Addr)

	json, err := json.Marshal(peers)

	if err != nil {
		return nil, err
	}

	s.Node.State.Propose("raft:peers", string(json))

	return res, nil
}

// ManagerRemoteRemove : adds nodes to the raft
func (s *RemoteServer) ManagerRemoteRemove(ctx context.Context, in *api.ManagerRemoteRemoveRequest) (*api.ManagerRemoteRemoveResponse, error) {
	if s.Node.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	var peers []string

	p, ok := s.Node.State.Lookup("raft:peers")

	if !ok {
		return nil, errors.New("invalid_raft_peers_state")
	}

	err := json.Unmarshal([]byte(p), &peers)

	if err != nil {
		return nil, err
	}

	if in.NodeId == 0 || int(in.NodeId) > len(peers) {
		return nil, errors.New("invalid_node_id")
	}

	cc := raftpb.ConfChange{
		Type:   raftpb.ConfChangeRemoveNode,
		NodeID: in.NodeId,
	}

	s.Node.Raft.ConfChangeC <- cc

	memberIndex := int(in.NodeId) - 1

	copy(peers[memberIndex:], peers[memberIndex+1:])

	peers[len(peers)-1] = ""

	peers = peers[:len(peers)-1]

	json, err := json.Marshal(peers)

	if err != nil {
		return nil, err
	}

	s.Node.State.Propose("raft:peers", string(json))

	res := &api.ManagerRemoteRemoveResponse{}

	log.Printf("Removed member from cluster %d", in.NodeId)

	return res, nil
}
