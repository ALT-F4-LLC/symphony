package cli

import (
	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/service"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// ClusterList : returns list of clusters
func ClusterList(apiserverAddr *string) {
	if apiserverAddr == nil {
		logrus.Fatal("Missing --apiserver-addr option. Check --help for more.")
	}

	conn := service.NewClientConnTCP(*apiserverAddr)

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), service.ContextTimeout)

	defer cancel()

	c := api.NewAPIServerClient(conn)

	options := &api.RequestClusters{}

	clusters, err := c.GetClusters(ctx, options)

	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info(clusters)
}

// ClusterNew : create a new cluster
func ClusterNew(apiserverAddr *string) {
	if apiserverAddr == nil {
		logrus.Fatal("Missing --apiserver-addr option. Check --help for more.")
	}

	conn := service.NewClientConnTCP(*apiserverAddr)

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), service.ContextTimeout)

	defer cancel()

	c := api.NewAPIServerClient(conn)

	options := &api.RequestNewCluster{}

	cluster, err := c.NewCluster(ctx, options)

	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info(cluster)
}
