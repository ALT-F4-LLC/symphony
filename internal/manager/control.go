package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/raft"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/raft/raftpb"
	"google.golang.org/grpc"
)

// controlServer : manager remote requests
type controlServer struct {
	Manager *Manager
}

func (s *controlServer) Init(ctx context.Context, in *api.ManagerControlInitRequest) (*api.ManagerControlInitResponse, error) {
	if s.Manager.Raft != nil {
		return nil, errors.New("invalid_raft_state")
	}

	httpStopC := make(chan struct{})

	raftAddr := fmt.Sprintf("http://%s", s.Manager.Flags.ListenRaftAddr.String())

	raftID := uint64(1)

	raftLeader := &api.RaftMember{
		Addr: raftAddr,
		Id:   raftID,
	}

	raftMembers := []*api.RaftMember{raftLeader}

	raftConfig := raft.Config{
		ConfigDir:  s.Manager.Flags.ConfigDir,
		HTTPStopC:  httpStopC,
		Join:       false,
		ListenAddr: s.Manager.Flags.ListenRaftAddr,
		Members:    raftMembers,
		NodeID:     raftID,
	}

	raft, state, err := raft.New(raftConfig)

	if err != nil {
		return nil, err
	}

	s.Manager.Raft = raft

	s.Manager.State = state

	saveErr := s.Manager.SaveRaftNodeID(raftID)

	if saveErr != nil {
		return nil, err
	}

	res := &api.ManagerControlInitResponse{}

	go raft.HTTPStopCHandler(s.Manager.Flags.ConfigDir, httpStopC)

	return res, nil
}

func (s *controlServer) Join(ctx context.Context, in *api.ManagerControlJoinRequest) (*api.ManagerControlJoinResponse, error) {
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

	addr := fmt.Sprintf("http://%s", s.Manager.Flags.ListenRaftAddr.String())

	opts := &api.ManagerRemoteJoinRequest{
		Addr: addr,
		Role: api.Role_MANAGER,
	}

	join, err := c.Join(ctx, opts)

	if err != nil {
		return nil, err
	}

	saveErr := s.Manager.SaveRaftNodeID(join.RaftId)

	if saveErr != nil {
		return nil, saveErr
	}

	httpStopC := make(chan struct{})

	raftConfig := raft.Config{
		ConfigDir:  s.Manager.Flags.ConfigDir,
		HTTPStopC:  httpStopC,
		Join:       true,
		ListenAddr: s.Manager.Flags.ListenRaftAddr,
		Members:    join.RaftMembers,
		NodeID:     join.RaftId,
	}

	raft, state, err := raft.New(raftConfig)

	if err != nil {
		return nil, err
	}

	_, raftMember := GetRaftMemberByID(join.RaftMembers, join.RaftId)

	if raftMember == nil {
		return nil, api.ProtoError(errors.New("invalid_raft_member"))
	}

	s.Manager.Raft = raft

	s.Manager.State = state

	res := &api.ManagerControlJoinResponse{}

	go raft.HTTPStopCHandler(s.Manager.Flags.ConfigDir, httpStopC)

	return res, nil
}

func (s *controlServer) Members(ctx context.Context, in *api.ManagerControlMembersRequest) (*api.ManagerControlMembersResponse, error) {
	if s.Manager.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	raft, err := s.Manager.GetRaftMembers()

	if err != nil {
		return nil, err
	}

	res := &api.ManagerControlMembersResponse{
		Raft: raft,
	}

	return res, nil
}

func (s *controlServer) Remove(ctx context.Context, in *api.ManagerControlRemoveRequest) (*api.ManagerControlRemoveResponse, error) {
	if s.Manager.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	members, err := s.Manager.GetRaftMembers()

	if err != nil {
		return nil, err
	}

	memberIndex, member := GetRaftMemberByID(members, in.RaftId)

	if member == nil {
		return nil, errors.New("invalid_member_id")
	}

	cc := raftpb.ConfChange{
		Type:   raftpb.ConfChangeRemoveNode,
		NodeID: member.Id,
	}

	s.Manager.Raft.ConfChangeC <- cc

	members = append(members[:*memberIndex], members[*memberIndex+1:]...)

	json, err := json.Marshal(members)

	if err != nil {
		return nil, err
	}

	s.Manager.State.Propose("members", string(json))

	res := &api.ManagerControlRemoveResponse{}

	log.Printf("Removed member from cluster %d", in.RaftId)

	// TODO : Make sure no members or raft data exist on remove

	return res, nil
}

// ControlServer : starts manager control server
func ControlServer(m *Manager) {
	socketPath := fmt.Sprintf("%s/control.sock", m.Flags.ConfigDir)

	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("unix", socketPath)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()

	control := &controlServer{
		Manager: m,
	}

	logrus.Info("Started manager control gRPC socket server.")

	api.RegisterManagerControlServer(server, control)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
