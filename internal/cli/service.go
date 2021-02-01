package cli

import (
	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// ServiceNewOptions : options for initializing a service
type ServiceNewOptions struct {
	ManagerAddr *string
	SocketPath  *string
}

// ServiceNew : initializes a service for use
func ServiceNew(opts ServiceNewOptions) {
	if opts.SocketPath == nil {
		logrus.Fatal("Missing --socket-path option. Check --help for more.")
	}

	conn := utils.NewClientConnSocket(*opts.SocketPath)

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewControlClient(conn)

	options := &api.RequestServiceNew{}

	service, err := c.ServiceNew(ctx, options)

	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info(service)
}
