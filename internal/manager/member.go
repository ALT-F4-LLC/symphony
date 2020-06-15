package manager

import (
	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
)

// GetGossipMemberByAddr : gets the index of a member in a set of peers
func GetGossipMemberByAddr(members []*api.GossipMember, addr string) (*int, *api.GossipMember) {
	for i, a := range members {
		if a.Addr == addr {
			return &i, a
		}
	}
	return nil, nil
}

// GetRaftMemberByAddr : gets the index of a member in a set of peers
func GetRaftMemberByAddr(members []*api.RaftMember, addr string) (*int, *api.RaftMember) {
	for i, a := range members {
		if a.Addr == addr {
			return &i, a
		}
	}
	return nil, nil
}

// GetRaftMemberByID : gets the index of a member in a set of peers
func GetRaftMemberByID(members []*api.RaftMember, raftID uint64) (*int, *api.RaftMember) {
	for i, a := range members {
		if a.Id == raftID {
			return &i, a
		}
	}
	return nil, nil
}

// GetServiceByAddr : gets the index of a member in a set of peers
func GetServiceByAddr(members []*api.ServiceMember, addr string) (*int, *api.ServiceMember) {
	for i, a := range members {
		if a.Addr == addr {
			return &i, a
		}
	}
	return nil, nil
}

// GetServiceByID : gets the index of a member in a set of peers
func GetServiceByID(members []*api.ServiceMember, id uuid.UUID) (*int, *api.ServiceMember) {
	for i, a := range members {
		if a.Id == id.String() {
			return &i, a
		}
	}
	return nil, nil
}
