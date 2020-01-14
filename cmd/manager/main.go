package main

import (
	"log"
	"time"

	logger "github.com/erkrnt/symphony/internal/pkg/logs"
	"go.etcd.io/etcd/raft/raftpb"

	"github.com/erkrnt/symphony/internal/pkg/flags"

	"github.com/erkrnt/symphony/internal/pkg/cluster"
)

func main() {
	l := logger.NewLogger()

	defer l.Sync()

	f, err := flags.GetManagerFlags(l)

	if err != nil {
		log.Panic(err)
	}

	proposeC := make(chan string)

	defer close(proposeC)

	confChangeC := make(chan raftpb.ConfChange)

	defer close(confChangeC)

	var kvs *cluster.KvStore

	commitC, errorC, node := cluster.NewRaftNode(confChangeC, f.ConfigDir, kvs, proposeC)

	kvs = cluster.NewStore(commitC, errorC, node, <-node.SnapshotterReady)

	time.Sleep(30 * time.Second)
}
