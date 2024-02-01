package framework

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

const (
	kubectlDeleteTimeout = "20m"
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
	opts := func(params *[]string) {
		*params = append(*params, "--timeout", kubectlDeleteTimeout)
	}
	if err := w.KubectlClient.DeleteManifest(ctx, w.ManagementClusterKubeconfigFile(), w.ClusterConfigLocation, opts); err != nil {
		w.T.Fatalf("Failed to delete workload cluster config: %s", err)
	}
	w.StopIfFailed()
}

// WaitForAvailableHardware waits for workload cluster hardware to be available.
func (w *WorkloadCluster) WaitForAvailableHardware() {
	ctx := context.Background()
	w.T.Logf("Waiting for workload cluster %s hardware to be available", w.ClusterName)
	err := retrier.Retry(240, 5*time.Second, func() error {
		return w.availableHardware(ctx)
	})
	if err != nil {
		w.T.Fatalf("Failed waiting for cluster hardware: %s", err)
	}
}

// WaitForKubeconfig waits for the kubeconfig for the workload cluster to be available and then writes it to disk.
func (w *WorkloadCluster) WaitForKubeconfig() {
	ctx := context.Background()
	w.T.Logf("Waiting for workload cluster %s kubeconfig to be available", w.ClusterName)
	err := retrier.Retry(120, 5*time.Second, func() error {
		return w.writeKubeconfigToDisk(ctx, fmt.Sprintf("%s-kubeconfig", w.ClusterName), w.KubeconfigFilePath())
	})
	if err != nil {
		w.T.Fatalf("Failed waiting for cluster kubeconfig: %s", err)
	}

	if len(w.ClusterConfig.AWSIAMConfigs) != 0 {
		w.T.Logf("Waiting for workload cluster %s iam auth kubeconfig to be available", w.ClusterName)
		err := retrier.Retry(120, 5*time.Second, func() error {
			return w.writeKubeconfigToDisk(ctx, fmt.Sprintf("%s-aws-iam-kubeconfig", w.ClusterName), w.iamAuthKubeconfigFilePath())
		})
		if err != nil {
			w.T.Fatalf("Failed waiting for cluster kubeconfig: %s", err)
		}
	}

	w.T.Logf("Waiting for workload cluster %s control plane to be ready", w.ClusterName)
	if err := w.KubectlClient.WaitForControlPlaneReady(ctx, w.managementCluster(), "15m", w.ClusterName); err != nil {
		w.T.Errorf("Failed waiting for control plane ready: %s", err)
	}
}

// ValidateClusterDelete verifies the cluster has been deleted.
func (w *WorkloadCluster) ValidateClusterDelete() {
	ctx := context.Background()
	w.T.Logf("Validating cluster deletion %s", w.ClusterName)
	clusterStateValidator := newClusterStateValidator(w.clusterStateValidationConfig)
	clusterStateValidator.WithValidations(
		validationsForClusterDoesNotExist()...,
	)
	if err := clusterStateValidator.Validate(ctx); err != nil {
		w.T.Fatalf("failed to validate cluster deletion %v", err)
	}
}

func (w *WorkloadCluster) writeKubeconfigToDisk(ctx context.Context, secretName string, filePath string) error {
	secret, err := w.KubectlClient.GetSecretFromNamespace(ctx, w.ManagementClusterKubeconfigFile(), secretName, constants.EksaSystemNamespace)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig for cluster: %s", err)
	}
	kubeconfig := secret.Data["value"]
	if err := w.Provider.UpdateKubeConfig(&kubeconfig, w.ClusterName); err != nil {
		return fmt.Errorf("failed to update kubeconfig for cluster: %s", err)
	}
	writer, err := filewriter.NewWriter(filepath.Dir(filePath))
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig to disk: %v", err)
	}

	_, err = writer.Write(filepath.Base(filePath), kubeconfig, func(op *filewriter.FileOptions) {
		op.IsTemp = false
	})
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig to disk: %v", err)
	}
	return err
}

func (w *WorkloadCluster) availableHardware(ctx context.Context) error {
	hardwareList, err := w.KubectlClient.GetUnprovisionedTinkerbellHardware(ctx, w.ManagementClusterKubeconfigFile(), constants.EksaSystemNamespace)
	if err != nil {
		return fmt.Errorf("failed to get unprovisioned hardware: %s", err)
	}

	cpHardwareRequired := w.ClusterConfig.Cluster.Spec.ControlPlaneConfiguration.Count
	var workerHardwareRequired int
	for _, workerNodeGroup := range w.ClusterConfig.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerHardwareRequired += *workerNodeGroup.Count
	}

	var cpHardwareAvailable int
	var workerHardwareAvailable int

	for _, hardware := range hardwareList {
		switch hardware.Labels[api.HardwareLabelTypeKeyName] {
		case api.ControlPlane:
			cpHardwareAvailable++
		case api.Worker:
			workerHardwareAvailable++
		}
	}

	if cpHardwareAvailable < cpHardwareRequired || workerHardwareAvailable < workerHardwareRequired {
		return errors.New("Insufficient hardware available for cluster")
	}

	return nil
}
