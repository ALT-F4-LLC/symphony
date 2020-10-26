package state

import (
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

// NewClient : creates new etcd client
func NewClient(config clientv3.Config) (*clientv3.Client, error) {
	client, err := clientv3.New(config)

	if err != nil {
		return nil, err
	}

	fields := logrus.Fields{
		"DialTimeout": config.DialTimeout,
		"Endpoints":   config.Endpoints,
	}

	logrus.WithFields(fields).Debug("New etcd client created.")

	return client, nil
}
