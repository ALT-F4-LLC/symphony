package apiserver

import (
	"net"
	"strings"

	"github.com/erkrnt/symphony/internal/service"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags : command line flags
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

	configDirPath, err := service.GetDirPath(configDir)

	if err != nil {
		return nil, err
	}

	listenAddr, err := service.GetListenAddr(15760, listenPort)

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
		"ConfigDir":     flags.ConfigDir,
		"EtcdEndpoints": flags.EtcdEndpoints,
		"ListenAddr":    flags.ListenAddr.String(),
		"Verbose":       flags.Verbose,
	}

	logrus.WithFields(fields).Info("Service command-line flags loaded.")

	return flags, nil
}
