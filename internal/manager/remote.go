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

	memberIndex, err := cluster.GetMemberIndex(in.Addr, peers)

	if err != nil {
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

	log.Printf("Existing member proposed to cluster %d", *memberIndex)

	nodeID := uint64(*memberIndex + 1)

	res := &api.ManagerRemoteJoinResponse{
		NodeId: nodeID,
		Peers:  s.Node.Raft.Peers,
	}

	s.Node.Raft.Peers = append(s.Node.Raft.Peers, in.Addr)

	return res, nil
}

func removeFromPeers(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// ManagerRemoteRemove : adds nodes to the raft
func (s *RemoteServer) ManagerRemoteRemove(ctx context.Context, in *api.ManagerRemoteRemoveRequest) (*api.ManagerRemoteRemoveResponse, error) {
	if s.Node.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	memberIndex, err := cluster.GetMemberIndex(in.Addr, s.Node.Raft.Peers)

	if err != nil {
		return nil, err
	}

	nodeID := uint64(*memberIndex + 1)

	cc := raftpb.ConfChange{
		Type:   raftpb.ConfChangeRemoveNode,
		NodeID: nodeID,
	}

	s.Node.Raft.ConfChangeC <- cc

	res := &api.ManagerRemoteRemoveResponse{}

	s.Node.Raft.Peers = removeFromPeers(s.Node.Raft.Peers, in.Addr)

	log.Printf("Removed member from cluster %d", *memberIndex)

	return res, nil
}
