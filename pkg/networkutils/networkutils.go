package networkutils

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"syscall"
	"time"
)

func IsPortValid(port string) bool {
	p, err := strconv.Atoi(port)
	return err == nil && p >= 1 && p <= 65535
}

func ValidateIP(ip string) error {
	if ip == "" {
		return fmt.Errorf("is required")
	}
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return fmt.Errorf("is invalid: %s", ip)
	}

	return nil
}

// IsIPInUse performs a best effort check to see if an IP address is in use. It is not completely
// reliable as testing if an IP is in use is inherently difficult, particularly with non-trivial
// network topologies.
func IsIPInUse(client NetClient, ip string) bool {
	// Dial and immediately close the connection if it was established as its superfluous for
	// our check. We use port 80 as its common and is more likely to get through firewalls
	// than other ports.
	conn, err := client.DialTimeout("tcp", net.JoinHostPort(ip, "80"), 500*time.Millisecond)
	if err == nil {
		conn.Close()
	}

	// If we establish a connection or we receive a response assume that address is in use.
	// The latter case covers situations like an IP in use but the port requested is not open.
	return err == nil || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET)
}

func IsPortInUse(client NetClient, host string, port string) bool {
	address := net.JoinHostPort(host, port)
	conn, err := client.DialTimeout("tcp", address, 500*time.Millisecond)
	if err == nil {
		conn.Close()
		return true
	}

	return false
}

func GetLocalIP() (net.IP, error) {
	conn, err := net.Dial("udp", "1.2.3.4:80")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve local ip: %v", err)
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP, nil
}
