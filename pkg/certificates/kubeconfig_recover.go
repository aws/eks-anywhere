package certificates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/eks-anywhere/pkg/logger"
)

type kubeconfigFetcher interface {
	RemoteScript(cluster string) string
	KubeconfigTempPath() string
}

type linuxKcfgFetcher struct{}

func (linuxKcfgFetcher) RemoteScript(cluster string) string {
	return fmt.Sprintf(`set -euo pipefail
export CLUSTER_NAME="%s"
export KUBECONFIG="/var/lib/kubeadm/admin.conf"
kubectl get secret ${CLUSTER_NAME}-kubeconfig -n eksa-system \
 -o jsonpath="{.data.value}" | base64 --decode > /tmp/user-admin.kubeconfig`,
		cluster)
}

func (linuxKcfgFetcher) KubeconfigTempPath() string { return "/tmp/user-admin.kubeconfig" }

// Bottlrtocket fecther.
type brKcfgFetcher struct{}

func brScriptCopyAdminConf() string {
	return `sudo sheltie <<'EOF'
cat /var/lib/kubeadm/admin.conf > \
/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/kubernetes-admin.kubeconfig
exit
EOF`
}

func brScriptEnsureKubectl() string {
	return `command -v kubectl >/dev/null 2>&1 || { \
curl -Lo /usr/local/bin/kubectl \
https://dl.k8s.io/release/$(curl -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl && \
chmod +x /usr/local/bin/kubectl; }`
}

func (brKcfgFetcher) RemoteScript(cluster string) string {
	return fmt.Sprintf(`set -euo pipefail
%s
export CLUSTER_NAME="%s"
export KUBECONFIG="/tmp/kubernetes-admin.kubeconfig"
%s
kubectl get secret ${CLUSTER_NAME}-kubeconfig -n eksa-system \
 -o jsonpath="{.data.value}" | base64 --decode > /tmp/user-admin.kubeconfig`,
		brScriptCopyAdminConf(),
		cluster,
		brScriptEnsureKubectl(),
	)
}

func (brKcfgFetcher) KubeconfigTempPath() string { return "/tmp/user-admin.kubeconfig" }

// RecoverExpiredKubeconfig recovers an expired kubeconfig from the cluster's control plane nodes.
func RecoverExpiredKubeconfig(ctx context.Context, cfg *RenewalConfig, ssh SSHRunner) (string, error) {
	if len(cfg.ControlPlane.Nodes) == 0 {
		return "", fmt.Errorf("no control-plane nodes in config")
	}

	// choose OS fetcher
	var fetcher kubeconfigFetcher
	// switch cfg.ControlPlane.OS {
	switch cfg.OS {
	case string(OSTypeBottlerocket):
		fetcher = brKcfgFetcher{}
	default:
		fetcher = linuxKcfgFetcher{}
	}

	// 2 minutes max time
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	for _, node := range cfg.ControlPlane.Nodes {
		logger.Info("Attempting kubeconfig recovery", "node", node)

		// 1) create /tmp/*.kubeconfig in node
		if err := ssh.RunCommand(ctx, node, fetcher.RemoteScript(cfg.ClusterName)); err != nil {
			logger.Info("Node failed, trying next", "node", node, "error", err)
			continue
		}

		// Download to local
		tmpDir, _ := os.MkdirTemp("", "eks-a-kubeconfig-*")
		dst := filepath.Join(tmpDir, "user-admin.kubeconfig")
		if err := ssh.DownloadFile(ctx, node, fetcher.KubeconfigTempPath(), dst); err != nil {
			logger.Info("Download failed, trying next", "node", node, "error", err)
			continue
		}

		logger.MarkPass("Recovered kubeconfig from node", "node", node, "path", dst)
		return dst, nil
	}
	return "", fmt.Errorf("failed to recover kubeconfig from all control-plane nodes")
}
