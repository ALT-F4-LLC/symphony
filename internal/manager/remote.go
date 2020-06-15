package manager

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/raft/raftpb"
	"google.golang.org/grpc"
)

type remoteServer struct {
	manager *Manager
}

// Join : joins node to the cluster
func (s *remoteServer) Join(ctx context.Context, in *api.ManagerRemoteJoinRequest) (*api.ManagerRemoteJoinResponse, error) {
	members, err := s.manager.GetServiceMembers()

	if err != nil {
		return nil, err
	}

	_, memberExists := GetServiceByAddr(members, in.Addr)

	if memberExists != nil {
		return nil, errors.New("invalid_service_addr")
	}

	serviceID := uuid.New()

	serviceMember := &api.ServiceMember{
		Addr: in.Addr,
		Id:   serviceID.String(),
		Type: in.Type,
	}

	members = append(members, serviceMember)

	membersJSON, err := json.Marshal(members)

	if err != nil {
		return nil, err
	}

	s.manager.State.Propose("service:members", string(membersJSON))

	res := &api.ManagerRemoteJoinResponse{
		Id: serviceID.String(),
	}

	return res, nil
}

// JoinGossip : adds nodes to the gossip
func (s *remoteServer) JoinGossip(ctx context.Context, in *api.ManagerRemoteJoinGossipRequest) (*api.ManagerRemoteJoinGossipResponse, error) {
	serviceMembers, err := s.manager.GetServiceMembers()

	if err != nil {
		return nil, err
	}

	serviceID, err := uuid.Parse(in.ServiceId)

	if err != nil {
		return nil, err
	}

	_, serviceMemberExists := GetServiceByID(serviceMembers, serviceID)

	if serviceMemberExists != nil {
		return nil, errors.New("invalid_service_id")
	}

	gossipMembers, err := s.manager.GetGossipMembers()

	if err != nil {
		return nil, err
	}

	_, gossipMemberExists := GetGossipMemberByAddr(gossipMembers, in.GossipAddr)

	if gossipMemberExists != nil {
		return nil, errors.New("invalid_gossip_addr")
	}

	gossipID := uuid.New()

	added := &api.GossipMember{
		Addr: in.GossipAddr,
		ID:   gossipID.String(),
	}

	gossipMembers = append(gossipMembers, added)

	gossipPeerAddr := in.GossipAddr

	if len(gossipMembers) > 1 {
		gossipPeerAddr = gossipMembers[0].Addr
	}

	gossipMembersJSON, err := json.Marshal(gossipMembers)

	if err != nil {
		return nil, err
	}

	s.manager.State.Propose("gossip:members", string(gossipMembersJSON))

	res := &api.ManagerRemoteJoinGossipResponse{
		GossipId:       gossipID.String(),
		GossipPeerAddr: gossipPeerAddr,
	}

	return res, nil
}

// JoinRaft : adds nodes to the raft
func (s *remoteServer) JoinRaft(ctx context.Context, in *api.ManagerRemoteJoinRaftRequest) (*api.ManagerRemoteJoinRaftResponse, error) {
	if s.manager.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	serviceMembers, err := s.manager.GetServiceMembers()

	if err != nil {
		return nil, err
	}

	serviceID, err := uuid.Parse(in.ServiceId)

	if err != nil {
		return nil, err
	}

	_, serviceMemberExists := GetServiceByID(serviceMembers, serviceID)

	if serviceMemberExists != nil {
		return nil, errors.New("invalid_service_id")
	}

	raftMembers, err := s.manager.GetRaftMembers()

	if err != nil {
		return nil, err
	}

	_, member := GetRaftMemberByAddr(raftMembers, in.RaftAddr)

	if member != nil {
		return nil, errors.New("invalid_raft_member")
	}

	// need a uniq ident for node ids so we will use the
	// raft commit index as it is also uniq and an index
	commitIndex := s.manager.Raft.Node.Status().Commit

	added := &api.RaftMember{
		Addr: in.RaftAddr,
		Id:   uint64(commitIndex),
	}

	cc := raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  added.Id,
		Context: []byte(in.RaftAddr),
	}

	s.manager.Raft.ConfChangeC <- cc

	log.Printf("New member proposed to cluster %d", added.Id)

	raftMembers = append(raftMembers, added)

	raftMembersJSON, err := json.Marshal(raftMembers)

	if err != nil {
		return nil, err
	}

	s.manager.State.Propose("raft:members", string(raftMembersJSON))

	res := &api.ManagerRemoteJoinRaftResponse{
		RaftId:      added.Id,
		RaftMembers: raftMembers,
	}

	return res, nil
}

// RemoteServer : starts Raft membership server
func RemoteServer(m *Manager) {
	lis, err := net.Listen("tcp", m.Flags.ListenRemoteAddr.String())

	if err != nil {
		log.Fatal("Failed to listen")
	}

	server := grpc.NewServer()

	remote := &remoteServer{
		manager: m,
	}

	logrus.Info("Started manager remote gRPC tcp server.")

	api.RegisterManagerRemoteServer(server, remote)

	if err := server.Serve(lis); err != nil {
		log.Fatal("Failed to serve")
	}
}
