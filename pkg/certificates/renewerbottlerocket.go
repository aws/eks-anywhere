package certificates

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	persistentCertDir     = "/var/lib/eks-anywhere/certificates"
	persistentEtcdCertDir = "etcd-certs"

	brEtcdCertDir           = "/var/lib/etcd"
	brControlPlaneCertDir   = "/var/lib/kubeadm/pki"
	brControlPlaneManifests = "/var/lib/kubeadm/manifests"
	brTempDir               = "/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp"
)

// BottlerocketRenewer implements OSRenewer for Bottlerocket systems.
type BottlerocketRenewer struct {
	osType string
	backup string
}

// NewBottlerocketRenewer creates a new BottlerocketRenewer.
func NewBottlerocketRenewer(backupDir string) *BottlerocketRenewer {
	return &BottlerocketRenewer{
		osType: string(OSTypeBottlerocket),
		backup: backupDir,
	}
}

// RenewControlPlaneCerts renews control plane certificates on a Bottlerocket node.
func (b *BottlerocketRenewer) RenewControlPlaneCerts(ctx context.Context, node string, config *RenewalConfig, component string, sshRunner SSHRunner) error {
	logger.V(4).Info("Processing control plane node", "node", node)

	// for renew control panel only.
	if component == constants.ControlPlaneComponent && len(config.Etcd.Nodes) > 0 {
		if err := b.loadCertsFromPersistentStorage(); err != nil {
			return fmt.Errorf("loading certificates from persistent storage: %v", err)
		}
	}

	// If we have external etcd nodes, first transfer certificates to the node
	if len(config.Etcd.Nodes) > 0 {
		if err := b.transferCertsToControlPlane(ctx, node, sshRunner); err != nil {
			return fmt.Errorf("transfer certificates to control plane node: %v", err)
		}
	}

	sessionCmds := buildBRSheltieCmd(
		buildBRImagePullCmd(),
		buildBRControlPlaneBackupCertsCmd(component, len(config.Etcd.Nodes) > 0, b.backup, brControlPlaneCertDir),
		buildBRControlPlaneRenewCertsCmd(),
		buildBRControlPlaneCopyCertsFromTmpCmd(),
		buildBRControlPlaneRestartPodsCmd(),
	)

	if _, err := sshRunner.RunCommand(ctx, node, sessionCmds); err != nil {
		return fmt.Errorf("renew control panel node certificates: %v", err)
	}

	logger.Info("Renewed certificates for control plane node", "node", node)
	return nil
}

