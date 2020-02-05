package cluster

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"go.etcd.io/etcd/etcdserver/api/rafthttp"
	"go.etcd.io/etcd/etcdserver/api/snap"
	stats "go.etcd.io/etcd/etcdserver/api/v2stats"
	"go.etcd.io/etcd/pkg/fileutil"
	"go.etcd.io/etcd/pkg/types"
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"go.etcd.io/etcd/wal"
	"go.etcd.io/etcd/wal/walpb"
	"go.uber.org/zap"
)

// RaftNode : a member of the cluster
type RaftNode struct {
	AppliedIndex     uint64
	CommitC          chan<- *string         // entries committed to log (k,v)
	ConfChangeC      chan raftpb.ConfChange // proposed cluster config changes
	ConfState        raftpb.ConfState
	ErrorC           chan error // errors from raft session
	GetSnapshot      func() ([]byte, error)
	ID               uint64
	Join             bool
	ListenAddr       *net.TCPAddr
	Node             raft.Node
	Peers            []string
	Transport        *rafthttp.Transport
	HTTPStopC        chan struct{} // signals http server to shutdown
	HTTPDoneC        chan struct{} // signals http server shutdown complete
	LastIndex        uint64        // index of log at start
	ProposeC         <-chan string // proposed messages (k,v)
	RaftStorage      *raft.MemoryStorage
	SnapshotIndex    uint64
	Snapshotter      *snap.Snapshotter
	SnapshotterDir   string
	SnapshotterReady chan *snap.Snapshotter
	SnapCount        uint64
	StopC            chan struct{} // signals proposal channel closed
	Wal              *wal.WAL
	WalDir           string
}

// StoppableListener : creates a stoppable listener
type StoppableListener struct {
	*net.TCPListener
	StopC <-chan struct{}
}

var defaultSnapshotCount uint64 = 10000

var snapshotCatchUpEntriesN uint64 = 10000

// Accept : accepts connections via stoppableListener
func (ln StoppableListener) Accept() (c net.Conn, err error) {
	connC := make(chan *net.TCPConn, 1)

	errC := make(chan error, 1)

	go func() {
		tc, err := ln.AcceptTCP()

		if err != nil {
			errC <- err
			return
		}

		connC <- tc
	}()

	select {
	case <-ln.StopC:
		return nil, errors.New("server stopped")
	case err := <-errC:
		return nil, err
	case tc := <-connC:
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(3 * time.Minute)
		return tc, nil
	}
}

// EntriesToApply : compiles entires to apply to raft
func (rn *RaftNode) EntriesToApply(ents []raftpb.Entry) (nents []raftpb.Entry) {
	if len(ents) == 0 {
		return ents
	}

	firstIdx := ents[0].Index

	if firstIdx > rn.AppliedIndex+1 {
		log.Fatalf("First index of committed entry[%d] should <= progress.AppliedIndex[%d]+1", firstIdx, rn.AppliedIndex)
	}

	if rn.AppliedIndex-firstIdx+1 < uint64(len(ents)) {
		nents = ents[rn.AppliedIndex-firstIdx+1:]
	}

	return nents
}

// IsIDRemoved : determines if the members ID has been removed
func (rn *RaftNode) IsIDRemoved(id uint64) bool { return false }

// LoadSnapshot : loads snapshot from directory
func (rn *RaftNode) LoadSnapshot() *raftpb.Snapshot {
	snapshot, err := rn.Snapshotter.Load()

	if err != nil && err != snap.ErrNoSnapshot {
		log.Fatalf("Error loading snapshot (%v)", err)
	}

	return snapshot
}

// MaybeTriggerSnapshot : see if should trigger snapshot change
func (rn *RaftNode) MaybeTriggerSnapshot() {
	if rn.AppliedIndex-rn.SnapshotIndex <= rn.SnapCount {
		return
	}

	log.Printf("Start snapshot [applied index: %d | last snapshot index: %d]", rn.AppliedIndex, rn.SnapshotIndex)

	data, err := rn.GetSnapshot()

	if err != nil {
		log.Fatal(err)
	}

	snap, err := rn.RaftStorage.CreateSnapshot(rn.AppliedIndex, &rn.ConfState, data)

	if err != nil {
		log.Fatal(err)
	}

	if err := rn.SaveSnap(snap); err != nil {
		log.Fatal(err)
	}

	compactIndex := uint64(1)

	if rn.AppliedIndex > snapshotCatchUpEntriesN {
		compactIndex = rn.AppliedIndex - snapshotCatchUpEntriesN
	}

	if err := rn.RaftStorage.Compact(compactIndex); err != nil {
		log.Fatal(err)
	}

	log.Printf("Compacted log at index %d", compactIndex)

	rn.SnapshotIndex = rn.AppliedIndex
}

