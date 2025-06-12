package certificates

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/certificates/bottlerocket"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	persistentCertDir     = "/var/lib/eks-anywhere/certificates"
	persistentEtcdCertDir = "etcd-certs"
)

// BottlerocketRenewer implements OSRenewer for Bottlerocket systems.
type BottlerocketRenewer struct {
	certPaths CertificatePaths
	osType    string
}

// NewBottlerocketRenewer creates a new BottlerocketRenewer.
func NewBottlerocketRenewer(certPaths CertificatePaths) *BottlerocketRenewer {
	return &BottlerocketRenewer{
		certPaths: certPaths,
		osType:    string(OSTypeBottlerocket),
	}
}

// RenewControlPlaneCerts renews control plane certificates on a Bottlerocket node.
func (b *BottlerocketRenewer) RenewControlPlaneCerts(ctx context.Context, node string, config *RenewalConfig, component string, sshRunner SSHRunner, backupDir string) error {
	logger.V(2).Info("Processing control plane node", "node", node)

	// for renew control panel only.
	if component == constants.ControlPlaneComponent && len(config.Etcd.Nodes) > 0 {
		if err := b.loadCertsFromPersistentStorage(backupDir); err != nil {
			return fmt.Errorf("loading certificates from persistent storage: %v", err)
		}
	}

	// If we have external etcd nodes, first transfer certificates to the node
	if len(config.Etcd.Nodes) > 0 {
		if err := b.transferCertsToControlPlane(ctx, node, sshRunner, backupDir); err != nil {
			return fmt.Errorf("transfer certificates to control plane node: %v", err)
		}
	}

	builder := bottlerocket.NewControlPlaneCommandBuilder(
		backupDir,
		bottlerocketControlPlaneCertDir,
		component,
		len(config.Etcd.Nodes) > 0,
	)
	commands := builder.BuildCommands()
	session := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\nEOF",
		commands.ShelliePrefix,
		commands.BackupCerts,
		commands.ImagePull,
		commands.RenewCerts,
		commands.CopyCerts,
		commands.RestartPods)

	if err := sshRunner.RunCommand(ctx, node, session); err != nil {
		return fmt.Errorf("renew control panel node certificates: %v", err)
	}

	if VerbosityLevel >= 1 {
		b.checkCertificates(ctx, node, sshRunner, commands)
	}

	logger.MarkPass("Renewed certificates for control plane node", "node", node)
	return nil
}