func (b *BottlerocketRenewer) transferCertsToControlPlane(
	ctx context.Context, node string, sshRunner SSHRunner,
) error {
	logger.V(4).Info("Transferring certificates to control-plane node", "node", node)

	crtB, err := os.ReadFile(filepath.Join(
		b.backup, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt"))
	if err != nil {
		return fmt.Errorf("read certificate file: %v", err)
	}
	keyB, err := os.ReadFile(filepath.Join(
		b.backup, tempLocalEtcdCertsDir, "apiserver-etcd-client.key"))
	if err != nil {
		return fmt.Errorf("read key file: %v", err)
	}

	sessionCmds := buildBRSheltieCmd(
		buildBRCreateTmpDirCmd(tempLocalEtcdCertsDir),
		buildBRWriteCertToTmpCmd(base64.StdEncoding.EncodeToString(crtB)),
		buildBRWriteKeyToTmpCmd(base64.StdEncoding.EncodeToString(keyB)),
		buildBRSetTmpCertPermissionsCmd(),
	)

	if _, err := sshRunner.RunCommand(ctx, node, sessionCmds); err != nil {
		return fmt.Errorf("transfer certificates: %v", err)
	}

	logger.V(4).Info("Certificates transferred", "node", node)
	return nil
}

// RenewEtcdCerts renews etcd certificates on a Bottlerocket node.
func (b *BottlerocketRenewer) RenewEtcdCerts(ctx context.Context, node string, sshRunner SSHRunner) error {
	logger.V(4).Info("Processing etcd node", "node", node)

	remoteTempDir := brTempDir

	if _, err := sshRunner.RunCommand(ctx, node, buildBRSheltieCmd(
		buildBRImagePullCmd(),
		buildBREtcdBackupCertsCmd(b.backup),
		buildBREtcdRenewCertsCmd(),
	)); err != nil {
		return fmt.Errorf("renew certificates: %v", err)
	}

	if _, err := sshRunner.RunCommand(ctx, node, buildBRSheltieCmd(
		buildBREtcdCopyCertsToTmpCmd(remoteTempDir),
	)); err != nil {
		return fmt.Errorf("copy certificates to tmp: %v", err)
	}

	if _, err := sshRunner.RunCommand(ctx, node, buildBRSheltieCmd(
		buildBREtcdCleanupTmpCmd(remoteTempDir),
	)); err != nil {
		return fmt.Errorf("cleanup temporary files: %v", err)
	}

	logger.Info("Renewed certificates for etcd node", "node", node)

	return nil
}

func (b *BottlerocketRenewer) CopyEtcdCerts(ctx context.Context, node string, sshRunner SSHRunner) error {
	logger.V(4).Info("Reading certificate from ETCD node", "node", node)
	logger.V(4).Info("Using backup directory", "path", b.backup)

	remoteTempDir := brTempDir

	if _, err := sshRunner.RunCommand(ctx, node, buildBRSheltieCmd(
		buildBREtcdCopyCertsToTmpCmd(remoteTempDir),
	)); err != nil {
		return fmt.Errorf("copy certificates to tmp: %v", err)
	}

	if _, err := sshRunner.RunCommand(ctx, node, buildBRListTmpFilesCmd(remoteTempDir)); err != nil {
		return fmt.Errorf("list certificate files: %v", err)
	}

	crtContent, err := sshRunner.RunCommand(ctx, node, buildBRReadTmpCertCmd(remoteTempDir))
	if err != nil {
		return fmt.Errorf("read certificate file: %v", err)
	}
	if len(crtContent) == 0 {
		return fmt.Errorf("certificate file is empty")
	}

	logger.V(4).Info("Reading key from ETCD node", "node", node)

	keyContent, err := sshRunner.RunCommand(ctx, node, buildBRReadTmpKeyCmd(remoteTempDir))
	if err != nil {
		return fmt.Errorf("read key file: %v", err)
	}
	if len(keyContent) == 0 {
		return fmt.Errorf("key file is empty")
	}

	destDir := filepath.Join(b.backup, tempLocalEtcdCertsDir)
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return fmt.Errorf("create local cert dir: %v", err)
	}

	crtPath := filepath.Join(destDir, "apiserver-etcd-client.crt")
	keyPath := filepath.Join(destDir, "apiserver-etcd-client.key")

	logger.V(4).Info("Writing certificates to:")
	logger.V(4).Info("Certificate", "path", crtPath)
	logger.V(4).Info("Key", "path", keyPath)

	if err := os.WriteFile(crtPath, []byte(crtContent), 0o600); err != nil {
		return fmt.Errorf("write certificate file: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte(keyContent), 0o600); err != nil {
		return fmt.Errorf("write key file: %v", err)
	}

	if _, err := sshRunner.RunCommand(ctx, node, buildBRSheltieCmd(
		buildBREtcdCleanupTmpCmd(remoteTempDir),
	)); err != nil {
		return fmt.Errorf("cleanup temporary files: %v", err)
	}

	logger.V(4).Info("Certificates copied successfully")
	logger.V(4).Info("Backup directory", "path", b.backup)
	logger.V(4).Info("Certificate path", "path", crtPath)
	logger.V(4).Info("Key path", "path", keyPath)

	if err := b.saveCertsToPersistentStorage(); err != nil {
		return fmt.Errorf("save certificates to persistent storage: %v", err)
	}

	return nil
}

// for renew control panel only.
func (b *BottlerocketRenewer) saveCertsToPersistentStorage() error {
	srcCrt := filepath.Join(b.backup, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt")
	srcKey := filepath.Join(b.backup, tempLocalEtcdCertsDir, "apiserver-etcd-client.key")

	destDir := filepath.Join(persistentCertDir, persistentEtcdCertDir)
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return fmt.Errorf("create persistent directory: %v", err)
	}

	destCrt := filepath.Join(destDir, "apiserver-etcd-client.crt")
	destKey := filepath.Join(destDir, "apiserver-etcd-client.key")

	if err := copyFile(srcCrt, destCrt); err != nil {
		return fmt.Errorf("copy certificate: %v", err)
	}
	if err := copyFile(srcKey, destKey); err != nil {
		return fmt.Errorf("copy key: %v", err)
	}

	return nil
}

func (b *BottlerocketRenewer) loadCertsFromPersistentStorage() error {
	srcDir := filepath.Join(persistentCertDir, persistentEtcdCertDir)
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("no etcd certificates found in persistent storage. Please run etcd certificate renewal first")
	}

	destDir := filepath.Join(b.backup, tempLocalEtcdCertsDir)
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return fmt.Errorf("create temporary directory: %v", err)
	}

	srcCrt := filepath.Join(srcDir, "apiserver-etcd-client.crt")
	srcKey := filepath.Join(srcDir, "apiserver-etcd-client.key")

	destCrt := filepath.Join(destDir, "apiserver-etcd-client.crt")
	destKey := filepath.Join(destDir, "apiserver-etcd-client.key")

	if err := copyFile(srcCrt, destCrt); err != nil {
		return fmt.Errorf("copy certificate: %v", err)
	}
	if err := copyFile(srcKey, destKey); err != nil {
		return fmt.Errorf("copy key: %v", err)
	}

	return nil
}

func copyFile(src, dest string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if err = os.WriteFile(dest, input, 0o600); err != nil {
		return err
	}

	return nil
}

func buildBRSheltieCmd(commands ...string) string {
	script := strings.Join(commands, "\n")

	fullCommand := fmt.Sprintf("sudo sheltie << 'EOF'\nset -euo pipefail\n%s\nEOF", script)
	return fullCommand
}