// OpenWAL : opens write access log
func (rn *RaftNode) OpenWAL(snapshot *raftpb.Snapshot) *wal.WAL {
	if !wal.Exist(rn.WalDir) {

		if err := os.Mkdir(rn.WalDir, 0750); err != nil {
			log.Fatalf("Cannot create dir for wal (%v)", err)
		}

		w, err := wal.Create(zap.NewExample(), rn.WalDir, nil)

		if err != nil {
			log.Fatalf("Create wal error (%v)", err)
		}

		w.Close()
	}

	walSnap := walpb.Snapshot{}

	if snapshot != nil {
		walSnap.Index, walSnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}

	log.Printf("Loading WAL at term %d and index %d", walSnap.Term, walSnap.Index)

	w, err := wal.Open(zap.NewExample(), rn.WalDir, walSnap)

	if err != nil {
		log.Fatalf("Error loading wal (%v)", err)
	}

	return w
}

// PublishEntries : writes committed log entries to commit channel and returns
// whether all entries could be published.
func (rn *RaftNode) PublishEntries(ents []raftpb.Entry) bool {
	for i := range ents {
		switch ents[i].Type {
		case raftpb.EntryNormal:
			if len(ents[i].Data) == 0 {
				// ignore empty messages
				break
			}

			s := string(ents[i].Data)

			select {
			case rn.CommitC <- &s:
			case <-rn.StopC:
				return false
			}
		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange

			if err := cc.Unmarshal(ents[i].Data); err != nil {
				log.Fatal(err)
			}

			rn.ConfState = *rn.Node.ApplyConfChange(cc)

			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if cc.NodeID != uint64(rn.ID) && len(cc.Context) > 0 {
					rn.Transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(rn.ID) {
					log.Println("I've been removed from the cluster! Shutting down.")
					return false
				}

				rn.Transport.RemovePeer(types.ID(cc.NodeID))
			}
		}

		// after commit, update appliedIndex
		rn.AppliedIndex = ents[i].Index

		// special nil commit to signal replay has finished
		if ents[i].Index == rn.LastIndex {
			select {
			case rn.CommitC <- nil:
			case <-rn.StopC:
				return false
			}
		}
	}

	return true
}

// PublishSnapshot : handles publishing the snapshot to the KvStore
func (rn *RaftNode) PublishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	log.Printf("Publishing snapshot at index %d", rn.SnapshotIndex)

	defer log.Printf("Finished publishing snapshot at index %d", rn.SnapshotIndex)

	if snapshotToSave.Metadata.Index <= rn.AppliedIndex {
		log.Fatalf("Snapshot index [%d] should > progress.AppliedIndex [%d]", snapshotToSave.Metadata.Index, rn.AppliedIndex)
	}

	rn.CommitC <- nil

	rn.AppliedIndex = snapshotToSave.Metadata.Index

	rn.ConfState = snapshotToSave.Metadata.ConfState

	rn.SnapshotIndex = snapshotToSave.Metadata.Index
}

// ReplayWAL : replays the write access log
func (rn *RaftNode) ReplayWAL() *wal.WAL {
	log.Printf("Replaying WAL of member %d", rn.ID)

	snapshot := rn.LoadSnapshot()

	w := rn.OpenWAL(snapshot)

	_, st, ents, err := w.ReadAll()

	if err != nil {
		log.Fatalf("Failed to read WAL (%v)", err)
	}

	rn.RaftStorage = raft.NewMemoryStorage()

	if snapshot != nil {
		rn.RaftStorage.ApplySnapshot(*snapshot)
	}

	rn.RaftStorage.SetHardState(st)

	// append to storage so raft starts at the right place in log
	rn.RaftStorage.Append(ents)

	// send nil once lastIndex is published so client knows commit channel is current
	if len(ents) > 0 {
		rn.LastIndex = ents[len(ents)-1].Index
	} else {
		rn.CommitC <- nil
	}

	return w
}

