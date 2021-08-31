package networkutils

import (
	"net"
	"time"
)

type NetClient interface {
	DialTimeout(network, address string, timeout time.Duration) (net.Conn, error)
}

type DefaultNetClient struct{}

func (n *DefaultNetClient) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(network, address, timeout)
}
