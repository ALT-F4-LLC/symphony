package cli

import (
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	GrpcContextTimeout = 5 * time.Second
)

func ClusterList(serviceAddr *string) {
	if serviceAddr == nil {
		logrus.Fatal("Missing APIServer address option. Check --help for more.")
	}

	conn := NewClient(*serviceAddr)

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), GrpcContextTimeout)

	defer cancel()

	client := api.NewAPIServerClient(conn)

	options := &api.RequestClusters{}

	clusters, err := client.GetClusters(ctx, options)

	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info(clusters)
}

func ClusterInit(serviceAddr *string) {
	if serviceAddr == nil {
		logrus.Fatal("Missing APIServer address option. Check --help for more.")
	}

	conn := NewClient(*serviceAddr)

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), GrpcContextTimeout)

	defer cancel()

	client := api.NewAPIServerClient(conn)

	options := &api.RequestNewCluster{}

	cluster, err := client.NewCluster(ctx, options)

	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info(cluster)
}
