package cluster

import (
	"errors"

	"github.com/erkrnt/symphony/api"
)

// GetMemberByAddr : gets the index of a member in a set of peers
func GetMemberByAddr(addr string, members []*api.Member) (*api.Member, *int, error) {
	for i, a := range members {
		if a.Addr == addr {
			return a, &i, nil
		}
	}
	return nil, nil, errors.New("invalid_member")
}

// GetMemberByID : gets the index of a member in a set of peers
func GetMemberByID(id uint64, members []*api.Member) (*api.Member, *int, error) {
	for i, a := range members {
		if a.ID == id {
			return a, &i, nil
		}
	}
	return nil, nil, errors.New("invalid_member")
}
