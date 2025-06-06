package certificates

import (
	"context"
	"fmt"
)

func (r *Renewer) renewControlPlaneCertsLinux(ctx context.Context, node string, config *RenewalConfig, component string) error {
	fmt.Printf("Processing control plane node: %s...\n", node)
	client, err := r.sshDialer("tcp", fmt.Sprintf("%s:22", node), r.sshConfig)
	if err != nil {
		return fmt.Errorf("connecting to node %s: %v", node, err)
	}
	defer client.Close()

	// Backup certificates, excluding etcd directory if component is control-plane
	var backupCmd string
	if component == componentControlPlane && len(config.Etcd.Nodes) > 0 {
		// When only updating control plane with external etcd, exclude etcd directory
		backupCmd = fmt.Sprintf(`
sudo mkdir -p '/etc/kubernetes/pki.bak_%[1]s'
cd %[2]s
for f in $(find . -type f ! -path './etcd/*'); do
    sudo mkdir -p $(dirname '/etc/kubernetes/pki.bak_%[1]s/'$f)
    sudo cp $f '/etc/kubernetes/pki.bak_%[1]s/'$f
done`, r.backupDir, ubuntuControlPlaneCertDir)
	} else {
		backupCmd = fmt.Sprintf("sudo cp -r '%s' '/etc/kubernetes/pki.bak_%s'",
			ubuntuControlPlaneCertDir, r.backupDir)
	}
	if err := r.runCommand(ctx, client, backupCmd); err != nil {
		return fmt.Errorf("backing up certificates: %v", err)
	}

	// Renew certificates
	fmt.Printf("Renewing certificates on node %s...\n", node)
	renewCmd := "sudo kubeadm certs renew all"
	if component == componentControlPlane && len(config.Etcd.Nodes) > 0 {
		// When only renewing control plane certs with external etcd,
		// we need to skip the etcd directory to preserve certificates
		renewCmd = `for cert in admin.conf apiserver apiserver-kubelet-client controller-manager.conf front-proxy-client scheduler.conf; do
            sudo kubeadm certs renew $cert
        done`
	}
	if err := r.runCommand(ctx, client, renewCmd); err != nil {
		return fmt.Errorf("renewing certificates: %v", err)
	}

	// Validate certificates
	fmt.Printf("Validating certificates on node %s...\n", node)
	validateCmd := "sudo kubeadm certs check-expiration"
	if err := r.runCommand(ctx, client, validateCmd); err != nil {
		return fmt.Errorf("validating certificates: %v", err)
	}

	// Restart
	fmt.Printf("Restarting control plane components on node %s...\n", node)
	restartCmd := fmt.Sprintf("sudo mkdir -p /tmp/manifests && "+
		"sudo mv %s/* /tmp/manifests/ && "+
		"sleep 20 && "+
		"sudo mv /tmp/manifests/* %s/",
		ubuntuControlPlaneManifests, ubuntuControlPlaneManifests)
	if err := r.runCommand(ctx, client, restartCmd); err != nil {
		return fmt.Errorf("restarting control plane components: %v", err)
	}

	fmt.Printf("✅ Completed renewing certificate for the control node: %s.\n", node)
	fmt.Printf("---------------------------------------------\n")
	return nil
}

func (r *Renewer) renewEtcdCertsLinux(ctx context.Context, node string) error {
	fmt.Printf("Processing etcd node: %s...\n", node)
	client, err := r.sshDialer("tcp", fmt.Sprintf("%s:22", node), r.sshConfig)
	if err != nil {
		return fmt.Errorf("connecting to node %s: %v", node, err)
	}
	defer client.Close()

	// Backup certificates
	fmt.Printf("# Backup certificates\n")
	backupCmd := fmt.Sprintf("cd %s && sudo cp -r pki pki.bak_%s && sudo rm -rf pki/* && sudo cp pki.bak_%s/ca.* pki/",
		ubuntuEtcdCertDir, r.backupDir, r.backupDir)
	if err := r.runCommand(ctx, client, backupCmd); err != nil {
		return fmt.Errorf("backing up certificates: %v", err)
	}

	// Renew certificates
	fmt.Printf("# Renew certificates\n")
	renewCmd := "sudo etcdadm join phase certificates http://eks-a-etcd-dumb-url"
	if err := r.runCommand(ctx, client, renewCmd); err != nil {
		return fmt.Errorf("renewing certificates: %v", err)
	}

	// Validate certificates
	fmt.Printf("# Validate certificates\n")
	validateCmd := fmt.Sprintf("sudo etcdctl --cacert=%s/pki/ca.crt "+
		"--cert=%s/pki/etcdctl-etcd-client.crt "+
		"--key=%s/pki/etcdctl-etcd-client.key "+
		"endpoint health",
		ubuntuEtcdCertDir, ubuntuEtcdCertDir, ubuntuEtcdCertDir)
	if err := r.runCommand(ctx, client, validateCmd); err != nil {
		return fmt.Errorf("validating certificates: %v", err)
	}

	// Copy certificates to local
	fmt.Printf("Copying certificates from node %s...\n", node)
	if err := r.copyEtcdCerts(ctx, client, node); err != nil {
		return fmt.Errorf("copying certificates: %v", err)
	}

	fmt.Printf("✅ Completed renewing certificate for the ETCD node: %s.\n", node)
	fmt.Printf("---------------------------------------------\n")
	return nil
}
