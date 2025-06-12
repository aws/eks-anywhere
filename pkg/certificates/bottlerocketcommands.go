// Package certificates provides utilities for managing certificates on Bottlerocket OS nodes.
package certificates

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/constants"
)

// ControlPlaneCommands contains shell commands for certificate operations on control plane nodes.
type ControlPlaneCommands struct {
	ShelliePrefix string
	BackupCerts   string
	ImagePull     string
	RenewCerts    string
	CheckCerts    string
	CopyCerts     string
	RestartPods   string
}

// ControlPlaneCommandBuilder builds shell commands for certificate operations on control plane nodes.
type ControlPlaneCommandBuilder struct {
	BackupDir       string
	CertDir         string
	Component       string
	HasExternalEtcd bool
}

// NewControlPlaneCommandBuilder creates a new builder for control plane certificate commands.
func NewControlPlaneCommandBuilder(backupDir, certDir, component string, hasExternalEtcd bool) *ControlPlaneCommandBuilder {
	return &ControlPlaneCommandBuilder{
		BackupDir:       backupDir,
		CertDir:         certDir,
		Component:       component,
		HasExternalEtcd: hasExternalEtcd,
	}
}

// BuildCommands creates a set of commands for control plane certificate operations.
func (b *ControlPlaneCommandBuilder) BuildCommands() *ControlPlaneCommands {
	return &ControlPlaneCommands{
		ShelliePrefix: b.buildShelliePrefix(),
		BackupCerts:   b.buildBackupCerts(),
		ImagePull:     b.buildImagePull(),
		RenewCerts:    b.buildRenewCerts(),
		CheckCerts:    b.buildCheckCerts(),
		CopyCerts:     b.buildCopyCerts(),
		RestartPods:   b.buildRestartPods(),
	}
}

// EtcdCommands contains shell commands for certificate operations on etcd nodes.
type EtcdCommands struct {
	ShelliePrefix string
	ImagePull     string
	BackupCerts   string
	RenewCerts    string
	CopyCerts     string
	Cleanup       string
}

// EtcdCommandBuilder builds shell commands for certificate operations on etcd nodes.
type EtcdCommandBuilder struct {
	BackupDir string
	TempDir   string
}

// NewEtcdCommandBuilder creates a new builder for etcd certificate commands.
func NewEtcdCommandBuilder(backupDir, tempDir string) *EtcdCommandBuilder {
	return &EtcdCommandBuilder{
		BackupDir: backupDir,
		TempDir:   tempDir,
	}
}

// BuildCommands creates a set of commands for etcd certificate operations.
func (b *EtcdCommandBuilder) BuildCommands() *EtcdCommands {
	return &EtcdCommands{
		ShelliePrefix: b.buildShelliePrefix(),
		ImagePull:     b.buildImagePull(),
		BackupCerts:   b.buildBackupCerts(),
		RenewCerts:    b.buildRenewCerts(),
		CopyCerts:     b.buildCopyCerts(),
		Cleanup:       b.buildCleanup(),
	}
}

// CertTransferCommands contains shell commands for transferring certificates to nodes.
type CertTransferCommands struct {
	ShelliePrefix    string
	CreateDir        string
	SetPermissions   string
	WriteCertificate string
	WriteKey         string
}

// CertTransferBuilder builds shell commands for transferring certificates to nodes.
type CertTransferBuilder struct {
	TempDir     string
	Certificate string
	Key         string
}

// NewCertTransferBuilder creates a new builder for certificate transfer commands.
func NewCertTransferBuilder(tempDir, certificate, key string) *CertTransferBuilder {
	return &CertTransferBuilder{
		TempDir:     tempDir,
		Certificate: certificate,
		Key:         key,
	}
}

// BuildCommands creates a set of commands for certificate transfer operations.
func (b *CertTransferBuilder) BuildCommands() *CertTransferCommands {
	return &CertTransferCommands{
		ShelliePrefix:    b.buildShelliePrefix(),
		CreateDir:        b.buildCreateDir(),
		SetPermissions:   b.buildSetPermissions(),
		WriteCertificate: b.buildWriteCertificate(),
		WriteKey:         b.buildWriteKey(),
	}
}

// CertReadCommands contains shell commands for reading certificates from nodes.
type CertReadCommands struct {
	ListFiles string
	ReadCert  string
	ReadKey   string
}

