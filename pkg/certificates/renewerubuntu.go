package certificates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/logger"
)

// LinuxRenewer implements OSRenewer for Linux-based systems (Ubuntu and RHEL).
type LinuxRenewer struct {
	certPaths CertificatePaths
	osType    OSType
}

// NewLinuxRenewer creates a new LinuxRenewer with the specified certificate paths and OS type.
func NewLinuxRenewer(paths CertificatePaths, osType OSType) *LinuxRenewer {
	return &LinuxRenewer{certPaths: paths, osType: osType}
}

// RenewControlPlaneCerts renews control plane certificates on a Linux node.
func (l *LinuxRenewer) RenewControlPlaneCerts(ctx context.Context, node string, cfg *RenewalConfig, component string, ssh SSHRunner, backupDir string) error {
	logger.V(2).Info(fmt.Sprintf("Processing node %s...", node))

	hasExternalEtcd := cfg != nil && len(cfg.Etcd.Nodes) > 0

	builder := NewLinuxControlPlaneCommandBuilder(l.certPaths, backupDir, component, hasExternalEtcd)
	cmds := builder.Build()

	if err := ssh.RunCommand(ctx, node, cmds.Backup); err != nil {
		return fmt.Errorf("backup certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node, cmds.Renew); err != nil {
		return fmt.Errorf("renew certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node, cmds.Validate); err != nil {
		return fmt.Errorf("validate certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node, cmds.Restart); err != nil {
		return fmt.Errorf("restart pods: %v", err)
	}

	logger.MarkPass(fmt.Sprintf("Renewed certificates for node %s", node))
	return nil
}

// RenewEtcdCerts renews etcd certificates on a Linux node.
func (l *LinuxRenewer) RenewEtcdCerts(ctx context.Context, node string, ssh SSHRunner, backupDir string) error {
	logger.V(2).Info("Processing etcd node", "os", l.osType, "node", node)

	builder := NewLinuxEtcdCommandBuilder(l.certPaths, backupDir)
	cmds := builder.Build()

	if err := ssh.RunCommand(ctx, node, cmds.Backup); err != nil {
		logger.MarkFail("Failed to backup certificates on node", "node", node)
		return fmt.Errorf("backup certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node, cmds.Renew); err != nil {
		logger.MarkFail("Failed to renew certificates on node", "node", node)
		return fmt.Errorf("renew certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node, cmds.Validate); err != nil {
		logger.MarkFail("Failed to validate certificates on node", "node", node)
		return fmt.Errorf("validate certs: %v", err)
	}

	if err := l.copyEtcdCerts(ctx, node, ssh, backupDir); err != nil {
		logger.MarkFail("Failed to copy certificates from node", "node", node)
		return fmt.Errorf("copy certs: %v", err)
	}

	logger.MarkPass(fmt.Sprintf("Renewed certificates for etcd node %s", node))
	return nil
}

func (l *LinuxRenewer) copyEtcdCerts(ctx context.Context, node string, ssh SSHRunner, backupDir string) error {
	etcdDir := l.certPaths.EtcdCertDir
	cat := func(file string) (string, error) {
		cmd := fmt.Sprintf("sudo cat %s/%s", etcdDir, file)
		return ssh.RunCommandWithOutput(ctx, node, cmd)
	}

	crt, err := cat("pki/apiserver-etcd-client.crt")
	if err != nil {
		logger.MarkFail("Failed to read certificate from node", "node", node)
		return fmt.Errorf("read crt: %v", err)
	}
	key, err := cat("pki/apiserver-etcd-client.key")
	if err != nil {
		logger.MarkFail("Failed to read key from node", "node", node)
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
	if err := os.WriteFile(filepath.Join(dstDir, "apiserver-etcd-client.crt"),
		[]byte(crt), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dstDir, "apiserver-etcd-client.key"),
		[]byte(key), 0o600); err != nil {
		return err
	}

	logger.V(2).Info("Copied etcd client certs", "path", dstDir)
	return nil
}
