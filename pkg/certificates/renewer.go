package certificates

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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
)

// Renewer handles the certificate renewal process for EKS Anywhere clusters.
type Renewer struct {
	backupDir       string
	kube            kubernetes.Client
	sshEtcd         SSHRunner
	sshControlPlane SSHRunner
	os              OSRenewer
}

// NewRenewer creates a new certificate renewer instance with a timestamped backup directory.
func NewRenewer(kube kubernetes.Client, osRenewer OSRenewer, cfg *RenewalConfig) (*Renewer, error) {
	ts := time.Now().Format(backupDirTimeFormat)
	backupDir := "certificate_backup_" + ts

	if err := os.MkdirAll(filepath.Join(backupDir, tempLocalEtcdCertsDir), 0o755); err != nil {
		return nil, fmt.Errorf("creating backup directory: %v", err)
	}

	// build sshRunner inside NewRenewer
	var sshEtcd SSHRunner
	if len(cfg.Etcd.Nodes) > 0 {
		var err error
		sshEtcd, err = createSSHRunner(cfg.Etcd.SSH)
		if err != nil {
			return nil, err
		}
	}

	sshControlPlane, err := createSSHRunner(cfg.ControlPlane.SSH)
	if err != nil {
		return nil, err
	}

	return &Renewer{
		backupDir:       backupDir,
		kube:            kube,
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

	return r.finishRenewal()
}

func (r *Renewer) renewEtcdCerts(ctx context.Context, cfg *RenewalConfig) error {
	logger.MarkPass("Starting etcd certificate renewal process")

	for _, node := range cfg.Etcd.Nodes {
		if err := r.os.RenewEtcdCerts(ctx, node, r.sshEtcd, r.backupDir); err != nil {
			return fmt.Errorf("renewing certificates for etcd node %s: %v", node, err)
		}
	}

	if err := r.updateAPIServerEtcdClientSecret(ctx, cfg.ClusterName); err != nil {
		return err
	}

	logger.MarkSuccess("Etcd certificate renewal process completed successfully.")
	return nil
}

func (r *Renewer) renewControlPlaneCerts(ctx context.Context, cfg *RenewalConfig, component string) error {
	logger.MarkPass("Starting control plane certificate renewal process")

	for _, node := range cfg.ControlPlane.Nodes {
		if err := r.os.RenewControlPlaneCerts(ctx, node, cfg, component, r.sshControlPlane, r.backupDir); err != nil {
			return fmt.Errorf("renewing certificates for control-plane node %s: %v", node, err)
		}
	}

	logger.MarkSuccess("Control plane certificate renewal process completed successfully.")
	return nil
}

func (r *Renewer) updateAPIServerEtcdClientSecret(ctx context.Context, clusterName string) error {
	logger.MarkPass("Updating apiserver-etcd-client secret", "cluster", clusterName)

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
	err = r.kube.Get(ctx, secretName, constants.EksaSystemNamespace, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(2).Info("Secret not found—skipping creation", "name", secretName)
			return nil
		}
		logger.V(2).Info("Failed to access Kubernetes API, skipping secret update", "error", err)
		return nil
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data["tls.crt"] = crtData
	secret.Data["tls.key"] = keyData
	if err = r.kube.Update(ctx, secret); err != nil {
		logger.V(2).Info("Failed to update secret, skipping", "error", err)
		return nil
	}

	logger.V(2).Info("Successfully updated secret", "name", secretName)
	return nil
}

func (r *Renewer) finishRenewal() error {
	logger.MarkPass("Cleaning up temporary files")
	return r.cleanup()
}

func (r *Renewer) cleanup() error {
	logger.V(2).Info("Cleaning up directory", "path", r.backupDir)
	chmodCmd := exec.Command("chmod", "-R", "u+w", r.backupDir)
	if err := chmodCmd.Run(); err != nil {
		return fmt.Errorf("changing permissions: %v", err)
	}
	return os.RemoveAll(r.backupDir)
}

func (r *Renewer) validateRenewalConfig(
	cfg *RenewalConfig,
	component string,
) (processEtcd, processControlPlane bool, err error) {
	processEtcd = ShouldProcessComponent(component, constants.EtcdComponent) &&
		len(cfg.Etcd.Nodes) > 0
	processControlPlane = ShouldProcessComponent(component, constants.ControlPlaneComponent)

	return processEtcd, processControlPlane, nil
}

// createSSHRunner creates a new SSH runner with environment variable handling.
func createSSHRunner(sshCfg SSHConfig) (SSHRunner, error) {
	runner, err := NewSSHRunner(sshCfg)
	if err != nil {
		return nil, err
	}
	return runner, nil
}
