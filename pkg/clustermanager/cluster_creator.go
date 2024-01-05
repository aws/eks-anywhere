package clustermanager

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

type ClusterCreator struct {
	Applier Applier
	// FS is a file system abstraction providing file creation and write capabilities.
	FS filewriter.FileWriter
}

func (cc ClusterCreator) GetKubeconfig(ctx context.Context, clusterSpec *cluster.Spec, managementCluster *types.Cluster) ([]byte, error) {
	kubeconfigSecret := &corev1.Secret{}

	err := retrier.New(
		cc.Applier.applyClusterTimeout,
		retrier.WithRetryPolicy(retrier.BackOffPolicy(cc.Applier.retryBackOff)),
	).Retry(func() error {
		client, err := cc.Applier.clientFactory.BuildClientFromKubeconfig(managementCluster.KubeconfigFile)
		if err != nil {
			return err
		}

		err = client.Get(ctx, fmt.Sprintf("%s-kubeconfig", clusterSpec.Cluster.Name), constants.EksaSystemNamespace, kubeconfigSecret)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return kubeconfigSecret.Data["value"], nil
}

func (cc ClusterCreator) CreateSync(ctx context.Context, spec *cluster.Spec, managementCluster *types.Cluster, provider providers.Provider) (*types.Cluster, error) {
	err := cc.Applier.Run(ctx, spec, *managementCluster)
	if err != nil {
		return nil, err
	}

	return cc.getWorkloadCluster(ctx, spec, managementCluster, provider)
}

func (cc ClusterCreator) getWorkloadCluster(ctx context.Context, clusterSpec *cluster.Spec, management *types.Cluster, provider providers.Provider) (*types.Cluster, error) {
	clusterName := clusterSpec.Cluster.Name

	workloadCluster := &types.Cluster{
		Name:               clusterName,
		ExistingManagement: management.ExistingManagement,
	}

	rawkubeconfig, err := cc.GetKubeconfig(ctx, clusterSpec, management)
	if err != nil {
		return nil, err
	}

	kubeconfigPath, err := cc.WriteKubeconfig(rawkubeconfig, clusterName, provider)
	if err != nil {
		return nil, err
	}
	workloadCluster.KubeconfigFile = kubeconfigPath

	return workloadCluster, nil
}

func (cc ClusterCreator) WriteKubeconfig(rawkubeconfig []byte, clusterName string, provider providers.Provider) (string, error) {
	err := provider.UpdateKubeConfig(&rawkubeconfig, clusterName)
	if err != nil {
		return "", err
	}

	kubeconfigPath, err := cc.FS.Write(
		kubeconfig.FormatWorkloadClusterKubeconfigFilename(clusterName),
		rawkubeconfig,
		filewriter.PersistentFile,
		filewriter.Permission0600,
	)

	if err != nil {
		return "", err
	}

	return kubeconfigPath, nil
}
