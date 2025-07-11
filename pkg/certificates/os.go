package certificates

import (
	"context"
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
	RenewControlPlaneCerts(ctx context.Context, node string, config *RenewalConfig, component string, sshRunner SSHRunner) error
	RenewEtcdCerts(ctx context.Context, node string, sshRunner SSHRunner) error
	CopyEtcdCerts(ctx context.Context, node string, sshRunner SSHRunner) error
	TransferCertsToControlPlane(ctx context.Context, node string, sshRunner SSHRunner) error
}

// BuildOSRenewer creates a new OSRenewer based on the OS type.
func BuildOSRenewer(osType string, backupDir string) OSRenewer {
	return osRenewerBuilders[osType](backupDir)
}

// Map of OS type to OSRenewer builder functions.
var osRenewerBuilders = map[string]func(backupDir string) OSRenewer{
	string(OSTypeLinux): func(backupDir string) OSRenewer {
		return NewLinuxRenewer(backupDir)
	},
	// comment for focus ubuntu pr
	string(OSTypeBottlerocket): func(backupDir string) OSRenewer {
		return NewBottlerocketRenewer(backupDir)
	},
}
