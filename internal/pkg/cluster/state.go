package cluster

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"
	"sync"

	"go.etcd.io/etcd/etcdserver/api/snap"
)

// RaftState : raft cluster state
type RaftState struct {
	Mutex       sync.RWMutex
	ProposeC    chan<- string // channel for proposing updates
	Snapshotter *snap.Snapshotter
	Store       map[string]string
}

// KeyVal : struct describing a key value
type KeyVal struct {
	Key string
	Val string
}

// GetSnapshot : gets the snapshot of a store
func (s *RaftState) GetSnapshot() ([]byte, error) {
	s.Mutex.RLock()

	defer s.Mutex.RUnlock()

	return json.Marshal(s.Store)
}

// Lookup : lookup key-value in store
func (s *RaftState) Lookup(key string) (string, bool) {
	s.Mutex.RLock()

	defer s.Mutex.RUnlock()

	v, ok := s.Store[key]

	return v, ok
}

// Propose : proposes changes to the state
func (s *RaftState) Propose(k string, v string) {
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(&KeyVal{k, v}); err != nil {
		log.Fatal(err)
	}

	s.ProposeC <- buf.String()
}

// ReadCommits : read commits in the commit channel
func (s *RaftState) ReadCommits(commitC <-chan *string, errorC <-chan error) {
	for data := range commitC {

		if data == nil {
			// done replaying log; new data incoming
			// OR signaled to load snapshot
			snapshot, err := s.Snapshotter.Load()

			if err == snap.ErrNoSnapshot {
				return
			}

			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)

			if err := s.RecoverFromSnapshot(snapshot.Data); err != nil {
				log.Fatal(err)
			}

			continue
		}

		var dataKv KeyVal

		dec := gob.NewDecoder(bytes.NewBufferString(*data))

		if err := dec.Decode(&dataKv); err != nil {
			log.Fatalf("Could not decode message (%v)", err)
		}

		s.Mutex.Lock()

		log.Printf("Locking state %s", dataKv.Val)

		s.Store[dataKv.Key] = dataKv.Val

		s.Mutex.Unlock()

		log.Printf("Unlocking state %s", dataKv.Val)
	}

	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}

// RecoverFromSnapshot : unmarshals data from snapshot
func (s *RaftState) RecoverFromSnapshot(snapshot []byte) error {
	var store map[string]string

	if err := json.Unmarshal(snapshot, &store); err != nil {
		return err
	}

	s.Mutex.Lock()

	defer s.Mutex.Unlock()

	s.Store = store

	return nil
}

// NewRaftState : creates a new store for the raft
func NewRaftState(commitC <-chan *string, errorC <-chan error, node *RaftNode, proposeC chan<- string, snapshotter *snap.Snapshotter) *RaftState {
	store := &RaftState{
		Snapshotter: snapshotter,
		Store:       make(map[string]string),
		ProposeC:    proposeC,
	}

	// replay log into store
	store.ReadCommits(commitC, errorC)

	// read commits from raft into KvStore map until error
	go store.ReadCommits(commitC, errorC)

	return store
}
