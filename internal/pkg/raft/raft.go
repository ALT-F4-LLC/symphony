package raft

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/erkrnt/symphony/api"
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

// Config : configuration for new raft
type Config struct {
	ConfigDir  string
	HTTPStopC  chan struct{}
	Join       bool
	ListenAddr *net.TCPAddr
	Members    []*api.RaftMember
	NodeID     uint64
}

// Raft : creates a new raft struct
type Raft struct {
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
	Members          []*api.RaftMember
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

var defaultSnapshotCount uint64 = 10000

var snapshotCatchUpEntriesN uint64 = 10000

// EntriesToApply : compiles entires to apply to raft
func (r *Raft) EntriesToApply(ents []raftpb.Entry) (nents []raftpb.Entry) {
	if len(ents) == 0 {
		return ents
	}

	firstIdx := ents[0].Index

	if firstIdx > r.AppliedIndex+1 {
		log.Fatalf("First index of committed entry[%d] should <= progress.AppliedIndex[%d]+1", firstIdx, r.AppliedIndex)
	}

	if r.AppliedIndex-firstIdx+1 < uint64(len(ents)) {
		nents = ents[r.AppliedIndex-firstIdx+1:]
	}

	return nents
}

// IsIDRemoved : determines if the members ID has been removed
func (r *Raft) IsIDRemoved(id uint64) bool { return false }

// LoadSnapshot : loads snapshot from directory
func (r *Raft) LoadSnapshot() *raftpb.Snapshot {
	snapshot, err := r.Snapshotter.Load()

	if err != nil && err != snap.ErrNoSnapshot {
		log.Fatalf("Error loading snapshot (%v)", err)
	}

	return snapshot
}

// MaybeTriggerSnapshot : see if should trigger snapshot change
func (r *Raft) MaybeTriggerSnapshot() {
	if r.AppliedIndex-r.SnapshotIndex <= r.SnapCount {
		return
	}

	log.Printf("Start snapshot [applied index: %d | last snapshot index: %d]", r.AppliedIndex, r.SnapshotIndex)

	data, err := r.GetSnapshot()

	if err != nil {
		log.Fatal(err)
	}

	snap, err := r.RaftStorage.CreateSnapshot(r.AppliedIndex, &r.ConfState, data)

	if err != nil {
		log.Fatal(err)
	}

	if err := r.SaveSnap(snap); err != nil {
		log.Fatal(err)
	}

	compactIndex := uint64(1)

	if r.AppliedIndex > snapshotCatchUpEntriesN {
		compactIndex = r.AppliedIndex - snapshotCatchUpEntriesN
	}

	if err := r.RaftStorage.Compact(compactIndex); err != nil {
		log.Fatal(err)
	}

	log.Printf("Compacted log at index %d", compactIndex)

	r.SnapshotIndex = r.AppliedIndex
}

// OpenWAL : opens write access log
func (r *Raft) OpenWAL(snapshot *raftpb.Snapshot) *wal.WAL {
	if !wal.Exist(r.WalDir) {

		if err := os.Mkdir(r.WalDir, 0750); err != nil {
			log.Fatalf("Cannot create dir for wal (%v)", err)
		}

		w, err := wal.Create(zap.NewExample(), r.WalDir, nil)

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

	w, err := wal.Open(zap.NewExample(), r.WalDir, walSnap)

	if err != nil {
		log.Fatalf("Error loading wal (%v)", err)
	}

	return w
}

// PublishEntries : writes committed log entries to commit channel and returns
// whether all entries could be published.
func (r *Raft) PublishEntries(ents []raftpb.Entry) bool {
	for i := range ents {
		switch ents[i].Type {
		case raftpb.EntryNormal:
			if len(ents[i].Data) == 0 {
				// ignore empty messages
				break
			}

			s := string(ents[i].Data)

			select {
			case r.CommitC <- &s:
			case <-r.StopC:
				return false
			}
		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange

			if err := cc.Unmarshal(ents[i].Data); err != nil {
				log.Fatal(err)
			}

			r.ConfState = *r.Node.ApplyConfChange(cc)

			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if cc.NodeID != uint64(r.ID) && len(cc.Context) > 0 {
					r.Transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(r.ID) {
					log.Println("I've been removed from the cluster! Shutting down.")

					return false
				}

				r.Transport.RemovePeer(types.ID(cc.NodeID))
			}
		}

		// after commit, update appliedIndex
		r.AppliedIndex = ents[i].Index

		// special nil commit to signal replay has finished
		if ents[i].Index == r.LastIndex {
			select {
			case r.CommitC <- nil:
			case <-r.StopC:
				return false
			}
		}
	}

	return true
}

// PublishSnapshot : handles publishing the snapshot to the KvStore
func (r *Raft) PublishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	log.Printf("Publishing snapshot at index %d", r.SnapshotIndex)

	defer log.Printf("Finished publishing snapshot at index %d", r.SnapshotIndex)

	if snapshotToSave.Metadata.Index <= r.AppliedIndex {
		log.Fatalf("Snapshot index [%d] should > progress.AppliedIndex [%d]", snapshotToSave.Metadata.Index, r.AppliedIndex)
	}

	r.CommitC <- nil

	r.AppliedIndex = snapshotToSave.Metadata.Index

	r.ConfState = snapshotToSave.Metadata.ConfState

	r.SnapshotIndex = snapshotToSave.Metadata.Index
}

// ReplayWAL : replays the write access log
func (r *Raft) ReplayWAL() *wal.WAL {
	log.Printf("Replaying WAL of member %d", r.ID)

	snapshot := r.LoadSnapshot()

	w := r.OpenWAL(snapshot)

	_, st, ents, err := w.ReadAll()

	if err != nil {
		log.Fatalf("Failed to read WAL (%v)", err)
	}

	r.RaftStorage = raft.NewMemoryStorage()

	if snapshot != nil {
		r.RaftStorage.ApplySnapshot(*snapshot)
	}

	r.RaftStorage.SetHardState(st)

	// append to storage so raft starts at the right place in log
	r.RaftStorage.Append(ents)

	// send nil once lastIndex is published so client knows commit channel is current
	if len(ents) > 0 {
		r.LastIndex = ents[len(ents)-1].Index
	} else {
		r.CommitC <- nil
	}

	return w
}

// SaveSnap : must save the snapshot index to the WAL before saving
// the snapshot to maintain the invariant that we only open the WAL
// at previously-saved snapshot indexes.
func (r *Raft) SaveSnap(snap raftpb.Snapshot) error {
	walSnap := walpb.Snapshot{
		Index: snap.Metadata.Index,
		Term:  snap.Metadata.Term,
	}

	if err := r.Wal.SaveSnapshot(walSnap); err != nil {
		return err
	}

	if err := r.Snapshotter.SaveSnap(snap); err != nil {
		return err
	}

	return r.Wal.ReleaseLockTo(snap.Metadata.Index)
}

// ServeChannels : serves raft channels for changes
func (r *Raft) ServeChannels() {
	snap, err := r.RaftStorage.Snapshot()

	if err != nil {
		log.Fatal(err)
	}

	r.AppliedIndex = snap.Metadata.Index

	r.ConfState = snap.Metadata.ConfState

	r.SnapshotIndex = snap.Metadata.Index

	defer r.Wal.Close()

	ticker := time.NewTicker(100 * time.Millisecond)

	defer ticker.Stop()

	// send proposals over raft
	go func() {
		confChangeCount := uint64(0)

		for r.ProposeC != nil && r.ConfChangeC != nil {
			select {
			case prop, ok := <-r.ProposeC:
				if !ok {
					r.ProposeC = nil
				} else {
					// blocks until accepted by raft state machine
					if err := r.Node.Propose(context.TODO(), []byte(prop)); err != nil {
						log.Fatal(err)
					}
				}
			case cc, ok := <-r.ConfChangeC:
				if !ok {
					r.ConfChangeC = nil
				} else {
					confChangeCount++

					cc.ID = confChangeCount

					if err := r.Node.ProposeConfChange(context.TODO(), cc); err != nil {
						log.Fatal(err)
					}
				}
			}
		}

		close(r.StopC)
	}()

	// event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			r.Node.Tick()

		// store raft entries to wal, then publish over commit channel
		case rd := <-r.Node.Ready():
			err := r.Wal.Save(rd.HardState, rd.Entries)

			if err != nil {
				log.Fatal(err)
			}

			if !raft.IsEmptySnap(rd.Snapshot) {
				r.SaveSnap(rd.Snapshot)

				r.RaftStorage.ApplySnapshot(rd.Snapshot)

				r.PublishSnapshot(rd.Snapshot)
			}

			r.RaftStorage.Append(rd.Entries)

			r.Transport.Send(rd.Messages)

			if ok := r.PublishEntries(r.EntriesToApply(rd.CommittedEntries)); !ok {
				r.Stop()
				return
			}

			r.MaybeTriggerSnapshot()

			r.Node.Advance()

		case err := <-r.Transport.ErrorC:
			r.WriteError(err)
			return

		case <-r.StopC:
			r.Stop()
			return
		}
	}
}

