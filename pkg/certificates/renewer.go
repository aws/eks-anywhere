package certificates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	tempLocalEtcdCertsDir = "etcd-client-certs"
	backupDirTimeFormat   = "2006-01-02T15_04_05"
	backupDirStr          = "certificate_backup_"
)

// Renewer handles the certificate renewal process for EKS Anywhere clusters.
type Renewer struct {
	BackupDir       string
	Kubectl         kubernetes.Client
	SshEtcd         SSHRunner
	SshControlPlane SSHRunner
	Os              OSRenewer
}

// NewRenewer creates a new certificate renewer instance with a timestamped backup directory.
func NewRenewer(kubectl kubernetes.Client, osType string, cfg *RenewalConfig) (*Renewer, error) {
	ts := time.Now().Format(backupDirTimeFormat)
	backupDir := backupDirStr + ts

	if err := os.MkdirAll(filepath.Join(backupDir, tempLocalEtcdCertsDir), 0o755); err != nil {
		return nil, fmt.Errorf("creating backup directory: %v", err)
	}

	osRenewer := BuildOSRenewer(osType, backupDir)

	var sshEtcd SSHRunner
	if len(cfg.Etcd.Nodes) > 0 {
		var err error
		sshEtcd, err = NewSSHRunner(cfg.Etcd.SSH)
		if err != nil {
			return nil, fmt.Errorf("building etcd ssh client: %v", err)
		}
	}

	sshControlPlane, err := NewSSHRunner(cfg.ControlPlane.SSH)
	if err != nil {
		return nil, fmt.Errorf("building control plane ssh client: %v", err)
	}

	return &Renewer{
		BackupDir:       backupDir,
		Kubectl:         kubectl,
		Os:              osRenewer,
		SshEtcd:         sshEtcd,
		SshControlPlane: sshControlPlane,
	}, nil
}

// RenewCertificates orchestrates the certificate renewal process for the specified component.
func (r *Renewer) RenewCertificates(ctx context.Context, cfg *RenewalConfig, component string) error {
	processEtcd, processControlPlane, err := r.validateRenewalConfig(cfg, component)
	if err != nil {
		return err
	}

	if processEtcd {
		if err := r.renewEtcdCerts(ctx, cfg); err != nil {
			return err
		}
	}

	if processControlPlane {
		if err := r.renewControlPlaneCerts(ctx, cfg, component); err != nil {
			return err
		}
	}

	if processEtcd {
		if err := r.updateAPIServerEtcdClientSecret(ctx, cfg.ClusterName); err != nil {
			return err
		}
	}

	logger.MarkSuccess("Successfully renewed cluster certificates")
	r.cleanup()
	return nil
}

func (r *Renewer) renewEtcdCerts(ctx context.Context, cfg *RenewalConfig) error {
	for _, node := range cfg.Etcd.Nodes {
		if err := r.Os.RenewEtcdCerts(ctx, node, r.SshEtcd); err != nil {
			return fmt.Errorf("renewing certificates for etcd node %s: %v", node, err)
		}
	}

	firstNode := cfg.Etcd.Nodes[0]
	logger.V(4).Info("Copying certificates from node", "node", firstNode)

	if err := r.Os.CopyEtcdCerts(ctx, firstNode, r.SshEtcd); err != nil {
		return fmt.Errorf("copying certificates from etcd node %s: %v", firstNode, err)
	}

	logger.MarkPass("Etcd certificate renewal process completed successfully.")
	return nil
}

func (r *Renewer) renewControlPlaneCerts(ctx context.Context, cfg *RenewalConfig, component string) error {
	for _, node := range cfg.ControlPlane.Nodes {
		if err := r.Os.RenewControlPlaneCerts(ctx, node, cfg, component, r.SshControlPlane); err != nil {
			return fmt.Errorf("renewing certificates for control-plane node %s: %v", node, err)
		}
	}

	logger.MarkPass("Control plane certificate renewal process completed successfully.")
	return nil
}

func (r *Renewer) updateAPIServerEtcdClientSecret(ctx context.Context, clusterName string) error {
	crtPath := filepath.Join(r.BackupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt")
	keyPath := filepath.Join(r.BackupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.key")
	crtData, err := os.ReadFile(crtPath)
	if err != nil {
		return fmt.Errorf("read certificate file: %v", err)
	}
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("read key file: %v", err)
	}

	secretName := fmt.Sprintf("%s-apiserver-etcd-client", clusterName)
	secret := &corev1.Secret{}
	err = r.Kubectl.Get(ctx, secretName, constants.EksaSystemNamespace, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(5).Info("secret not found, please manually update the secret", "name", secretName)
			return nil
		}
		logger.V(5).Info("cannot access Kubernetes API, please manually update the secret", "error", err)
		return nil
	}
	secret.Data["tls.crt"] = crtData
	secret.Data["tls.key"] = keyData
	if err = r.Kubectl.Update(ctx, secret); err != nil {
		return fmt.Errorf("updating secret %s: %v", secretName, err)
	}

	logger.V(4).Info("Successfully updated secret", "name", secretName)
	return nil
}

func (r *Renewer) cleanup() {
	logger.V(4).Info("Cleaning up backup directory", "path", r.BackupDir)

	if err := os.RemoveAll(r.BackupDir); err != nil {
		logger.Error(err, "cleaning up certificate backup directory, please cleanup manually", "path", r.BackupDir)
	} else {
		logger.V(4).Info("Successfully cleaned up backup directory", "path", r.BackupDir)
	}
}

func (r *Renewer) validateRenewalConfig(
	cfg *RenewalConfig,
	_ string,
) (processEtcd, processControlPlane bool, err error) {

	processEtcd = len(cfg.Etcd.Nodes) > 0
	processControlPlane = true

	return processEtcd, processControlPlane, nil
}