// SaveSnap : must save the snapshot index to the WAL before saving
// the snapshot to maintain the invariant that we only open the WAL
// at previously-saved snapshot indexes.
func (rn *RaftNode) SaveSnap(snap raftpb.Snapshot) error {
	walSnap := walpb.Snapshot{
		Index: snap.Metadata.Index,
		Term:  snap.Metadata.Term,
	}

	if err := rn.Wal.SaveSnapshot(walSnap); err != nil {
		return err
	}

	if err := rn.Snapshotter.SaveSnap(snap); err != nil {
		return err
	}

	return rn.Wal.ReleaseLockTo(snap.Metadata.Index)
}

// ServeChannels : serves raft channels for changes
func (rn *RaftNode) ServeChannels() {
	snap, err := rn.RaftStorage.Snapshot()

	if err != nil {
		log.Fatal(err)
	}

	rn.AppliedIndex = snap.Metadata.Index

	rn.ConfState = snap.Metadata.ConfState

	rn.SnapshotIndex = snap.Metadata.Index

	defer rn.Wal.Close()

	ticker := time.NewTicker(100 * time.Millisecond)

	defer ticker.Stop()

	// send proposals over raft
	go func() {
		confChangeCount := uint64(0)

		for rn.ProposeC != nil && rn.ConfChangeC != nil {
			select {
			case prop, ok := <-rn.ProposeC:
				if !ok {
					rn.ProposeC = nil
				} else {
					// blocks until accepted by raft state machine
					if err := rn.Node.Propose(context.TODO(), []byte(prop)); err != nil {
						log.Fatal(err)
					}
				}
			case cc, ok := <-rn.ConfChangeC:
				if !ok {
					rn.ConfChangeC = nil
				} else {
					confChangeCount++

					cc.ID = confChangeCount

					if err := rn.Node.ProposeConfChange(context.TODO(), cc); err != nil {
						log.Fatal(err)
					}
				}
			}
		}

		close(rn.StopC)
	}()

	// event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			rn.Node.Tick()

		// store raft entries to wal, then publish over commit channel
		case rd := <-rn.Node.Ready():
			err := rn.Wal.Save(rd.HardState, rd.Entries)

			if err != nil {
				log.Fatal(err)
			}

			if !raft.IsEmptySnap(rd.Snapshot) {
				rn.SaveSnap(rd.Snapshot)

				rn.RaftStorage.ApplySnapshot(rd.Snapshot)

				rn.PublishSnapshot(rd.Snapshot)
			}

			rn.RaftStorage.Append(rd.Entries)

			rn.Transport.Send(rd.Messages)

			if ok := rn.PublishEntries(rn.EntriesToApply(rd.CommittedEntries)); !ok {
				rn.Stop()
				return
			}

			rn.MaybeTriggerSnapshot()

			rn.Node.Advance()

		case err := <-rn.Transport.ErrorC:
			rn.WriteError(err)
			return

		case <-rn.StopC:
			rn.Stop()
			return
		}
	}
}

// ServeRaft : serves the raft
func (rn *RaftNode) ServeRaft() {
	ln, err := newStoppableListener(rn.ListenAddr.String(), rn.HTTPStopC)

	if err != nil {
		log.Fatalf("Failed to listen rafthttp (%v)", err)
	}

	err = (&http.Server{Handler: rn.Transport.Handler()}).Serve(ln)

	if err != nil {
		log.Fatal(err)
	}

	select {
	case <-rn.HTTPStopC:
	default:
		log.Fatalf("Failed to serve raft (%v)", err)
	}

	close(rn.HTTPDoneC)
}

