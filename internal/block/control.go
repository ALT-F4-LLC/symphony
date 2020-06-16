package block

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/erkrnt/symphony/internal/pkg/gossip"
	"github.com/google/uuid"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type controlServer struct {
	block      *Block
	memberlist *memberlist.Memberlist
}

func (s *controlServer) Init(ctx context.Context, in *api.BlockControlInitRequest) (*api.BlockControlInitResponse, error) {
	joinAddr, err := net.ResolveTCPAddr("tcp", in.Addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(joinAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	r := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	serviceAddr := fmt.Sprintf("%s", s.block.Flags.listenServiceAddr.String())

	opts := &api.ManagerRemoteInitRequest{
		ServiceAddr: serviceAddr,
		ServiceType: api.ServiceType_BLOCK,
	}

	join, err := r.Init(ctx, opts)

	if err != nil {
		return nil, err
	}

	serviceID, err := uuid.Parse(join.ServiceID)

	if err != nil {
		return nil, err
	}

	key := &config.Key{
		ServiceID: serviceID,
	}

	saveErr := key.Save(s.block.Flags.configDir)

	if saveErr != nil {
		st := status.New(codes.Internal, saveErr.Error())

		return nil, st.Err()
	}

	gossipID := uuid.New()

	gossipMember := &gossip.Member{
		ServiceAddr: opts.ServiceAddr,
		ServiceID:   serviceID.String(),
		ServiceType: opts.ServiceType.String(),
	}

	memberlist, err := gossip.NewMemberList(gossipID, gossipMember, s.block.Flags.listenGossipAddr.Port)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	_, joinErr := memberlist.Join([]string{join.GossipAddr})

	if joinErr != nil {
		st := status.New(codes.Internal, joinErr.Error())

		return nil, st.Err()
	}

	res := &api.BlockControlInitResponse{
		ServiceID: serviceID.String(),
	}

	return res, nil
}

func (s *controlServer) Leave(ctx context.Context, in *api.BlockControlLeaveRequest) (*api.BlockControlLeaveResponse, error) {

	res := &api.BlockControlLeaveResponse{}

	return res, nil
}

// ControlServer : starts block control server
func ControlServer(b *Block) {
	socketPath := fmt.Sprintf("%s/control.sock", b.Flags.configDir)

	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("unix", socketPath)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	cs := &controlServer{
		block: b,
	}

	logrus.Info("Started block control gRPC socket server.")

	api.RegisterBlockControlServer(s, cs)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
