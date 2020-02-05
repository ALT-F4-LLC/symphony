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
	getCommand          = kingpin.Command("get", "Get a value from the raft state.")
	getCommandKey       = getCommand.Arg("key", "Specific key to retrieve for value.").Required().String()
	initCommand         = kingpin.Command("init", "Initializes a new cluster with optional peers list (minimum of 3)")
	initCommandJoinAddr = initCommand.Flag("join-addr", "Sets value to use for state.").String()
	initCommandPeers    = initCommand.Flag("peers", "Sets value to use for state.").String()
	joinCommand         = kingpin.Command("join", "Joins an existing initialized cluster.")
	joinCommandAddr     = joinCommand.Arg("addr", "Manager address of existing initialized cluster.").Required().String()
	setCommand          = kingpin.Command("set", "Sets a value into the raft state.")
	setCommandKey       = setCommand.Arg("key", "Specific key to set for value.").Required().String()
	setCommandValue     = setCommand.Arg("value", "Specific value to set for key.").Required().String()
	removeCommand       = kingpin.Command("remove", "Removes a node from the cluster.")
	removeCommandAddr   = removeCommand.Arg("addr", "Raft address of node to be removed from the cluster.").Required().String()
	socket              = kingpin.Flag("socket", "Sets the socket connection for the client.").String()
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

func getCommandHandler() {
	if *getCommandKey == "" {
		log.Fatal("Invalid parameters")
	}

	conn := newConnection()

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlGetValueRequest{
		Key: *getCommandKey,
	}

	res, err := c.ManagerControlGetValue(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(res.Value)
}

func initCommandHandler() {
	if *initCommandJoinAddr != "" && *initCommandPeers != "" {
		log.Fatal("Cannot use --join-addr and --peers flags together.")
	}

	conn := newConnection()

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlInitializeRequest{}

	if *initCommandJoinAddr != "" {
		opts.JoinAddr = *initCommandJoinAddr
	}

	if *initCommandPeers != "" {
		opts.Peers = strings.Split(*initCommandPeers, ",")
	}

	log.Print(opts)

	_, initErr := c.ManagerControlInitialize(ctx, opts)

	if initErr != nil {
		log.Fatal(initErr)
	}
}

func joinCommandHandler() {
	conn := newConnection()

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlJoinRequest{JoinAddr: *joinCommandAddr}

	_, joinErr := c.ManagerControlJoin(ctx, opts)

	if joinErr != nil {
		log.Fatal(joinErr)
	}
}

func setCommandHandler() {
	if *setCommandKey == "" || *setCommandValue == "" {
		log.Fatal("Invalid parameters")
	}

	conn := newConnection()

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlSetValueRequest{
		Key:   *setCommandKey,
		Value: *setCommandValue,
	}

	res, err := c.ManagerControlSetValue(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(res.Value)
}

func main() {
	switch kingpin.Parse() {
	case getCommand.FullCommand():
		getCommandHandler()
	case initCommand.FullCommand():
		initCommandHandler()
	case joinCommand.FullCommand():
		joinCommandHandler()
	case setCommand.FullCommand():
		log.Print("Needs to be implemented")
	case setCommand.FullCommand():
		setCommandHandler()
	}
}