// CertReadBuilder builds shell commands for reading certificates from nodes.
type CertReadBuilder struct {
	TempDir string
}

// NewCertReadBuilder creates a new builder for certificate read commands.
func NewCertReadBuilder(tempDir string) *CertReadBuilder {
	return &CertReadBuilder{
		TempDir: tempDir,
	}
}

// BuildCommands creates a set of commands for certificate read operations.
func (b *CertReadBuilder) BuildCommands() *CertReadCommands {
	return &CertReadCommands{
		ListFiles: b.buildListFiles(),
		ReadCert:  b.buildReadCert(),
		ReadKey:   b.buildReadKey(),
	}
}

func (b *ControlPlaneCommandBuilder) buildShelliePrefix() string {
	return `set -euo pipefail
sudo sheltie << 'EOF'`
}

func (b *ControlPlaneCommandBuilder) buildBackupCerts() string {
	if b.Component == constants.ControlPlaneComponent && b.HasExternalEtcd {
		return fmt.Sprintf(`mkdir -p '/etc/kubernetes/pki.bak_%[1]s'
cd %[2]s
for f in $(find . -type f ! -path './etcd/*'); do
    mkdir -p $(dirname '/etc/kubernetes/pki.bak_%[1]s/'$f)
    cp $f '/etc/kubernetes/pki.bak_%[1]s/'$f
done`, b.BackupDir, b.CertDir)
	}
	return fmt.Sprintf("cp -r '%s' '/etc/kubernetes/pki.bak_%s'",
		b.CertDir, b.BackupDir)
}

func (b *ControlPlaneCommandBuilder) buildImagePull() string {
	return `IMAGE_ID=$(apiclient get | apiclient exec admin jq -r '.settings["host-containers"]["kubeadm-bootstrap"].source')
ctr image pull ${IMAGE_ID}`
}

func (b *ControlPlaneCommandBuilder) buildRenewCerts() string {
	return `ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs renew all`
}

func (b *ControlPlaneCommandBuilder) buildCheckCerts() string {
	return `ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs check-expiration`
}

func (b *ControlPlaneCommandBuilder) buildCopyCerts() string {
	return `if [ -d "/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs" ]; then
    echo "Source certificates:"
    ls -l /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs/
    
    echo "Destination before copy:"
    ls -l /var/lib/kubeadm/pki/server-etcd-client.crt || true
    ls -l /var/lib/kubeadm/pki/apiserver-etcd-client.key || true
    
    cp -v /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs/apiserver-etcd-client.crt /var/lib/kubeadm/pki/server-etcd-client.crt || {
        echo "❌ Failed to copy certificate"
        exit 1
    }
    cp -v /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs/apiserver-etcd-client.key /var/lib/kubeadm/pki/apiserver-etcd-client.key || {
        echo "❌ Failed to copy key"
        exit 1
    }
    
    chmod 600 /var/lib/kubeadm/pki/server-etcd-client.crt || {
        echo "❌ Failed to set certificate permissions"
        exit 1
    }
    chmod 600 /var/lib/kubeadm/pki/apiserver-etcd-client.key || {
        echo "❌ Failed to set key permissions"
        exit 1
    }
    
    echo "Destination after copy:"
    ls -l /var/lib/kubeadm/pki/server-etcd-client.crt
    ls -l /var/lib/kubeadm/pki/apiserver-etcd-client.key
    
    echo "✅ Certificates copied successfully"
else
    echo "❌ Source directory does not exist"
    ls -l /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/
    exit 1
fi`
}

func (b *ControlPlaneCommandBuilder) buildRestartPods() string {
	return `apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' \
 | xargs -n1 -I{} apiclient set settings.kubernetes.static-pods.{}.enabled=false
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' \
 | xargs -n1 -I{} apiclient set settings.kubernetes.static-pods.{}.enabled=true`
}

func (b *EtcdCommandBuilder) buildShelliePrefix() string {
	return `set -euo pipefail
sudo sheltie << 'EOF'`
}

func (b *EtcdCommandBuilder) buildImagePull() string {
	return `IMAGE_ID=$(apiclient get | apiclient exec admin jq -r '.settings["host-containers"]["kubeadm-bootstrap"].source')
ctr image pull ${IMAGE_ID}`
}

