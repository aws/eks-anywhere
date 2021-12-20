package networkutils_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/networkutils"
)

var (
	validPorts   = []string{"443", "8080", "32000"}
	invalidPorts = []string{"", "443a", "abc", "0", "123456"}
)

func TestIsPortValidExpectValid(t *testing.T) {
	for _, port := range validPorts {
		if !networkutils.IsPortValid(port) {
			t.Fatalf("Expected port %s to be valid", port)
		}
	}
}

func TestIsPortValidExpectInvalid(t *testing.T) {
	for _, port := range invalidPorts {
		if networkutils.IsPortValid(port) {
			t.Fatalf("Expected port %s to be invalid", port)
		}
	}
}
