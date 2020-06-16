package manager

import (
	"net"
	"strings"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type flags struct {
	configPath        string
	etcdEndpoints     []string
	listenGossipAddr  *net.TCPAddr
	listenServiceAddr *net.TCPAddr
	verbose           bool
}

var (
	configDir         = kingpin.Flag("config-dir", "Sets configuration directory for manager.").Default(".").String()
	etcdEndpoints     = kingpin.Flag("etcd-endpoints", "Sets the etcd endpoints list.").Required().String()
	listenGossipPort  = kingpin.Flag("listen-gossip-port", "Sets the remote gossip listener port.").Int()
	listenServicePort = kingpin.Flag("listen-service-port", "Sets the remote service port.").Int()
	verbose           = kingpin.Flag("verbose", "Sets the lowest level of service output.").Bool()
)

func getFlags() (*flags, error) {
	kingpin.Parse()

	configPath, err := config.GetDirPath(configDir)

	if err != nil {
		return nil, err
	}

	ip, err := config.GetOutboundIP()

	if err != nil {
		return nil, err
	}

	listenGossipAddr, err := config.GetListenAddr(37065, ip, listenGossipPort)

	if err != nil {
		return nil, err
	}

	listenServiceAddr, err := config.GetListenAddr(15760, ip, listenServicePort)

	if err != nil {
		return nil, err
	}

	flags := &flags{
		configPath:        *configPath,
		etcdEndpoints:     strings.Split(*etcdEndpoints, ","),
		listenGossipAddr:  listenGossipAddr,
		listenServiceAddr: listenServiceAddr,
		verbose:           *verbose,
	}

	fields := logrus.Fields{
		"configPath":        flags.configPath,
		"etcdEndpoints":     flags.etcdEndpoints,
		"listenGossipAddr":  flags.listenGossipAddr.String(),
		"listenServiceAddr": flags.listenServiceAddr.String(),
		"verbose":           flags.verbose,
	}

	logrus.WithFields(fields).Info("Loaded command-line flags")

	return flags, nil
}
