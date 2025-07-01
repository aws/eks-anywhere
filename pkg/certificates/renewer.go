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
	backupDir       string
	kubectl         kubernetes.Client
	sshEtcd         SSHRunner
	sshControlPlane SSHRunner
	os              OSRenewer
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
		backupDir:       backupDir,
		kubectl:         kubectl,
		os:              osRenewer,
		sshEtcd:         sshEtcd,
		sshControlPlane: sshControlPlane,
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

	logger.MarkSuccess("Successfully renewed cluster certificates")
	r.cleanup()
	return nil
}

func (r *Renewer) renewEtcdCerts(ctx context.Context, cfg *RenewalConfig) error {
	for _, node := range cfg.Etcd.Nodes {
		if err := r.os.RenewEtcdCerts(ctx, node, r.sshEtcd); err != nil {
			return fmt.Errorf("renewing certificates for etcd node %s: %v", node, err)
		}
	}

	firstNode := cfg.Etcd.Nodes[0]
	logger.V(4).Info("Copying certificates from node", "node", firstNode)

	if err := r.os.CopyEtcdCerts(ctx, firstNode, r.sshEtcd); err != nil {
		return fmt.Errorf("copying certificates from etcd node %s: %v", firstNode, err)
	}

	if err := r.updateAPIServerEtcdClientSecret(ctx, cfg.ClusterName); err != nil {
		return err
	}

	logger.MarkPass("Etcd certificate renewal process completed successfully.")
	return nil
}

func (r *Renewer) renewControlPlaneCerts(ctx context.Context, cfg *RenewalConfig, component string) error {
	for _, node := range cfg.ControlPlane.Nodes {
		if err := r.os.RenewControlPlaneCerts(ctx, node, cfg, component, r.sshControlPlane); err != nil {
			return fmt.Errorf("renewing certificates for control-plane node %s: %v", node, err)
		}
	}

	logger.MarkPass("Control plane certificate renewal process completed successfully.")
	return nil
}

func (r *Renewer) updateAPIServerEtcdClientSecret(ctx context.Context, clusterName string) error {

	crtPath := filepath.Join(r.backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt")
	keyPath := filepath.Join(r.backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.key")
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
	err = r.kubectl.Get(ctx, secretName, constants.EksaSystemNamespace, secret)
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
	if err = r.kubectl.Update(ctx, secret); err != nil {
		return fmt.Errorf("failed to update secret %s: %v", secretName, err)
	}

	logger.V(4).Info("Successfully updated secret", "name", secretName)
	return nil
}

func (r *Renewer) cleanup() {
	logger.V(4).Info("Cleaning up backup directory", "path", r.backupDir)

	if err := os.RemoveAll(r.backupDir); err != nil {
		logger.Error(err, "cleaning up certificate backup directory, please cleanup manually", "path", r.backupDir)
	} else {
		logger.V(4).Info("Successfully cleaned up backup directory", "path", r.backupDir)
	}
}

func (r *Renewer) validateRenewalConfig(
	cfg *RenewalConfig,
	component string,
) (processEtcd, processControlPlane bool, err error) {
	processEtcd = shouldProcessComponent(component, constants.EtcdComponent) &&
		len(cfg.Etcd.Nodes) > 0
	processControlPlane = shouldProcessComponent(component, constants.ControlPlaneComponent)

	return processEtcd, processControlPlane, nil
}
