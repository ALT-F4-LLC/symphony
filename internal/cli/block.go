package cli

import (
	"context"
	"log"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/sirupsen/logrus"
)

// BlockNewPhysicalVolumeOptions : options for initializing a service
type BlockNewPhysicalVolumeOptions struct {
	DeviceName  *string
	ManagerAddr *string
	ServiceID   *string
}

// BlockNewPhysicalVolume : handles creation of new physical volume
func BlockNewPhysicalVolume(opts BlockNewPhysicalVolumeOptions) {
	if opts.ManagerAddr == nil {
		logrus.Fatal("Missing --manager-addr option. Check --help for more.")
	}

	if *opts.DeviceName == "" || *opts.ServiceID == "" {
		logrus.Fatal("Missing --device-name or --service-id option. Check --help for more.")
	}

	conn, err := utils.NewClientConnTcp(*opts.ManagerAddr)

	if err != nil {
		logrus.Fatal(err)
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewManagerClient(conn)

	physicalVolumeOpts := &api.RequestNewPhysicalVolume{
		DeviceName: *opts.DeviceName,
		ServiceID:  *opts.ServiceID,
	}

	physicalVolume, err := c.NewPhysicalVolume(ctx, physicalVolumeOpts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*physicalVolume)
}