// ServeRaft : serves the raft
func (r *Raft) ServeRaft() {
	ln, err := newStoppableListener(r.ListenAddr.String(), r.HTTPStopC)

	if err != nil {
		log.Fatalf("Failed to listen rafthttp (%v)", err)
	}

	server := &http.Server{Handler: r.Transport.Handler()}

	server.Serve(ln)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	select {
	case <-r.HTTPStopC:
	default:
		log.Fatalf("Failed to serve raft (%v)", err)
	}

	close(r.HTTPDoneC)
}

// StartRaft : starts the raft
func (r *Raft) StartRaft() {
	if !fileutil.Exist(r.SnapshotterDir) {
		if err := os.Mkdir(r.SnapshotterDir, 0750); err != nil {
			log.Fatalf("Cannot create dir for snapshot (%v)", err)
		}
	}

	r.Snapshotter = snap.New(zap.NewExample(), r.SnapshotterDir)

	r.SnapshotterReady <- r.Snapshotter

	oldwal := walExists(r.WalDir)

	r.Wal = r.ReplayWAL()

	var peers []raft.Peer

	membersLength := len(r.Members)

	peers = make([]raft.Peer, membersLength)

	for i := range peers {
		peers[i] = raft.Peer{
			ID:      r.Members[i].Id,
			Context: []byte(r.Members[i].Addr),
		}
	}

	c := &raft.Config{
		ID:              r.ID,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         r.RaftStorage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}

	if oldwal || r.Join {
		r.Node = raft.RestartNode(c)
	} else {
		r.Node = raft.StartNode(c, peers)
	}

	r.Transport = &rafthttp.Transport{
		Logger:      zap.NewExample(),
		ID:          types.ID(r.ID),
		ClusterID:   0x1000,
		Raft:        r,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(strconv.Itoa(int(r.ID))),
		ErrorC:      make(chan error),
	}

	r.Transport.Start()

	for i, m := range r.Members {
		if r.ID != m.Id {
			r.Transport.AddPeer(types.ID(m.Id), []string{r.Members[i].Addr})
		}
	}

	go r.ServeRaft()

	go r.ServeChannels()
}

