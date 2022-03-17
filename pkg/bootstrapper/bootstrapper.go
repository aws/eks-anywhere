package bootstrapper

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Bootstrapper struct {
	clusterClient ClusterClient
}

type ClusterClient interface {
	CreateBootstrapCluster(ctx context.Context, clusterSpec *cluster.Spec, opts ...BootstrapClusterClientOption) (kubeconfig string, err error)
	DeleteBootstrapCluster(ctx context.Context, cluster *types.Cluster) error
	WithExtraDockerMounts() BootstrapClusterClientOption
	WithEnv(env map[string]string) BootstrapClusterClientOption
	WithDefaultCNIDisabled() BootstrapClusterClientOption
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	GetClusters(ctx context.Context, cluster *types.Cluster) ([]types.CAPICluster, error)
	GetKubeconfig(ctx context.Context, clusterName string) (string, error)
	ClusterExists(ctx context.Context, clusterName string) (bool, error)
	ValidateClustersCRD(ctx context.Context, cluster *types.Cluster) error
	CreateNamespace(ctx context.Context, kubeconfig string, namespace string) error
	GetNamespace(ctx context.Context, kubeconfig string, namespace string) error
}

type (
	BootstrapClusterClientOption func() error
	BootstrapClusterOption       func(b *Bootstrapper) BootstrapClusterClientOption
)

func New(clusterClient ClusterClient) *Bootstrapper {
	return &Bootstrapper{
		clusterClient: clusterClient,
	}
}

func (b *Bootstrapper) CreateBootstrapCluster(ctx context.Context, clusterSpec *cluster.Spec, opts ...BootstrapClusterOption) (*types.Cluster, error) {
	kubeconfigFile, err := b.clusterClient.CreateBootstrapCluster(ctx, clusterSpec, b.getClientOptions(opts)...)
	if err != nil {
		return nil, fmt.Errorf("error creating bootstrap cluster: %v, try rerunning with --force-cleanup to force delete previously created bootstrap cluster", err)
	}

	c := &types.Cluster{
		Name:           clusterSpec.Cluster.Name,
		KubeconfigFile: kubeconfigFile,
	}

	err = b.clusterClient.GetNamespace(ctx, c.KubeconfigFile, constants.EksaSystemNamespace)
	if err != nil {
		if err := b.clusterClient.CreateNamespace(ctx, c.KubeconfigFile, constants.EksaSystemNamespace); err != nil {
			return nil, err
		}
	}

	err = cluster.ApplyExtraObjects(ctx, b.clusterClient, c, clusterSpec)
	if err != nil {
		return nil, fmt.Errorf("error applying extra objects to bootstrap cluster: %v", err)
	}

	return c, nil
}

func (b *Bootstrapper) DeleteBootstrapCluster(ctx context.Context, cluster *types.Cluster, isUpgrade bool) error {
	clusterExists, err := b.clusterClient.ClusterExists(ctx, cluster.Name)
	if err != nil {
		return fmt.Errorf("error deleting bootstrap cluster: %v", err)
	}
	if !clusterExists {
		logger.V(4).Info("Skipping delete bootstrap cluster, cluster doesn't exist")
		return nil
	}
	mgmtCluster, err := b.managementInCluster(ctx, cluster)
	if err != nil {
		return fmt.Errorf("error deleting bootstrap cluster: %v", err)
	}

	if mgmtCluster != nil {
		if isUpgrade || mgmtCluster.Status.Phase == "Provisioned" {
			return errors.New("error deleting bootstrap cluster: management cluster in bootstrap cluster")
		}
	}

	return b.clusterClient.DeleteBootstrapCluster(ctx, cluster)
}

func (b *Bootstrapper) managementInCluster(ctx context.Context, cluster *types.Cluster) (*types.CAPICluster, error) {
	if cluster.KubeconfigFile == "" {
		kubeconfig, err := b.clusterClient.GetKubeconfig(ctx, cluster.Name)
		if err != nil {
			return nil, fmt.Errorf("error fetching bootstrap cluster's kubeconfig: %v", err)
		}
		cluster.KubeconfigFile = kubeconfig
	}
	err := b.clusterClient.ValidateClustersCRD(ctx, cluster)
	if err == nil {
		clusters, err := b.clusterClient.GetClusters(ctx, cluster)
		if err != nil {
			return nil, err
		}
		if len(clusters) != 0 {
			return &clusters[0], nil
		}
	}
	return nil, nil
}

func (b *Bootstrapper) getClientOptions(opts []BootstrapClusterOption) []BootstrapClusterClientOption {
	clientOpts := make([]BootstrapClusterClientOption, 0, len(opts))

	for _, o := range opts {
		clientOpts = append(clientOpts, o(b))
	}

	return clientOpts
}

func WithExtraDockerMounts() BootstrapClusterOption {
	return func(b *Bootstrapper) BootstrapClusterClientOption {
		return b.clusterClient.WithExtraDockerMounts()
	}
}

func WithEnv(env map[string]string) BootstrapClusterOption {
	return func(b *Bootstrapper) BootstrapClusterClientOption {
		return b.clusterClient.WithEnv(env)
	}
}

func WithDefaultCNIDisabled() BootstrapClusterOption {
	return func(b *Bootstrapper) BootstrapClusterClientOption {
		return b.clusterClient.WithDefaultCNIDisabled()
	}
}
