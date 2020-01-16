package cluster

import (
	"fmt"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"go.etcd.io/etcd/etcdserver/api/snap"
	"go.etcd.io/etcd/raft/raftpb"
)

// NewRaft : returns a new key-value store and raft node
func NewRaft(flags *config.Flags) (*RaftNode, *RaftState) {
	commitC := make(chan *string)

	confChangeC := make(chan raftpb.ConfChange)

	errorC := make(chan error)

	proposeC := make(chan string)

	var state *RaftState

	node := NewRaftNode(commitC, confChangeC, flags, errorC, proposeC, state)

	state = NewRaftState(commitC, errorC, node, proposeC, <-node.SnapshotterReady)

	return node, state
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

// NewRaftNode : creates a new raft member
func NewRaftNode(commitC chan<- *string, confChangeC <-chan raftpb.ConfChange, flags *config.Flags, errorC chan<- error, proposeC <-chan string, state *RaftState) *RaftNode {
	getSnapshot := func() ([]byte, error) { return state.GetSnapshot() }

	id := 1

	join := false

	peers := []string{"http://127.0.0.1:12379"}

	if flags.JoinAddr != nil {
		join = true
	}

	member := &RaftNode{
		CommitC:          commitC,
		ConfChangeC:      confChangeC,
		ErrorC:           errorC,
		ID:               id,
		Join:             join,
		Peers:            peers,
		GetSnapshot:      getSnapshot,
		HTTPDoneC:        make(chan struct{}),
		HTTPStopC:        make(chan struct{}),
		SnapshotterDir:   fmt.Sprintf("%s/node-%d-snapshot", flags.ConfigDir, id),
		SnapshotterReady: make(chan *snap.Snapshotter, 1),
		SnapCount:        defaultSnapshotCount,
		StopC:            make(chan struct{}),
		ProposeC:         proposeC,
		WalDir:           fmt.Sprintf("%s/node-%d", flags.ConfigDir, id),
	}

	go member.StartRaft()

	return member
}
