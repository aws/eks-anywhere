package certificates

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	tempLocalEtcdCertsDir = "etcd-client-certs"

	ubuntuEtcdCertDir           = "/etc/etcd"
	ubuntuControlPlaneCertDir   = "/etc/kubernetes/pki"
	ubuntuControlPlaneManifests = "/etc/kubernetes/manifests"

	bottlerocketEtcdCertDir         = "/var/lib/etcd"
	bottlerocketControlPlaneCertDir = "/var/lib/kubeadm/pki"
	bottlerocketTmpDir              = "/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp"

	componentEtcd         = "etcd"
	componentControlPlane = "control-plane"
)

type sshDialer func(network, addr string, config *ssh.ClientConfig) (sshClient, error)

type Renewer struct {
	backupDir  string
	sshConfig  *ssh.ClientConfig
	sshKeyPath string // store SSH key path from config
	kubeClient kubernetes.Interface
	sshDialer  sshDialer
}

func NewRenewer() (*Renewer, error) {
	backupDate := time.Now().Format("20060102_150405")
	backupDir := fmt.Sprintf("certificate_backup_%s", backupDate)
	fmt.Printf("Creating backup directory: %s\n", backupDir)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %v", err)
	}

	etcdCertsPath := filepath.Join(backupDir, tempLocalEtcdCertsDir)
	fmt.Printf("Creating etcd certs directory: %s\n", etcdCertsPath)

	if err := os.MkdirAll(etcdCertsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create etcd certs directory: %v", err)
	}

	r := &Renewer{
		backupDir: backupDir,
		sshDialer: func(network, addr string, config *ssh.ClientConfig) (sshClient, error) {
			return ssh.Dial(network, addr, config)
		},
	}
	return r, nil
}

func (r *Renewer) RenewCertificates(ctx context.Context, cluster *types.Cluster, config *RenewalConfig, component string) error {
	if component != "" && component != componentEtcd && component != componentControlPlane {
		return fmt.Errorf("invalid component %q, must be either %q or %q", component, componentEtcd, componentControlPlane)
	}

	fmt.Printf("✅ Checking if Kubernetes API server is reachable...\n")
	if err := r.initKubeClient(); err != nil {
		return fmt.Errorf("failed to initialize kubernetes client: %v", err)
	}

	if err := r.checkAPIServerReachability(ctx); err != nil {
		return fmt.Errorf("API server health check failed: %v", err)
	}

	fmt.Printf("✅ Backing up kubeadm-config ConfigMap...\n")
	if err := r.backupKubeadmConfig(ctx); err != nil {
		return fmt.Errorf("failed to backup kubeadm config: %v", err)
	}

	if component == componentEtcd || component == "" {
		if len(config.Etcd.Nodes) > 0 {
			fmt.Printf("Starting etcd certificate renewal process...\n")
			if err := r.renewEtcdCerts(ctx, config); err != nil {
				return fmt.Errorf("failed to renew etcd certificates: %v", err)
			}
			fmt.Printf("🎉 Etcd certificate renewal process completed successfully.\n")
		} else {
			fmt.Printf("Cluster does not have external ETCD.\n")
		}
	}

	if component == componentControlPlane || component == "" {
		if len(config.ControlPlane.Nodes) == 0 {
			return fmt.Errorf("❌ Error: No control plane node IPs found")
		}
		fmt.Printf("Starting control plane certificate renewal process...\n")
		if err := r.renewControlPlaneCerts(ctx, config, component); err != nil {
			return fmt.Errorf("failed to renew control plane certificates: %v", err)
		}
		fmt.Printf("🎉 Control plane certificate renewal process completed successfully.\n")
	}

	fmt.Printf("✅ Cleaning up temporary files...\n")
	if err := r.cleanup(); err != nil {
		fmt.Printf("❌ API server unreachable — skipping cleanup to preserve debug data.\n")
		return err
	}
	fmt.Printf("✅ All temporary files removed.\n")
	return nil
}

func (r *Renewer) initKubeClient() error {
	if r.kubeClient != nil {
		return nil
	}

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	r.kubeClient = clientset
	return nil
}

