package certificates

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/constants"
)

func buildBRImagePullCmd() []string {
	return []string{
		`IMAGE_ID=$(apiclient get | apiclient exec admin jq -r '.settings["host-containers"]["kubeadm-bootstrap"].source')
ctr image pull ${IMAGE_ID}`,
	}
}

func buildBRControlPlaneBackupCertsCmd(component string, hasExternalEtcd bool, backupDir, certDir string) []string {
	var script string
	if component == constants.ControlPlaneComponent && hasExternalEtcd {
		script = fmt.Sprintf(`mkdir -p '/etc/kubernetes/pki.bak_%[1]s'
cp -r %[2]s/* '/etc/kubernetes/pki.bak_%[1]s/'
rm -rf '/etc/kubernetes/pki.bak_%[1]s/etcd'`, backupDir, certDir)
	} else {
		script = fmt.Sprintf("cp -r '%s' '/etc/kubernetes/pki.bak_%s'", certDir, backupDir)
	}
	return []string{script}
}

func buildBRControlPlaneRenewCertsCmd() []string {
	script := `ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs renew all`
	return []string{script}
}

func buildBRControlPlaneCheckCertsCmd() []string {
	script := `ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs check-expiration`
	return []string{script}
}

func buildBRControlPlaneCopyCertsFromTmpCmd() []string {
	script := `if [ -d "/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs" ]; then
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
	return []string{script}
}

func buildBRControlPlaneRestartPodsCmd() []string {
	script := `
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=false
sleep 10
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=true
`
	return []string{script}
}

func buildBREtcdBackupCertsCmd(backupDir string) []string {
	script := fmt.Sprintf(`cp -r /var/lib/etcd/pki /var/lib/etcd/pki.bak_%[1]s
rm /var/lib/etcd/pki/*
cp /var/lib/etcd/pki.bak_%[1]s/ca.* /var/lib/etcd/pki
echo "✅ Certs backedup"`, backupDir)
	return []string{script}
}

func buildBREtcdRenewCertsCmd() []string {
	script := `ctr run \
--mount type=bind,src=/var/lib/etcd/pki,dst=/etc/etcd/pki,options=rbind:rw \
--net-host \
--rm \
${IMAGE_ID} tmp-cert-renew \
/opt/bin/etcdadm join phase certificates http://eks-a-etcd-dumb-url --init-system kubelet`
	return []string{script}
}

func buildBREtcdCopyCertsToTmpCmd(tempDir string) []string {
	script := fmt.Sprintf(`echo "Source files in /var/lib/etcd/pki/:"
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
ls -l %[1]s/apiserver-etcd-client.*`, tempDir)
	return []string{script}
}

func buildBREtcdCleanupTmpCmd(tempDir string) []string {
	script := fmt.Sprintf(`rm -f %s/apiserver-etcd-client.*`, tempDir)
	return []string{script}
}

// Cert Transfer and Read

func buildBRCreateTmpDirCmd(dirName string) []string {
	script := fmt.Sprintf(`echo "Creating directory..."
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
ls -l /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/`, dirName)
	return []string{script}
}

func buildBRWriteCertToTmpCmd(certBase64 string) []string {
	script := fmt.Sprintf(`echo "Writing certificate file..."
cat <<'CRT_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.crt"
%s
CRT_END
if [ $? -ne 0 ]; then
    echo "❌ Failed to write certificate file"
    exit 1
fi`, certBase64)
	return []string{script}
}

func buildBRWriteKeyToTmpCmd(keyBase64 string) []string {
	script := fmt.Sprintf(`echo "Writing key file..."
cat <<'KEY_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.key"
%s
KEY_END
if [ $? -ne 0 ]; then
    echo "❌ Failed to write key file"
    exit 1
fi`, keyBase64)
	return []string{script}
}

func buildBRSetTmpCertPermissionsCmd() []string {
	script := `echo "Setting permissions..."
chmod 600 "${TARGET_DIR}/apiserver-etcd-client.crt" || {
    echo "❌ Failed to set permissions on certificate"
    exit 1
}
chmod 600 "${TARGET_DIR}/apiserver-etcd-client.key" || {
    echo "❌ Failed to set permissions on key"
    exit 1
}`
	return []string{script}
}

func buildBRListTmpFilesCmd(tempDir string) []string {
	script := fmt.Sprintf(`sudo sheltie << 'EOF'
echo "Checking source files:"
ls -l %s/apiserver-etcd-client.*
exit
EOF`, tempDir)
	return []string{script}
}

func buildBRReadTmpCertCmd(tempDir string) []string {
	script := fmt.Sprintf(`sudo sheltie << 'EOF'
cat %s/apiserver-etcd-client.crt
exit
EOF`, tempDir)
	return []string{script}
}

func buildBRReadTmpKeyCmd(tempDir string) []string {
	script := fmt.Sprintf(`sudo sheltie << 'EOF'
cat %s/apiserver-etcd-client.key
exit
EOF`, tempDir)
	return []string{script}
}
