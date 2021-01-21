package block

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Block : block service
type Block struct {
	ErrorC chan error

	Flags *Flags
	Key   *service.Key
}

const (
	GRPC_RECONNECT_MAX_ATTEMPTS    = 3
	GRPC_RECONNECT_TIMEOUT_SECONDS = 5
)

var GRPC_RECONNECT_ATTEMPTS = 0

// New : creates new block service
func New() (*Block, error) {
	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	if flags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	key, err := service.GetKey(flags.ConfigDir)

	if err != nil {
		return nil, err
	}

	block := &Block{
		ErrorC: make(chan error),
		Flags:  flags,
		Key:    key,
	}

	return block, nil
}

// Start : handles start of block service
func (b *Block) Start() {
	if b.Key.ServiceID != nil {
		err := b.restart()

		if err != nil {
			b.ErrorC <- err
		}
	}

	go b.startControl()

	go b.startHealth()
}

func (b *Block) handleGRPCReconnect(err error) {
	st, ok := status.FromError(err)

	if ok && codes.Code(st.Code()) == codes.Unavailable {
		GRPC_RECONNECT_ATTEMPTS = GRPC_RECONNECT_ATTEMPTS + 1

		go b.listenGRPC()

		return
	}

	b.ErrorC <- err

	return
}

func (b *Block) startControl() {
	socketPath := fmt.Sprintf("%s/control.sock", b.Flags.ConfigDir)

	if err := os.RemoveAll(socketPath); err != nil {
		b.ErrorC <- err
	}

	lis, err := net.Listen("unix", socketPath)

	if err != nil {
		b.ErrorC <- err
	}

	grpcServer := grpc.NewServer()

	controlServer := &GRPCServerControl{
		Block: b,
	}

	api.RegisterControlServer(grpcServer, controlServer)

	logrus.Info("Started block control gRPC socket server.")

	serveErr := grpcServer.Serve(lis)

	if serveErr != nil {
		b.ErrorC <- serveErr
	}
}

func (b *Block) startHealth() {
	listenAddr := b.Flags.HealthServiceAddr.String()

	listen, err := net.Listen("tcp", listenAddr)

	if err != nil {
		b.ErrorC <- err
	}

	grpcServer := grpc.NewServer()

	healthServer := &service.GRPCServerHealth{}

	api.RegisterHealthServer(grpcServer, healthServer)

	logrus.Info("Started health grpc server.")

	serveErr := grpcServer.Serve(listen)

	if serveErr != nil {
		b.ErrorC <- serveErr
	}
}

func (b *Block) listenGRPC() {
	if GRPC_RECONNECT_ATTEMPTS >= 1 && GRPC_RECONNECT_ATTEMPTS <= GRPC_RECONNECT_MAX_ATTEMPTS {
		timeout := time.Duration(GRPC_RECONNECT_ATTEMPTS * GRPC_RECONNECT_TIMEOUT_SECONDS)

		time.Sleep(timeout * time.Second)

		output := fmt.Sprintf("Service reconnection to APIServer attempt %d...", GRPC_RECONNECT_ATTEMPTS)

		logrus.Info(output)
	}

	if GRPC_RECONNECT_ATTEMPTS > GRPC_RECONNECT_MAX_ATTEMPTS {
		err := errors.New("Service unable to connect to APIServer")

		b.ErrorC <- err

		return
	}

	conn := service.NewClientConnTCP(b.Flags.APIServerAddr.String())

	defer conn.Close()

	client := api.NewAPIServerClient(conn)

	in := &api.RequestState{
		ServiceID: b.Key.ServiceID.String(),
	}

	pvs, err := client.StatePhysicalVolumes(context.Background(), in)

	if err != nil && GRPC_RECONNECT_ATTEMPTS <= GRPC_RECONNECT_MAX_ATTEMPTS {
		b.handleGRPCReconnect(err)

		return
	}

	vgs, err := client.StateVolumeGroups(context.Background(), in)

	if err != nil && GRPC_RECONNECT_ATTEMPTS <= GRPC_RECONNECT_MAX_ATTEMPTS {
		b.handleGRPCReconnect(err)

		return
	}

	lvs, err := client.StateLogicalVolumes(context.Background(), in)

	if err != nil && GRPC_RECONNECT_ATTEMPTS <= GRPC_RECONNECT_MAX_ATTEMPTS {
		b.handleGRPCReconnect(err)

		return
	}

	GRPC_RECONNECT_ATTEMPTS = 0

	logrus.Info("Service successfully connected to APIServer")

	for {
		pv, err := pvs.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			b.handleGRPCReconnect(err)

			return
		}

		if pv != nil {
			logrus.Print(pv)
		}
	}

	for {
		vg, err := vgs.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			b.handleGRPCReconnect(err)

			return
		}

		if vg != nil {
			logrus.Print(vg)
		}
	}

	for {
		lv, err := lvs.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			b.handleGRPCReconnect(err)

			return
		}

		if lv != nil {
			logrus.Print(lv)
		}
	}
}

func (b *Block) restart() error {
	apiserverAddr := b.Flags.APIServerAddr

	if apiserverAddr == nil {
		return errors.New("invalid_apiserver_addr")
	}

	conn := service.NewClientConnTCP(apiserverAddr.String())

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), service.ContextTimeout)

	defer cancel()

	c := api.NewAPIServerClient(conn)

	serviceOptions := &api.RequestService{
		ServiceID: b.Key.ServiceID.String(),
	}

	service, err := c.GetService(ctx, serviceOptions)

	if err != nil {
		return err
	}

	if service == nil {
		return errors.New("invalid_service_id")
	}

	go b.listenGRPC()

	return nil
}