// StartRaft : starts the raft
func (rn *RaftNode) StartRaft() {
	if !fileutil.Exist(rn.SnapshotterDir) {

		if err := os.Mkdir(rn.SnapshotterDir, 0750); err != nil {
			log.Fatalf("Cannot create dir for snapshot (%v)", err)
		}

	}

	rn.Snapshotter = snap.New(zap.NewExample(), rn.SnapshotterDir)

	rn.SnapshotterReady <- rn.Snapshotter

	oldwal := walExists(rn.WalDir)

	rn.Wal = rn.ReplayWAL()

	rpeers := make([]raft.Peer, len(rn.Peers))

	for i := range rpeers {
		rpeers[i] = raft.Peer{ID: uint64(i + 1), Context: []byte(rn.Peers[i])}
	}

	c := &raft.Config{
		ID:              rn.ID,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         rn.RaftStorage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}

	if oldwal || rn.Join {
		rn.Node = raft.RestartNode(c)
	} else {
		rn.Node = raft.StartNode(c, rpeers)
	}

	rn.Transport = &rafthttp.Transport{
		Logger:      zap.NewExample(),
		ID:          types.ID(rn.ID),
		ClusterID:   0x1000,
		Raft:        rn,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(strconv.Itoa(int(rn.ID))),
		ErrorC:      make(chan error),
	}

	rn.Transport.Start()

	for i := range rn.Peers {
		if i+1 != int(rn.ID) {
			rn.Transport.AddPeer(types.ID(i+1), []string{rn.Peers[i]})
		}
	}

	go rn.ServeRaft()

	go rn.ServeChannels()
}

// Stop : closes http, closes all channels, and stops raft.
func (rn *RaftNode) Stop() {
	rn.StopHTTP()

	close(rn.CommitC)

	close(rn.ErrorC)

	rn.Node.Stop()
}

// StopHTTP : closes http
func (rn *RaftNode) StopHTTP() {
	rn.Transport.Stop()

	close(rn.HTTPStopC)

	<-rn.HTTPDoneC
}

// Process : processes member messages
func (rn *RaftNode) Process(ctx context.Context, message raftpb.Message) error {
	return rn.Node.Step(ctx, message)
}

// ReportUnreachable : determines if the member is unreachable
func (rn *RaftNode) ReportUnreachable(id uint64) {}

// ReportSnapshot : reports snapshot status
func (rn *RaftNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {}

// WriteError : writes error to all listeners
func (rn *RaftNode) WriteError(err error) {
	rn.StopHTTP()

	close(rn.CommitC)

	rn.ErrorC <- err

	close(rn.ErrorC)

	rn.Node.Stop()
}

// GetMemberIndex : gets the index of a member in a set of peers
func GetMemberIndex(addr string, peers []string) (*int, error) {
	for i, a := range peers {
		if a == addr {
			return &i, nil
		}
	}
	return nil, errors.New("invalid_member")
}

// NewRaft : returns a new key-value store and raft node
func NewRaft(flags *config.Flags, join bool, nodeID uint64, peers []string) (*RaftNode, *RaftState, error) {
	commitC := make(chan *string)

	confChangeC := make(chan raftpb.ConfChange)

	errorC := make(chan error)

	proposeC := make(chan string)

	var state *RaftState

	node, err := NewRaftNode(commitC, confChangeC, errorC, flags, join, nodeID, peers, proposeC, state)

	if err != nil {
		return nil, nil, err
	}

	state = NewRaftState(commitC, errorC, node, proposeC, <-node.SnapshotterReady)

	return node, state, nil
}

// NewRaftNode : creates a new raft member
func NewRaftNode(commitC chan<- *string, confChangeC chan raftpb.ConfChange, errorC chan error, flags *config.Flags, join bool, nodeID uint64, peers []string, proposeC <-chan string, state *RaftState) (*RaftNode, error) {
	getSnapshot := func() ([]byte, error) { return state.GetSnapshot() }

	member := &RaftNode{
		CommitC:          commitC,
		ConfChangeC:      confChangeC,
		ErrorC:           errorC,
		ID:               nodeID,
		Join:             join,
		ListenAddr:       flags.ListenRaftAddr,
		Peers:            peers,
		GetSnapshot:      getSnapshot,
		HTTPDoneC:        make(chan struct{}),
		HTTPStopC:        make(chan struct{}),
		SnapshotterDir:   fmt.Sprintf("%s/node-%d-snapshot", flags.ConfigDir, nodeID),
		SnapshotterReady: make(chan *snap.Snapshotter, 1),
		SnapCount:        defaultSnapshotCount,
		StopC:            make(chan struct{}),
		ProposeC:         proposeC,
		WalDir:           fmt.Sprintf("%s/node-%d", flags.ConfigDir, nodeID),
	}

	go member.StartRaft()

	return member, nil
}
