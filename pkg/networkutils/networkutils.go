package networkutils

import (
	"fmt"
	"net"
	"strconv"
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
