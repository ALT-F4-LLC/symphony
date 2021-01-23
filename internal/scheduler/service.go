package scheduler

import (
	"context"
	"io"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/sirupsen/logrus"
)

// Scheduler : scheduler service
type Scheduler struct {
	ErrorC chan error
	Flags  *Flags
}

// New : creates new scheduler service
func New() (*Scheduler, error) {
	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	if flags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	scheduler := &Scheduler{
		Flags: flags,
	}

	return scheduler, nil
}

// Start : handles start of scheduler service
func (s *Scheduler) Start() {
	s.listenResources()
}

func (s *Scheduler) listenLogicalVolumes(client api.APIServerClient) (func() error, error) {
	in := &api.RequestState{}

	lvs, err := client.StateLogicalVolumes(context.Background(), in)

	if err != nil {
		return nil, err
	}

	listener := func() error {
		for {
			lv, err := lvs.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			if lv != nil {
				logrus.Print(lv)
			}
		}

		return nil
	}

	return listener, nil
}

func (s *Scheduler) listenPhysicalVolumes(client api.APIServerClient) (func() error, error) {
	in := &api.RequestState{}

	pvs, err := client.StatePhysicalVolumes(context.Background(), in)

	if err != nil {
		return nil, err
	}

	listener := func() error {
		for {
			pv, err := pvs.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			if pv != nil {
				logrus.Print(pv)
			}
		}

		return nil
	}

	return listener, nil
}

func (s *Scheduler) listenResources() {
	apiserverAddr := s.Flags.APIServerAddr.String()

	logicalVolumes := utils.APIServerConn{
		Address: apiserverAddr,
		ErrorC:  s.ErrorC,
		Handler: s.listenLogicalVolumes,
	}

	physicalVolumes := utils.APIServerConn{
		Address: apiserverAddr,
		ErrorC:  s.ErrorC,
		Handler: s.listenPhysicalVolumes,
	}

	volumeGroups := utils.APIServerConn{
		Address: apiserverAddr,
		ErrorC:  s.ErrorC,
		Handler: s.listenVolumeGroups,
	}

	go logicalVolumes.Listen()

	go physicalVolumes.Listen()

	go volumeGroups.Listen()
}

func (s *Scheduler) listenVolumeGroups(client api.APIServerClient) (func() error, error) {
	in := &api.RequestState{}

	vgs, err := client.StateVolumeGroups(context.Background(), in)

	if err != nil {
		return nil, err
	}

	listener := func() error {
		for {
			vg, err := vgs.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			if vg != nil {
				logrus.Print(vg)
			}
		}

		return nil
	}

	return listener, nil
}
