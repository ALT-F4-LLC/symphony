package cli

import (
	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/service"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// ServiceNewOptions : options for initializing a service
type ServiceNewOptions struct {
	APIServerAddr *string
	SocketPath    *string
}

// ServiceNew : initializes a service for use
func ServiceNew(opts ServiceNewOptions) {
	if opts.SocketPath == nil {
		logrus.Fatal("Missing --socket-path option. Check --help for more.")
	}

	conn := service.NewClientConnSocket(*opts.SocketPath)

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), service.ContextTimeout)

	defer cancel()

	c := api.NewControlClient(conn)

	options := &api.RequestServiceNew{}

	service, err := c.ServiceNew(ctx, options)

	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info(service)
}
