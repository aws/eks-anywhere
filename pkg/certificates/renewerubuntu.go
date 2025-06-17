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

// LinuxRenewer implements OSRenewer for Linux-based systems (Ubuntu and RHEL).
type LinuxRenewer struct {
	osType OSType
}

// NewLinuxRenewer creates a new renewer for Linux-based operating systems.
func NewLinuxRenewer() *LinuxRenewer { return &LinuxRenewer{osType: OSTypeLinux} }

// RenewControlPlaneCerts renews control plane certificates on a Linux node.
func (l *LinuxRenewer) RenewControlPlaneCerts(ctx context.Context, node string, cfg *RenewalConfig, component string, ssh SSHRunner, backupDir string) error {
	logger.V(2).Info(fmt.Sprintf("Processing node %s...", node))

	hasExternalEtcd := cfg != nil && len(cfg.Etcd.Nodes) > 0

	if err := ssh.RunCommand(ctx, node,
		buildCPBackupCmd(component, hasExternalEtcd, backupDir)); err != nil {
		return fmt.Errorf("backup certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node,
		buildCPRenewCmd(component, hasExternalEtcd)); err != nil {
		return fmt.Errorf("renew certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node, []string{"sudo kubeadm certs check-expiration"}); err != nil {
		return fmt.Errorf("validate certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node, buildCPRestartCmd()); err != nil {
		return fmt.Errorf("restart pods: %v", err)
	}

	logger.MarkPass(fmt.Sprintf("Renewed certificates for node %s", node))
	return nil
}

// RenewEtcdCerts renews etcd certificates on a Linux node.
func (l *LinuxRenewer) RenewEtcdCerts(
	ctx context.Context, node string, ssh SSHRunner, backupDir string,
) error {
	logger.V(2).Info("Processing etcd node", "os", l.osType, "node", node)

	if err := ssh.RunCommand(ctx, node,
		buildEtcdBackupCmd(backupDir)); err != nil {
		return fmt.Errorf("backup certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node,
		[]string{"sudo etcdadm join phase certificates http://eks-a-etcd-dumb-url"}); err != nil {
		return fmt.Errorf("renew certs: %v", err)
	}
	if err := ssh.RunCommand(ctx, node,
		buildEtcdValidateCmd()); err != nil {
		return fmt.Errorf("validate certs: %v", err)
	}

	if err := l.copyEtcdCerts(ctx, node, ssh, backupDir); err != nil {
		return err
	}
	logger.MarkPass("Renewed etcd certificates", "node", node)
	return nil
}

func (l *LinuxRenewer) copyEtcdCerts(ctx context.Context, node string, ssh SSHRunner, backupDir string) error {
	// etcdDir := l.certPaths.EtcdCertDir
	etcdDir := linuxEtcdCertDir
	cat := func(file string) (string, error) {
		cmd := fmt.Sprintf("sudo cat %s/%s", etcdDir, file)
		return ssh.RunCommandWithOutput(ctx, node, []string{cmd})
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

func buildCPBackupCmd(component string, hasExternalEtcd bool, backup string) []string {
	if component == constants.ControlPlaneComponent && hasExternalEtcd {
		script := fmt.Sprintf(`
sudo mkdir -p '/etc/kubernetes/pki.bak_%[1]s'
cd %[2]s
for f in $(find . -type f ! -path './etcd/*'); do
    sudo mkdir -p $(dirname '/etc/kubernetes/pki.bak_%[1]s/'$f)
    sudo cp $f '/etc/kubernetes/pki.bak_%[1]s/'$f
done`, backup, linuxControlPlaneCertDir)
		return []string{script}
	}
	cmd := fmt.Sprintf("sudo cp -r '%s' '/etc/kubernetes/pki.bak_%s'",
		linuxControlPlaneCertDir, backup)
	return []string{cmd}
}

func buildCPRenewCmd(component string, hasExternalEtcd bool) []string {
	if component == constants.ControlPlaneComponent && hasExternalEtcd {
		// When only renewing control plane certs with external etcd,
		// we need to skip the etcd directory to preserve certificates
		script := `for cert in admin.conf apiserver apiserver-kubelet-client controller-manager.conf front-proxy-client scheduler.conf; do
            sudo kubeadm certs renew $cert
done`
		return []string{script}
	}
	return []string{"sudo kubeadm certs renew all"}
}

func buildCPRestartCmd() []string {
	cmd := fmt.Sprintf("sudo mkdir -p /tmp/manifests && "+
		"sudo mv %s/* /tmp/manifests/ && "+
		"sleep 20 && "+
		"sudo mv /tmp/manifests/* %s/",
		linuxControlPlaneManifests, linuxControlPlaneManifests)
	return []string{cmd}
}

func buildEtcdBackupCmd(backup string) []string {
	cmd := fmt.Sprintf("cd %s && sudo cp -r pki pki.bak_%s && sudo rm -rf pki/* && sudo cp pki.bak_%s/ca.* pki/",
		linuxEtcdCertDir, backup, backup)
	return []string{cmd}
}

func buildEtcdValidateCmd() []string {
	cmd := fmt.Sprintf("sudo etcdctl --cacert=%s/pki/ca.crt "+
		"--cert=%s/pki/etcdctl-etcd-client.crt "+
		"--key=%s/pki/etcdctl-etcd-client.key "+
		"endpoint health",
		linuxEtcdCertDir, linuxEtcdCertDir, linuxEtcdCertDir)
	return []string{cmd}
}
