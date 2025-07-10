package certificates

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	brEtcdCertDir           = "/var/lib/etcd"
	brControlPlaneCertDir   = "/var/lib/kubeadm/pki"
	brControlPlaneManifests = "/var/lib/kubeadm/manifests"
	brTempDir               = "/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp"
)

// BottlerocketRenewer implements OSRenewer for Bottlerocket systems.
type BottlerocketRenewer struct {
	osType OSType
	backup string
}

// NewBottlerocketRenewer creates a new BottlerocketRenewer.
func NewBottlerocketRenewer(backupDir string) *BottlerocketRenewer {
	return &BottlerocketRenewer{
		osType: OSTypeBottlerocket,
		backup: backupDir,
	}
}

// RenewControlPlaneCerts renews certificates for control plane nodes.
func (b *BottlerocketRenewer) RenewControlPlaneCerts(
	ctx context.Context,
	node string,
	cfg *RenewalConfig,
	component string,
	ssh SSHRunner,
) error {
	logger.V(0).Info("Processing control-plane node", "node", node)

	hasExternalEtcd := cfg != nil && len(cfg.Etcd.Nodes) > 0

	if hasExternalEtcd {
		if err := b.transferCertsToControlPlane(ctx, node, ssh); err != nil {
			return fmt.Errorf("transferring certificates to control plane node: %v", err)
		}
	}

	sessionCmds := buildBRSheltieCmd(
		buildBRImagePullCmd(),
		buildBRControlPlaneBackupCertsCmd(component, hasExternalEtcd, b.backup, brControlPlaneCertDir),
		buildBRControlPlaneRenewCertsCmd(),
		buildBRControlPlaneCheckCertsCmd(),
		buildBRControlPlaneCopyCertsFromTmpCmd(),
		buildBRControlPlaneRestartPodsCmd(),
	)

	if _, err := ssh.RunCommand(ctx, node, sessionCmds); err != nil {
		return fmt.Errorf("renewing control plane certificates: %v", err)
	}

	logger.V(0).Info("Renewed control-plane certificates", "node", node)
	return nil
}

