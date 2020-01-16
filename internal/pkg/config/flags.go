package config

import (
	"net"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags : command line flags for service
type Flags struct {
	ConfigDir       string
	JoinAddr        *net.TCPAddr
	ListenRemoteAPI *net.TCPAddr
}

var (
	configDir       = kingpin.Flag("config-dir", "Sets configuration directory for manager.").Default(".").String()
	joinAddr        = kingpin.Flag("join-addr", "Sets existing Raft remote manager address to join.").String()
	listenRemoteAPI = kingpin.Flag("listen-remote-api", "Sets the remote API address for nodes.").Default("127.0.0.1:32389").String()
)

// GetFlags : gets struct of flags from command line
func GetFlags() (*Flags, error) {
	kingpin.Parse()

	configPath, err := filepath.Abs(*configDir)

	if err != nil {
		logrus.Fatal(err)
	}

	_, statErr := os.Stat(configPath)

	if statErr != nil {
		if err := os.MkdirAll(configPath, 0750); err != nil {
			logrus.Fatal(err)
		}
	}

	nodeFlags := &Flags{
		ConfigDir: configPath,
	}

	if *joinAddr != "" {
		join, err := net.ResolveTCPAddr("tcp", *joinAddr)

		if err != nil {
			return nil, err
		}

		nodeFlags.JoinAddr = join
	}

	if *listenRemoteAPI != "" {
		remoteAPI, err := net.ResolveTCPAddr("tcp", *listenRemoteAPI)

		if err != nil {
			return nil, err
		}

		nodeFlags.ListenRemoteAPI = remoteAPI
	}

	fields := logrus.Fields{
		"ConfigDir":       nodeFlags.ConfigDir,
		"JoinAddr":        nodeFlags.JoinAddr.String(),
		"ListenRemoteAPI": nodeFlags.ListenRemoteAPI.String(),
	}

	logrus.WithFields(fields).Debug("Loaded command-line flags")

	return nodeFlags, nil
}
