package manager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// ControlServer : manager remote requests
type ControlServer struct {
	Node *cluster.Node
}

func (s *ControlServer) startRaftAndRespond(join bool, members []*api.Member, nodeID uint64) (*api.ManagerControlInitRes, error) {
	raft, state, err := cluster.NewRaft(s.Node.Flags, join, members, nodeID)

	if err != nil {
		return nil, err
	}

	s.Node.RaftNode = raft

	s.Node.RaftState = state

	res := &api.ManagerControlInitRes{}

	return res, nil
}

// Get : gets a specified key value in raft
func (s *ControlServer) Get(ctx context.Context, in *api.ManagerControlGetReq) (*api.ManagerControlGetRes, error) {
	if s.Node.RaftState == nil {
		return nil, errors.New("invalid_state_request")
	}

	p, ok := s.Node.RaftState.Lookup(in.Key)

	if !ok {
		return nil, errors.New("invalid_key_lookup")
	}

	res := &api.ManagerControlGetRes{
		Value: p,
	}

	return res, nil
}

// Init : initializes a manager for a cluster
func (s *ControlServer) Init(ctx context.Context, in *api.ManagerControlInitReq) (*api.ManagerControlInitRes, error) {
	if in.JoinAddr != "" && in.Members != nil {
		return nil, errors.New("invalid_init_request")
	}

	if s.Node.RaftNode != nil {
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

		join, joinErr := c.Init(ctx, &api.ManagerRemoteInitReq{Addr: addr})

		if joinErr != nil {
			return nil, joinErr
		}

		saveErr := s.Node.SaveRaftNodeID(join.MemberId)

		if saveErr != nil {
			return nil, saveErr
		}

		return s.startRaftAndRespond(false, join.Members, join.MemberId)
	}

	nodeID := uint64(1)

	members := []*api.Member{&api.Member{ID: nodeID, Addr: addr}}

	if len(in.Members) > 1 {
		_, _, err := cluster.GetMemberByAddr(addr, in.Members)

		if err == nil {
			return nil, errors.New("invalid_peer_list")
		}

		members = append(members, in.Members...)
	}

	err := s.Node.SaveRaftNodeID(nodeID)

	if err != nil {
		return nil, err
	}

	return s.startRaftAndRespond(false, members, nodeID)
}

// Join : joins a manager to a existing cluster
func (s *ControlServer) Join(ctx context.Context, in *api.ManagerControlJoinReq) (*api.ManagerControlJoinRes, error) {
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

	join, joinErr := c.Join(ctx, &api.ManagerRemoteJoinReq{Addr: addr})

	if joinErr != nil {
		return nil, joinErr
	}

	saveErr := s.Node.SaveRaftNodeID(join.MemberId)

	if saveErr != nil {
		return nil, saveErr
	}

	raft, state, err := cluster.NewRaft(s.Node.Flags, true, join.Members, join.MemberId)

	if err != nil {
		return nil, err
	}

	s.Node.RaftNode = raft

	s.Node.RaftState = state

	res := &api.ManagerControlJoinRes{}

	return res, nil
}

// Set : sets a specified key value in raft
func (s *ControlServer) Set(ctx context.Context, in *api.ManagerControlSetReq) (*api.ManagerControlSetRes, error) {
	if s.Node.RaftState == nil {
		return nil, errors.New("invalid_state_request")
	}

	s.Node.RaftState.Propose(in.Key, in.Value)

	res := &api.ManagerControlSetRes{
		Value: in.Value,
	}

	return res, nil
}

// Remove : Removes a manager to a existing cluster
func (s *ControlServer) Remove(ctx context.Context, in *api.ManagerControlRemoveReq) (*api.ManagerControlRemoveRes, error) {
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

	_, rmErr := c.Remove(ctx, &api.ManagerRemoteRemoveReq{MemberId: in.MemberId})

	if rmErr != nil {
		return nil, rmErr
	}

	res := &api.ManagerControlRemoveRes{}

	return res, nil
}

// StartControlServer : starts manager control server
func StartControlServer(node *cluster.Node) {
	socketPath := fmt.Sprintf("%s/control.sock", node.Flags.ConfigDir)

	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("unix", socketPath)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	cs := &ControlServer{
		Node: node,
	}

	logrus.Info("Started manager control gRPC socket server.")

	api.RegisterManagerControlServer(s, cs)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
