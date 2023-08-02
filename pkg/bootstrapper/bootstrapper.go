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
	Apply(ctx context.Context, cluster *types.Cluster, data []byte) error
	CreateNamespace(ctx context.Context, kubeconfig, namespace string) error
	GetCAPIClusterCRD(ctx context.Context, cluster *types.Cluster) error
	GetCAPIClusters(ctx context.Context, cluster *types.Cluster) ([]types.CAPICluster, error)
	KindClusterExists(ctx context.Context, clusterName string) (bool, error)
	GetKindClusterKubeconfig(ctx context.Context, clusterName string) (string, error)
	CreateBootstrapCluster(ctx context.Context, clusterSpec *cluster.Spec, opts ...BootstrapClusterClientOption) (string, error)
	DeleteKindCluster(ctx context.Context, cluster *types.Cluster) error
	WithExtraDockerMounts() BootstrapClusterClientOption
	WithExtraPortMappings([]int) BootstrapClusterClientOption
	WithEnv(env map[string]string) BootstrapClusterClientOption
}

type (
	BootstrapClusterClientOption func() error
	BootstrapClusterOption       func(b *Bootstrapper) BootstrapClusterClientOption
)

// New constructs a new bootstrapper.
func New(clusterClient ClusterClient) *Bootstrapper {
	return &Bootstrapper{
		clusterClient: clusterClient,
	}
}

func (b *Bootstrapper) CreateBootstrapCluster(ctx context.Context, clusterSpec *cluster.Spec, opts ...BootstrapClusterOption) (*types.Cluster, error) {
	kubeconfigFile, err := b.clusterClient.CreateBootstrapCluster(ctx, clusterSpec, b.getClientOptions(opts)...)
	if err != nil {
		return nil, fmt.Errorf("creating bootstrap cluster: %v", err)
	}

	c := &types.Cluster{
		Name:           clusterSpec.Cluster.Name,
		KubeconfigFile: kubeconfigFile,
	}

	if err = b.clusterClient.CreateNamespace(ctx, c.KubeconfigFile, constants.EksaSystemNamespace); err != nil {
		return nil, err
	}

	return c, nil
}

func (b *Bootstrapper) DeleteBootstrapCluster(ctx context.Context, cluster *types.Cluster, operationType constants.Operation, isForceCleanup bool) error {
	clusterExists, err := b.clusterClient.KindClusterExists(ctx, cluster.Name)
	if err != nil {
		return fmt.Errorf("deleting bootstrap cluster: %v", err)
	}
	if !clusterExists {
		logger.V(4).Info("Skipping delete bootstrap cluster, cluster doesn't exist")
		return nil
	}
	mgmtCluster, err := b.managementInCluster(ctx, cluster)
	if err != nil {
		return fmt.Errorf("deleting bootstrap cluster: %v", err)
	}

	if mgmtCluster != nil {
		if !isForceCleanup && (operationType == constants.Upgrade || mgmtCluster.Status.Phase == "Provisioned") {
			return errors.New("error deleting bootstrap cluster: management cluster in bootstrap cluster")
		}
	}

	return b.clusterClient.DeleteKindCluster(ctx, cluster)
}

func (b *Bootstrapper) managementInCluster(ctx context.Context, cluster *types.Cluster) (*types.CAPICluster, error) {
	if cluster.KubeconfigFile == "" {
		kubeconfig, err := b.clusterClient.GetKindClusterKubeconfig(ctx, cluster.Name)
		if err != nil {
			return nil, fmt.Errorf("fetching bootstrap cluster's kubeconfig: %v", err)
		}
		cluster.KubeconfigFile = kubeconfig
	}
	err := b.clusterClient.GetCAPIClusterCRD(ctx, cluster)
	if err == nil {
		clusters, err := b.clusterClient.GetCAPIClusters(ctx, cluster)
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

func WithExtraPortMappings(ports []int) BootstrapClusterOption {
	return func(b *Bootstrapper) BootstrapClusterClientOption {
		return b.clusterClient.WithExtraPortMappings(ports)
	}
}

func WithEnv(env map[string]string) BootstrapClusterOption {
	return func(b *Bootstrapper) BootstrapClusterClientOption {
		return b.clusterClient.WithEnv(env)
	}
}