func (b *BottlerocketRenewer) transferCertsToControlPlane(
	ctx context.Context, node string, ssh SSHRunner,
) error {
	logger.V(4).Info("Transferring certificates to control-plane node", "node", node)

	crtB, err := os.ReadFile(filepath.Join(
		b.backup, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt"))
	if err != nil {
		return fmt.Errorf("reading certificate file: %v", err)
	}
	keyB, err := os.ReadFile(filepath.Join(
		b.backup, tempLocalEtcdCertsDir, "apiserver-etcd-client.key"))
	if err != nil {
		return fmt.Errorf("reading key file: %v", err)
	}

	sessionCmds := buildBRSheltieCmd(
		buildBRCreateTmpDirCmd(tempLocalEtcdCertsDir),
		buildBRWriteCertToTmpCmd(base64.StdEncoding.EncodeToString(crtB)),
		buildBRWriteKeyToTmpCmd(base64.StdEncoding.EncodeToString(keyB)),
		buildBRSetTmpCertPermissionsCmd(),
	)

	if _, err := ssh.RunCommand(ctx, node, sessionCmds); err != nil {
		return fmt.Errorf("transfering certificates: %v", err)
	}

	logger.V(4).Info("Certificates transferred", "node", node)
	return nil
}

// RenewEtcdCerts renews etcd certificates on a Bottlerocket node.
func (b *BottlerocketRenewer) RenewEtcdCerts(ctx context.Context, node string, ssh SSHRunner) error {
	logger.V(0).Info("Processing etcd node", "os", b.osType, "node", node)

	remoteTempDir := brTempDir

	if _, err := ssh.RunCommand(ctx, node, buildBRSheltieCmd(
		buildBRImagePullCmd(),
		buildBREtcdBackupCertsCmd(b.backup),
		buildBREtcdRenewCertsCmd(),
	)); err != nil {
		return fmt.Errorf("renewing certificates: %v", err)
	}

	if _, err := ssh.RunCommand(ctx, node, buildBRSheltieCmd(
		buildBREtcdRenewChecksCmd(),
	)); err != nil {
		return fmt.Errorf("validating etcd certificates: %v", err)
	}

	if _, err := ssh.RunCommand(ctx, node, buildBRSheltieCmd(
		buildBREtcdCopyCertsToTmpCmd(remoteTempDir),
	)); err != nil {
		return fmt.Errorf("copying certificates to tmp: %v", err)
	}

	if _, err := ssh.RunCommand(ctx, node, buildBRSheltieCmd(
		buildBREtcdCleanupTmpCmd(remoteTempDir),
	)); err != nil {
		return fmt.Errorf("cleanup temporary files: %v", err)
	}

	logger.Info("Renewed certificates for etcd node", "node", node)

	return nil
}

func (b *BottlerocketRenewer) CopyEtcdCerts(ctx context.Context, node string, ssh SSHRunner) error {
	logger.V(4).Info("Reading certificate from ETCD node", "node", node)
	logger.V(4).Info("Using backup directory", "path", b.backup)

	remoteTempDir := brTempDir

	if _, err := ssh.RunCommand(ctx, node, buildBRSheltieCmd(
		buildBREtcdCopyCertsToTmpCmd(remoteTempDir),
	)); err != nil {
		return fmt.Errorf("copying certificates to tmp: %v", err)
	}

	if _, err := ssh.RunCommand(ctx, node, buildBRListTmpFilesCmd(remoteTempDir)); err != nil {
		return fmt.Errorf("listing certificate files: %v", err)
	}

	crtContent, err := ssh.RunCommand(ctx, node, buildBRReadTmpCertCmd(remoteTempDir))
	if err != nil {
		return fmt.Errorf("reading certificate file: %v", err)
	}

	if len(crtContent) == 0 {
		return fmt.Errorf("certificate file is empty")
	}

	logger.V(4).Info("Reading key from ETCD node", "node", node)

	keyContent, err := ssh.RunCommand(ctx, node, buildBRReadTmpKeyCmd(remoteTempDir))
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

	if _, err := ssh.RunCommand(ctx, node, buildBRSheltieCmd(
		buildBREtcdCleanupTmpCmd(remoteTempDir),
	)); err != nil {
		return fmt.Errorf("cleanup temporary files: %v", err)
	}

	logger.V(4).Info("Certificates copied successfully")
	logger.V(4).Info("Backup directory", "path", b.backup)
	logger.V(4).Info("Certificate path", "path", crtPath)
	logger.V(4).Info("Key path", "path", keyPath)

	return nil
}

func buildBRSheltieCmd(commands ...string) string {
	script := strings.Join(commands, "\n")

	fullCommand := fmt.Sprintf("sudo sheltie << 'EOF'\nset -euo pipefail\n%s\nEOF", script)
	return fullCommand
}

func buildBRImagePullCmd() string {
	return `IMAGE_ID=$(apiclient get | apiclient exec admin jq -r '.settings["host-containers"]["kubeadm-bootstrap"].source')
ctr image pull ${IMAGE_ID}`
}

func buildBRControlPlaneBackupCertsCmd(_ string, hasExternalEtcd bool, backupDir, certDir string) string {
	var script string
	if hasExternalEtcd {
		script = fmt.Sprintf(`mkdir -p '/etc/kubernetes/pki.bak_%[1]s'
cp -r %[2]s/* '/etc/kubernetes/pki.bak_%[1]s/'
rm -rf '/etc/kubernetes/pki.bak_%[1]s/etcd'`, backupDir, certDir)
	} else {
		script = fmt.Sprintf("cp -r '%s' '/etc/kubernetes/pki.bak_%s'", certDir, backupDir)
	}
	return script
}

func buildBRControlPlaneRenewCertsCmd() string {
	script := `ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs renew all`
	return script
}

func buildBRControlPlaneCheckCertsCmd() string {
	script := `ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs check-expiration`
	return script
}

func buildBRControlPlaneCopyCertsFromTmpCmd() string {
	script := `if [ -d "/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs" ]; then
    cp /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs/apiserver-etcd-client.crt /var/lib/kubeadm/pki/server-etcd-client.crt || exit 1
    cp /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs/apiserver-etcd-client.key /var/lib/kubeadm/pki/apiserver-etcd-client.key || exit 1
    chmod 600 /var/lib/kubeadm/pki/server-etcd-client.crt || exit 1
    chmod 600 /var/lib/kubeadm/pki/apiserver-etcd-client.key || exit 1
    rm -rf /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs
fi`
	return script
}

func buildBRControlPlaneRestartPodsCmd() string {
	script := `
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=false
sleep 10
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=true
`
	return script
}

func buildBREtcdBackupCertsCmd(backupDir string) string {
	script := fmt.Sprintf(`cp -r /var/lib/etcd/pki /var/lib/etcd/pki.bak_%[1]s
rm /var/lib/etcd/pki/*
cp /var/lib/etcd/pki.bak_%[1]s/ca.* /var/lib/etcd/pki`, backupDir)
	return script
}

func buildBREtcdRenewCertsCmd() string {
	script := `ctr run \
--mount type=bind,src=/var/lib/etcd/pki,dst=/etc/etcd/pki,options=rbind:rw \
--net-host \
--rm \
${IMAGE_ID} tmp-cert-renew \
/opt/bin/etcdadm join phase certificates http://eks-a-etcd-dumb-url --init-system kubelet`
	return script
}

func buildBREtcdRenewChecksCmd() string {
	script := `ETCD_CONTAINER_ID=$(ctr -n k8s.io c ls | grep -w "etcd-io" | cut -d " " -f1 | tail -1)
ctr -n k8s.io t exec --exec-id etcd ${ETCD_CONTAINER_ID} etcdctl \
     --cacert=/var/lib/etcd/pki/ca.crt \
     --cert=/var/lib/etcd/pki/server.crt \
     --key=/var/lib/etcd/pki/server.key \
     member list`
	return script
}

func buildBREtcdCopyCertsToTmpCmd(tempDir string) string {
	script := fmt.Sprintf(`cp /var/lib/etcd/pki/apiserver-etcd-client.* %[1]s/ || exit 1
chmod 600 %[1]s/apiserver-etcd-client.crt || exit 1
chmod 600 %[1]s/apiserver-etcd-client.key || exit 1`, tempDir)
	return script
}

func buildBREtcdCleanupTmpCmd(tempDir string) string {
	script := fmt.Sprintf(`rm -f %s/apiserver-etcd-client.*`, tempDir)
	return script
}

func buildBRCreateTmpDirCmd(dirName string) string {
	script := fmt.Sprintf(`TARGET_DIR="/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/%[1]s"
mkdir -p "${TARGET_DIR}" || exit 1
chmod 755 "${TARGET_DIR}" || exit 1`, dirName)
	return script
}

func buildBRWriteCertToTmpCmd(certBase64 string) string {
	script := fmt.Sprintf(`cat <<'CRT_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.crt"
%s
CRT_END
[ $? -eq 0 ] || exit 1`, certBase64)
	return script
}

func buildBRWriteKeyToTmpCmd(keyBase64 string) string {
	script := fmt.Sprintf(`cat <<'KEY_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.key"
%s
KEY_END
[ $? -eq 0 ] || exit 1`, keyBase64)
	return script
}

func buildBRSetTmpCertPermissionsCmd() string {
	script := `chmod 600 "${TARGET_DIR}/apiserver-etcd-client.crt" || exit 1
chmod 600 "${TARGET_DIR}/apiserver-etcd-client.key" || exit 1`
	return script
}

func buildBRListTmpFilesCmd(tempDir string) string {
	script := fmt.Sprintf(`sudo sheltie << 'EOF'
ls -l %s/apiserver-etcd-client.*
exit
EOF`, tempDir)
	return script
}

func buildBRReadTmpCertCmd(tempDir string) string {
	script := fmt.Sprintf(`sudo sheltie << 'EOF'
cat %s/apiserver-etcd-client.crt
exit
EOF`, tempDir)
	return script
}

func buildBRReadTmpKeyCmd(tempDir string) string {
	script := fmt.Sprintf(`sudo sheltie << 'EOF'
cat %s/apiserver-etcd-client.key
exit
EOF`, tempDir)
	return script
}
