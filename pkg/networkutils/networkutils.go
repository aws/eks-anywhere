package networkutils

import (
	"fmt"
	"net"
	"strconv"
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

// IsIPInUse performs a soft check to see if there are any services listening on a selection of common ports at
// ip by trying to establish a TCP connection. Ports checked include: 22, 23, 80, 443 and 6443 (Kubernetes API Server).
// Each connection attempt allows up-to 500ms for a response.
//
// todo(chrisdoherty) change to an icmp approach to eliminate the need for ports.
func IsIPInUse(client NetClient, ip string) bool {
	ports := []string{"22", "23", "80", "443", "6443"}
	for _, port := range ports {
		address := net.JoinHostPort(ip, port)
		conn, err := client.DialTimeout("tcp", address, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
	}

	return false
}
