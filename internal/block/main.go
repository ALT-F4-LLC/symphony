package main

import (
	"context"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/raft/raftpb"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type blockServer struct{}

// RaftNodeConfig : used to configure raft from discovery endpoints
type RaftNodeConfig struct {
	ID    int
	Join  bool
	Peers []string
}

// Discover : handles main service discovery on join
func Discover(flags *Flags) (*RaftNodeConfig, error) {
	var id int

	var join = false

	var peers []string

	if flags.Join != nil {
		// TODO: Validate that join is a proper address

		// TODO: If flag is join - reach out over grpc
		// and join the server/inject the node id into
		// raft node configuration

		joinAddr, err := url.ParseRequestURI(*flags.Join)

		if err != nil {
			return nil, err
		}

		conn, err := grpc.Dial(joinAddr.String(), grpc.WithInsecure())

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		defer cancel()

		fields := &protobuff.GetServiceByHostnameFields{Hostname: hostname}

		service, err := c.GetServiceByHostname(ctx, fields)

		if err != nil {
			return nil, err
		}

		logrus.WithFields(logrus.Fields{"ID": service.ID}).Debug("getServiceByHostname")

		join = true
	}

	return nil, RaftNodeConfig{
		ID:    id,
		Join:  join,
		Peers: peers,
	}
}

func main() {
	flags := getFlags()

	if *flags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	proposeC := make(chan string)

	defer close(proposeC)

	configChangeC := make(chan raftpb.ConfChange)

	defer close(configChangeC)

	var kvs *service.KVStore

	getSnapshot := func() ([]byte, error) { return kvs.GetSnapshot() }

	var id int

	var join = false

	var peers []string

	if flags.Join != nil {
		// TODO: Validate that join is a proper address

		// TODO: If flag is join - reach out over grpc
		// and join the server/inject the node id into
		// raft node configuration

		joinAddr, err := url.ParseRequestURI(*flags.Join)

		if err != nil {
			panic(err)
		}

		join = true
	}

	node := service.NewRaftNode(id, peers, join, getSnapshot, proposeC, configChangeC)

	kvs = service.NewKVStore(<-node.Raft.SnapshotterReady, proposeC, node.CommitC, node.ErrorC)

	time.Sleep(30 * time.Second)

	// client := protobuff.NewManagerClient(conn)

	// service, err := service.Handshake(client, flags.Hostname, "block")

	// if err != nil {
	// 	panic(err)
	// }

	// lis, err := net.Listen("tcp", port)

	// if err != nil {
	// 	logrus.Fatal("Failed to listen: %v", err)
	// }

	// s := grpc.NewServer()

	// protobuff.RegisterBlockServer(s, &blockServer{})

	// logrus.WithFields(logrus.Fields{"ID": service.ID, "Port": port}).Info("Started block service.")

	// if err := s.Serve(lis); err != nil {
	// 	logrus.Fatal("Failed to serve: %v", err)
	// }
}
