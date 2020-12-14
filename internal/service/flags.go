package service

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	// ContextTimeout : default context timeout
	ContextTimeout = 5 * time.Second

	// DialTimeout : default dial timeout
	DialTimeout = 5 * time.Second
)

// GetDirPath : resolves full path for directory
func GetDirPath(dir *string) (*string, error) {
	path, err := filepath.Abs(*dir)

	if err != nil {
		return nil, err
	}

	_, statErr := os.Stat(path)

	if statErr != nil {
		if err := os.MkdirAll(path, 0750); err != nil {
			return nil, err
		}
	}

	return &path, nil
}

// GetListenAddr : returns the TCP listen addr
func GetListenAddr(ipAddr string, port int) (*net.TCPAddr, error) {
	listenAddr := fmt.Sprintf("%s:%d", ipAddr, port)

	listenAddrTCP, err := net.ResolveTCPAddr("tcp", listenAddr)

	if err != nil {
		return nil, err
	}

	return listenAddrTCP, nil
}

// GetOutboundIP : get preferred outbound ip of this machine
func GetOutboundIP() (*net.IP, error) {
	conn, err := net.Dial("udp", "1.1.1.1:80")

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)

	ip := addr.IP

	return &ip, nil
}
