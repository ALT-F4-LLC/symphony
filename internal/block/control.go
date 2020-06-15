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
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type controlServer struct {
	block *Block
}

func (s *controlServer) Join(ctx context.Context, in *api.BlockControlJoinRequest) (*api.BlockControlJoinResponse, error) {
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

	serviceAddr := fmt.Sprintf("%s", s.block.Flags.listenRemoteAddr.String())

	serviceJoinOpts := &api.ManagerRemoteJoinRequest{
		Addr: serviceAddr,
		Type: api.ServiceType_BLOCK,
	}

	serviceJoin, err := c.Join(ctx, serviceJoinOpts)

	if err != nil {
		return nil, err
	}

	serviceID, err := uuid.Parse(serviceJoin.Id)

	if err != nil {
		return nil, err
	}

	saveServiceIDErr := s.block.SaveServiceID(serviceID)

	if saveServiceIDErr != nil {
		return nil, saveServiceIDErr
	}

	addr := s.block.Flags.listenGossipAddr

	opts := &api.ManagerRemoteJoinGossipRequest{
		GossipAddr: addr.String(),
		ServiceId:  serviceID.String(),
	}

	join, err := c.JoinGossip(ctx, opts)

	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(join.GossipId)

	if err != nil {
		return nil, err
	}

	memberlist, gossipErr := gossip.NewMemberList(id, addr.Port)

	if gossipErr != nil {
		return nil, err
	}

	log.Print(join.GossipPeerAddr)

	if join.GossipPeerAddr != addr.String() {
		_, err := memberlist.Join([]string{join.GossipPeerAddr})

		if err != nil {
			return nil, err
		}
	}

	res := &api.BlockControlJoinResponse{}

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
