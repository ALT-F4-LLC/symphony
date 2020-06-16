package block

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type controlServer struct {
	block *Block
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

	serviceAddr := fmt.Sprintf("%s", s.block.Flags.listenAddr.String())

	serviceJoinOpts := &api.ManagerRemoteInitRequest{
		Addr: serviceAddr,
		Type: api.ServiceType_BLOCK,
	}

	serviceJoin, err := c.Init(ctx, serviceJoinOpts)

	if err != nil {
		return nil, err
	}

	serviceID, err := uuid.Parse(serviceJoin.Id)

	if err != nil {
		return nil, err
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
