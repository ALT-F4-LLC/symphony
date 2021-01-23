package service

import (
	"github.com/sirupsen/logrus"
)

// Service : cloud service definition
type Service struct {
	ErrorC chan error

	Flags *Flags
	Key   *Key
}

// New : create a new cloud service
func New() (*Service, error) {
	flags, err := GetFlags()

	if err != nil {
		return nil, err
	}

	if flags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	key, err := GetKey(flags.ConfigDir)

	if err != nil {
		return nil, err
	}

	service := &Service{
		ErrorC: make(chan error),

		Flags: flags,
		Key:   key,
	}

	return service, nil
}
