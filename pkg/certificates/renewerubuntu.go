package certificates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/constants"
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
}

// NewLinuxRenewer creates a new renewer for Linux-based operating systems.
func NewLinuxRenewer() *LinuxRenewer { return &LinuxRenewer{osType: OSTypeLinux} }

// RenewControlPlaneCerts renews certificates for control plane nodes.
func (l *LinuxRenewer) RenewControlPlaneCerts(
	ctx context.Context,
	node string,
	cfg *RenewalConfig,
	component string,
	ssh SSHRunner,
	backupDir string,
) error {
	logger.V(2).Info("Processing control-plane node", "node", node)

	hasExternalEtcd := cfg != nil && len(cfg.Etcd.Nodes) > 0

	if err := ssh.RunCommand(ctx, node, buildCPBackupCmd(component, hasExternalEtcd, backupDir)); err != nil {
		return fmt.Errorf("backup certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node, buildCPRenewCmd(component, hasExternalEtcd)); err != nil {
		return fmt.Errorf("renew certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node, []string{"sudo", "kubeadm", "certs", "check-expiration"}); err != nil {
		return fmt.Errorf("validate certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node, buildCPRestartCmd()); err != nil {
		return fmt.Errorf("restart pods: %v", err)
	}

	logger.MarkPass("Renewed control-plane certificates", "node", node)
	return nil
}

// RenewEtcdCerts renews certificates for etcd nodes.
func (l *LinuxRenewer) RenewEtcdCerts(
	ctx context.Context,
	node string,
	ssh SSHRunner,
	backupDir string,
) error {
	logger.V(2).Info("Processing etcd node", "os", l.osType, "node", node)

	if err := ssh.RunCommand(ctx, node, buildEtcdBackupCmd(backupDir)); err != nil {
		return fmt.Errorf("backup certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node,
		[]string{"sudo", "etcdadm", "join", "phase", "certificates", "http://eks-a-etcd-dumb-url"}); err != nil {
		return fmt.Errorf("renew certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node, buildEtcdValidateCmd()); err != nil {
		return fmt.Errorf("validate certs: %v", err)
	}
	if err := l.copyEtcdCerts(ctx, node, ssh, backupDir); err != nil {
		return err
	}

	logger.MarkPass("Renewed etcd certificates", "node", node)
	return nil
}

func (l *LinuxRenewer) copyEtcdCerts(
	ctx context.Context,
	node string,
	ssh SSHRunner,
	backupDir string,
) error {
	cat := func(file string) (string, error) {
		return ssh.RunCommandWithOutput(ctx, node,
			[]string{"sudo", "cat", filepath.Join(linuxEtcdCertDir, file)})
	}

	crt, err := cat("pki/apiserver-etcd-client.crt")
	if err != nil {
		logger.MarkFail("Failed to read certificate", "node", node)
		return fmt.Errorf("read crt: %v", err)
	}
	key, err := cat("pki/apiserver-etcd-client.key")
	if err != nil {
		logger.MarkFail("Failed to read key", "node", node)
		return fmt.Errorf("read key: %v", err)
	}
	if crt == "" || key == "" {
		logger.MarkFail("Certificate or key is empty")
		return fmt.Errorf("etcd client cert or key is empty")
	}

	dstDir := filepath.Join(backupDir, tempLocalEtcdCertsDir)
	if err := os.MkdirAll(dstDir, 0o700); err != nil {
		logger.MarkFail("Failed to create directory", "path", dstDir)
		return fmt.Errorf("mkdir %s: %v", dstDir, err)
	}
	if err := os.WriteFile(filepath.Join(dstDir, "apiserver-etcd-client.crt"), []byte(crt), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dstDir, "apiserver-etcd-client.key"), []byte(key), 0o600); err != nil {
		return err
	}

	logger.V(2).Info("Copied etcd client certs", "path", dstDir)
	return nil
}

func buildCPBackupCmd(component string, hasExternalEtcd bool, backup string) []string {
	backupPath := fmt.Sprintf("/etc/kubernetes/pki.bak_%s", backup)

	if component == constants.ControlPlaneComponent && hasExternalEtcd {
		script := fmt.Sprintf(
			"cp -r %[1]s '%[2]s' && rm -rf '%[2]s/etcd'",
			linuxControlPlaneCertDir,
			backupPath,
		)
		return []string{"sudo", "sh", "-c", script}
	}
	return []string{
		"sudo", "cp", "-r", linuxControlPlaneCertDir,
		fmt.Sprintf("/etc/kubernetes/pki.bak_%s", backup),
	}
}

func buildCPRenewCmd(component string, hasExternalEtcd bool) []string {
	if component == constants.ControlPlaneComponent && hasExternalEtcd {
		script := `
for cert in admin.conf apiserver apiserver-kubelet-client controller-manager.conf front-proxy-client scheduler.conf; do
    kubeadm certs renew "$cert"
done`
		return []string{"sudo", "sh", "-c", script}
	}
	return []string{"sudo", "kubeadm", "certs", "renew", "all"}
}

func buildCPRestartCmd() []string {
	script := fmt.Sprintf(
		"mkdir -p /tmp/manifests && mv %s/* /tmp/manifests/ && sleep 20 && mv /tmp/manifests/* %s/",
		linuxControlPlaneManifests, linuxControlPlaneManifests,
	)
	return []string{"sudo", "sh", "-c", script}
}

func buildEtcdBackupCmd(backup string) []string {
	script := fmt.Sprintf(
		"cd %[1]s && cp -r pki pki.bak_%[2]s && rm -rf pki/* && cp pki.bak_%[2]s/ca.* pki/",
		linuxEtcdCertDir, backup,
	)
	return []string{"sudo", "sh", "-c", script}
}

func buildEtcdValidateCmd() []string {
	return []string{
		"sudo", "etcdctl",
		"--cacert=" + filepath.Join(linuxEtcdCertDir, "pki", "ca.crt"),
		"--cert=" + filepath.Join(linuxEtcdCertDir, "pki", "etcdctl-etcd-client.crt"),
		"--key=" + filepath.Join(linuxEtcdCertDir, "pki", "etcdctl-etcd-client.key"),
		"endpoint", "health",
	}
}
