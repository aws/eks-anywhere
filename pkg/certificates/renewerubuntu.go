package certificates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	linuxEtcdCertDir           = "/etc/etcd"
	linuxControlPlaneCertDir   = "/etc/kubernetes/pki"
	linuxControlPlaneManifests = "/etc/kubernetes/manifests"
	linuxTempDir               = "/tmp"
)

// LinuxRenewer implements OSRenewer for Linux-based systems (Ubuntu / RHEL).
type LinuxRenewer struct {
	osType OSType
	backup string
}

// NewLinuxRenewer creates a new renewer for Linux-based operating systems.
func NewLinuxRenewer(backupDir string) *LinuxRenewer {
	return &LinuxRenewer{osType: OSTypeLinux, backup: backupDir}
}

// RenewControlPlaneCerts renews certificates for control plane nodes.
func (l *LinuxRenewer) RenewControlPlaneCerts(
	ctx context.Context,
	node string,
	cfg *RenewalConfig,
	component string,
	ssh SSHRunner,
) error {
	logger.V(0).Info("Processing control-plane node", "node", node)

	hasExternalEtcd := cfg != nil && len(cfg.Etcd.Nodes) > 0

	if _, err := ssh.RunCommand(ctx, node, l.backupControlPlaneCerts(component, hasExternalEtcd, l.backup)); err != nil {
		return fmt.Errorf("backing up control plane certs: %v", err)
	}
	if _, err := ssh.RunCommand(ctx, node, l.renewControlPlaneCerts(component, hasExternalEtcd)); err != nil {
		return fmt.Errorf("renewing control plane certs: %v", err)
	}

	if !hasExternalEtcd {
		if _, err := ssh.RunCommand(ctx, node, "sudo kubeadm certs check-expiration"); err != nil {
			return fmt.Errorf("validating control plane certs: %v", err)
		}
	} else {
		logger.V(4).Info("Skipping kubeadm cert validation for external etcd setup", "node", node)
	}

	if hasExternalEtcd {
		if err := l.transferCertsToControlPlane(ctx, node, ssh); err != nil {
			return fmt.Errorf("transferring certificates to control plane node: %v", err)
		}
		if _, err := ssh.RunCommand(ctx, node, l.copyExternalEtcdCerts(hasExternalEtcd)); err != nil {
			return fmt.Errorf("copying etcd client certs: %v", err)
		}
	}

	if _, err := ssh.RunCommand(ctx, node, l.restartControlPlaneStaticPods()); err != nil {
		return fmt.Errorf("restarting control plane pods: %v", err)
	}

	logger.V(0).Info("Renewed control-plane certificates", "node", node)
	return nil
}

// RenewEtcdCerts renews certificates for etcd nodes.
func (l *LinuxRenewer) RenewEtcdCerts(
	ctx context.Context,
	node string,
	ssh SSHRunner,
) error {
	logger.V(0).Info("Processing etcd node", "os", l.osType, "node", node)

	if _, err := ssh.RunCommand(ctx, node, l.backupEtcdCerts()); err != nil {
		return fmt.Errorf("backing up etcd certs: %v", err)
	}
	if _, err := ssh.RunCommand(ctx, node,
		"sudo etcdadm join phase certificates http://eks-a-etcd-dumb-url"); err != nil {
		return fmt.Errorf("renewing etcd certs: %v", err)
	}
	if _, err := ssh.RunCommand(ctx, node, l.validateEtcdCerts()); err != nil {
		return fmt.Errorf("validating etcd certs: %v", err)
	}
	logger.V(0).Info("Renewed etcd certificates", "node", node)
	return nil
}

// CopyEtcdCerts copies the etcd certificates from the specified node to the local machine.
func (l *LinuxRenewer) CopyEtcdCerts(
	ctx context.Context,
	node string,
	ssh SSHRunner,
) error {
	cat := func(file string) (string, error) {
		return ssh.RunCommand(ctx, node,
			fmt.Sprintf("sudo cat %s", filepath.Join(linuxEtcdCertDir, file)))
	}

	crt, err := cat("pki/apiserver-etcd-client.crt")
	if err != nil {
		return fmt.Errorf("read crt: %v", err)
	}
	key, err := cat("pki/apiserver-etcd-client.key")
	if err != nil {
		return fmt.Errorf("read key: %v", err)
	}
	if crt == "" || key == "" {
		return fmt.Errorf("etcd client cert or key is empty")
	}

	dstDir := filepath.Join(l.backup, tempLocalEtcdCertsDir)
	if err := os.MkdirAll(dstDir, 0o700); err != nil {
		return fmt.Errorf("creating etcd backup direcotry %s: %v", dstDir, err)
	}
	if err := os.WriteFile(filepath.Join(dstDir, "apiserver-etcd-client.crt"), []byte(crt), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dstDir, "apiserver-etcd-client.key"), []byte(key), 0o600); err != nil {
		return err
	}

	logger.V(4).Info("Copied etcd client certs", "path", dstDir)
	return nil
}

