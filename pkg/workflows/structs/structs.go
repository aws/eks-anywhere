package structs

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

// ClusterCreate creates a Kubernetes cluster that is immediately usable for simple workloads.
type ClusterCreate struct {
	Cluster        interfaces.Cluster
	ClusterCreator interfaces.ClusterCreator

	// FS is a file system abstraction providing file creation and write capabilities.
	FS filewriter.FileWriter
}

// GetWorkloadCluster returns the types.Cluster object for a newly created workload cluster.
func (cc ClusterCreate) GetWorkloadCluster(ctx context.Context, clusterSpec *cluster.Spec, management *types.Cluster, provider providers.Provider) (*types.Cluster, error) {
	clusterName := clusterSpec.Cluster.Name

	workloadCluster := &types.Cluster{
		Name:               clusterName,
		ExistingManagement: management.ExistingManagement,
	}

	var buf bytes.Buffer

	if err := cc.Cluster.WriteKubeconfig(ctx, &buf, management); err != nil {
		return nil, err
	}

	rawKubeconfig := buf.Bytes()

	// The Docker provider wants to update the kubeconfig to patch the server address before
	// we write it to disk. This is to ensure we can communicate with the cluster even when
	// hosted inside a Docker Desktop VM.
	if err := provider.UpdateKubeConfig(&rawKubeconfig, clusterName); err != nil {
		return nil, err
	}

	kubeconfigFile, err := cc.FS.Write(
		kubeconfig.FormatWorkloadClusterKubeconfigFilename(clusterName),
		rawKubeconfig,
		filewriter.PersistentFile,
		filewriter.Permission0600,
	)
	if err != nil {
		return nil, fmt.Errorf("writing workload kubeconfig: %v", err)
	}
	workloadCluster.KubeconfigFile = kubeconfigFile

	return workloadCluster, nil
}
