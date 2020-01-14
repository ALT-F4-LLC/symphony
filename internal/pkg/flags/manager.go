package flags

import (
	"net"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"gopkg.in/alecthomas/kingpin.v2"
)

// ManagerFlags : command line flags for service
type ManagerFlags struct {
	ConfigDir       string
	JoinAddr        *net.TCPAddr
	ListenRemoteAPI *net.TCPAddr
}

// GetManagerFlags : gets struct of flags from command line
func GetManagerFlags(logger *zap.Logger) (*ManagerFlags, error) {
	var (
		configDir       = kingpin.Flag("config-dir", "Sets configuration directory for manager.").Default(".").String()
		joinAddr        = kingpin.Flag("join-addr", "Sets existing Raft remote manager address to join.").String()
		listenRemoteAPI = kingpin.Flag("listen-remote-api", "Sets the remote API address for nodes.").Default("127.0.0.1:32389").String()
	)

	kingpin.Parse()

	configPath, err := filepath.Abs(*configDir)

	if err != nil {
		logger.Panic(err.Error())
	}

	_, statErr := os.Stat(configPath)

	if statErr != nil {
		if err := os.MkdirAll(configPath, 0750); err != nil {
			logger.Panic(err.Error())
		}
	}

	flags := &ManagerFlags{
		ConfigDir: configPath,
	}

	if *joinAddr != "" {
		join, err := net.ResolveTCPAddr("tcp", *joinAddr)

		if err != nil {
			return nil, err
		}

		flags.JoinAddr = join
	}

	if *listenRemoteAPI != "" {
		remoteAPI, err := net.ResolveTCPAddr("tcp", *listenRemoteAPI)

		if err != nil {
			return nil, err
		}

		flags.ListenRemoteAPI = remoteAPI
	}

	logger.Info(
		"FLAGS: Loaded command-line options",
		zap.String("ConfigDir", flags.ConfigDir),
		zap.String("JoinAddr", flags.JoinAddr.String()),
		zap.String("ListenRemoteAPI", flags.ListenRemoteAPI.String()),
	)

	return flags, nil
}
