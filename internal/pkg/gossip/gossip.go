package gossip

import (
	"encoding/json"
	"sync"

	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
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
	ServiceAddr string
	ServiceID   string
	ServiceType string
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
func NewMemberList(member *Member, port int) (*memberlist.Memberlist, error) {
	data, err := json.Marshal(member)

	if err != nil {
		return nil, err
	}

	fields := logrus.Fields{"metadata": string(data)}

	logrus.WithFields(fields).Debug("Gossip memberlist metadata created.")

	config := memberlist.DefaultLocalConfig()

	config.AdvertisePort = port

	config.Delegate = &Delegate{Meta: data}

	config.BindPort = port

	config.Name = member.ServiceID

	list, err := memberlist.Create(config)

	if err != nil {
		return nil, err
	}

	for _, member := range list.Members() {
		fields := logrus.Fields{"addr": member.Addr, "name": member.Name}

		logrus.WithFields(fields).Debug("Gossip memberlist updated.")
	}

	return list, nil
}
