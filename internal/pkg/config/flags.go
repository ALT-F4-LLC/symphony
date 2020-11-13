package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
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
func GetListenAddr(defaultPort int, overridePort *int) (*net.TCPAddr, error) {
	if *overridePort != 0 {
		defaultPort = *overridePort
	}

	listenAddr := fmt.Sprintf(":%d", defaultPort)

	listenTCPAddr, err := net.ResolveTCPAddr("tcp", listenAddr)

	if err != nil {
		return nil, err
	}

	return listenTCPAddr, nil
}
