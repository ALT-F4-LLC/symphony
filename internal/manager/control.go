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
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/raft/raftpb"
	"google.golang.org/grpc"
)

// controlServer : manager remote requests
type controlServer struct {
	Manager *Manager
}

// Init : initializes manager node and raft
func (s *controlServer) Init(ctx context.Context, in *api.ManagerControlInitRequest) (*api.ManagerControlInitResponse, error) {
	if s.Manager.Raft != nil {
		return nil, errors.New("invalid_raft_state")
	}

	httpStopC := make(chan struct{})

	raftAddr := fmt.Sprintf("http://%s", s.Manager.Flags.ListenRaftAddr.String())

	raftID := uint64(1)

	raftMember := &api.RaftMember{
		Addr: raftAddr,
		Id:   raftID,
	}

	raftMembers := []*api.RaftMember{raftMember}

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

	raftMembersJSON, err := json.Marshal(raftMembers)

	if err != nil {
		return nil, err
	}

	state.Propose("raft:members", string(raftMembersJSON))

	saveRaftIDErr := s.Manager.SaveRaftID(raftID)

	if saveRaftIDErr != nil {
		return nil, saveRaftIDErr
	}

	serviceID := uuid.New()

	serviceAddr := fmt.Sprintf("%s", s.Manager.Flags.ListenRemoteAddr.String())

	service := &api.ServiceMember{
		Id:   serviceID.String(),
		Addr: serviceAddr,
		Type: api.ServiceType_MANAGER,
	}

	serviceMembers := []*api.ServiceMember{service}

	serviceMembersJSON, err := json.Marshal(serviceMembers)

	if err != nil {
		return nil, err
	}

	state.Propose("service:members", string(serviceMembersJSON))

	saveServiceIDErr := s.Manager.SaveServiceID(serviceID)

	if saveServiceIDErr != nil {
		return nil, saveServiceIDErr
	}

	res := &api.ManagerControlInitResponse{}

	go raft.HTTPStopCHandler(s.Manager.Flags.ConfigDir, httpStopC)

	return res, nil
}

// Join : joins a manager to the cluster
func (s *controlServer) Join(ctx context.Context, in *api.ManagerControlJoinRequest) (*api.ManagerControlJoinResponse, error) {
	remoteAddr, err := net.ResolveTCPAddr("tcp", in.RemoteAddr)

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

	serviceAddr := fmt.Sprintf("%s", s.Manager.Flags.ListenRemoteAddr.String())

	serviceJoinOpts := &api.ManagerRemoteJoinRequest{
		Addr: serviceAddr,
		Type: api.ServiceType_MANAGER,
	}

	serviceJoin, err := c.Join(ctx, serviceJoinOpts)

	if err != nil {
		return nil, err
	}

	serviceID, err := uuid.Parse(serviceJoin.Id)

	if err != nil {
		return nil, err
	}

	saveServiceIDErr := s.Manager.SaveServiceID(serviceID)

	if saveServiceIDErr != nil {
		return nil, saveServiceIDErr
	}

	httpStopC := make(chan struct{})

	raftAddr := fmt.Sprintf("http://%s", s.Manager.Flags.ListenRaftAddr.String())

	raftJoinOpts := &api.ManagerRemoteJoinRaftRequest{
		RaftAddr:  raftAddr,
		ServiceId: serviceJoin.Id,
	}

	raftJoin, err := c.JoinRaft(ctx, raftJoinOpts)

	raftConfig := raft.Config{
		ConfigDir:  s.Manager.Flags.ConfigDir,
		HTTPStopC:  httpStopC,
		Join:       true,
		ListenAddr: s.Manager.Flags.ListenRaftAddr,
		Members:    raftJoin.RaftMembers,
		NodeID:     raftJoin.RaftId,
	}

	raft, state, err := raft.New(raftConfig)

	if err != nil {
		return nil, err
	}

	s.Manager.Raft = raft

	s.Manager.State = state

	saveRaftIDErr := s.Manager.SaveRaftID(raftJoin.RaftId)

	if saveRaftIDErr != nil {

		return nil, saveRaftIDErr
	}

	res := &api.ManagerControlJoinResponse{}

	go raft.HTTPStopCHandler(s.Manager.Flags.ConfigDir, httpStopC)

	return res, nil
}

func (s *controlServer) Members(ctx context.Context, in *api.ManagerControlMembersRequest) (*api.ManagerControlMembersResponse, error) {
	if s.Manager.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	gossip, err := s.Manager.GetGossipMembers()

	if err != nil {
		return nil, err
	}

	raft, err := s.Manager.GetRaftMembers()

	if err != nil {
		return nil, err
	}

	res := &api.ManagerControlMembersResponse{
		Gossip: gossip,
		Raft:   raft,
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

	s.Manager.State.Propose("raft:members", string(json))

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
