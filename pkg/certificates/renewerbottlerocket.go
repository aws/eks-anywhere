package certificates

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/certificates/bottlerocket"
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

	builder := bottlerocket.NewControlPlaneCommandBuilder(
		r.backupDir,
		bottlerocketControlPlaneCertDir,
		component,
		len(config.Etcd.Nodes) > 0,
	)
	commands := builder.BuildCommands()
	session := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\nEOF",
		commands.ShelliePrefix,
		commands.BackupCerts,
		commands.ImagePull,
		commands.RenewCerts,
		commands.CopyCerts,
		commands.RestartPods)

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

	builder := bottlerocket.NewCertTransferBuilder(tempLocalEtcdCertsDir, crtBase64, keyBase64)
	commands := builder.BuildCommands()
	session := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\nEOF",
		commands.ShelliePrefix,
		commands.CreateDir,
		commands.WriteCertificate,
		commands.WriteKey,
		commands.SetPermissions)

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

	builder := bottlerocket.NewEtcdCommandBuilder(r.backupDir, bottlerocketTmpDir)
	commands := builder.BuildCommands()

	// first session: backup and renew certificates
	firstSession := fmt.Sprintf("%s\n%s\n%s\n%s\nEOF",
		commands.ShelliePrefix,
		commands.ImagePull,
		commands.BackupCerts,
		commands.RenewCerts)

	if err := r.runCommand(ctx, client, firstSession); err != nil {
		return fmt.Errorf("failed to renew certificates: %v", err)
	}

	// second sheltie session for copying certs
	secondSession := fmt.Sprintf("%s\n%s\nEOF",
		commands.ShelliePrefix,
		commands.CopyCerts)

	if err := r.runCommand(ctx, client, secondSession); err != nil {
		return fmt.Errorf("failed to copy certificates2 to tmp: %v", err)
	}

	// copy certificates to local
	fmt.Printf("Copying certificates from node %s...\n", node)
	if err := r.copyEtcdCerts(ctx, client, node); err != nil {
		return fmt.Errorf("failed to copy certificates3: %v", err)
	}

	// third sheltie session for cleanup
	thirdSession := fmt.Sprintf("%s\n%s\nEOF",
		commands.ShelliePrefix,
		commands.Cleanup)

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

	builder := bottlerocket.NewCertReadBuilder(bottlerocketTmpDir)
	commands := builder.BuildCommands()

	if err := r.runCommand(ctx, client, commands.ListFiles); err != nil {
		return fmt.Errorf("failed to list certificate files: %v", err)
	}

	crtContent, err := r.runCommandWithOutput(ctx, client, commands.ReadCert)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %v", err)
	}

	if len(crtContent) == 0 {
		return fmt.Errorf("certificate file is empty")
	}

	fmt.Printf("Reading key from ETCD node %s...\n", node)
	keyContent, err := r.runCommandWithOutput(ctx, client, commands.ReadKey)
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

	if err := os.WriteFile(crtPath, []byte(crtContent), 0600); err != nil {
		return fmt.Errorf("failed to write certificate file: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte(keyContent), 0600); err != nil {
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

// for renew control panel only
func (r *Renewer) saveCertsToPersistentStorage() error {
	srcCrt := filepath.Join(r.backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt")
	srcKey := filepath.Join(r.backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.key")

	destDir := filepath.Join(persistentCertDir, persistentEtcdCertDir)
	if err := os.MkdirAll(destDir, 0700); err != nil {
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
	if err := os.MkdirAll(destDir, 0700); err != nil {
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

	if err = os.WriteFile(dest, input, 0600); err != nil {
		return err
	}

	return nil
}