func (r *Renewer) checkAPIServerReachability(ctx context.Context) error {
	for i := 0; i < 5; i++ {
		_, err := r.kubeClient.Discovery().ServerVersion()
		if err == nil {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("kubernetes API server is not reachable")
}

func (r *Renewer) backupKubeadmConfig(ctx context.Context) error {
	cm, err := r.kubeClient.CoreV1().ConfigMaps("kube-system").Get(ctx, "kubeadm-config", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get kubeadm-config: %v", err)
	}

	backupPath := filepath.Join(r.backupDir, "kubeadm-config.yaml")
	if err := os.WriteFile(backupPath, []byte(cm.Data["ClusterConfiguration"]), 0600); err != nil {
		return fmt.Errorf("failed to write kubeadm config backup: %v", err)
	}

	return nil
}

func (r *Renewer) renewEtcdCerts(ctx context.Context, config *RenewalConfig) error {

	if err := r.initSSHConfig(config.Etcd.SSHUser, config.Etcd.SSHKey, config.Etcd.SSHPasswd); err != nil {
		return fmt.Errorf("failed to initialize SSH config: %v", err)
	}

	for _, node := range config.Etcd.Nodes {
		if err := r.renewEtcdNodeCerts(ctx, node, config.Etcd); err != nil {
			return fmt.Errorf("failed to renew certificates for etcd node %s: %v", node, err)
		}
	}

	if err := r.updateAPIServerEtcdClientSecret(ctx, config.ClusterName); err != nil {
		return fmt.Errorf("failed to update apiserver-etcd-client secret: %v", err)
	}

	return nil
}

func (r *Renewer) renewControlPlaneCerts(ctx context.Context, config *RenewalConfig, component string) error {
	if err := r.initSSHConfig(config.ControlPlane.SSHUser, config.ControlPlane.SSHKey, config.ControlPlane.SSHPasswd); err != nil {
		return fmt.Errorf("failed to initialize SSH config: %v", err)
	}

	// Renew certificate for each control plane node
	for _, node := range config.ControlPlane.Nodes {
		if err := r.renewControlPlaneNodeCerts(ctx, node, config, component); err != nil {
			return fmt.Errorf("failed to renew certificates for control plane node %s: %v", node, err)
		}
	}

	return nil
}

func (r *Renewer) initSSHConfig(user, keyPath string, passwd string) error {
	r.sshKeyPath = keyPath // Store SSH key path
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read SSH key: %v", err)
	}

	var signer ssh.Signer
	signer, err = ssh.ParsePrivateKey(key)
	if err != nil {
		if err.Error() == "ssh: this private key is passphrase protected" {
			if passwd == "" {
				fmt.Printf("Enter passphrase for SSH key '%s': ", keyPath)
				var passphrase []byte
				passphrase, err = term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return fmt.Errorf("failed to read passphrase: %v", err)
				}
				fmt.Println() // Print newline after password input
				passwd = string(passphrase)
			}
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passwd))
			if err != nil {
				return fmt.Errorf("failed to parse SSH key with passphrase: %v", err)
			}
		} else {
			return fmt.Errorf("failed to parse SSH key: %v", err)
		}
	}

	r.sshConfig = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	return nil
}

func (r *Renewer) renewEtcdNodeCerts(ctx context.Context, node string, config NodeConfig) error {
	switch config.OS {
	case "ubuntu", "rhel", "redhat":
		return r.renewEtcdCertsLinux(ctx, node)
	case "bottlerocket":
		return r.renewEtcdCertsBottlerocket(ctx, node)
	default:
		return fmt.Errorf("unsupported OS: %s", config.OS)
	}
}

func (r *Renewer) renewControlPlaneNodeCerts(ctx context.Context, node string, config *RenewalConfig, component string) error {
	switch config.ControlPlane.OS {
	case "ubuntu", "rhel", "redhat":
		return r.renewControlPlaneCertsLinux(ctx, node, config, component)
	case "bottlerocket":
		return r.renewControlPlaneCertsBottlerocket(ctx, node, config, component)
	default:
		return fmt.Errorf("unsupported OS: %s", config.ControlPlane.OS)
	}
}

func (r *Renewer) runCommand(ctx context.Context, client sshClient, cmd string) error {
	done := make(chan error, 1)
	go func() {
		session, err := client.NewSession()
		if err != nil {
			done <- fmt.Errorf("failed to create session: %v", err)
			return
		}
		defer session.Close()
		// print shell session progress
		session.Stdout = os.Stdout
		session.Stderr = os.Stderr

		done <- session.Run(cmd)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("command cancelled: %v", ctx.Err())
	case err := <-done:
		if err != nil {
			return fmt.Errorf("command failed: %v", err)
		}
		return nil
	}
}

func (r *Renewer) runCommandWithOutput(ctx context.Context, client sshClient, cmd string) (string, error) {
	type result struct {
		output string
		err    error
	}
	done := make(chan result, 1)

	go func() {
		session, err := client.NewSession()
		if err != nil {
			done <- result{"", fmt.Errorf("failed to create session: %v", err)}
			return
		}
		defer session.Close()

		output, err := session.Output(cmd)
		if err != nil {
			done <- result{"", fmt.Errorf("command failed: %v", err)}
			return
		}
		done <- result{strings.TrimSpace(string(output)), nil}
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("command cancelled: %v", ctx.Err())
	case res := <-done:
		return res.output, res.err
	}
}

func (r *Renewer) cleanup() error {
	fmt.Printf("Cleaning up directory: %s\n", r.backupDir)

	chmodCmd := exec.Command("chmod", "-R", "u+w", r.backupDir)
	if err := chmodCmd.Run(); err != nil {
		return fmt.Errorf("failed to change permissions: %v", err)
	}

	return os.RemoveAll(r.backupDir)
}
