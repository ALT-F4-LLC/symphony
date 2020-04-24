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
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/google/uuid"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/raft/raftpb"
	"google.golang.org/grpc"
)

// controlServer : manager remote requests
type controlServer struct {
	Node *cluster.Node
}

func createRaftMembers(listenRaftAddr *net.TCPAddr) []*api.RaftMember {
	addr := fmt.Sprintf("http://%s", listenRaftAddr.String())

	leader := &api.RaftMember{
		Addr: addr,
		Id:   uint64(1),
	}

	members := []*api.RaftMember{leader}

	return members
}

func getTCPAddr(addr string) (*net.TCPAddr, error) {
	var joinGossipAddr *net.TCPAddr

	if addr != "" {
		addr, err := net.ResolveTCPAddr("tcp", addr)

		if err != nil {
			return nil, err
		}

		joinGossipAddr = addr
	}

	return joinGossipAddr, nil
}

func newMemberList(flags *config.Flags, gossipID uuid.UUID, raftID uint64) (*memberlist.Memberlist, error) {
	data := api.GossipMember{
		RaftId: raftID,
	}

	meta, err := json.Marshal(data)

	if err != nil {
		return nil, err
	}

	config := memberlist.DefaultLocalConfig()

	config.AdvertisePort = flags.ListenGossipAddr.Port

	config.Delegate = &cluster.GossipDelegate{Meta: meta}

	config.BindPort = flags.ListenGossipAddr.Port

	config.Name = gossipID.String()

	list, err := memberlist.Create(config)

	if err != nil {
		return nil, err
	}

	for _, member := range list.Members() {
		fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
	}

	return list, nil
}

func (s *controlServer) Init(ctx context.Context, in *api.ManagerControlInitRequest) (*api.ManagerControlInitResponse, error) {
	if s.Node.Raft != nil {
		return nil, errors.New("invalid_raft_state")
	}

	httpStopC := make(chan struct{})

	raftAddr := fmt.Sprintf("http://%s", s.Node.Flags.ListenRaftAddr.String())

	raftID := uint64(1)

	raftLeader := &api.RaftMember{
		Addr: raftAddr,
		Id:   raftID,
	}

	raftMembers := []*api.RaftMember{raftLeader}

	raft, state, err := cluster.NewRaft(s.Node.Flags, httpStopC, false, raftMembers, raftID)

	if err != nil {
		return nil, err
	}

	// TODO : Potentially remove on init and have gossip be created when
	// -----> workers join the cluster

	raftLeaderGossipID := uuid.New()

	list, err := newMemberList(s.Node.Flags, raftLeaderGossipID, raftLeader.Id)

	if err != nil {
		return nil, err
	}

	s.Node.Memberlist = list

	s.Node.Raft = raft

	s.Node.State = state

	saveErr := s.Node.SaveRaftNodeID(raftID)

	if saveErr != nil {
		return nil, err
	}

	res := &api.ManagerControlInitResponse{}

	go raft.HTTPStopCHandler(s.Node.Flags.ConfigDir, httpStopC, list)

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

	addr := fmt.Sprintf("http://%s", s.Node.Flags.ListenRaftAddr.String())

	opts := &api.ManagerRemoteJoinRequest{
		Addr: addr,
		Role: api.Role_MANAGER,
	}

	join, err := c.Join(ctx, opts)

	if err != nil {
		return nil, err
	}

	saveErr := s.Node.SaveRaftNodeID(join.RaftId)

	if saveErr != nil {
		return nil, saveErr
	}

	httpStopC := make(chan struct{})

	raft, state, err := cluster.NewRaft(s.Node.Flags, httpStopC, true, join.RaftMembers, join.RaftId)

	if err != nil {
		return nil, err
	}

	gossipAddr, err := getTCPAddr(join.GossipAddr)

	if err != nil {
		return nil, err
	}

	gossipID, err := uuid.Parse(join.GossipId)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	_, raftMember := cluster.GetRaftMemberByID(join.RaftMembers, join.RaftId)

	if raftMember == nil {
		return nil, api.ProtoError(errors.New("invalid_raft_member"))
	}

	list, err := newMemberList(s.Node.Flags, gossipID, raftMember.Id)

	if err != nil {
		return nil, err
	}

	_, joinErr := list.Join([]string{gossipAddr.String()})

	if joinErr != nil {
		return nil, joinErr
	}

	s.Node.Raft = raft

	s.Node.State = state

	s.Node.Memberlist = list

	res := &api.ManagerControlJoinResponse{}

	go raft.HTTPStopCHandler(s.Node.Flags.ConfigDir, httpStopC, list)

	return res, nil
}

func (s *controlServer) Members(ctx context.Context, in *api.ManagerControlMembersRequest) (*api.ManagerControlMembersResponse, error) {
	if s.Node.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	gossip, err := s.Node.GetGossipMembers()

	if err != nil {
		return nil, err
	}

	raft, err := s.Node.GetRaftMembers()

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
	if s.Node.Raft == nil {
		return nil, errors.New("invalid_raft_state")
	}

	members, err := s.Node.GetRaftMembers()

	if err != nil {
		return nil, err
	}

	memberIndex, member := cluster.GetRaftMemberByID(members, in.RaftId)

	if member == nil {
		return nil, errors.New("invalid_member_id")
	}

	cc := raftpb.ConfChange{
		Type:   raftpb.ConfChangeRemoveNode,
		NodeID: member.Id,
	}

	s.Node.Raft.ConfChangeC <- cc

	members = append(members[:*memberIndex], members[*memberIndex+1:]...)

	json, err := json.Marshal(members)

	if err != nil {
		return nil, err
	}

	s.Node.State.Propose("members", string(json))

	res := &api.ManagerControlRemoveResponse{}

	log.Printf("Removed member from cluster %d", in.RaftId)

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

	server := grpc.NewServer()

	control := &controlServer{
		Node: node,
	}

	logrus.Info("Started manager control gRPC socket server.")

	api.RegisterManagerControlServer(server, control)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
