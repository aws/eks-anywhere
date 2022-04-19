package networkutils_test

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/aws/eks-anywhere/pkg/networkutils"
)

type DummyNetClient struct{}

func (n *DummyNetClient) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	// add dummy case for coverage
	if address == "255.255.255.255:22" {
		return &net.IPConn{}, nil
	}
	return nil, errors.New("")
}

func TestGenerateUniqueIP(t *testing.T) {
	cidrBlock := "1.2.3.4/16"

	ipgen := networkutils.NewIPGenerator(&DummyNetClient{})
	ip, err := ipgen.GenerateUniqueIP(cidrBlock)
	if err != nil {
		t.Fatalf("GenerateUniqueIP() ip = %v error: %v", ip, err)
	}
}
