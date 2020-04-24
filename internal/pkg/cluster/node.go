package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
)

// Node : manager node
type Node struct {
	Flags      *config.Flags
	Key        *config.Key
	Memberlist *memberlist.Memberlist
	Raft       *Raft
	State      *State

	mu sync.Mutex
}

// GetGossipMembers : retrieves gossip member list
func (n *Node) GetGossipMembers() ([]*api.GossipMember, error) {
	var members []*api.GossipMember

	list := n.Memberlist.Members()

	for _, m := range list {
		var member *api.GossipMember

		if err := json.Unmarshal(m.Meta, &member); err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	return members, nil
}

// GetRaftMembers : retrieves raft member list
func (n *Node) GetRaftMembers() ([]*api.RaftMember, error) {
	var members []*api.RaftMember

	m, ok := n.State.Lookup("members")

	if !ok {
		members = n.Raft.Members
	} else {
		err := json.Unmarshal([]byte(m), &members)

		if err != nil {
			return nil, err
		}
	}

	return members, nil
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

	n := &Node{
		Flags: flags,
		Key:   key,
	}

	if key.RaftNodeID != 0 {
		join := true

		httpStopC := make(chan struct{})

		raft, state, err := NewRaft(n.Flags, httpStopC, join, nil, key.RaftNodeID)

		if err != nil {
			return nil, err
		}

		var members []api.RaftMember

		m, ok := state.Lookup("members")

		if !ok {
			log.Print("we are not okay")
		}

		if err := json.Unmarshal([]byte(m), &members); err != nil {
			return nil, err
		}

		log.Print(members)

		// TODO : Once raft is re-initialized -> get a node from the members list to join gossip

		go func() {
			<-httpStopC

			log.Print("left the raft after a restart")
		}()

		n.Raft = raft

		n.State = state
	}

	return n, nil
}
