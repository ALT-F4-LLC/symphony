package cluster

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

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

// Member : a member of the cluster
type Member struct {
	AppliedIndex     uint64
	CommitC          chan<- *string           // entries committed to log (k,v)
	ConfChangeC      <-chan raftpb.ConfChange // proposed cluster config changes
	ConfState        raftpb.ConfState
	ErrorC           chan<- error // errors from raft session
	ID               int
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
	Store            *KvStore
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
func (m *Member) EntriesToApply(ents []raftpb.Entry) (nents []raftpb.Entry) {
	if len(ents) == 0 {
		return ents
	}

	firstIdx := ents[0].Index

	if firstIdx > m.AppliedIndex+1 {
		log.Fatalf("First index of committed entry[%d] should <= progress.AppliedIndex[%d]+1", firstIdx, m.AppliedIndex)
	}

	if m.AppliedIndex-firstIdx+1 < uint64(len(ents)) {
		nents = ents[m.AppliedIndex-firstIdx+1:]
	}

	return nents
}

// LoadSnapshot : loads snapshot from directory
func (m *Member) LoadSnapshot() *raftpb.Snapshot {
	snapshot, err := m.Snapshotter.Load()

	if err != nil && err != snap.ErrNoSnapshot {
		log.Fatalf("Error loading snapshot (%v)", err)
	}

	return snapshot
}

// MaybeTriggerSnapshot : see if should trigger snapshot change
func (m *Member) MaybeTriggerSnapshot() {
	if m.AppliedIndex-m.SnapshotIndex <= m.SnapCount {
		return
	}

	log.Printf("Start snapshot [applied index: %d | last snapshot index: %d]", m.AppliedIndex, m.SnapshotIndex)

	data, err := m.Store.GetSnapshot()

	if err != nil {
		log.Panic(err)
	}

	snap, err := m.RaftStorage.CreateSnapshot(m.AppliedIndex, &m.ConfState, data)

	if err != nil {
		panic(err)
	}

	if err := m.SaveSnap(snap); err != nil {
		panic(err)
	}

	compactIndex := uint64(1)

	if m.AppliedIndex > snapshotCatchUpEntriesN {
		compactIndex = m.AppliedIndex - snapshotCatchUpEntriesN
	}

	if err := m.RaftStorage.Compact(compactIndex); err != nil {
		panic(err)
	}

	log.Printf("Compacted log at index %d", compactIndex)

	m.SnapshotIndex = m.AppliedIndex
}

// OpenWAL : opens write access log
func (m *Member) OpenWAL(snapshot *raftpb.Snapshot) *wal.WAL {
	if !wal.Exist(m.WalDir) {

		if err := os.Mkdir(m.WalDir, 0750); err != nil {
			log.Fatalf("Cannot create dir for wal (%v)", err)
		}

		w, err := wal.Create(zap.NewExample(), m.WalDir, nil)

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

	w, err := wal.Open(zap.NewExample(), m.WalDir, walSnap)

	if err != nil {
		log.Fatalf("Error loading wal (%v)", err)
	}

	return w
}

// ReplayWAL : replays the write access log
func (m *Member) ReplayWAL() *wal.WAL {
	log.Printf("Replaying WAL of member %d", m.ID)

	snapshot := m.LoadSnapshot()

	w := m.OpenWAL(snapshot)

	_, st, ents, err := w.ReadAll()

	if err != nil {
		log.Fatalf("Failed to read WAL (%v)", err)
	}

	m.RaftStorage = raft.NewMemoryStorage()

	if snapshot != nil {
		m.RaftStorage.ApplySnapshot(*snapshot)
	}

	m.RaftStorage.SetHardState(st)

	// append to storage so raft starts at the right place in log
	m.RaftStorage.Append(ents)

	// send nil once lastIndex is published so client knows commit channel is current
	if len(ents) > 0 {
		m.LastIndex = ents[len(ents)-1].Index
	} else {
		m.CommitC <- nil
	}

	return w
}

// PublishEntries : writes committed log entries to commit channel and returns
// whether all entries could be published.
func (m *Member) PublishEntries(ents []raftpb.Entry) bool {
	for i := range ents {
		switch ents[i].Type {
		case raftpb.EntryNormal:
			if len(ents[i].Data) == 0 {
				// ignore empty messages
				break
			}

			s := string(ents[i].Data)

			select {
			case m.CommitC <- &s:
			case <-m.StopC:
				return false
			}
		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange

			cc.Unmarshal(ents[i].Data)

			m.ConfState = *m.Node.ApplyConfChange(cc)

			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if len(cc.Context) > 0 {
					m.Transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(m.ID) {
					log.Println("I've been removed from the cluster! Shutting down.")
					return false
				}

				m.Transport.RemovePeer(types.ID(cc.NodeID))
			}
		}

		// after commit, update appliedIndex
		m.AppliedIndex = ents[i].Index

		// special nil commit to signal replay has finished
		if ents[i].Index == m.LastIndex {
			select {
			case m.CommitC <- nil:
			case <-m.StopC:
				return false
			}
		}
	}

	return true
}

// PublishSnapshot : handles publishing the snapshot to the KvStore
func (m *Member) PublishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	log.Printf("Publishing snapshot at index %d", m.SnapshotIndex)

	defer log.Printf("Finished publishing snapshot at index %d", m.SnapshotIndex)

	if snapshotToSave.Metadata.Index <= m.AppliedIndex {
		log.Fatalf("Snapshot index [%d] should > progress.AppliedIndex [%d]", snapshotToSave.Metadata.Index, m.AppliedIndex)
	}

	m.CommitC <- nil

	m.AppliedIndex = snapshotToSave.Metadata.Index

	m.ConfState = snapshotToSave.Metadata.ConfState

	m.SnapshotIndex = snapshotToSave.Metadata.Index
}

// SaveSnap : must save the snapshot index to the WAL before saving
// the snapshot to maintain the invariant that we only open the WAL
// at previously-saved snapshot indexes.
func (m *Member) SaveSnap(snap raftpb.Snapshot) error {
	walSnap := walpb.Snapshot{
		Index: snap.Metadata.Index,
		Term:  snap.Metadata.Term,
	}

	if err := m.Wal.SaveSnapshot(walSnap); err != nil {
		return err
	}

	if err := m.Snapshotter.SaveSnap(snap); err != nil {
		return err
	}

	return m.Wal.ReleaseLockTo(snap.Metadata.Index)
}

// ServeChannels : serves raft channels for changes
func (m *Member) ServeChannels() {
	snap, err := m.RaftStorage.Snapshot()

	if err != nil {
		log.Fatal(err)
	}

	m.AppliedIndex = snap.Metadata.Index

	m.ConfState = snap.Metadata.ConfState

	m.SnapshotIndex = snap.Metadata.Index

	defer m.Wal.Close()

	ticker := time.NewTicker(100 * time.Millisecond)

	defer ticker.Stop()

	// send proposals over raft
	go func() {
		confChangeCount := uint64(0)

		for m.ProposeC != nil && m.ConfChangeC != nil {
			select {
			case prop, ok := <-m.ProposeC:
				if !ok {
					m.ProposeC = nil
				} else {
					// blocks until accepted by raft state machine
					m.Node.Propose(context.TODO(), []byte(prop))
				}
			case cc, ok := <-m.ConfChangeC:
				if !ok {
					m.ConfChangeC = nil
				} else {
					confChangeCount++

					cc.ID = confChangeCount

					m.Node.ProposeConfChange(context.TODO(), cc)
				}
			}
		}

		close(m.StopC)
	}()

	// event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			m.Node.Tick()

		// store raft entries to wal, then publish over commit channel
		case rd := <-m.Node.Ready():
			m.Wal.Save(rd.HardState, rd.Entries)

			if !raft.IsEmptySnap(rd.Snapshot) {
				m.SaveSnap(rd.Snapshot)

				m.RaftStorage.ApplySnapshot(rd.Snapshot)

				m.PublishSnapshot(rd.Snapshot)
			}

			m.RaftStorage.Append(rd.Entries)

			m.Transport.Send(rd.Messages)

			if ok := m.PublishEntries(m.EntriesToApply(rd.CommittedEntries)); !ok {
				m.Stop()
				return
			}

			m.MaybeTriggerSnapshot()

			m.Node.Advance()

		case err := <-m.Transport.ErrorC:
			m.WriteError(err)
			return

		case <-m.StopC:
			m.Stop()
			return
		}
	}
}

