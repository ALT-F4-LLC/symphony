package block

import (
	"context"
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

// ControlServer : block socket requests
type ControlServer struct {
	Node *cluster.Node
}

// Join : joins a manager to a existing cluster
func (s *ControlServer) Join(ctx context.Context, in *api.BlockControlJoinReq) (*api.BlockControlJoinRes, error) {
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

	s.Node.Raft = raft

	s.Node.State = state

	res := &api.BlockControlJoinRes{}

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

	logrus.Info("Started block control gRPC socket server.")

	api.RegisterBlockControlServer(s, cs)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
