package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/erkrnt/symphony/api"
	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	socket              = kingpin.Flag("socket", "Sets the socket connection for the client.").String()
	clusterInitCommand  = kingpin.Command("init", "Initializes a new cluster with optional peers list (minimum of 3)")
	clusterJoinCommand  = kingpin.Command("join", "Joins an existing initialized cluster.")
	clusterPeersCommand = kingpin.Command("peers", "Returns a list of all active peers in state.")
	clusterInitJoinAddr = clusterInitCommand.Flag("join-addr", "Sets value to use for state.").String()
	clusterInitPeers    = clusterInitCommand.Flag("peers", "Sets value to use for state.").String()
	clusterJoinAddr     = clusterJoinCommand.Arg("addr", "Manager address of existing initialized cluster.").Required().String()
)

func newConnection() *grpc.ClientConn {
	abs, err := filepath.Abs(".")

	if err != nil {
		log.Fatal(err)
	}

	s := fmt.Sprintf("%s/control.sock", abs)

	if *socket != "" {
		abs, err := filepath.Abs(*socket)

		if err != nil {
			log.Fatal(err)
		}

		s = abs
	}

	conn, err := grpc.Dial(fmt.Sprintf("unix://%s", s), grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	return conn
}

func clusterInitCommandHandler() {
	if *clusterInitJoinAddr != "" && *clusterInitPeers != "" {
		log.Fatal("Cannot use --join-addr and --peers flags together.")
	}

	conn := newConnection()

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlInitializeRequest{}

	if *clusterInitJoinAddr != "" {
		opts.JoinAddr = *clusterInitJoinAddr
	}

	if *clusterInitPeers != "" {
		opts.Peers = strings.Split(*clusterInitPeers, ",")
	}

	log.Print(opts)

	_, initErr := c.ManagerControlInitialize(ctx, opts)

	if initErr != nil {
		log.Fatal(initErr)
	}
}

func clusterJoinCommandHandler() {
	conn := newConnection()

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlJoinRequest{JoinAddr: *clusterJoinAddr}

	_, joinErr := c.ManagerControlJoin(ctx, opts)

	if joinErr != nil {
		log.Fatal(joinErr)
	}
}

func clusterPeersCommandHandler() {
	conn := newConnection()

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlPeersRequest{}

	res, err := c.ManagerControlPeers(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(res.Peers)
}

func main() {
	switch kingpin.Parse() {
	case clusterInitCommand.FullCommand():
		clusterInitCommandHandler()
	case clusterJoinCommand.FullCommand():
		clusterJoinCommandHandler()
	case clusterPeersCommand.FullCommand():
		clusterPeersCommandHandler()
	}
}
