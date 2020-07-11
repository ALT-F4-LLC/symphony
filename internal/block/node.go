package block

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/erkrnt/symphony/internal/pkg/gossip"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Block : block node
type Block struct {
	ErrorC chan string

	flags      *flags
	key        *config.Key
	memberlist *memberlist.Memberlist
}

// Start : handles start of manager service
func (b *Block) Start() {
	key, err := b.key.Get(b.flags.configDir)

	if err != nil {
		b.ErrorC <- err.Error()
	}

	if key.ClusterID != nil && key.ServiceID != nil {
		restartErr := b.restart(key)

		if restartErr != nil {
			b.ErrorC <- restartErr.Error()
		}
	}

	go b.listenControl()

	go b.listenRemote()
}

func (b *Block) listenControl() {
	socketPath := fmt.Sprintf("%s/control.sock", b.flags.configDir)

	if err := os.RemoveAll(socketPath); err != nil {
		b.ErrorC <- err.Error()
	}

	lis, err := net.Listen("unix", socketPath)

	if err != nil {
		b.ErrorC <- err.Error()
	}

	s := grpc.NewServer()

	cs := &endpoints{
		block: b,
	}

	api.RegisterBlockServer(s, cs)

	logrus.Info("Started block control gRPC socket server.")

	serveErr := s.Serve(lis)

	if serveErr != nil {
		b.ErrorC <- serveErr.Error()
	}
}

func (b *Block) listenRemote() {
	lis, err := net.Listen("tcp", b.flags.listenServiceAddr.String())

	if err != nil {
		b.ErrorC <- err.Error()
	}

	s := grpc.NewServer()

	server := &endpoints{
		block: b,
	}

	api.RegisterBlockServer(s, server)

	logrus.Info("Started block remote gRPC tcp server.")

	serveErr := s.Serve(lis)

	if serveErr != nil {
		b.ErrorC <- err.Error()
	}
}

func (b *Block) restart(key *config.Key) error {
	managers := key.Endpoints

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
			ClusterID: key.ClusterID.String(),
			ServiceID: key.ServiceID.String(),
		}

		join, err := r.ServiceJoin(ctx, opts)

		if err != nil {
			logrus.Debug("Restart join failed to remote peer... skipping.")
			continue
		}

		gossipMember := &gossip.Member{
			ServiceAddr: b.flags.listenServiceAddr.String(),
			ServiceID:   join.ServiceID,
			ServiceType: api.ServiceType_BLOCK.String(),
		}

		listenGossipAddr := b.flags.listenGossipAddr

		memberlist, err := gossip.NewMemberList(gossipMember, listenGossipAddr.Port)

		if err != nil {
			return err
		}

		b.memberlist = memberlist

		_, joinErr := memberlist.Join([]string{join.GossipAddr})

		if joinErr != nil {
			return err
		}

		return nil
	}

	return errors.New("invalid_manager_endpoints")
}

// New : creates new block node
func New() (*Block, error) {
	errorC := make(chan string)

	key := &config.Key{}

	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	if flags.verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	block := &Block{
		ErrorC: errorC,

		flags: flags,
		key:   key,
	}

	return block, nil
}