// Stop : closes http, closes all channels, and stops raft.
func (r *Raft) Stop() {
	r.StopHTTP()

	close(r.CommitC)

	close(r.ErrorC)

	r.Node.Stop()
}

// StopHTTP : closes http
func (r *Raft) StopHTTP() {
	r.Transport.Stop()

	close(r.HTTPStopC)

	<-r.HTTPDoneC
}

// Process : processes member messages
func (r *Raft) Process(ctx context.Context, message raftpb.Message) error {
	return r.Node.Step(ctx, message)
}

// ReportUnreachable : determines if the member is unreachable
func (r *Raft) ReportUnreachable(id uint64) {}

// ReportSnapshot : reports snapshot status
func (r *Raft) ReportSnapshot(id uint64, status raft.SnapshotStatus) {}

// WriteError : writes error to all listeners
func (r *Raft) WriteError(err error) {
	r.StopHTTP()

	close(r.CommitC)

	r.ErrorC <- err

	close(r.ErrorC)

	r.Node.Stop()
}

// HTTPStopCHandler : handles leaving gossip protocol on raft leave
func (r *Raft) HTTPStopCHandler(configDir string, httpStopC chan struct{}) {
	<-httpStopC

	// TODO : Need to properly close listener to prevent sigfault on restart of raft

	rmWal := os.RemoveAll(r.WalDir)

	if rmWal != nil {
		log.Fatal(rmWal)
	}

	rmSnap := os.RemoveAll(r.SnapshotterDir)

	if rmSnap != nil {
		log.Fatal(rmSnap)
	}

	rmKeyPath := fmt.Sprintf("%s/key.json", configDir)

	rmKey := os.RemoveAll(rmKeyPath)

	if rmKey != nil {
		log.Fatal(rmKey)
	}

	log.Print("Successfully left gossip and raft protocols")
}

// New : returns a new key-value store and raft node
func New(config Config) (*Raft, *State, error) {
	commitC := make(chan *string)

	confChangeC := make(chan raftpb.ConfChange)

	errorC := make(chan error)

	proposeC := make(chan string)

	var state *State

	getSnapshot := func() ([]byte, error) { return state.GetSnapshot() }

	log.Print(config.NodeID)

	log.Print(config.Members)

	node := &Raft{
		CommitC:          commitC,
		ConfChangeC:      confChangeC,
		ErrorC:           errorC,
		ID:               config.NodeID,
		Join:             config.Join,
		ListenAddr:       config.ListenAddr,
		Members:          config.Members,
		GetSnapshot:      getSnapshot,
		HTTPDoneC:        make(chan struct{}),
		HTTPStopC:        config.HTTPStopC,
		SnapshotterDir:   fmt.Sprintf("%s/node-%d-snapshot", config.ConfigDir, config.NodeID),
		SnapshotterReady: make(chan *snap.Snapshotter, 1),
		SnapCount:        defaultSnapshotCount,
		StopC:            make(chan struct{}),
		ProposeC:         proposeC,
		WalDir:           fmt.Sprintf("%s/node-%d", config.ConfigDir, config.NodeID),
	}

	go node.StartRaft()

	state = NewState(commitC, errorC, proposeC, <-node.SnapshotterReady)

	return node, state, nil
}
