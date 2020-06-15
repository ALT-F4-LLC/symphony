package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/raft"
	"github.com/google/uuid"
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

// GetGossipMembers : gets members from the gossip protocol
func (m *Manager) GetGossipMembers() ([]*api.GossipMember, error) {
	var members []*api.GossipMember

	state, ok := m.State.Lookup("gossip:members")

	if !ok {
		members = make([]*api.GossipMember, 0)
	} else {
		err := json.Unmarshal([]byte(state), &members)

		if err != nil {
			return nil, err
		}
	}

	return members, nil
}

// GetRaftMembers : retrieves raft member list
func (m *Manager) GetRaftMembers() ([]*api.RaftMember, error) {
	var members []*api.RaftMember

	state, ok := m.State.Lookup("raft:members")

	if !ok {
		return nil, errors.New("invalid_raft_state")
	} else {
		err := json.Unmarshal([]byte(state), &members)

		if err != nil {
			return nil, err
		}
	}

	return members, nil
}

// GetServiceMembers : retrieves services list
func (m *Manager) GetServiceMembers() ([]*api.ServiceMember, error) {
	var services []*api.ServiceMember

	state, ok := m.State.Lookup("service:members")

	if !ok {
		return nil, errors.New("invalid_raft_state")
	} else {
		err := json.Unmarshal([]byte(state), &services)

		if err != nil {
			return nil, err
		}
	}

	return services, nil
}

// SaveRaftID : sets the RAFT_ID in key.json
func (m *Manager) SaveRaftID(id uint64) error {
	m.Key.RaftID = id

	m.mu.Lock()

	keyJSON, _ := json.Marshal(m.Key)

	path := fmt.Sprintf("%s/key.json", m.Flags.ConfigDir)

	err := ioutil.WriteFile(path, keyJSON, 0644)

	defer m.mu.Unlock()

	if err != nil {
		return err
	}

	fields := logrus.Fields{
		"RAFT_ID": m.Key.RaftID,
	}

	logrus.WithFields(fields).Info("Updated key.json file")

	return nil
}

// SaveServiceID : sets the SERVICE_ID in key.json
func (m *Manager) SaveServiceID(id uuid.UUID) error {
	m.Key.ServiceID = id

	m.mu.Lock()

	keyJSON, err := json.Marshal(m.Key)

	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/key.json", m.Flags.ConfigDir)

	writeErr := ioutil.WriteFile(path, keyJSON, 0644)

	if writeErr != nil {
		return writeErr
	}

	defer m.mu.Unlock()

	fields := logrus.Fields{
		"SERVICE_ID": m.Key.ServiceID,
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

	if key.RaftID != 0 {
		join := true

		httpStopC := make(chan struct{})

		config := raft.Config{
			ConfigDir:  n.Flags.ConfigDir,
			HTTPStopC:  httpStopC,
			Join:       join,
			Members:    nil,
			NodeID:     key.RaftID,
			ListenAddr: n.Flags.ListenRaftAddr,
		}

		raft, state, err := raft.New(config)

		if err != nil {
			return nil, err
		}

		var members []api.RaftMember

		m, ok := state.Lookup("raft:members")

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
