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
	JoinAddr         *net.TCPAddr
	Peers            []string
	ListenRaftAddr   *net.TCPAddr
	ListenRemoteAddr *net.TCPAddr
}

var (
	configDir        = kingpin.Flag("config-dir", "Sets configuration directory for manager.").Default(".").String()
	joinAddr         = kingpin.Flag("join-addr", "Sets existing Raft remote manager address to join.").String()
	peersList        = kingpin.Flag("peers", "Sets the initial custer size (minimum of 2 other nodes required).").String()
	listenRaftAddr   = kingpin.Flag("listen-raft-addr", "Sets the raft address.").String()
	listenRemoteAddr = kingpin.Flag("listen-remote-addr", "Sets the remote API address.").String()
)

// GetOutboundIP : get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
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

	if *joinAddr != "" {
		join, err := net.ResolveTCPAddr("tcp", *joinAddr)

		if err != nil {
			return nil, err
		}

		nodeFlags.JoinAddr = join
	}

	if *listenRaftAddr != "" {
		addr, err := net.ResolveTCPAddr("tcp", *listenRaftAddr)

		if err != nil {
			return nil, err
		}

		nodeFlags.ListenRaftAddr = addr
	} else {
		ip := GetOutboundIP()

		tcpAddr := fmt.Sprintf("%s:15760", ip.String())

		addr, err := net.ResolveTCPAddr("tcp", tcpAddr)

		if err != nil {
			return nil, err
		}

		nodeFlags.ListenRaftAddr = addr
	}

	if *listenRemoteAddr != "" {
		addr, err := net.ResolveTCPAddr("tcp", *listenRemoteAddr)

		if err != nil {
			return nil, err
		}

		nodeFlags.ListenRemoteAddr = addr
	} else {
		ip := GetOutboundIP()

		tcpAddr := fmt.Sprintf("%s:27242", ip.String())

		addr, err := net.ResolveTCPAddr("tcp", tcpAddr)

		if err != nil {
			return nil, err
		}

		nodeFlags.ListenRemoteAddr = addr
	}

	if peersList != nil {
		peers, err := GetPeersFromString(peersList)

		if err != nil {
			return nil, err
		}

		nodeFlags.Peers = peers
	}

	fields := logrus.Fields{
		"ConfigDir":        nodeFlags.ConfigDir,
		"JoinAddr":         nodeFlags.JoinAddr.String(),
		"Peers":            nodeFlags.Peers,
		"ListenRaftAddr":   nodeFlags.ListenRaftAddr.String(),
		"ListenRemoteAddr": nodeFlags.ListenRemoteAddr.String(),
	}

	logrus.WithFields(fields).Info("Loaded command-line flags")

	return nodeFlags, nil
}
