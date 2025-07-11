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
	logger.V(0).Info("Renewing control-plane certificates", "node", node)

	hasExternalEtcd := cfg != nil && len(cfg.Etcd.Nodes) > 0

	if hasExternalEtcd {
		if err := b.transferCertsToControlPlane(ctx, node, ssh); err != nil {
			return fmt.Errorf("transferring certificates to control plane node: %v", err)
		}
	}

	sessionCmds := b.sheltie(
		b.pullContainerImage(),
		b.backupControlPlaneCerts(component, hasExternalEtcd, b.backup, brControlPlaneCertDir),
		b.renewControlPlaneCerts(),
		b.checkControlPlaneCerts(),
		b.copyExternalEtcdCerts(),
		b.restartControlPlaneStaticPods(),
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

	sessionCmds := b.sheltie(
		b.createTempDirectoryAndWriteCerts(
			tempLocalEtcdCertsDir,
			base64.StdEncoding.EncodeToString(crtB),
			base64.StdEncoding.EncodeToString(keyB),
		),
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

	if _, err := ssh.RunCommand(ctx, node, b.sheltie(
		b.pullContainerImage(),
		b.backupEtcdCerts(b.backup),
		b.renewEtcdCerts(),
	)); err != nil {
		return fmt.Errorf("renewing certificates: %v", err)
	}

	if _, err := ssh.RunCommand(ctx, node, b.sheltie(
		b.validateEtcdCerts(),
	)); err != nil {
		return fmt.Errorf("validating etcd certificates: %v", err)
	}

	if _, err := ssh.RunCommand(ctx, node, b.sheltie(
		b.copyEtcdCertsToTemp(remoteTempDir),
	)); err != nil {
		return fmt.Errorf("copying certificates to tmp: %v", err)
	}

	if _, err := ssh.RunCommand(ctx, node, b.sheltie(
		b.cleanupEtcdTempFiles(remoteTempDir),
	)); err != nil {
		return fmt.Errorf("cleanup temporary files: %v", err)
	}

	logger.Info("Renewed certificates for etcd node", "node", node)

	return nil
}

func (b *BottlerocketRenewer) CopyEtcdCerts(ctx context.Context, node string, ssh SSHRunner) error {

	remoteTempDir := brTempDir

	if _, err := ssh.RunCommand(ctx, node, b.sheltie(
		b.copyEtcdCertsToTemp(remoteTempDir),
	)); err != nil {
		return fmt.Errorf("copying certificates to tmp: %v", err)
	}

	crtPath := filepath.Join(remoteTempDir, "apiserver-etcd-client.crt")
	crtContent, err := ssh.RunCommand(ctx, node, b.readTempFile(crtPath))
	if err != nil {
		return fmt.Errorf("reading certificate file: %v", err)
	}

	if len(crtContent) == 0 {
		return fmt.Errorf("certificate file is empty")
	}

	logger.V(4).Info("Reading key from ETCD node", "node", node)

	keyPath := filepath.Join(remoteTempDir, "apiserver-etcd-client.key")
	keyContent, err := ssh.RunCommand(ctx, node, b.readTempFile(keyPath))
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

	localCrtPath := filepath.Join(destDir, "apiserver-etcd-client.crt")
	localKeyPath := filepath.Join(destDir, "apiserver-etcd-client.key")

	if err := os.WriteFile(localCrtPath, []byte(crtContent), 0o600); err != nil {
		return fmt.Errorf("write certificate file: %v", err)
	}
	if err := os.WriteFile(localKeyPath, []byte(keyContent), 0o600); err != nil {
		return fmt.Errorf("write key file: %v", err)
	}

	if _, err := ssh.RunCommand(ctx, node, b.sheltie(
		b.cleanupEtcdTempFiles(remoteTempDir),
	)); err != nil {
		return fmt.Errorf("cleanup temporary files: %v", err)
	}

	return nil
}

func (b *BottlerocketRenewer) sheltie(commands ...string) string {
	script := strings.Join(commands, "\n")

	fullCommand := fmt.Sprintf("sudo sheltie << 'EOF'\nset -euo pipefail\n%s\nEOF", script)
	return fullCommand
}

func (b *BottlerocketRenewer) pullContainerImage() string {
	return `IMAGE_ID=$(apiclient get | apiclient exec admin jq -r '.settings["host-containers"]["kubeadm-bootstrap"].source')
ctr image pull ${IMAGE_ID}`
}

func (b *BottlerocketRenewer) backupControlPlaneCerts(_ string, hasExternalEtcd bool, backupDir, certDir string) string {

	backupPath := fmt.Sprintf("/var/lib/kubeadm/pki.bak_%s", backupDir)

	if hasExternalEtcd {
		return fmt.Sprintf("mkdir -p '%s' && cp -r %s/* '%s/' && rm -rf '%s/etcd'",
			backupPath, certDir, backupPath, backupPath)
	}

	return fmt.Sprintf("cp -r '%s' '%s'", certDir, backupPath)
}

func (b *BottlerocketRenewer) renewControlPlaneCerts() string {
	script := `ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs renew all`
	return script
}

func (b *BottlerocketRenewer) checkControlPlaneCerts() string {
	script := `ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs check-expiration`
	return script
}

func (b *BottlerocketRenewer) copyExternalEtcdCerts() string {
	script := fmt.Sprintf(`if [ -d "/tmp/%[1]s" ]; then
    cp /tmp/%[1]s/apiserver-etcd-client.crt %[2]s/server-etcd-client.crt
    cp /tmp/%[1]s/apiserver-etcd-client.key %[2]s/apiserver-etcd-client.key
    rm -rf /tmp/%[1]s
fi`, tempLocalEtcdCertsDir, brControlPlaneCertDir)
	return script
}

func (b *BottlerocketRenewer) restartControlPlaneStaticPods() string {
	script := `
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=false
sleep 10
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=true
`
	return script
}

func (b *BottlerocketRenewer) backupEtcdCerts(backupDir string) string {
	script := fmt.Sprintf(`cp -r /var/lib/etcd/pki /var/lib/etcd/pki.bak_%[1]s
rm /var/lib/etcd/pki/*
cp /var/lib/etcd/pki.bak_%[1]s/ca.* /var/lib/etcd/pki`, backupDir)
	return script
}

func (b *BottlerocketRenewer) renewEtcdCerts() string {
	script := `ctr run \
--mount type=bind,src=/var/lib/etcd/pki,dst=/etc/etcd/pki,options=rbind:rw \
--net-host \
--rm \
${IMAGE_ID} tmp-cert-renew \
/opt/bin/etcdadm join phase certificates http://eks-a-etcd-dumb-url --init-system kubelet`
	return script
}

func (b *BottlerocketRenewer) validateEtcdCerts() string {
	script := `ETCD_CONTAINER_ID=$(ctr -n k8s.io c ls | grep -w "etcd-io" | cut -d " " -f1 | tail -1)
ctr -n k8s.io t exec --exec-id etcd ${ETCD_CONTAINER_ID} etcdctl \
     --cacert=/var/lib/etcd/pki/ca.crt \
     --cert=/var/lib/etcd/pki/server.crt \
     --key=/var/lib/etcd/pki/server.key \
     member list`
	return script
}

func (b *BottlerocketRenewer) copyEtcdCertsToTemp(tempDir string) string {
	script := fmt.Sprintf(`cp /var/lib/etcd/pki/apiserver-etcd-client.* %[1]s/ 
chmod 600 %[1]s/apiserver-etcd-client.crt
chmod 600 %[1]s/apiserver-etcd-client.key`, tempDir)
	return script
}

func (b *BottlerocketRenewer) cleanupEtcdTempFiles(tempDir string) string {
	script := fmt.Sprintf(`rm -f %s/apiserver-etcd-client.*`, tempDir)
	return script
}

func (b *BottlerocketRenewer) createTempDirectoryAndWriteCerts(dirName, certBase64, keyBase64 string) string {
	script := fmt.Sprintf(`TARGET_DIR="/tmp/%[1]s"
mkdir -p "${TARGET_DIR}"
chmod 755 "${TARGET_DIR}"

cat <<'CRT_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.crt"
%[2]s
CRT_END
[ $? -eq 0 ]

cat <<'KEY_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.key"
%[3]s
KEY_END
[ $? -eq 0 ]

chmod 600 "${TARGET_DIR}/apiserver-etcd-client.crt"
chmod 600 "${TARGET_DIR}/apiserver-etcd-client.key"`, dirName, certBase64, keyBase64)
	return script
}

func (b *BottlerocketRenewer) readTempFile(filePath string) string {
	script := fmt.Sprintf(`sudo sheltie << 'EOF'
cat %s
exit
EOF`, filePath)
	return script
}