func (b *BottlerocketRenewer) transferCertsToControlPlane(ctx context.Context, node string, sshRunner SSHRunner, backupDir string) error {
	logger.V(2).Info("Transferring certificates to control plane node", "node", node)

	srcCrt := filepath.Join(backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt")
	crtContent, err := os.ReadFile(srcCrt)
	if err != nil {
		return fmt.Errorf("read certificate file: %v", err)
	}

	srcKey := filepath.Join(backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.key")
	keyContent, err := os.ReadFile(srcKey)
	if err != nil {
		return fmt.Errorf("read key file: %v", err)
	}

	crtBase64 := base64.StdEncoding.EncodeToString(crtContent)
	keyBase64 := base64.StdEncoding.EncodeToString(keyContent)

	builder := bottlerocket.NewCertTransferBuilder(tempLocalEtcdCertsDir, crtBase64, keyBase64)
	commands := builder.BuildCommands()
	session := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\nEOF",
		commands.ShelliePrefix,
		commands.CreateDir,
		commands.WriteCertificate,
		commands.WriteKey,
		commands.SetPermissions)

	if err := sshRunner.RunCommand(ctx, node, session); err != nil {
		return fmt.Errorf("transfer certificates: %v", err)
	}

	logger.V(2).Info("External certificates transferred to control plane node", "node", node)
	return nil
}

// RenewEtcdCerts renews etcd certificates on a Bottlerocket node.
func (b *BottlerocketRenewer) RenewEtcdCerts(ctx context.Context, node string, sshRunner SSHRunner, backupDir string) error {
	logger.V(2).Info("Processing etcd node", "node", node)

	builder := bottlerocket.NewEtcdCommandBuilder(backupDir, bottlerocketTmpDir)
	commands := builder.BuildCommands()

	// first session: backup and renew certificates
	firstSession := fmt.Sprintf("%s\n%s\n%s\n%s\nEOF",
		commands.ShelliePrefix,
		commands.ImagePull,
		commands.BackupCerts,
		commands.RenewCerts)

	if err := sshRunner.RunCommand(ctx, node, firstSession); err != nil {
		return fmt.Errorf("renew certificates: %v", err)
	}

	// second sheltie session for copying certs
	secondSession := fmt.Sprintf("%s\n%s\nEOF",
		commands.ShelliePrefix,
		commands.CopyCerts)

	if err := sshRunner.RunCommand(ctx, node, secondSession); err != nil {
		return fmt.Errorf("copy certificates2 to tmp: %v", err)
	}

	// copy certificates to local
	logger.V(2).Info("Copying certificates from node", "node", node)

	if err := b.copyEtcdCerts(ctx, node, sshRunner, backupDir); err != nil {
		return fmt.Errorf("copy certificates3: %v", err)
	}

	// third sheltie session for cleanup
	thirdSession := fmt.Sprintf("%s\n%s\nEOF",
		commands.ShelliePrefix,
		commands.Cleanup)

	if err := sshRunner.RunCommand(ctx, node, thirdSession); err != nil {
		return fmt.Errorf("cleanup temporary files: %v", err)
	}

	logger.MarkPass("Renewed certificates for etcd node", "node", node)

	// save etcd cert for control panel renew
	if err := b.saveCertsToPersistentStorage(backupDir); err != nil {
		return fmt.Errorf("save certificates to persistent storage: %v", err)
	}

	return nil
}

func (b *BottlerocketRenewer) copyEtcdCerts(ctx context.Context, node string, sshRunner SSHRunner, backupDir string) error {

	logger.V(2).Info("Reading certificate from ETCD node", "node", node)
	logger.V(2).Info("Using backup directory", "path", backupDir)

	builder := bottlerocket.NewCertReadBuilder(bottlerocketTmpDir)
	commands := builder.BuildCommands()

	if err := sshRunner.RunCommand(ctx, node, commands.ListFiles); err != nil {
		return fmt.Errorf("list certificate files: %v", err)
	}

	crtContent, err := sshRunner.RunCommandWithOutput(ctx, node, commands.ReadCert)
	if err != nil {
		return fmt.Errorf("read certificate file: %v", err)
	}

	if len(crtContent) == 0 {
		return fmt.Errorf("certificate file is empty")
	}

	logger.V(2).Info("Reading key from ETCD node", "node", node)

	keyContent, err := sshRunner.RunCommandWithOutput(ctx, node, commands.ReadKey)
	if err != nil {
		return fmt.Errorf("read key file: %v", err)
	}

	if len(keyContent) == 0 {
		return fmt.Errorf("key file is empty")
	}

	crtPath := filepath.Join(backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt")
	keyPath := filepath.Join(backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.key")

	logger.V(2).Info("Writing certificates to:")
	logger.V(2).Info("Certificate", "path", crtPath)
	logger.V(2).Info("Key", "path", keyPath)

	if err := os.WriteFile(crtPath, []byte(crtContent), 0o600); err != nil {
		return fmt.Errorf("write certificate file: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte(keyContent), 0o600); err != nil {
		return fmt.Errorf("write key file: %v", err)
	}

	logger.V(2).Info("Certificates copied successfully")
	logger.V(2).Info("Backup directory", "path", backupDir)
	logger.V(2).Info("Certificate path", "path", crtPath)
	logger.V(2).Info("Key path", "path", keyPath)

	return nil
}

// for renew control panel only.
func (b *BottlerocketRenewer) saveCertsToPersistentStorage(backupDir string) error {
	srcCrt := filepath.Join(backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt")
	srcKey := filepath.Join(backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.key")

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

func (b *BottlerocketRenewer) loadCertsFromPersistentStorage(backupDir string) error {
	srcDir := filepath.Join(persistentCertDir, persistentEtcdCertDir)
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("no etcd certificates found in persistent storage. Please run etcd certificate renewal first")
	}

	destDir := filepath.Join(backupDir, tempLocalEtcdCertsDir)
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

func (b *BottlerocketRenewer) checkCertificates(ctx context.Context, node string, sshRunner SSHRunner, commands *bottlerocket.ControlPlaneCommands) {
	checkSession := fmt.Sprintf("%s\n%s\n%s\nEOF", commands.ShelliePrefix, commands.ImagePull, commands.CheckCerts)
	output, err := sshRunner.RunCommandWithOutput(ctx, node, checkSession)
	if err != nil {
		logger.Info(fmt.Sprintf("Certificate check failed: %v", err), "node", node)
		if output != "" {
			logger.Info("Certificate check partial output:", "node", node)
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if line != "" {
					logger.Info("  " + line)
				}
			}
		}
	} else {
		logger.Info("Certificate check results:", "node", node)
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if line != "" {
				logger.Info("  " + line)
			}
		}
	}
}
