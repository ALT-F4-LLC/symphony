package manager

import (
	"net"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags : command line flags for service
type Flags struct {
	ConfigDir        string
	ListenRaftAddr   *net.TCPAddr
	ListenRemoteAddr *net.TCPAddr
}

var (
	configDir        = kingpin.Flag("config-dir", "Sets configuration directory for manager.").Default(".").String()
	listenRaftPort   = kingpin.Flag("listen-raft-port", "Sets the raft listener port.").Int()
	listenRemotePort = kingpin.Flag("listen-remote-port", "Sets the remote gRPC listener port.").Int()
)

// GetFlags : gets struct of flags from command line
func GetFlags() (*Flags, error) {
	kingpin.Parse()

	configPath, err := config.GetDirPath(configDir)

	if err != nil {
		return nil, err
	}

	flags := &Flags{
		ConfigDir: *configPath,
	}

	ip := config.GetOutboundIP()

	listenRaftAddr, err := config.GetListenAddr(15760, ip, listenRaftPort)

	if err != nil {
		return nil, err
	}

	flags.ListenRaftAddr = listenRaftAddr

	listenRemoteAddr, err := config.GetListenAddr(27242, ip, listenRemotePort)

	if err != nil {
		return nil, err
	}

	flags.ListenRemoteAddr = listenRemoteAddr

	fields := logrus.Fields{
		"ConfigDir":        flags.ConfigDir,
		"ListenRaftAddr":   flags.ListenRaftAddr.String(),
		"ListenRemoteAddr": flags.ListenRemoteAddr.String(),
	}

	logrus.WithFields(fields).Info("Loaded command-line flags")

	return flags, nil
}
