package certificates

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	persistentCertDir     = "/var/lib/eks-anywhere/certificates"
	persistentEtcdCertDir = "etcd-certs"
)

func (r *Renewer) renewControlPlaneCertsBottlerocket(ctx context.Context, node string, config *RenewalConfig, component string) error {
	fmt.Printf("Processing control plane node: %s...\n", node)

	// for renew control panel only
	if component == componentControlPlane && len(config.Etcd.Nodes) > 0 {
		if err := r.loadCertsFromPersistentStorage(); err != nil {
			return fmt.Errorf("failed to load certificates from persistent storage: %v", err)
		}
	}

	client, err := r.sshDialer("tcp", fmt.Sprintf("%s:22", node), r.sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to node %s: %v", node, err)
	}
	defer client.Close()

	// If we have external etcd nodes, first transfer certificates to the node
	if len(config.Etcd.Nodes) > 0 {
		if err := r.transferCertsToControlPlane(ctx, node); err != nil {
			return fmt.Errorf("failed to transfer certificates to control plane node: %v", err)
		}
	}

	// Single sheltie session for all control plane operations
	var backupCmd string
	if component == componentControlPlane && len(config.Etcd.Nodes) > 0 {
		backupCmd = fmt.Sprintf(`mkdir -p '/etc/kubernetes/pki.bak_%[1]s'
cd %[2]s
for f in $(find . -type f ! -path './etcd/*'); do
    mkdir -p $(dirname '/etc/kubernetes/pki.bak_%[1]s/'$f)
    cp $f '/etc/kubernetes/pki.bak_%[1]s/'$f
done`, r.backupDir, bottlerocketControlPlaneCertDir)
	} else {
		backupCmd = fmt.Sprintf("cp -r '%s' '/etc/kubernetes/pki.bak_%s'",
			bottlerocketControlPlaneCertDir, r.backupDir)
	}

	session := fmt.Sprintf(`set -euo pipefail
sudo sheltie << 'EOF'
set -x

# 1. backup
%[1]s

# 2. pull image
IMAGE_ID=$(apiclient get | apiclient exec admin jq -r '.settings["host-containers"]["kubeadm-bootstrap"].source')
ctr image pull ${IMAGE_ID}

# 3. kubeadm renew
ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs renew all

# 4. kubeadm check
ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-check \
/opt/bin/kubeadm certs check-expiration

# 5. copy etcd certs
if [ -d "/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs" ]; then
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
fi



# 6. restart pods
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' \
 | xargs -n1 -I{} apiclient set settings.kubernetes.static-pods.{}.enabled=false
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' \
 | xargs -n1 -I{} apiclient set settings.kubernetes.static-pods.{}.enabled=true
EOF`, backupCmd)

	if err := r.runCommand(ctx, client, session); err != nil {
		return fmt.Errorf("failed to renew control panel node certificates: %v", err)
	}

	fmt.Printf("✅ Completed renewing certificate for the control node: %s.\n", node)
	fmt.Printf("---------------------------------------------\n")
	return nil
}

