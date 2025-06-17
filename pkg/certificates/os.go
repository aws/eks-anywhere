package certificates

import (
	"context"
	"fmt"
)

// OSType represents the type of operating system.
type OSType string

const (
	// OSTypeLinux represents Linux-based operating systems.
	OSTypeLinux OSType = "linux"
	// OSTypeBottlerocket represents Bottlerocket OS.
	OSTypeBottlerocket OSType = "bottlerocket"
)

// OSRenewer defines the interface for OS-specific certificate renewal operations.
type OSRenewer interface {
	// RenewControlPlaneCerts renews control plane certificates on a node
	RenewControlPlaneCerts(ctx context.Context, node string, config *RenewalConfig, component string, sshRunner SSHRunner, backupDir string) error

	// RenewEtcdCerts renews etcd certificates on a node
	RenewEtcdCerts(ctx context.Context, node string, sshRunner SSHRunner, backupDir string) error
}

// BuildOSRenewer creates a new OSRenewer based on the OS type.
func BuildOSRenewer(osType string) (OSRenewer, error) {
	osBuilder, ok := osRenewerBuilders[osType]
	if !ok {
		return nil, fmt.Errorf("unsupported OS type: %s", osType)
	}

	return osBuilder(), nil
}

// Map of OS type to OSRenewer builder functions.
var osRenewerBuilders = map[string]func() OSRenewer{
	string(OSTypeLinux): func() OSRenewer {
		return NewLinuxRenewer()
	},
	string(OSTypeBottlerocket): func() OSRenewer {
		return NewBottlerocketRenewer()
	},
}
