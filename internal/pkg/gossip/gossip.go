package gossip

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/hashicorp/memberlist"
)

// Delegate : generic gossip protocol delegate
type Delegate struct {
	Meta        []byte
	broadcasts  [][]byte
	msgs        [][]byte
	mu          sync.Mutex
	remoteState []byte
	state       []byte
}

// Member : defines gossip member metadata
type Member struct {
	id string
}

// GetBroadcasts : gets broadcasts from gossip protocol
func (d *Delegate) GetBroadcasts(overhead, limit int) [][]byte {
	d.mu.Lock()

	defer d.mu.Unlock()

	b := d.broadcasts

	d.broadcasts = nil

	return b
}

// LocalState : returns local state from gossip protocol
func (d *Delegate) LocalState(join bool) []byte {
	d.mu.Lock()

	defer d.mu.Unlock()

	return d.state
}

// NodeMeta : returs node meta from gossip protocol
func (d *Delegate) NodeMeta(limit int) []byte {
	d.mu.Lock()

	defer d.mu.Unlock()

	return d.Meta
}

// NotifyMsg : process notify message from gossip protocol
func (d *Delegate) NotifyMsg(msg []byte) {
	d.mu.Lock()

	defer d.mu.Unlock()

	cp := make([]byte, len(msg))

	copy(cp, msg)

	d.msgs = append(d.msgs, cp)
}

// MergeRemoteState : merges remote state into local state
func (d *Delegate) MergeRemoteState(buf []byte, join bool) {
	d.mu.Lock()

	defer d.mu.Unlock()

	d.remoteState = buf
}

// NewMemberList : creates memberlist for gossip protocol
func NewMemberList(id uuid.UUID, port int) (*memberlist.Memberlist, error) {
	data := Member{
		id: id.String(),
	}

	meta, err := json.Marshal(data)

	if err != nil {
		return nil, err
	}

	config := memberlist.DefaultLocalConfig()

	config.AdvertisePort = port

	config.Delegate = &Delegate{Meta: meta}

	config.BindPort = port

	config.Name = id.String()

	list, err := memberlist.Create(config)

	if err != nil {
		return nil, err
	}

	for _, member := range list.Members() {
		fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
	}

	return list, nil
}