func (r *Renewer) transferCertsToControlPlane(ctx context.Context, node string) error {
	fmt.Printf("Transferring certificates to control plane node: %s...\n", node)

	client, err := r.sshDialer("tcp", fmt.Sprintf("%s:22", node), r.sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to node %s: %v", node, err)
	}
	defer client.Close()

	srcCrt := filepath.Join(r.backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt")
	crtContent, err := os.ReadFile(srcCrt)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %v", err)
	}

	srcKey := filepath.Join(r.backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.key")
	keyContent, err := os.ReadFile(srcKey)
	if err != nil {
		return fmt.Errorf("failed to read key file: %v", err)
	}

	crtBase64 := base64.StdEncoding.EncodeToString(crtContent)
	keyBase64 := base64.StdEncoding.EncodeToString(keyContent)

	session := fmt.Sprintf(`
sudo sheltie << 'EOF'
set -x 

echo "Creating directory..."
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
ls -l /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/

echo "Writing certificate file..."
cat <<'CRT_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.crt"
%[2]s
CRT_END
if [ $? -ne 0 ]; then
    echo "❌ Failed to write certificate file"
    exit 1
fi

echo "Writing key file..."
cat <<'KEY_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.key"
%[3]s
KEY_END
if [ $? -ne 0 ]; then
    echo "❌ Failed to write key file"
    exit 1
fi

echo "Setting permissions..."
chmod 600 "${TARGET_DIR}/apiserver-etcd-client.crt" || {
    echo "❌ Failed to set permissions on certificate"
    exit 1
}
chmod 600 "${TARGET_DIR}/apiserver-etcd-client.key" || {
    echo "❌ Failed to set permissions on key"
    exit 1
}

echo "Verifying files..."
ls -l "${TARGET_DIR}"/*
echo "✅ Files transferred successfully"
exit
EOF`, tempLocalEtcdCertsDir, crtBase64, keyBase64)

	if err := r.runCommand(ctx, client, session); err != nil {
		return fmt.Errorf("failed to transfer certificates: %v", err)
	}

	fmt.Printf("External certificates transferred to control plane node: %s.\n", node)
	return nil
}

func (r *Renewer) renewEtcdCertsBottlerocket(ctx context.Context, node string) error {
	fmt.Printf("Processing etcd node: %s...\n", node)

	client, err := r.sshDialer("tcp", fmt.Sprintf("%s:22", node), r.sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to node %s: %v", node, err)
	}
	defer client.Close()

	// First sheltie session for certificate renewal
	firstSession := fmt.Sprintf(`set -euo pipefail
sudo sheltie << 'EOF'
set -x
# Get image ID and pull it
IMAGE_ID=$(apiclient get | apiclient exec admin jq -r '.settings["host-containers"]["kubeadm-bootstrap"].source')
ctr image pull ${IMAGE_ID}

# Backup certs
cp -r /var/lib/etcd/pki /var/lib/etcd/pki.bak_%[1]s
rm /var/lib/etcd/pki/*
cp /var/lib/etcd/pki.bak_%[1]s/ca.* /var/lib/etcd/pki
echo "✅ Certs backedup"

# Recreate certificates
ctr run \
--mount type=bind,src=/var/lib/etcd/pki,dst=/etc/etcd/pki,options=rbind:rw \
--net-host \
--rm \
${IMAGE_ID} tmp-cert-renew \
/opt/bin/etcdadm join phase certificates http://eks-a-etcd-dumb-url --init-system kubelet
exit
EOF`, r.backupDir)

	if err := r.runCommand(ctx, client, firstSession); err != nil {
		return fmt.Errorf("failed to renew certificates: %v", err)
	}

	// Second sheltie session for copying certs
	secondSession := fmt.Sprintf(`set -euo pipefail
sudo sheltie << 'EOF'
set -x
echo "Source files in /var/lib/etcd/pki/:"
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
ls -l %[1]s/apiserver-etcd-client.*
exit
EOF`, bottlerocketTmpDir)

	if err := r.runCommand(ctx, client, secondSession); err != nil {
		return fmt.Errorf("failed to copy certificates2 to tmp: %v", err)
	}

	// Copy certificates to local
	fmt.Printf("Copying certificates from node %s...\n", node)
	if err := r.copyEtcdCerts(ctx, client, node); err != nil {
		return fmt.Errorf("failed to copy certificates3: %v", err)
	}

	// Third sheltie session for cleanup
	thirdSession := fmt.Sprintf(`set -euo pipefail
sudo sheltie << 'EOF'
set -x
rm -f %s/apiserver-etcd-client.*
exit
EOF`, bottlerocketTmpDir)

	if err := r.runCommand(ctx, client, thirdSession); err != nil {
		return fmt.Errorf("failed to cleanup temporary files: %v", err)
	}

	fmt.Printf("✅ Completed renewing certificate for the ETCD node: %s.\n", node)
	fmt.Printf("---------------------------------------------\n")

	// save etcd cert for control panel renew
	if err := r.saveCertsToPersistentStorage(); err != nil {
		return fmt.Errorf("failed to save certificates to persistent storage: %v", err)
	}

	return nil
}

func (r *Renewer) copyEtcdCerts(ctx context.Context, client sshClient, node string) error {
	fmt.Printf("Reading certificate from ETCD node %s...\n", node)
	fmt.Printf("Using backup directory: %s\n", r.backupDir)

	debugCmd := fmt.Sprintf(`
sudo sheltie << 'EOF'
echo "Checking source files:"
ls -l %s/apiserver-etcd-client.*
exit
EOF`, bottlerocketTmpDir)
	if err := r.runCommand(ctx, client, debugCmd); err != nil {
		return fmt.Errorf("failed to list certificate files: %v", err)
	}

	crtCmd := fmt.Sprintf(`
sudo sheltie << 'EOF'
cat %s/apiserver-etcd-client.crt
exit
EOF`, bottlerocketTmpDir)
	crtContent, err := r.runCommandWithOutput(ctx, client, crtCmd)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %v", err)
	}

	if len(crtContent) == 0 {
		return fmt.Errorf("certificate file is empty")
	}

	fmt.Printf("Reading key from ETCD node %s...\n", node)
	keyCmd := fmt.Sprintf(`
sudo sheltie << 'EOF'
cat %s/apiserver-etcd-client.key
exit
EOF`, bottlerocketTmpDir)
	keyContent, err := r.runCommandWithOutput(ctx, client, keyCmd)
	if err != nil {
		return fmt.Errorf("failed to read key file: %v", err)
	}

	if len(keyContent) == 0 {
		return fmt.Errorf("key file is empty")
	}

	crtPath := filepath.Join(r.backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt")
	keyPath := filepath.Join(r.backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.key")

	fmt.Printf("Writing certificates to:\n")
	fmt.Printf("Certificate: %s\n", crtPath)
	fmt.Printf("Key: %s\n", keyPath)

	if err := os.WriteFile(crtPath, []byte(crtContent), 0o600); err != nil {
		return fmt.Errorf("failed to write certificate file: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte(keyContent), 0o600); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}

	fmt.Printf("✅ Certificates copied successfully:\n")
	fmt.Printf("Backup directory: %s\n", r.backupDir)
	fmt.Printf("Certificate path: %s\n", crtPath)
	fmt.Printf("Key path: %s\n", keyPath)

	return nil
}

func (r *Renewer) updateAPIServerEtcdClientSecret(ctx context.Context, clusterName string) error {
	fmt.Printf("Updating %s-apiserver-etcd-client secret...\n", clusterName)

	if err := r.ensureNamespaceExists(ctx, "eksa-system"); err != nil {
		return fmt.Errorf("ensuring eksa-system namespace exists: %v", err)
	}

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

	// get current sercet or create
	secretName := fmt.Sprintf("%s-apiserver-etcd-client", clusterName)
	secret, err := r.kubeClient.CoreV1().Secrets("eksa-system").Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get secret %s: %v", secretName, err)
		}

		// if sercet not exist, create
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: "eksa-system",
			},
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				"tls.crt": crtData,
				"tls.key": keyData,
			},
		}

		_, err = r.kubeClient.CoreV1().Secrets("eksa-system").Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create secret %s: %v", secretName, err)
		}
	} else {
		// if sercet exist, renew it
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}

		secret.Data["tls.crt"] = crtData
		secret.Data["tls.key"] = keyData

		_, err = r.kubeClient.CoreV1().Secrets("eksa-system").Update(ctx, secret, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update secret %s: %v", secretName, err)
		}
	}

	fmt.Printf("✅ Successfully updated %s secret.\n", secretName)
	return nil
}

