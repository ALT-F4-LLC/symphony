package config

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags : command line flags for service
type Flags struct {
	ConfigDir        string
	ListenGossipAddr *net.TCPAddr
	ListenRaftAddr   *net.TCPAddr
	ListenRemoteAddr *net.TCPAddr
}

var (
	configDir        = kingpin.Flag("config-dir", "Sets configuration directory for manager.").Default(".").String()
	listenGossipPort = kingpin.Flag("listen-gossip-port", "Sets the gossip listener port.").Int()
	listenRaftPort   = kingpin.Flag("listen-raft-port", "Sets the raft listener port.").Int()
	listenRemotePort = kingpin.Flag("listen-remote-port", "Sets the remote gRPC listener port.").Int()
)

// GetListenAddr : returns the TCP listen addr
func GetListenAddr(defaultPort int, ip net.IP, overridePort *int) (*net.TCPAddr, error) {
	if *overridePort != 0 {
		defaultPort = *overridePort
	}

	listenAddr := fmt.Sprintf("%s:%d", ip.String(), defaultPort)

	listenTCPAddr, err := net.ResolveTCPAddr("tcp", listenAddr)

	if err != nil {
		return nil, err
	}

	return listenTCPAddr, nil
}

// GetOutboundIP : get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "1.1.1.1:80")

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// GetPeersFromString : parses a peers list to generate an array
func GetPeersFromString(list *string) ([]string, error) {
	peers := strings.Split(*list, ",")

	if len(peers) < 1 {
		return nil, errors.New("--peers list requires minimum 2 other peers")
	}

	return nil, nil
}

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

	ip := GetOutboundIP()

	listenGossipAddr, err := GetListenAddr(7946, ip, listenGossipPort)

	if err != nil {
		return nil, err
	}

	nodeFlags.ListenGossipAddr = listenGossipAddr

	listenRaftAddr, err := GetListenAddr(15760, ip, listenRaftPort)

	if err != nil {
		return nil, err
	}

	nodeFlags.ListenRaftAddr = listenRaftAddr

	listenRemoteAddr, err := GetListenAddr(27242, ip, listenRemotePort)

	if err != nil {
		return nil, err
	}

	nodeFlags.ListenRemoteAddr = listenRemoteAddr

	fields := logrus.Fields{
		"ConfigDir":        nodeFlags.ConfigDir,
		"ListenGossipAddr": nodeFlags.ListenGossipAddr.String(),
		"ListenRaftAddr":   nodeFlags.ListenRaftAddr.String(),
		"ListenRemoteAddr": nodeFlags.ListenRemoteAddr.String(),
	}

	logrus.WithFields(fields).Info("Loaded command-line flags")

	return nodeFlags, nil
}
