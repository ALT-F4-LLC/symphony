package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/raft"
	"github.com/sirupsen/logrus"
)

// Manager : manager node
type Manager struct {
	Flags *Flags
	Key   *Key
	Raft  *raft.Raft
	State *raft.State

	mu sync.Mutex
}

// GetRaftMembers : retrieves raft member list
func (m *Manager) GetRaftMembers() ([]*api.RaftMember, error) {
	var members []*api.RaftMember

	state, ok := m.State.Lookup("members")

	if !ok {
		members = m.Raft.Members
	} else {
		err := json.Unmarshal([]byte(state), &members)

		if err != nil {
			return nil, err
		}
	}

	return members, nil
}

// SaveRaftNodeID : sets the RAFT_NODE_ID in key.json
func (m *Manager) SaveRaftNodeID(nodeID uint64) error {
	m.Key.RaftNodeID = nodeID

	m.mu.Lock()

	keyJSON, _ := json.Marshal(m.Key)

	path := fmt.Sprintf("%s/key.json", m.Flags.ConfigDir)

	err := ioutil.WriteFile(path, keyJSON, 0644)

	defer m.mu.Unlock()

	if err != nil {
		return err
	}

	fields := logrus.Fields{
		"RAFT_NODE_ID": m.Key.RaftNodeID,
	}

	logrus.WithFields(fields).Info("Updated key.json file")

	return nil
}

// New : creates a new manager struct
func New() (*Manager, error) {
	flags, err := GetFlags()

	if err != nil {
		return nil, err
	}

	key, err := GetKey(flags.ConfigDir)

	if err != nil {
		return nil, err
	}

	n := &Manager{
		Flags: flags,
		Key:   key,
	}

	if key.RaftNodeID != 0 {
		join := true

		httpStopC := make(chan struct{})

		config := raft.Config{
			ConfigDir:  n.Flags.ConfigDir,
			HTTPStopC:  httpStopC,
			Join:       join,
			Members:    nil,
			NodeID:     key.RaftNodeID,
			ListenAddr: n.Flags.ListenRaftAddr,
		}

		raft, state, err := raft.New(config)

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

		go func() {
			<-httpStopC

			log.Print("left the raft after a restart")
		}()

		n.Raft = raft

		n.State = state
	}

	return n, nil
}
