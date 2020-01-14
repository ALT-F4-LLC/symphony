package main

import (
	"github.com/erkrnt/symphony/internal/manager"
	"go.etcd.io/etcd/raft/raftpb"

	"github.com/erkrnt/symphony/internal/pkg/cluster"
)

func main() {
	m, err := manager.NewManager()

	if err != nil {
		m.Logger.Fatal(err.Error())
	}

	proposeC := make(chan string)

	defer close(proposeC)

	confChangeC := make(chan raftpb.ConfChange)

	defer close(confChangeC)

	var kvs *cluster.KvStore

	commitC, errorC, member := cluster.NewRaftMember(confChangeC, m.Flags.ConfigDir, kvs, proposeC)

	kvs = cluster.NewStore(commitC, errorC, member, <-member.SnapshotterReady)

	m.Member = member

	m.Store = kvs

	manager.StartRaftMembershipServer(m)
}
