package discovery

import (
	"context"
	"net/url"
	"time"

	"github.com/erkrnt/symphony/api"
	"google.golang.org/grpc"
)

// JoinRaftMembership : joins the node to the cluster
func JoinRaftMembership(addr string, joinAddr *string) (*api.JoinResponse, error) {
	joinURL, err := url.ParseRequestURI(*joinAddr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(joinURL.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	c := api.NewRaftMembershipClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	fields := &api.JoinRequest{Addr: addr}

	res, err := c.Join(ctx, fields)

	if err != nil {
		return nil, err
	}

	return res, nil
}