func (b *EtcdCommandBuilder) buildBackupCerts() string {
	return fmt.Sprintf(`cp -r /var/lib/etcd/pki /var/lib/etcd/pki.bak_%[1]s
rm /var/lib/etcd/pki/*
cp /var/lib/etcd/pki.bak_%[1]s/ca.* /var/lib/etcd/pki
echo "✅ Certs backedup"`, b.BackupDir)
}

func (b *EtcdCommandBuilder) buildRenewCerts() string {
	return `ctr run \
--mount type=bind,src=/var/lib/etcd/pki,dst=/etc/etcd/pki,options=rbind:rw \
--net-host \
--rm \
${IMAGE_ID} tmp-cert-renew \
/opt/bin/etcdadm join phase certificates http://eks-a-etcd-dumb-url --init-system kubelet`
}

func (b *EtcdCommandBuilder) buildCopyCerts() string {
	return fmt.Sprintf(`echo "Source files in /var/lib/etcd/pki/:"
ls -l /var/lib/etcd/pki/apiserver-etcd-client.*

echo "Copying certificates to %[1]s..."
cp /var/lib/etcd/pki/apiserver-etcd-client.* %[1]s || { 
    echo "❌ Failed to copy certs to tmp"
    echo "Source files:"
    ls -l /var/lib/etcd/pki/apiserver-etcd-client.*
    echo "Destination directory:"
    ls -l %[1]s
    exit 1
}

echo "Setting permissions..."
chmod 600 %[1]s/apiserver-etcd-client.crt || { 
    echo "❌ Failed to chmod certificate"
    ls -l %[1]s/apiserver-etcd-client.crt
    exit 1
}
chmod 600 %[1]s/apiserver-etcd-client.key || { 
    echo "❌ Failed to chmod key"
    ls -l %[1]s/apiserver-etcd-client.key
    exit 1
}

echo "Verifying copied files..."
ls -l %[1]s/apiserver-etcd-client.*`, b.TempDir)
}

func (b *EtcdCommandBuilder) buildCleanup() string {
	return fmt.Sprintf(`rm -f %s/apiserver-etcd-client.*`, b.TempDir)
}

func (b *CertTransferBuilder) buildShelliePrefix() string {
	return `sudo sheltie << 'EOF'`
}

func (b *CertTransferBuilder) buildCreateDir() string {
	return fmt.Sprintf(`echo "Creating directory..."
TARGET_DIR="/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/%[1]s"
mkdir -p "${TARGET_DIR}" || {
    echo "❌ Failed to create directory"
    exit 1
}

chmod 755 "${TARGET_DIR}" || {
    echo "❌ Failed to set directory permissions"
    exit 1
}

echo "Verifying directory:"
ls -ld "${TARGET_DIR}"
ls -l /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/`, b.TempDir)
}

func (b *CertTransferBuilder) buildSetPermissions() string {
	return `echo "Setting permissions..."
chmod 600 "${TARGET_DIR}/apiserver-etcd-client.crt" || {
    echo "❌ Failed to set permissions on certificate"
    exit 1
}
chmod 600 "${TARGET_DIR}/apiserver-etcd-client.key" || {
    echo "❌ Failed to set permissions on key"
    exit 1
}`
}

func (b *CertTransferBuilder) buildWriteCertificate() string {
	return fmt.Sprintf(`echo "Writing certificate file..."
cat <<'CRT_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.crt"
%s
CRT_END
if [ $? -ne 0 ]; then
    echo "❌ Failed to write certificate file"
    exit 1
fi`, b.Certificate)
}

func (b *CertTransferBuilder) buildWriteKey() string {
	return fmt.Sprintf(`echo "Writing key file..."
cat <<'KEY_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.key"
%s
KEY_END
if [ $? -ne 0 ]; then
    echo "❌ Failed to write key file"
    exit 1
fi`, b.Key)
}

func (b *CertReadBuilder) buildListFiles() string {
	return fmt.Sprintf(`sudo sheltie << 'EOF'
echo "Checking source files:"
ls -l %s/apiserver-etcd-client.*
exit
EOF`, b.TempDir)
}

func (b *CertReadBuilder) buildReadCert() string {
	return fmt.Sprintf(`sudo sheltie << 'EOF'
cat %s/apiserver-etcd-client.crt
exit
EOF`, b.TempDir)
}

func (b *CertReadBuilder) buildReadKey() string {
	return fmt.Sprintf(`sudo sheltie << 'EOF'
cat %s/apiserver-etcd-client.key
exit
EOF`, b.TempDir)
}
