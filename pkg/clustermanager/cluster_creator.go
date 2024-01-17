package clustermanager

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
)

// ClusterApplier is responsible for applying the cluster spec to the cluster.
type ClusterApplier interface {
	Run(ctx context.Context, spec *cluster.Spec, managementCluster types.Cluster) error
}

// ClusterCreator is responsible for applying the cluster config and writing the kubeconfig file.
type ClusterCreator struct {
	ClusterApplier
	kubeconfigWriter kubeconfig.Writer
	fs               filewriter.FileWriter
}

// NewClusterCreator creates a ClusterCreator.
func NewClusterCreator(applier ClusterApplier, kubeconfigWriter kubeconfig.Writer, fs filewriter.FileWriter) *ClusterCreator {
	return &ClusterCreator{
		ClusterApplier:   applier,
		kubeconfigWriter: kubeconfigWriter,
		fs:               fs,
	}
}

// CreateSync creates a workload cluster using the EKS-A controller and returns the types.Cluster object for that cluster.
func (cc ClusterCreator) CreateSync(ctx context.Context, spec *cluster.Spec, managementCluster *types.Cluster) (*types.Cluster, error) {
	if err := cc.Run(ctx, spec, *managementCluster); err != nil {
		return nil, err
	}

	return cc.buildClusterAccess(ctx, spec.Cluster.Name, managementCluster)
}

func (cc ClusterCreator) buildClusterAccess(ctx context.Context, clusterName string, management *types.Cluster) (*types.Cluster, error) {
	cluster := &types.Cluster{
		Name: clusterName,
	}

	fsOptions := []filewriter.FileOptionsFunc{filewriter.PersistentFile, filewriter.Permission0600}
	fh, path, err := cc.fs.Create(
		kubeconfig.FormatWorkloadClusterKubeconfigFilename(clusterName),
		fsOptions...,
	)
	if err != nil {
		return nil, err
	}

	defer fh.Close()

	err = cc.kubeconfigWriter.WriteKubeconfig(ctx, clusterName, management.KubeconfigFile, fh)
	if err != nil {
		return nil, err
	}

	cluster.KubeconfigFile = path

	return cluster, nil
}
