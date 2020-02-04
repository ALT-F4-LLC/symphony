package manager

import (
	"context"
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

// ManagerRemoteInit : adds nodes to the raft
func (s *RemoteServer) ManagerRemoteInit(ctx context.Context, in *api.ManagerRemoteInitRequest) (*api.ManagerRemoteInitResponse, error) {
	if s.Node.Raft == nil {
		return nil, errors.New("invalid_raft_cluster")
	}

	if in.Addr == "" {
		return nil, errors.New("invalid_raft_addr")
	}

	membersLength := len(s.Node.Raft.Peers)

	if membersLength > 3 {
		return nil, errors.New("invalid_raft_state") // cluster has already been initialized
	}

	memberIndex, err := cluster.GetMemberIndex(in.Addr, s.Node.Raft.Peers)

	if membersLength <= 3 && err != nil {
		return nil, errors.New("invalid_raft_member") // does not exist in initial cluster peers
	}

	res := &api.ManagerRemoteInitResponse{
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

	memberIndex, err := cluster.GetMemberIndex(in.Addr, s.Node.Raft.Peers)

	if err != nil {
		nodeID := uint64(len(s.Node.Raft.Peers) + 1)

		cc := raftpb.ConfChange{
			Type:    raftpb.ConfChangeAddNode,
			NodeID:  nodeID,
			Context: []byte(in.Addr),
		}

		// --> DO WE NEED TO FIRE THIS ON ALL SERVERS? s.Node.Raft.ConfChangeC <- cc

		s.Node.Raft.ConfChangeC <- cc

		log.Printf("New member proposed to cluster %d", nodeID)

		res := &api.ManagerRemoteJoinResponse{
			NodeId: nodeID,
			Peers:  s.Node.Raft.Peers,
		}

		s.Node.Raft.Peers = append(s.Node.Raft.Peers, in.Addr)

		return res, nil
	}

	log.Printf("Existing member proposed to cluster %d", *memberIndex)

	s.Node.Raft.Peers = append(s.Node.Raft.Peers, in.Addr)

	nodeID := uint64(*memberIndex + 1)

	res := &api.ManagerRemoteJoinResponse{
		NodeId: nodeID,
		Peers:  s.Node.Raft.Peers,
	}

	return res, nil
}
