package bootstrapper

import (
	"context"
	"time"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

// KindClient is a Kind client.
type KindClient interface {
	CreateBootstrapCluster(ctx context.Context, clusterSpec *cluster.Spec, opts ...BootstrapClusterClientOption) (kubeconfig string, err error)
	DeleteBootstrapCluster(ctx context.Context, cluster *types.Cluster) error
	WithExtraDockerMounts() BootstrapClusterClientOption
	WithExtraPortMappings([]int) BootstrapClusterClientOption
	WithEnv(env map[string]string) BootstrapClusterClientOption
	GetKubeconfig(ctx context.Context, clusterName string) (string, error)
	ClusterExists(ctx context.Context, clusterName string) (bool, error)
}

// KubernetesClient is a Kubernetes client.
type KubernetesClient interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	GetClusters(ctx context.Context, cluster *types.Cluster) ([]types.CAPICluster, error)
	ValidateClustersCRD(ctx context.Context, cluster *types.Cluster) error
	CreateNamespaceIfNotPresent(ctx context.Context, kubeconfig string, namespace string) error
}

// RetrierClientOpt allows to customize a RetrierClient
// on construction.
type RetrierClientOpt func(*RetrierClient)

// WithRetrierClientRetrier allows to use a custom retrier.
func WithRetrierClientRetrier(retrier retrier.Retrier) RetrierClientOpt {
	return func(u *RetrierClient) {
		u.retrier = retrier
	}
}

// RetrierClient wraps kind and kubernetes APIs around a retrier.
type RetrierClient struct {
	KindClient
	k8s     KubernetesClient
	retrier retrier.Retrier
}

// NewRetrierClient constructs a new RetrierClient.
func NewRetrierClient(kind KindClient, k8s KubernetesClient, opts ...RetrierClientOpt) RetrierClient {
	c := &RetrierClient{
		k8s:        k8s,
		KindClient: kind,
		retrier:    *retrier.NewWithMaxRetries(10, 5*time.Second),
	}

	for _, opt := range opts {
		opt(c)
	}

	return *c
}

// Apply creates/updates the data objects for a cluster.
func (c RetrierClient) Apply(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.retrier.Retry(
		func() error {
			return c.k8s.ApplyKubeSpecFromBytes(ctx, cluster, data)
		},
	)
}

// CreateNamespace creates a namespace if the namespace does not exist.
func (c RetrierClient) CreateNamespace(ctx context.Context, kubeconfig, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.k8s.CreateNamespaceIfNotPresent(ctx, kubeconfig, namespace)
		},
	)
}

// GetCAPIClusterCRD gets the capi cluster crd in a K8s cluster.
func (c RetrierClient) GetCAPIClusterCRD(ctx context.Context, cluster *types.Cluster) error {
	return c.retrier.Retry(
		func() error {
			return c.k8s.ValidateClustersCRD(ctx, cluster)
		},
	)
}

// GetCAPIClusters gets all the capi clusters in a K8s cluster.
func (c RetrierClient) GetCAPIClusters(ctx context.Context, cluster *types.Cluster) ([]types.CAPICluster, error) {
	clusters := []types.CAPICluster{}
	err := c.retrier.Retry(
		func() error {
			var err error
			clusters, err = c.k8s.GetClusters(ctx, cluster)
			return err
		},
	)
	if err != nil {
		return nil, err
	}

	return clusters, nil
}

// KindClusterExists checks whether a kind cluster exists by a cluster name.
func (c RetrierClient) KindClusterExists(ctx context.Context, clusterName string) (bool, error) {
	var exists bool
	err := c.retrier.Retry(
		func() error {
			var err error
			exists, err = c.KindClient.ClusterExists(ctx, clusterName)
			return err
		},
	)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetKindClusterKubeconfig gets the kubeconfig for a kind cluster by cluster name.
func (c RetrierClient) GetKindClusterKubeconfig(ctx context.Context, clusterName string) (string, error) {
	var kubeconfig string
	err := c.retrier.Retry(
		func() error {
			var err error
			kubeconfig, err = c.KindClient.GetKubeconfig(ctx, clusterName)
			return err
		},
	)
	if err != nil {
		return "", err
	}

	return kubeconfig, nil
}

// DeleteKindCluster deletes a kind cluster by cluster name.
func (c RetrierClient) DeleteKindCluster(ctx context.Context, cluster *types.Cluster) error {
	return c.retrier.Retry(
		func() error {
			return c.KindClient.DeleteBootstrapCluster(ctx, cluster)
		},
	)
}
