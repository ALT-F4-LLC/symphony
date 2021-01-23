package utils

import (
	"fmt"
	"net"
)

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
