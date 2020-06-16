package manager

import (
	"net"
	"strings"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags : command line flags for service
type Flags struct {
	ConfigPath    string
	EtcdEndpoints []string
	ListenAddr    *net.TCPAddr
}

var (
	configDir     = kingpin.Flag("config-dir", "Sets configuration directory for manager.").Default(".").String()
	etcdEndpoints = kingpin.Flag("etcd-endpoints", "Sets the etcd endpoints list.").Required().String()
	listenPort    = kingpin.Flag("listen-port", "Sets the remote gRPC listener port.").Int()
)

// GetFlags : gets struct of flags from command line
func GetFlags() (*Flags, error) {
	kingpin.Parse()

	configPath, err := config.GetDirPath(configDir)

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

	flags := &Flags{
		ConfigPath:    *configPath,
		EtcdEndpoints: strings.Split(*etcdEndpoints, ","),
		ListenAddr:    listenAddr,
	}

	fields := logrus.Fields{
		"ConfigPath":    flags.ConfigPath,
		"EtcdEndpoints": flags.EtcdEndpoints,
		"ListenAddr":    flags.ListenAddr.String(),
	}

	logrus.WithFields(fields).Info("Loaded command-line flags")

	return flags, nil
}
