package manager

import (
	"context"
	"encoding/json"
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

func startRaftAndReturnResponse(server *ControlServer, join bool, nodeID uint64, peers []string) (*api.ManagerControlInitializeResponse, error) {
	raft, state, err := cluster.NewRaft(server.Node.Flags, join, server.Node.Key, nodeID, peers)

	if err != nil {
		return nil, err
	}

	server.Node.Raft = raft

	server.Node.State = state

	res := &api.ManagerControlInitializeResponse{}

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

		join, joinErr := c.ManagerRemoteInit(ctx, &api.ManagerRemoteInitRequest{Addr: addr})

		if joinErr != nil {
			return nil, joinErr
		}

		return startRaftAndReturnResponse(s, false, join.NodeId, join.Peers)
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

	return startRaftAndReturnResponse(s, false, nodeID, peers)
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

	raft, state, err := cluster.NewRaft(s.Node.Flags, true, s.Node.Key, join.NodeId, join.Peers)

	if err != nil {
		return nil, err
	}

	s.Node.Raft = raft

	s.Node.State = state

	res := &api.ManagerControlJoinResponse{}

	return res, nil
}

// ManagerControlPeers : returns a list of manager peers
func (s *ControlServer) ManagerControlPeers(ctx context.Context, in *api.ManagerControlPeersRequest) (*api.ManagerControlPeersResponse, error) {
	if s.Node.State == nil {
		return nil, errors.New("invalid_state_request")
	}

	p, ok := s.Node.State.Lookup("raft:peers")

	if !ok {
		return nil, errors.New("invalid_raft_peers")
	}

	var peers []string

	if err := json.Unmarshal([]byte(p), &peers); err != nil {
		return nil, err
	}

	res := &api.ManagerControlPeersResponse{
		Peers: peers,
	}

	return res, nil
}
