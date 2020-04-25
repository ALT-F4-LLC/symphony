package manager

import (
	"github.com/erkrnt/symphony/api"
)

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
