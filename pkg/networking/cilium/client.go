package cilium

import (
	"context"
	"time"

	v1 "k8s.io/api/apps/v1"

	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	// DaemonSetName is the default name for the Cilium DS installed in EKS-A clusters.
	DaemonSetName = "cilium"
	// PreflightDaemonSetName is the default name for the Cilium preflight DS installed
	// in EKS-A clusters during Cilium upgrades.
	PreflightDaemonSetName  = "cilium-pre-flight-check"
	DeploymentName          = "cilium-operator"
	PreflightDeploymentName = "cilium-pre-flight-check"
	// ConfigMapName is the default name for the Cilium ConfigMap
	// containing Cilium's configuration.
	ConfigMapName = "cilium-config"
	// ServiceName is the default name for the Cilium Service installed in EKS-A clusters.
	ServiceName = "cilium-agent"
)

// Client allows to interact with the Kubernetes API.
type Client interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	DeleteKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	GetDaemonSet(ctx context.Context, name, namespace, kubeconfig string) (*v1.DaemonSet, error)
	GetDeployment(ctx context.Context, name, namespace, kubeconfig string) (*v1.Deployment, error)
	RolloutRestartDaemonSet(ctx context.Context, name, namespace, kubeconfig string) error
}

// RetrierClient wraps basic kubernetes API operations around a retrier.
type RetrierClient struct {
	client  Client
	retrier *retrier.Retrier
}

// RetrierClientOpt allows to customize a RetrierClient
// on construction.
type RetrierClientOpt func(*RetrierClient)

// RetrierClientRetrier allows to use a custom retrier.
func RetrierClientRetrier(retrier *retrier.Retrier) RetrierClientOpt {
	return func(u *RetrierClient) {
		u.retrier = retrier
	}
}

// NewRetrier constructs a new RetrierClient.
func NewRetrier(client Client, opts ...RetrierClientOpt) *RetrierClient {
	c := &RetrierClient{
		client:  client,
		retrier: retrier.New(5 * time.Minute),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Apply creates/updates the objects provided by the yaml document in the cluster.
func (c *RetrierClient) Apply(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.retrier.Retry(
		func() error {
			return c.client.ApplyKubeSpecFromBytes(ctx, cluster, data)
		},
	)
}

// Delete deletes the objects defined in the yaml document from the cluster.
func (c *RetrierClient) Delete(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.retrier.Retry(
		func() error {
			return c.client.DeleteKubeSpecFromBytes(ctx, cluster, data)
		},
	)
}

// WaitForPreflightDaemonSet blocks until the Cilium preflight DS installed during upgrades
// becomes ready or until the timeout expires.
func (c *RetrierClient) WaitForPreflightDaemonSet(ctx context.Context, cluster *types.Cluster) error {
	return c.retrier.Retry(
		func() error {
			return c.checkPreflightDaemonSetReady(ctx, cluster)
		},
	)
}

func (c *RetrierClient) checkPreflightDaemonSetReady(ctx context.Context, cluster *types.Cluster) error {
	ciliumDaemonSet, err := c.client.GetDaemonSet(ctx, DaemonSetName, namespace, cluster.KubeconfigFile)
	if err != nil {
		return err
	}

	if err := CheckDaemonSetReady(ciliumDaemonSet); err != nil {
		return err
	}

	preflightDaemonSet, err := c.client.GetDaemonSet(ctx, PreflightDaemonSetName, namespace, cluster.KubeconfigFile)
	if err != nil {
		return err
	}

	if err := CheckPreflightDaemonSetReady(ciliumDaemonSet, preflightDaemonSet); err != nil {
		return err
	}

	return nil
}

// WaitForPreflightDeployment blocks until the Cilium preflight Deployment installed during upgrades
// becomes ready or until the timeout expires.
func (c *RetrierClient) WaitForPreflightDeployment(ctx context.Context, cluster *types.Cluster) error {
	return c.retrier.Retry(
		func() error {
			return c.checkPreflightDeploymentReady(ctx, cluster)
		},
	)
}

func (c *RetrierClient) checkPreflightDeploymentReady(ctx context.Context, cluster *types.Cluster) error {
	preflightDeployment, err := c.client.GetDeployment(ctx, PreflightDeploymentName, namespace, cluster.KubeconfigFile)
	if err != nil {
		return err
	}

	if err := CheckDeploymentReady(preflightDeployment); err != nil {
		return err
	}

	return nil
}

// WaitForCiliumDaemonSet blocks until the Cilium DS installed as part of the default
// Cilium installation becomes ready or until the timeout expires.
func (c *RetrierClient) WaitForCiliumDaemonSet(ctx context.Context, cluster *types.Cluster) error {
	return c.retrier.Retry(
		func() error {
			return c.checkCiliumDaemonSetReady(ctx, cluster)
		},
	)
}

// RolloutRestartCiliumDaemonSet triggers a rollout restart of the Cilium DS installed
// as part of the default Cilium installation.
func (c *RetrierClient) RolloutRestartCiliumDaemonSet(ctx context.Context, cluster *types.Cluster) error {
	return c.retrier.Retry(
		func() error {
			return c.client.RolloutRestartDaemonSet(ctx, DaemonSetName, namespace, cluster.KubeconfigFile)
		},
	)
}

func (c *RetrierClient) checkCiliumDaemonSetReady(ctx context.Context, cluster *types.Cluster) error {
	daemonSet, err := c.client.GetDaemonSet(ctx, DaemonSetName, namespace, cluster.KubeconfigFile)
	if err != nil {
		return err
	}

	if err := CheckDaemonSetReady(daemonSet); err != nil {
		return err
	}

	return nil
}

// WaitForCiliumDeployment blocks until the Cilium Deployment installed as part of the default
// Cilium installation becomes ready or until the timeout expires.
func (c *RetrierClient) WaitForCiliumDeployment(ctx context.Context, cluster *types.Cluster) error {
	return c.retrier.Retry(
		func() error {
			return c.checkCiliumDeploymentReady(ctx, cluster)
		},
	)
}

func (c *RetrierClient) checkCiliumDeploymentReady(ctx context.Context, cluster *types.Cluster) error {
	deployment, err := c.client.GetDeployment(ctx, DeploymentName, namespace, cluster.KubeconfigFile)
	if err != nil {
		return err
	}

	if err := CheckDeploymentReady(deployment); err != nil {
		return err
	}

	return nil
}
