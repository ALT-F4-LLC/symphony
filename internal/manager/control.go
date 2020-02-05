package manager

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"google.golang.org/grpc"
)

// ControlServer : manager remote requests
type ControlServer struct {
	Node *Node
}

// saveRaftNodeID : sets the RAFT_NODE_ID in key.json
func (s *ControlServer) saveRaftNodeID(nodeID uint64) error {
	s.Node.Key.RaftNodeID = nodeID

	err := s.Node.SaveKey()

	if err != nil {
		return err
	}

	return nil
}

func (s *ControlServer) startRaftAndRespond(join bool, nodeID uint64, peers []string) (*api.ManagerControlInitializeResponse, error) {
	raft, state, err := cluster.NewRaft(s.Node.Flags, join, nodeID, peers)

	if err != nil {
		return nil, err
	}

	s.Node.Raft = raft

	s.Node.State = state

	res := &api.ManagerControlInitializeResponse{}

	return res, nil
}

// ManagerControlGetValue : gets a specified key value in raft
func (s *ControlServer) ManagerControlGetValue(ctx context.Context, in *api.ManagerControlGetValueRequest) (*api.ManagerControlGetValueResponse, error) {
	if s.Node.State == nil {
		return nil, errors.New("invalid_state_request")
	}

	p, ok := s.Node.State.Lookup(in.Key)

	if !ok {
		return nil, errors.New("invalid_key_lookup")
	}

	res := &api.ManagerControlGetValueResponse{
		Value: p,
	}

	return res, nil
}

// ManagerControlInitialize : initializes a manager for a cluster
func (s *ControlServer) ManagerControlInitialize(ctx context.Context, in *api.ManagerControlInitializeRequest) (*api.ManagerControlInitializeResponse, error) {
	if in.JoinAddr != "" && in.Peers != nil {
		return nil, errors.New("invalid_init_request")
	}

	if s.Node.Raft != nil {
		return nil, errors.New("invalid_raft_status")
	}

	addr := fmt.Sprintf("http://%s", s.Node.Flags.ListenRaftAddr.String())

	if in.JoinAddr != "" {
		joinAddr, err := net.ResolveTCPAddr("tcp", in.JoinAddr)

		if err != nil {
			return nil, err
		}

		conn, err := grpc.Dial(joinAddr.String(), grpc.WithInsecure())

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		c := api.NewManagerRemoteClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		defer cancel()

		join, joinErr := c.ManagerRemoteInitialize(ctx, &api.ManagerRemoteInitializeRequest{Addr: addr})

		if joinErr != nil {
			return nil, joinErr
		}

		saveErr := s.saveRaftNodeID(join.NodeId)

		if saveErr != nil {
			return nil, saveErr
		}

		return s.startRaftAndRespond(false, join.NodeId, join.Peers)
	}

	var nodeID uint64

	var peers []string

	nodeID = 1

	peers = []string{addr}

	if len(in.Peers) > 1 {
		_, err := cluster.GetMemberIndex(addr, in.Peers)

		if err == nil {
			return nil, errors.New("invalid_peer_list")
		}

		peers = append(peers, in.Peers...)
	}

	err := s.saveRaftNodeID(nodeID)

	if err != nil {
		return nil, err
	}

	return s.startRaftAndRespond(false, nodeID, peers)
}

// ManagerControlJoin : joins a manager to a existing cluster
func (s *ControlServer) ManagerControlJoin(ctx context.Context, in *api.ManagerControlJoinRequest) (*api.ManagerControlJoinResponse, error) {
	joinAddr, err := net.ResolveTCPAddr("tcp", in.JoinAddr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(joinAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	c := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	addr := fmt.Sprintf("http://%s", s.Node.Flags.ListenRaftAddr.String())

	join, joinErr := c.ManagerRemoteJoin(ctx, &api.ManagerRemoteJoinRequest{Addr: addr})

	if joinErr != nil {
		return nil, joinErr
	}

	saveErr := s.saveRaftNodeID(join.NodeId)

	if saveErr != nil {
		return nil, saveErr
	}

	raft, state, err := cluster.NewRaft(s.Node.Flags, true, join.NodeId, join.Peers)

	if err != nil {
		return nil, err
	}

	s.Node.Raft = raft

	s.Node.State = state

	res := &api.ManagerControlJoinResponse{}

	return res, nil
}

// ManagerControlSetValue : sets a specified key value in raft
func (s *ControlServer) ManagerControlSetValue(ctx context.Context, in *api.ManagerControlSetValueRequest) (*api.ManagerControlSetValueResponse, error) {
	if s.Node.State == nil {
		return nil, errors.New("invalid_state_request")
	}

	s.Node.State.Propose(in.Key, in.Value)

	res := &api.ManagerControlSetValueResponse{
		Value: in.Value,
	}

	return res, nil
}

// ManagerControlRemove : Removes a manager to a existing cluster
func (s *ControlServer) ManagerControlRemove(ctx context.Context, in *api.ManagerControlRemoveRequest) (*api.ManagerControlRemoveResponse, error) {
	remoteAddr, err := net.ResolveTCPAddr("tcp", s.Node.Flags.ListenRemoteAddr.String())

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(remoteAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	c := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	addr := fmt.Sprintf("http://%s", in.Addr)

	_, rmErr := c.ManagerRemoteRemove(ctx, &api.ManagerRemoteRemoveRequest{Addr: addr})

	if rmErr != nil {
		return nil, rmErr
	}

	res := &api.ManagerControlRemoveResponse{}

	return res, nil
}
