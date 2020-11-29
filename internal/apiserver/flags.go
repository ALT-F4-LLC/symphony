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
	PeerAddr      *net.TCPAddr
	Verbose       bool
}

var (
	configDir     = kingpin.Flag("config-dir", "Sets configuration directory for manager.").Default(".").String()
	etcdEndpoints = kingpin.Flag("etcd-endpoints", "Sets the etcd endpoints list.").Required().String()
	listenPort    = kingpin.Flag("listen-port", "Sets the remote service port.").Int()
	peerAddr      = kingpin.Flag("peer-addr", "Sets the peer address for the service.").Required().String()
	verbose       = kingpin.Flag("verbose", "Sets the lowest level of service output.").Bool()
)

func getFlags() (*Flags, error) {
	kingpin.Parse()

	configDirPath, err := service.GetDirPath(configDir)

	if err != nil {
		return nil, err
	}

	ip, err := service.GetOutboundIP()

	if err != nil {
		return nil, err
	}

	listenAddr, err := service.GetListenAddr(15760, ip, listenPort)

	if err != nil {
		return nil, err
	}

	peerAddrTCP, err := net.ResolveTCPAddr("tcp", *peerAddr)

	if err != nil {
		return nil, err
	}

	flags := &Flags{
		ConfigDir:     *configDirPath,
		EtcdEndpoints: strings.Split(*etcdEndpoints, ","),
		ListenAddr:    listenAddr,
		PeerAddr:      peerAddrTCP,
		Verbose:       *verbose,
	}

	fields := logrus.Fields{
		"configDir":     flags.ConfigDir,
		"etcdEndpoints": flags.EtcdEndpoints,
		"listenAddr":    flags.ListenAddr.String(),
		"peerAddr":      peerAddrTCP.String(),
		"verbose":       flags.Verbose,
	}

	logrus.WithFields(fields).Info("Service command-line flags loaded.")

	return flags, nil
}
