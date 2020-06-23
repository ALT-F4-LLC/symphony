package block

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/erkrnt/symphony/internal/pkg/gossip"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type controlServer struct {
	block *Block
}

func (s *controlServer) Init(ctx context.Context, in *api.BlockControlInitRequest) (*api.BlockControlInitResponse, error) {
	initAddr, err := net.ResolveTCPAddr("tcp", in.ServiceAddr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(initAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	r := api.NewManagerRemoteClient(conn)

	serviceAddr := fmt.Sprintf("%s", s.block.flags.listenServiceAddr.String())

	opts := &api.ManagerRemoteInitRequest{
		ServiceAddr: serviceAddr,
		ServiceType: api.ServiceType_BLOCK,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	init, err := r.Init(ctx, opts)

	if err != nil {
		return nil, err
	}

	clusterID, err := uuid.Parse(init.ClusterID)

	if err != nil {
		return nil, err
	}

	serviceID, err := uuid.Parse(init.ServiceID)

	if err != nil {
		return nil, err
	}

	key := &config.Key{
		ClusterID: &clusterID,
		Endpoints: init.Endpoints,
		ServiceID: &serviceID,
	}

	saveErr := key.Save(s.block.flags.configDir)

	if saveErr != nil {
		st := status.New(codes.Internal, saveErr.Error())

		return nil, st.Err()
	}

	gossipMember := &gossip.Member{
		ServiceAddr: opts.ServiceAddr,
		ServiceID:   serviceID.String(),
		ServiceType: opts.ServiceType.String(),
	}

	memberlist, err := gossip.NewMemberList(gossipMember, s.block.flags.listenGossipAddr.Port)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	_, joinErr := memberlist.Join([]string{init.GossipAddr})

	if joinErr != nil {
		st := status.New(codes.Internal, joinErr.Error())

		return nil, st.Err()
	}

	s.block.memberlist = memberlist

	res := &api.BlockControlInitResponse{
		ClusterID: clusterID.String(),
		ServiceID: serviceID.String(),
	}

	return res, nil
}
