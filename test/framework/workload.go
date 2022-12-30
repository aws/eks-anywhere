package framework

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

type WorkloadCluster struct {
	*ClusterE2ETest
	ManagementClusterKubeconfigFile func() string
}

type WorkloadClusters map[string]*WorkloadCluster

func (w *WorkloadCluster) CreateCluster(opts ...CommandOpt) {
	opts = append(opts, withKubeconfig(w.ManagementClusterKubeconfigFile()))
	w.createCluster(opts...)
}

func (w *WorkloadCluster) UpgradeCluster(clusterOpts []ClusterE2ETestOpt, commandOpts ...CommandOpt) {
	commandOpts = append(commandOpts, withKubeconfig(w.ManagementClusterKubeconfigFile()))
	w.upgradeCluster(clusterOpts, commandOpts...)
}

func (w *WorkloadCluster) DeleteCluster(opts ...CommandOpt) {
	opts = append(opts, withKubeconfig(w.ManagementClusterKubeconfigFile()))
	w.deleteCluster(opts...)
}

// ApplyClusterManifest uses client-side logic to create/update objects defined in a cluster yaml manifest.
func (w *WorkloadCluster) ApplyClusterManifest() {
	ctx := context.Background()
	w.T.Logf("Applying workload cluster %s spec located at %s", w.ClusterName, w.ClusterConfigLocation)
	if err := w.KubectlClient.ApplyManifest(ctx, w.ManagementClusterKubeconfigFile(), w.ClusterConfigLocation); err != nil {
		w.T.Fatalf("Failed to apply workload cluster config: %s", err)
	}
	w.StopIfFailed()
}

// DeleteClusterWithKubectl uses client-side logic to delete a cluster.
func (w *WorkloadCluster) DeleteClusterWithKubectl() {
	ctx := context.Background()
	w.T.Logf("Deleting workload cluster %s with kubectl", w.ClusterName)
	if err := w.KubectlClient.DeleteCluster(ctx, w.managementCluster(), w.cluster()); err != nil {
		w.T.Fatalf("Failed to delete workload cluster config: %s", err)
	}
	w.StopIfFailed()
}

// WaitForKubeconfig waits for the kubeconfig for the workload cluster to be available and then writes it to disk.
func (w *WorkloadCluster) WaitForKubeconfig() {
	ctx := context.Background()
	w.T.Logf("Waiting for workload cluster %s kubeconfig to be available", w.ClusterName)
	err := retrier.Retry(60, 5*time.Second, func() error {
		return w.writeKubeconfigToDisk(ctx)
	})
	if err != nil {
		w.T.Fatalf("Failed waiting for cluster kubeconfig: %s", err)
	}
}

func (w *WorkloadCluster) writeKubeconfigToDisk(ctx context.Context) error {
	secret, err := w.KubectlClient.GetSecretFromNamespace(ctx, w.ManagementClusterKubeconfigFile(), fmt.Sprintf("%s-kubeconfig", w.ClusterName), constants.EksaSystemNamespace)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig for cluster: %s", err)
	}
	kubeconfig := secret.Data["value"]
	writer, err := filewriter.NewWriter(w.ClusterConfigFolder)
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig to disk: %v", err)
	}

	_, err = writer.Write(filepath.Base(w.kubeconfigFilePath()), kubeconfig, func(op *filewriter.FileOptions) {
		op.IsTemp = false
	})
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig to disk: %v", err)
	}
	return err
}
