package certificates

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const tempLocalEtcdCertsDir = "etcd-client-certs"

// Renewer handles the certificate renewal process for EKS Anywhere clusters.
type Renewer struct {
	backupDir string
	kube      kubernetes.Client
	ssh       SSHRunner
	os        OSRenewer
}

// NewRenewer creates a new certificate renewer instance with a timestamped backup directory.
func NewRenewer(kube kubernetes.Client, sshRunner SSHRunner, osRenewer OSRenewer) (*Renewer, error) {
	ts := time.Now().Format("20060102_150405")
	backupDir := "certificate_backup_" + ts

	if err := os.MkdirAll(filepath.Join(backupDir, tempLocalEtcdCertsDir), 0o755); err != nil {
		return nil, fmt.Errorf("creating backup directory: %v", err)
	}
	return &Renewer{
		backupDir: backupDir,
		kube:      kube,
		ssh:       sshRunner,
		os:        osRenewer,
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
		if err := r.os.RenewEtcdCerts(ctx, node, r.ssh, r.backupDir); err != nil {
			return fmt.Errorf("renewing certificates for etcd node %s: %v", node, err)
		}
	}

	if err := r.updateAPIServerEtcdClientSecret(ctx, cfg.ClusterName); err != nil {
		logger.MarkWarning("Failed to update apiserver-etcd-client secret", "error", err)
		logger.Info("You may need to manually update the secret after the API server is reachable")
		logger.Info("Use kubectl edit secret to update the secret", "command", fmt.Sprintf("kubectl edit secret %s-apiserver-etcd-client -n eksa-system", cfg.ClusterName))

	}

	logger.MarkSuccess("Etcd certificate renewal process completed successfully.")
	return nil
}

func (r *Renewer) renewControlPlaneCerts(ctx context.Context, cfg *RenewalConfig, component string) error {
	logger.MarkPass("Starting control plane certificate renewal process")

	for _, node := range cfg.ControlPlane.Nodes {
		if err := r.os.RenewControlPlaneCerts(ctx, node, cfg, component, r.ssh, r.backupDir); err != nil {
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
		return fmt.Errorf("failed to read certificate file: %v", err)
	}
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key file: %v", err)
	}

	crtBase64 := base64.StdEncoding.EncodeToString(crtData)
	keyBase64 := base64.StdEncoding.EncodeToString(keyData)

	secretName := fmt.Sprintf("%s-apiserver-etcd-client", clusterName)

	patchData := fmt.Sprintf(`{"data":{"tls.crt":"%s","tls.key":"%s"}}`, crtBase64, keyBase64)

	cmd := exec.Command("kubectl", "patch", "secret", secretName,
		"-n", constants.EksaSystemNamespace,
		"--type=merge",
		"-p", patchData)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "NotFound") {
			createCmd := exec.Command("kubectl", "create", "secret", "tls",
				secretName,
				"-n", constants.EksaSystemNamespace,
				"--cert", crtPath,
				"--key", keyPath)

			if output, err := createCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to create secret %s: %v, output: %s", secretName, err, string(output))
			}
		} else {
			return fmt.Errorf("failed to update secret %s: %v, output: %s", secretName, err, string(output))
		}
	}

	logger.V(2).Info("Successfully updated secret", "name", secretName)
	return nil
}

func (r *Renewer) ensureNamespaceExists(ctx context.Context, namespace string) error {
	ns := &corev1.Namespace{}
	err := r.kube.Get(ctx, namespace, "", ns)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("checking namespace %s: %v", namespace, err)
		}
		ns.Name = namespace
		if err = r.kube.Create(ctx, ns); err != nil {
			return fmt.Errorf("create namespace %s: %v", namespace, err)
		}
		logger.Info("Created namespace", "name", namespace)
	}
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

func (r *Renewer) validateRenewalConfig(cfg *RenewalConfig, component string) (processEtcd, processControlPlane bool, err error) {
	processEtcd = ShouldProcessComponent(component, constants.EtcdComponent) && len(cfg.Etcd.Nodes) > 0
	processControlPlane = ShouldProcessComponent(component, constants.ControlPlaneComponent)

	if processEtcd {
		if err := r.ssh.InitSSHConfig(cfg.Etcd.SSH); err != nil {
			return false, false, fmt.Errorf("initializing SSH config for etcd: %v", err)
		}
	}

	if processControlPlane {
		if err := r.ssh.InitSSHConfig(cfg.ControlPlane.SSH); err != nil {
			return false, false, fmt.Errorf("initializing SSH config for control-plane: %v", err)
		}
	}

	return processEtcd, processControlPlane, nil
}
