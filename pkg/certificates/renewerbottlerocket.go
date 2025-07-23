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
	brEtcdPkiDir            = "/var/lib/etcd/pki"
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

	shellCommands := b.sheltie(
		b.pullContainerImage(),
		b.backupControlPlaneCerts(component, hasExternalEtcd, brControlPlaneCertDir),
		b.renewControlPlaneCerts(),
		b.checkControlPlaneCerts(),
		b.restartControlPlaneStaticPods(),
	)

	if _, err := ssh.RunCommand(ctx, node, shellCommands); err != nil {
		return fmt.Errorf("renewing control plane certificates: %v", err)
	}

	logger.V(0).Info("Renewed control-plane certificates", "node", node)
	return nil
}

// TransferCertsToControlPlaneFromLocal transfers etcd client certificates to a control plane node.
func (b *BottlerocketRenewer) TransferCertsToControlPlaneFromLocal(
	ctx context.Context, node string, ssh SSHRunner,
) error {
	certificateBytes, err := os.ReadFile(filepath.Join(
		b.backup, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt"))
	if err != nil {
		return fmt.Errorf("reading certificate file from the admin machine: %v", err)
	}
	keyBytes, err := os.ReadFile(filepath.Join(
		b.backup, tempLocalEtcdCertsDir, "apiserver-etcd-client.key"))
	if err != nil {
		return fmt.Errorf("reading key file from the admin machine: %v", err)
	}

	shellCommands := b.sheltie(
		b.createTempDirectory(tempLocalEtcdCertsDir),
		b.writeCertToTemp(base64.StdEncoding.EncodeToString(certificateBytes)),
		b.writeKeyToTemp(base64.StdEncoding.EncodeToString(keyBytes)),
		b.copyExternalEtcdCerts(),
	)

	if _, err := ssh.RunCommand(ctx, node, shellCommands, WithSSHLogging(false)); err != nil {
		return fmt.Errorf("transfering certificates to control plane: %v", err)
	}

	// if _, err := ssh.RunCommand(ctx, node, b.copyExternalEtcdCerts()); err != nil {
	// 	return fmt.Errorf("copying etcd client certs: %v", err)
	// }
	return nil
}

// RenewEtcdCerts renews etcd certificates on a Bottlerocket node.
func (b *BottlerocketRenewer) RenewEtcdCerts(ctx context.Context, node string, ssh SSHRunner) error {
	logger.V(0).Info("Renewing etcd certificates", "node", node)

	if _, err := ssh.RunCommand(ctx, node, b.sheltie(
		b.pullContainerImage(),
		b.backupEtcdCerts(),
		b.renewEtcdCerts(),
	)); err != nil {
		return fmt.Errorf("renewing etcd certificates: %v", err)
	}

	if _, err := ssh.RunCommand(ctx, node, b.sheltie(
		b.validateEtcdCerts(),
	)); err != nil {
		return fmt.Errorf("validating etcd certificates: %v", err)
	}

	logger.V(0).Info("Renewed etcd certificates", "node", node)

	return nil
}

// CopyEtcdCertsToLocal copies the etcd certificates from the specified node to the local machine.
func (b *BottlerocketRenewer) CopyEtcdCertsToLocal(ctx context.Context, node string, ssh SSHRunner) error {

	if _, err := ssh.RunCommand(ctx, node, b.sheltie(
		b.copyEtcdCertsToTemp(brTempDir),
	)); err != nil {
		return fmt.Errorf("copying etcd certificates to tmp: %v", err)
	}

	certificatePath := filepath.Join(brTempDir, "apiserver-etcd-client.crt")
	certificateContent, err := ssh.RunCommand(ctx, node, b.readTempFile(certificatePath), WithSSHLogging(false))
	if err != nil {
		return fmt.Errorf("reading etcd certificate file: %v", err)
	}

	if len(certificateContent) == 0 {
		return fmt.Errorf("etcd certificate file is empty")
	}

	keyFilePath := filepath.Join(brTempDir, "apiserver-etcd-client.key")
	keyContent, err := ssh.RunCommand(ctx, node, b.readTempFile(keyFilePath), WithSSHLogging(false))
	if err != nil {
		return fmt.Errorf("reading etcd key file: %v", err)
	}
	if len(keyContent) == 0 {
		return fmt.Errorf("etcd key file is empty")
	}

	localCertificateDir := filepath.Join(b.backup, tempLocalEtcdCertsDir)
	if err := os.MkdirAll(localCertificateDir, 0o700); err != nil {
		return fmt.Errorf("creating local etcd cert dir: %v", err)
	}

	localCertificatePath := filepath.Join(localCertificateDir, "apiserver-etcd-client.crt")
	localKeyFilePath := filepath.Join(localCertificateDir, "apiserver-etcd-client.key")

	if err := os.WriteFile(localCertificatePath, []byte(certificateContent), 0o600); err != nil {
		return fmt.Errorf("writing etcd certificate file: %v", err)
	}
	if err := os.WriteFile(localKeyFilePath, []byte(keyContent), 0o600); err != nil {
		return fmt.Errorf("writing etcd key file: %v", err)
	}

	if _, err := ssh.RunCommand(ctx, node, b.sheltie(
		b.cleanupEtcdTempFiles(brTempDir),
	)); err != nil {
		return fmt.Errorf("cleaning up temporary etcd files: %v", err)
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

func (b *BottlerocketRenewer) backupControlPlaneCerts(_ string, hasExternalEtcd bool, certDir string) string {

	backupPath := fmt.Sprintf("/var/lib/kubeadm/pki.bak_%s", b.backup)

	if hasExternalEtcd {
		return fmt.Sprintf("mkdir -p '%s' && cp -r %s/* '%s/' && rm -rf '%s/etcd'",
			backupPath, certDir, backupPath, backupPath)
	}

	return fmt.Sprintf("cp -r '%s' '%s'", certDir, backupPath)
}

func (b *BottlerocketRenewer) renewControlPlaneCerts() string {
	renewCerts := `ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs renew all`
	return renewCerts
}

func (b *BottlerocketRenewer) checkControlPlaneCerts() string {
	checkCerts := `ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs check-expiration`
	return checkCerts
}

func (b *BottlerocketRenewer) copyExternalEtcdCerts() string {
	copyCerts := fmt.Sprintf(`
		mkdir -p %[2]s
		cp /tmp/%[1]s/apiserver-etcd-client.crt %[2]s/server-etcd-client.crt
		cp /tmp/%[1]s/apiserver-etcd-client.key %[2]s/apiserver-etcd-client.key
		rm -rf /tmp/%[1]s
		`, tempLocalEtcdCertsDir, brControlPlaneCertDir)
	return copyCerts
}

func (b *BottlerocketRenewer) restartControlPlaneStaticPods() string {
	restartPods := `
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=false
sleep 10
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=true
`
	return restartPods
}

func (b *BottlerocketRenewer) backupEtcdCerts() string {
	backupCerts := fmt.Sprintf(`cp -r %[1]s %[1]s.bak_%[2]s
rm %[1]s/*
cp %[1]s.bak_%[2]s/ca.* %[1]s`, brEtcdPkiDir, b.backup)
	return backupCerts
}

func (b *BottlerocketRenewer) renewEtcdCerts() string {
	renewCerts := fmt.Sprintf(`ctr run \
--mount type=bind,src=%s,dst=/etc/etcd/pki,options=rbind:rw \
--net-host \
--rm \
${IMAGE_ID} tmp-cert-renew \
/opt/bin/etcdadm join phase certificates http://eks-a-etcd-dumb-url --init-system kubelet`, brEtcdPkiDir)
	return renewCerts
}

func (b *BottlerocketRenewer) validateEtcdCerts() string {
	validateCerts := fmt.Sprintf(`ETCD_CONTAINER_ID=$(ctr -n k8s.io c ls | grep -w "etcd-io" | cut -d " " -f1 | tail -1)
ctr -n k8s.io t exec --exec-id etcd ${ETCD_CONTAINER_ID} etcdctl \
     --cacert=%[1]s/ca.crt \
     --cert=%[1]s/server.crt \
     --key=%[1]s/server.key \
     member list`, brEtcdPkiDir)
	return validateCerts
}

func (b *BottlerocketRenewer) copyEtcdCertsToTemp(tempDir string) string {
	copyCerts := fmt.Sprintf(`cp %[1]s/apiserver-etcd-client.* %[2]s/ 
chmod 600 %[2]s/apiserver-etcd-client.key`, brEtcdPkiDir, tempDir)
	return copyCerts
}

func (b *BottlerocketRenewer) cleanupEtcdTempFiles(tempDir string) string {
	cleanup := fmt.Sprintf(`rm -f %s/apiserver-etcd-client.*`, tempDir)
	return cleanup
}

func (b *BottlerocketRenewer) createTempDirectory(tempLocalEtcdCertsDir string) string {
	tempDirectory := fmt.Sprintf(`TARGET_DIR="/tmp/%s"
mkdir -p "${TARGET_DIR}"
`, tempLocalEtcdCertsDir)
	return tempDirectory
}

func (b *BottlerocketRenewer) writeCertToTemp(certificateBytes64 string) string {
	certToTemp := fmt.Sprintf(`cat <<'CRT_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.crt"
%s
CRT_END
`, certificateBytes64)
	return certToTemp
}

func (b *BottlerocketRenewer) writeKeyToTemp(keyBytes64 string) string {
	keyToTemp := fmt.Sprintf(`cat <<'KEY_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.key"
%s
KEY_END
`, keyBytes64)
	return keyToTemp
}

func (b *BottlerocketRenewer) readTempFile(filePath string) string {
	readFile := fmt.Sprintf(`sudo sheltie << 'EOF'
cat %s
exit
EOF`, filePath)
	return readFile
}