// For workload cluster, if there is no eksa-system name space, create it.
func (r *Renewer) ensureNamespaceExists(ctx context.Context, namespace string) error {
	_, err := r.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			_, err = r.kubeClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create namespace %s: %v", namespace, err)
			}
			fmt.Printf("Created namespace %s\n", namespace)
		} else {
			return fmt.Errorf("failed to check namespace %s: %v", namespace, err)
		}
	}
	return nil
}

// for renew control panel only.
func (r *Renewer) saveCertsToPersistentStorage() error {
	srcCrt := filepath.Join(r.backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt")
	srcKey := filepath.Join(r.backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.key")

	destDir := filepath.Join(persistentCertDir, persistentEtcdCertDir)
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return fmt.Errorf("failed to create persistent directory: %v", err)
	}

	destCrt := filepath.Join(destDir, "apiserver-etcd-client.crt")
	destKey := filepath.Join(destDir, "apiserver-etcd-client.key")

	if err := copyFile(srcCrt, destCrt); err != nil {
		return fmt.Errorf("failed to copy certificate: %v", err)
	}
	if err := copyFile(srcKey, destKey); err != nil {
		return fmt.Errorf("failed to copy key: %v", err)
	}

	return nil
}

func (r *Renewer) loadCertsFromPersistentStorage() error {
	srcDir := filepath.Join(persistentCertDir, persistentEtcdCertDir)
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("no etcd certificates found in persistent storage. Please run etcd certificate renewal first")
	}

	destDir := filepath.Join(r.backupDir, tempLocalEtcdCertsDir)
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}

	srcCrt := filepath.Join(srcDir, "apiserver-etcd-client.crt")
	srcKey := filepath.Join(srcDir, "apiserver-etcd-client.key")

	destCrt := filepath.Join(destDir, "apiserver-etcd-client.crt")
	destKey := filepath.Join(destDir, "apiserver-etcd-client.key")

	if err := copyFile(srcCrt, destCrt); err != nil {
		return fmt.Errorf("failed to copy certificate: %v", err)
	}
	if err := copyFile(srcKey, destKey); err != nil {
		return fmt.Errorf("failed to copy key: %v", err)
	}

	return nil
}

func copyFile(src, dest string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if err = os.WriteFile(dest, input, 0o600); err != nil {
		return err
	}

	return nil
}
