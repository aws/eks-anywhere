package cilium

import (
	"context"
	"time"

	v1 "k8s.io/api/apps/v1"

	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	DaemonSetName           = "cilium"
	PreflightDaemonSetName  = "cilium-pre-flight-check"
	DeploymentName          = "cilium-operator"
	PreflightDeploymentName = "cilium-pre-flight-check"
)

type Client interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	DeleteKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	GetDaemonSet(ctx context.Context, name, namespace, kubeconfig string) (*v1.DaemonSet, error)
	GetDeployment(ctx context.Context, name, namespace, kubeconfig string) (*v1.Deployment, error)
	RolloutRestartDaemonSet(ctx context.Context, name, namespace, kubeconfig string) error
}

type retrierClient struct {
	Client
	*retrier.Retrier
}

func newRetrier(client Client) *retrierClient {
	return &retrierClient{
		Client:  client,
		Retrier: retrier.New(5 * time.Minute),
	}
}

func (c *retrierClient) Apply(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.Retry(
		func() error {
			return c.ApplyKubeSpecFromBytes(ctx, cluster, data)
		},
	)
}

func (c *retrierClient) Delete(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.Retry(
		func() error {
			return c.DeleteKubeSpecFromBytes(ctx, cluster, data)
		},
	)
}

func (c *retrierClient) WaitForPreflightDaemonSet(ctx context.Context, cluster *types.Cluster) error {
	return c.Retry(
		func() error {
			return c.checkPreflightDaemonSetReady(ctx, cluster)
		},
	)
}

func (c *retrierClient) checkPreflightDaemonSetReady(ctx context.Context, cluster *types.Cluster) error {
	ciliumDaemonSet, err := c.GetDaemonSet(ctx, DaemonSetName, namespace, cluster.KubeconfigFile)
	if err != nil {
		return err
	}

	if err := CheckDaemonSetReady(ciliumDaemonSet); err != nil {
		return err
	}

	preflightDaemonSet, err := c.GetDaemonSet(ctx, PreflightDaemonSetName, namespace, cluster.KubeconfigFile)
	if err != nil {
		return err
	}

	if err := CheckPreflightDaemonSetReady(ciliumDaemonSet, preflightDaemonSet); err != nil {
		return err
	}

	return nil
}

func (c *retrierClient) WaitForPreflightDeployment(ctx context.Context, cluster *types.Cluster) error {
	return c.Retry(
		func() error {
			return c.checkPreflightDeploymentReady(ctx, cluster)
		},
	)
}

func (c *retrierClient) checkPreflightDeploymentReady(ctx context.Context, cluster *types.Cluster) error {
	preflightDeployment, err := c.GetDeployment(ctx, PreflightDeploymentName, namespace, cluster.KubeconfigFile)
	if err != nil {
		return err
	}

	if err := CheckDeploymentReady(preflightDeployment); err != nil {
		return err
	}

	return nil
}

func (c *retrierClient) WaitForCiliumDaemonSet(ctx context.Context, cluster *types.Cluster) error {
	return c.Retry(
		func() error {
			return c.checkCiliumDaemonSetReady(ctx, cluster)
		},
	)
}

func (c *retrierClient) RolloutRestartCiliumDaemonSet(ctx context.Context, cluster *types.Cluster) error {
	return c.Retry(
		func() error {
			return c.RolloutRestartDaemonSet(ctx, DaemonSetName, namespace, cluster.KubeconfigFile)
		},
	)
}

func (c *retrierClient) checkCiliumDaemonSetReady(ctx context.Context, cluster *types.Cluster) error {
	daemonSet, err := c.GetDaemonSet(ctx, DaemonSetName, namespace, cluster.KubeconfigFile)
	if err != nil {
		return err
	}

	if err := CheckDaemonSetReady(daemonSet); err != nil {
		return err
	}

	return nil
}

func (c *retrierClient) WaitForCiliumDeployment(ctx context.Context, cluster *types.Cluster) error {
	return c.Retry(
		func() error {
			return c.checkCiliumDeploymentReady(ctx, cluster)
		},
	)
}

func (c *retrierClient) checkCiliumDeploymentReady(ctx context.Context, cluster *types.Cluster) error {
	deployment, err := c.GetDeployment(ctx, DeploymentName, namespace, cluster.KubeconfigFile)
	if err != nil {
		return err
	}

	if err := CheckDeploymentReady(deployment); err != nil {
		return err
	}

	return nil
}
