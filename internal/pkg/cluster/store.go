package cluster

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"
	"sync"

	"go.etcd.io/etcd/etcdserver/api/snap"
)

// KvStore : cluster store struct
type KvStore struct {
	Mutex       sync.RWMutex
	Snapshotter *snap.Snapshotter
	Store       map[string]string
}

// KeyVal : struct describing a key value
type KeyVal struct {
	Key string
	Val string
}

// GetSnapshot : gets the snapshot of a store
func (s *KvStore) GetSnapshot() ([]byte, error) {
	s.Mutex.RLock()

	defer s.Mutex.RUnlock()

	return json.Marshal(s.Store)
}

// NewStore : creates a new store for the raft
func NewStore(commitC <-chan *string, errorC <-chan error, node *Member, snapshotter *snap.Snapshotter) *KvStore {
	store := &KvStore{Snapshotter: snapshotter}

	// replay log into store
	store.ReadCommits(commitC, errorC)

	// read commits from raft into KvStore map until error
	go store.ReadCommits(commitC, errorC)

	return store
}

// RecoverFromSnapshot : unmarshals data from snapshot
func (s *KvStore) RecoverFromSnapshot(snapshot []byte) error {
	var store map[string]string

	if err := json.Unmarshal(snapshot, &store); err != nil {
		return err
	}

	s.Mutex.Lock()

	defer s.Mutex.Unlock()

	s.Store = store

	return nil
}

// ReadCommits : read commits in the commit channel
func (s *KvStore) ReadCommits(commitC <-chan *string, errorC <-chan error) {
	for data := range commitC {
		if data == nil {
			// done replaying log; new data incoming
			// OR signaled to load snapshot
			snapshot, err := s.Snapshotter.Load()

			if err == snap.ErrNoSnapshot {
				log.Print("ErrNoSnapshot")
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

		s.Store[dataKv.Key] = dataKv.Val

		s.Mutex.Unlock()
	}

	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}
