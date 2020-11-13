package apiserver

import (
	"net"
	"strings"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Flags struct {
	ConfigDir     string
	EtcdEndpoints []string
	ListenAddr    *net.TCPAddr
	Verbose       bool
}

var (
	configDir     = kingpin.Flag("config-dir", "Sets configuration directory for manager.").Default(".").String()
	etcdEndpoints = kingpin.Flag("etcd-endpoints", "Sets the etcd endpoints list.").Required().String()
	listenPort    = kingpin.Flag("listen-port", "Sets the remote service port.").Int()
	verbose       = kingpin.Flag("verbose", "Sets the lowest level of service output.").Bool()
)

func getFlags() (*Flags, error) {
	kingpin.Parse()

	configDirPath, err := config.GetDirPath(configDir)

	if err != nil {
		return nil, err
	}

	listenAddr, err := config.GetListenAddr(15760, listenPort)

	if err != nil {
		return nil, err
	}

	flags := &Flags{
		ConfigDir:     *configDirPath,
		EtcdEndpoints: strings.Split(*etcdEndpoints, ","),
		ListenAddr:    listenAddr,
		Verbose:       *verbose,
	}

	fields := logrus.Fields{
		"configDir":     flags.ConfigDir,
		"etcdEndpoints": flags.EtcdEndpoints,
		"listenAddr":    flags.ListenAddr.String(),
		"verbose":       flags.Verbose,
	}

	logrus.WithFields(fields).Info("Service command-line flags loaded.")

	return flags, nil
}
