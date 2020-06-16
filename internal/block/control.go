package block

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/erkrnt/symphony/api"
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

	c := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	addr := fmt.Sprintf("%s", s.block.Flags.listenServiceAddr.String())

	opts := &api.ManagerRemoteInitRequest{
		Addr: addr,
		Type: api.ServiceType_BLOCK,
	}

	join, err := c.Init(ctx, opts)

	if err != nil {
		return nil, err
	}

	serviceID, err := uuid.Parse(join.Id)

	if err != nil {
		return nil, err
	}

	gossipID := uuid.New()

	gossipMember := &gossip.Member{
		Id:          gossipID.String(),
		ServiceAddr: addr,
		ServiceId:   serviceID.String(),
		ServiceType: opts.Type.String(),
	}

	memberlist, err := gossip.NewMemberList(gossipMember, s.block.Flags.listenGossipAddr.Port)

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
		Id: serviceID.String(),
	}

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
