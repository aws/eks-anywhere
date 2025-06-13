package certificates

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/constants"
)

// LinuxControlPlaneCommands holds the commands needed for control plane certificate operations.
type LinuxControlPlaneCommands struct {
	Backup   string
	Renew    string
	Validate string
	Restart  string
}

// LinuxControlPlaneCommandBuilder builds commands for control plane certificate operations.
type LinuxControlPlaneCommandBuilder struct {
	Paths           CertificatePaths
	BackupDir       string
	Component       string
	HasExternalEtcd bool
}

// NewLinuxControlPlaneCommandBuilder creates a new builder for control plane certificate commands.
func NewLinuxControlPlaneCommandBuilder(paths CertificatePaths,
	backupDir, component string, hasExternalEtcd bool,
) *LinuxControlPlaneCommandBuilder {
	return &LinuxControlPlaneCommandBuilder{
		Paths:           paths,
		BackupDir:       backupDir,
		Component:       component,
		HasExternalEtcd: hasExternalEtcd,
	}
}

// Build creates a set of commands for control plane certificate operations.
func (b *LinuxControlPlaneCommandBuilder) Build() *LinuxControlPlaneCommands {
	return &LinuxControlPlaneCommands{
		Backup:   b.buildBackup(),
		Renew:    b.buildRenew(),
		Validate: "sudo kubeadm certs check-expiration",
		Restart:  b.buildRestart(),
	}
}

// Backup certificates, excluding etcd directory if component is control-plane.
func (b *LinuxControlPlaneCommandBuilder) buildBackup() string {
	if b.Component == constants.ControlPlaneComponent && b.HasExternalEtcd {
		// When only updating control plane with external etcd, exclude etcd directory
		return fmt.Sprintf(`
sudo mkdir -p '/etc/kubernetes/pki.bak_%[1]s'
cd %[2]s
for f in $(find . -type f ! -path './etcd/*'); do
    sudo mkdir -p $(dirname '/etc/kubernetes/pki.bak_%[1]s/'$f)
    sudo cp $f '/etc/kubernetes/pki.bak_%[1]s/'$f
done`, b.BackupDir, b.Paths.ControlPlaneCertDir)
	}
	return fmt.Sprintf("sudo cp -r '%s' '/etc/kubernetes/pki.bak_%s'",
		b.Paths.ControlPlaneCertDir, b.BackupDir)
}

func (b *LinuxControlPlaneCommandBuilder) buildRenew() string {
	if b.Component == constants.ControlPlaneComponent && b.HasExternalEtcd {
		// When only renewing control plane certs with external etcd,
		// we need to skip the etcd directory to preserve certificates
		return `for cert in admin.conf apiserver apiserver-kubelet-client controller-manager.conf front-proxy-client scheduler.conf; do
            sudo kubeadm certs renew $cert
done`
	}
	return "sudo kubeadm certs renew all"
}

func (b *LinuxControlPlaneCommandBuilder) buildRestart() string {
	return fmt.Sprintf("sudo mkdir -p /tmp/manifests && "+
		"sudo mv %s/* /tmp/manifests/ && "+
		"sleep 20 && "+
		"sudo mv /tmp/manifests/* %s/",
		b.Paths.ControlPlaneManifests, b.Paths.ControlPlaneManifests)
}

// LinuxEtcdCommands holds the commands needed for etcd certificate operations.
type LinuxEtcdCommands struct {
	Backup   string
	Renew    string
	Validate string
}

// LinuxEtcdCommandBuilder builds commands for etcd certificate operations.
type LinuxEtcdCommandBuilder struct {
	Paths     CertificatePaths
	BackupDir string
}

// NewLinuxEtcdCommandBuilder creates a new builder for etcd certificate commands.
func NewLinuxEtcdCommandBuilder(paths CertificatePaths, backupDir string) *LinuxEtcdCommandBuilder {
	return &LinuxEtcdCommandBuilder{
		Paths:     paths,
		BackupDir: backupDir,
	}
}

// Build creates a set of commands for etcd certificate operations.
func (b *LinuxEtcdCommandBuilder) Build() *LinuxEtcdCommands {
	return &LinuxEtcdCommands{
		Backup: b.buildBackup(),
		Renew:  "sudo etcdadm join phase certificates http://eks-a-etcd-dumb-url",
		Validate: fmt.Sprintf("sudo etcdctl --cacert=%s/pki/ca.crt "+
			"--cert=%s/pki/etcdctl-etcd-client.crt "+
			"--key=%s/pki/etcdctl-etcd-client.key "+
			"endpoint health",
			b.Paths.EtcdCertDir, b.Paths.EtcdCertDir, b.Paths.EtcdCertDir),
	}
}

func (b *LinuxEtcdCommandBuilder) buildBackup() string {
	return fmt.Sprintf("cd %s && sudo cp -r pki pki.bak_%s && sudo rm -rf pki/* && sudo cp pki.bak_%s/ca.* pki/",
		b.Paths.EtcdCertDir, b.BackupDir, b.BackupDir)
}
