package certificates

import (
	"context"
	"fmt"
)

// OSType represents the type of operating system.
type OSType string

const (
	// OSTypeUbuntu represents Ubuntu OS.
	OSTypeUbuntu OSType = "ubuntu"
	// OSTypeRHEL represents RHEL OS.
	OSTypeRHEL OSType = "redhat"
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

// Map of OS type to OSRenewer builder functions
// var osRenewerBuilders = map[string]func() OSRenewer{
// 	string(OSTypeUbuntu): func() OSRenewer {
// 		return NewLinuxRenewer(GetCertificatePaths(string(OSTypeUbuntu)))
// 	},
// 	string(OSTypeRHEL): func() OSRenewer {
// 		return NewLinuxRenewer(GetCertificatePaths(string(OSTypeRHEL)))
// 	},
// 	string(OSTypeBottlerocket): func() OSRenewer {
// 		return NewBottlerocketRenewer(GetCertificatePaths(string(OSTypeBottlerocket)))
// 	},
// }

var osRenewerBuilders = map[string]func() OSRenewer{
	string(OSTypeUbuntu): func() OSRenewer {
		return NewLinuxRenewer(GetCertificatePaths(string(OSTypeUbuntu)), OSTypeUbuntu)
	},
	string(OSTypeRHEL): func() OSRenewer {
		return NewLinuxRenewer(GetCertificatePaths(string(OSTypeRHEL)), OSTypeRHEL)
	},
	string(OSTypeBottlerocket): func() OSRenewer {
		return NewBottlerocketRenewer(GetCertificatePaths(string(OSTypeBottlerocket)))
	},
}

// CertificatePaths contains the paths to certificate directories for different OS types.
type CertificatePaths struct {
	// EtcdCertDir is the directory containing etcd certificates
	EtcdCertDir string
	// ControlPlaneCertDir is the directory containing control plane certificates
	ControlPlaneCertDir string
	// ControlPlaneManifests is the directory containing control plane manifests
	ControlPlaneManifests string
	// TempDir is a temporary directory used for operations
	TempDir string
}

// GetCertificatePaths returns the appropriate certificate paths for the given OS type.
func GetCertificatePaths(osType string) CertificatePaths {
	switch osType {
	case string(OSTypeUbuntu), string(OSTypeRHEL):
		return CertificatePaths{
			EtcdCertDir:           ubuntuEtcdCertDir,
			ControlPlaneCertDir:   ubuntuControlPlaneCertDir,
			ControlPlaneManifests: ubuntuControlPlaneManifests,
			TempDir:               "/tmp",
		}
	case string(OSTypeBottlerocket):
		return CertificatePaths{
			EtcdCertDir:           bottlerocketEtcdCertDir,
			ControlPlaneCertDir:   bottlerocketControlPlaneCertDir,
			ControlPlaneManifests: "/var/lib/kubeadm/manifests",
			TempDir:               bottlerocketTmpDir,
		}
	default:
		// Default to Linux paths
		return CertificatePaths{
			EtcdCertDir:           ubuntuEtcdCertDir,
			ControlPlaneCertDir:   ubuntuControlPlaneCertDir,
			ControlPlaneManifests: ubuntuControlPlaneManifests,
			TempDir:               "/tmp",
		}
	}
}
