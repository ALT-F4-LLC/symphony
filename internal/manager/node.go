package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Node : manager node
type Node struct {
	Flags *config.Flags
	Key   *config.Key
	Raft  *cluster.RaftNode
	State *cluster.RaftState

	mu sync.Mutex
}

const joinTokenManagerKey = "jointoken:manager"

const joinTokenWorkerKey = "jointoken:worker"

// FindOrCreateJoinTokens : looks up join tokens in the raft
func (n *Node) FindOrCreateJoinTokens() (*JoinTokens, error) {
	var jtm string
	var jtw string

	jtm, jtmOk := n.State.Lookup(joinTokenManagerKey)

	if !jtmOk {
		jtm = GenerateToken()

		n.State.Propose("jointoken:manager", jtm)

		log.Printf("Generated new manager join token: %s", jtm)
	}

	jtw, jtwOk := n.State.Lookup(joinTokenWorkerKey)

	if !jtwOk {
		jtw = GenerateToken()

		n.State.Propose("jointoken:worker", jtw)

		log.Printf("Generated new worker join token: %s", jtw)
	}

	jt := &JoinTokens{
		Manager: jtm,
		Worker:  jtw,
	}

	return jt, nil
}

// FindOrCreateMembers : finds or creates peers into state
func (n *Node) FindOrCreateMembers() ([]*api.Member, error) {
	m, ok := n.State.Lookup("members")

	if !ok {
		peersJSON, _ := json.Marshal(n.Raft.Members)

		n.State.Propose("members", string(peersJSON))

		return n.Raft.Members, nil
	}

	var members []*api.Member

	if err := json.Unmarshal([]byte(m), &members); err != nil {
		return nil, err
	}

	return members, nil
}

// SaveKey : Saves the key file current state
func (n *Node) SaveKey() error {
	n.mu.Lock()

	keyJSON, _ := json.Marshal(n.Key)

	path := fmt.Sprintf("%s/key.json", n.Flags.ConfigDir)

	err := ioutil.WriteFile(path, keyJSON, 0644)

	defer n.mu.Unlock()

	if err != nil {
		return err
	}

	fields := logrus.Fields{
		"RAFT_NODE_ID": n.Key.RaftNodeID,
	}

	logrus.WithFields(fields).Info("Updated key.json file")

	return nil
}

// StartControlServer : starts manager control server
func (n *Node) StartControlServer() {
	socketPath := fmt.Sprintf("%s/control.sock", n.Flags.ConfigDir)

	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("unix", socketPath)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	cs := &ControlServer{
		Node: n,
	}

	logrus.Info("Started manager control gRPC socket server.")

	api.RegisterManagerControlServer(s, cs)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// StartRemoteServer : starts Raft memebership server
func (n *Node) StartRemoteServer() {
	lis, err := net.Listen("tcp", n.Flags.ListenRemoteAddr.String())

	if err != nil {
		log.Fatal("Failed to listen")
	}

	s := grpc.NewServer()

	server := &RemoteServer{
		Node: n,
	}

	logrus.Info("Started manager remote gRPC tcp server.")

	api.RegisterManagerRemoteServer(s, server)

	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve")
	}
}

// NewNode : creates a new manager struct
func NewNode() (*Node, error) {
	flags, err := config.GetFlags()

	if err != nil {
		return nil, err
	}

	key, err := config.GetKey(flags.ConfigDir)

	if err != nil {
		return nil, err
	}

	m := &Node{
		Flags: flags,
		Key:   key,
	}

	if key.RaftNodeID != 0 {
		join := true

		raft, state, err := cluster.NewRaft(m.Flags, join, nil, key.RaftNodeID)

		if err != nil {
			return nil, err
		}

		m.Raft = raft

		m.State = state
	}

	return m, nil
}