// ServeRaft : serves the raft
func (m *Member) ServeRaft() {
	url, err := url.Parse(m.Peers[m.ID-1])

	if err != nil {
		log.Fatal(err)
	}

	ln, err := newStoppableListener(url.Host, m.HTTPStopC)

	if err != nil {
		log.Fatalf("Failed to listen rafthttp (%v)", err)
	}

	err = (&http.Server{Handler: m.Transport.Handler()}).Serve(ln)

	if err != nil {
		log.Fatal(err)
	}

	select {
	case <-m.HTTPStopC:
	default:
		log.Fatalf("Failed to serve raft (%v)", err)
	}

	close(m.HTTPDoneC)
}

// StartRaft : starts the raft
func (m *Member) StartRaft() {
	if !fileutil.Exist(m.SnapshotterDir) {

		if err := os.Mkdir(m.SnapshotterDir, 0750); err != nil {
			log.Fatalf("Cannot create dir for snapshot (%v)", err)
		}

	}

	m.Snapshotter = snap.New(zap.NewExample(), m.SnapshotterDir)

	m.SnapshotterReady <- m.Snapshotter

	oldwal := walExists(m.WalDir)

	m.Wal = m.ReplayWAL()

	rpeers := make([]raft.Peer, len(m.Peers))

	for i := range rpeers {
		rpeers[i] = raft.Peer{ID: uint64(i + 1)}
	}

	c := &raft.Config{
		ID:              uint64(m.ID),
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         m.RaftStorage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}

	if oldwal {
		m.Node = raft.RestartNode(c)
	} else {
		m.Node = raft.StartNode(c, rpeers)
	}

	m.Transport = &rafthttp.Transport{
		Logger:      zap.NewExample(),
		ID:          types.ID(m.ID),
		ClusterID:   0x1000,
		Raft:        m,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(strconv.Itoa(m.ID)),
		ErrorC:      make(chan error),
	}

	m.Transport.Start()

	for i := range m.Peers {

		if i+1 != m.ID {
			m.Transport.AddPeer(types.ID(i+1), []string{m.Peers[i]})
		}

	}

	go m.ServeRaft()

	go m.ServeChannels()
}

