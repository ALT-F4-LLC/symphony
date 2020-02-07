package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
)

// Node : manager node
type Node struct {
	Flags     *config.Flags
	Key       *config.Key
	RaftNode  *RaftNode
	RaftState *RaftState

	mu sync.Mutex
}

// SaveRaftNodeID : sets the RAFT_NODE_ID in key.json
func (n *Node) SaveRaftNodeID(nodeID uint64) error {
	n.Key.RaftNodeID = nodeID

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

		raft, state, err := NewRaft(m.Flags, join, nil, key.RaftNodeID)

		if err != nil {
			return nil, err
		}

		m.RaftNode = raft

		m.RaftState = state
	}

	return m, nil
}
