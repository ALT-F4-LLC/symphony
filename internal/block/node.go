package block

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Block : block node
type Block struct {
	Node config.Node

	flags *flags
}

// Start : handles start of manager service
func (b *Block) Start() {
	if b.Node.Key.ClusterID != nil && b.Node.Key.ServiceID != nil {
		restartErr := b.restart()

		if restartErr != nil {
			b.Node.ErrorC <- restartErr.Error()
		}
	}

	go b.listenControl()

	go b.listenRemote()
}

func (b *Block) listenControl() {
	socketPath := fmt.Sprintf("%s/control.sock", b.flags.configDir)

	if err := os.RemoveAll(socketPath); err != nil {
		b.Node.ErrorC <- err.Error()
	}

	lis, err := net.Listen("unix", socketPath)

	if err != nil {
		b.Node.ErrorC <- err.Error()
	}

	s := grpc.NewServer()

	cs := &endpoints{
		block: b,
	}

	api.RegisterBlockServer(s, cs)

	logrus.Info("Started block control gRPC socket server.")

	serveErr := s.Serve(lis)

	if serveErr != nil {
		b.Node.ErrorC <- serveErr.Error()
	}
}

func (b *Block) listenRemote() {
	lis, err := net.Listen("tcp", b.flags.listenAddr.String())

	if err != nil {
		b.Node.ErrorC <- err.Error()
	}

	s := grpc.NewServer()

	server := &endpoints{
		block: b,
	}

	api.RegisterBlockServer(s, server)

	logrus.Info("Started block remote gRPC tcp server.")

	serveErr := s.Serve(lis)

	if serveErr != nil {
		b.Node.ErrorC <- err.Error()
	}
}

func (b *Block) restart() error {
	managers := b.Node.Key.Endpoints

	for _, manager := range managers {
		joinAddr, err := net.ResolveTCPAddr("tcp", manager)

		if err != nil {
			return err
		}

		conn, err := grpc.Dial(joinAddr.String(), grpc.WithInsecure())

		if err != nil {
			logrus.Debug("Restart join failed to remote peer... skipping.")
			continue
		}

		defer conn.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		r := api.NewManagerClient(conn)

		opts := &api.ManagerServiceJoinRequest{
			ClusterID: b.Node.Key.ClusterID.String(),
			ServiceID: b.Node.Key.ServiceID.String(),
		}

		join, err := r.ServiceJoin(ctx, opts)

		if err != nil {
			logrus.Debug("Restart join failed to remote peer... skipping.")
			continue
		}

		restartSerfErr := b.Node.RestartSerf(b.flags.listenAddr)

		if restartSerfErr != nil {
			return restartSerfErr
		}

		_, serfErr := b.Node.Serf.Join([]string{join.SerfAddress}, true)

		if serfErr != nil {
			return serfErr
		}

		return nil
	}

	return errors.New("invalid_manager_endpoints")
}

// New : creates new block node
func New() (*Block, error) {
	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	if flags.verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	key, err := config.NewKey(flags.configDir)

	if err != nil {
		return nil, err
	}

	serfInstance, err := cluster.NewSerf(flags.listenAddr, key.ServiceID, api.ServiceType_BLOCK)

	if err != nil {
		return nil, err
	}

	node := config.Node{
		Key:  key,
		Serf: serfInstance,
	}

	if err != nil {
		return nil, err
	}

	block := &Block{
		Node: node,

		flags: flags,
	}

	return block, nil
}