func (l *LinuxRenewer) backupControlPlaneCerts(_ string, hasExternalEtcd bool, backup string) string {
	backupPath := fmt.Sprintf("/etc/kubernetes/pki.bak_%s", backup)
	if hasExternalEtcd {
		return fmt.Sprintf("sudo sh -c 'cp -r %s \"%s\" && rm -rf \"%s/etcd\"'",
			linuxControlPlaneCertDir, backupPath, backupPath)
	}
	return fmt.Sprintf("sudo cp -r %s %s", linuxControlPlaneCertDir, backupPath)
}

func (l *LinuxRenewer) renewControlPlaneCerts(_ string, hasExternalEtcd bool) string {
	if hasExternalEtcd {
		return "sudo sh -c 'for cert in admin.conf apiserver apiserver-kubelet-client controller-manager.conf front-proxy-client scheduler.conf; do kubeadm certs renew \"$cert\"; done'"
	}
	return "sudo kubeadm certs renew all"
}

func (l *LinuxRenewer) restartControlPlaneStaticPods() string {
	return fmt.Sprintf("sudo sh -c 'mkdir -p /tmp/manifests && mv %s/* /tmp/manifests/ && sleep 20 && mv /tmp/manifests/* %s/'",
		linuxControlPlaneManifests, linuxControlPlaneManifests)
}

func (l *LinuxRenewer) backupEtcdCerts() string {
	return fmt.Sprintf("sudo sh -c 'cd %[1]s && cp -r pki pki.bak_%[2]s && rm -rf pki/* && cp pki.bak_%[2]s/ca.* pki/'",
		linuxEtcdCertDir, l.backup)
}

func (l *LinuxRenewer) validateEtcdCerts() string {
	return fmt.Sprintf("sudo etcdctl --cacert=%[1]s/pki/ca.crt --cert=%[1]s/pki/etcdctl-etcd-client.crt --key=%[1]s/pki/etcdctl-etcd-client.key member list",
		linuxEtcdCertDir)
}

func (l *LinuxRenewer) transferCertsToControlPlane(
	ctx context.Context, node string, ssh SSHRunner,
) error {
	logger.V(4).Info("Transferring certificates to control-plane node", "node", node)

	crtPath := filepath.Join(l.backup, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt")
	keyPath := filepath.Join(l.backup, tempLocalEtcdCertsDir, "apiserver-etcd-client.key")

	crtData, err := os.ReadFile(crtPath)
	if err != nil {
		return fmt.Errorf("reading certificate file: %v", err)
	}
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("reading key file: %v", err)
	}

	crtCmd := fmt.Sprintf("sudo tee %s/apiserver-etcd-client.crt > /dev/null << 'EOF'\n%s\nEOF", linuxTempDir, string(crtData))
	if _, err := ssh.RunCommand(ctx, node, crtCmd); err != nil {
		return fmt.Errorf("copying certificate: %v", err)
	}

	keyCmd := fmt.Sprintf("sudo tee %s/apiserver-etcd-client.key > /dev/null << 'EOF'\n%s\nEOF", linuxTempDir, string(keyData))
	if _, err := ssh.RunCommand(ctx, node, keyCmd); err != nil {
		return fmt.Errorf("copying key: %v", err)
	}

	logger.V(4).Info("Certificates transferred", "node", node)
	return nil
}

func (l *LinuxRenewer) copyExternalEtcdCerts(hasExternalEtcd bool) string {
	if hasExternalEtcd {
		return fmt.Sprintf("sudo sh -c 'if [ -f %[1]s/apiserver-etcd-client.crt ]; then cp %[1]s/apiserver-etcd-client.* %[2]s/ && rm -f %[1]s/apiserver-etcd-client.*; fi'",
			linuxTempDir, linuxControlPlaneCertDir)
	}
	return "true"
}
