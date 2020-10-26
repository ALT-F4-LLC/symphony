package manager

import (
	"net"
	"strings"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type flags struct {
	configDir     string
	etcdEndpoints []string
	listenAddr    *net.TCPAddr
	verbose       bool
}

var (
	configDir     = kingpin.Flag("config-dir", "Sets configuration directory for manager.").Default(".").String()
	etcdEndpoints = kingpin.Flag("etcd-endpoints", "Sets the etcd endpoints list.").Required().String()
	listenPort    = kingpin.Flag("listen-port", "Sets the remote service port.").Int()
	verbose       = kingpin.Flag("verbose", "Sets the lowest level of service output.").Bool()
)

func getFlags() (*flags, error) {
	kingpin.Parse()

	configDirPath, err := config.GetDirPath(configDir)

	if err != nil {
		return nil, err
	}

	ip, err := config.GetOutboundIP()

	if err != nil {
		return nil, err
	}

	listenAddr, err := config.GetListenAddr(15760, ip, listenPort)

	if err != nil {
		return nil, err
	}

	flags := &flags{
		configDir:     *configDirPath,
		etcdEndpoints: strings.Split(*etcdEndpoints, ","),
		listenAddr:    listenAddr,
		verbose:       *verbose,
	}

	fields := logrus.Fields{
		"configDir":     flags.configDir,
		"etcdEndpoints": flags.etcdEndpoints,
		"listenAddr":    flags.listenAddr.String(),
		"verbose":       flags.verbose,
	}

	logrus.WithFields(fields).Info("Service command-line flags loaded.")

	return flags, nil
}