// Stop : closes http, closes all channels, and stops raft.
func (m *Member) Stop() {
	m.StopHTTP()

	close(m.CommitC)

	close(m.ErrorC)

	m.Node.Stop()
}

// StopHTTP : closes http
func (m *Member) StopHTTP() {
	m.Transport.Stop()

	close(m.HTTPStopC)

	<-m.HTTPDoneC
}

// NewRaftMember : creates a new raft member
func NewRaftMember(confChangeC <-chan raftpb.ConfChange, configDir string, kvs *KvStore, proposeC <-chan string) (<-chan *string, <-chan error, *Member) {
	commitC := make(chan *string)

	errorC := make(chan error)

	id := 1

	peers := []string{"http://127.0.0.1:12379"}

	member := &Member{
		CommitC:          commitC,
		ConfChangeC:      confChangeC,
		ID:               id,
		Peers:            peers,
		HTTPDoneC:        make(chan struct{}),
		HTTPStopC:        make(chan struct{}),
		SnapshotterDir:   fmt.Sprintf("%s/node-%d-snapshot", configDir, id),
		SnapshotterReady: make(chan *snap.Snapshotter, 1),
		SnapCount:        defaultSnapshotCount,
		Store:            kvs,
		StopC:            make(chan struct{}),
		ProposeC:         proposeC,
		WalDir:           fmt.Sprintf("%s/node-%d", configDir, id),
	}

	go member.StartRaft()

	return commitC, errorC, member
}

// Process : processes member messages
func (m *Member) Process(ctx context.Context, message raftpb.Message) error {
	return m.Node.Step(ctx, message)
}

// IsIDRemoved : determines if the members ID has been removed
func (m *Member) IsIDRemoved(id uint64) bool { return false }

// ReportUnreachable : determines if the member is unreachable
func (m *Member) ReportUnreachable(id uint64) {}

// ReportSnapshot : reports snapshot status
func (m *Member) ReportSnapshot(id uint64, status raft.SnapshotStatus) {}

// WriteError : writes error to all listeners
func (m *Member) WriteError(err error) {
	m.StopHTTP()

	close(m.CommitC)

	m.ErrorC <- err

	close(m.ErrorC)

	m.Node.Stop()
}
