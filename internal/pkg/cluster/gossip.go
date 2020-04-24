package cluster

import "sync"

// GossipDelegate : generic gossip protocol delegate
type GossipDelegate struct {
	Meta        []byte
	broadcasts  [][]byte
	msgs        [][]byte
	mu          sync.Mutex
	remoteState []byte
	state       []byte
}

// GetBroadcasts : gets broadcasts from gossip protocol
func (d *GossipDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	d.mu.Lock()

	defer d.mu.Unlock()

	b := d.broadcasts

	d.broadcasts = nil

	return b
}

// LocalState : returns local state from gossip protocol
func (d *GossipDelegate) LocalState(join bool) []byte {
	d.mu.Lock()

	defer d.mu.Unlock()

	return d.state
}

// NodeMeta : returs node meta from gossip protocol
func (d *GossipDelegate) NodeMeta(limit int) []byte {
	d.mu.Lock()

	defer d.mu.Unlock()

	return d.Meta
}

// NotifyMsg : process notify message from gossip protocol
func (d *GossipDelegate) NotifyMsg(msg []byte) {
	d.mu.Lock()

	defer d.mu.Unlock()

	cp := make([]byte, len(msg))

	copy(cp, msg)

	d.msgs = append(d.msgs, cp)
}

// MergeRemoteState : merges remote state into local state
func (d *GossipDelegate) MergeRemoteState(buf []byte, join bool) {
	d.mu.Lock()

	defer d.mu.Unlock()

	d.remoteState = buf
}
