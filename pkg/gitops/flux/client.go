package flux

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	maxRetries          = 5
	backOffPeriod       = 5 * time.Second
	reconcileAnnotation = "kustomize.toolkit.fluxcd.io/reconcile"
)

// FluxClient is an interface that abstracts the basic commands of flux executable.
type FluxClient interface {
	BootstrapGithub(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error
	BootstrapGit(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig, cliConfig *config.CliConfig) error
	Uninstall(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error
	Reconcile(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error
}

// KubeClient is an interface that abstracts the basic commands of kubectl executable.
type KubeClient interface {
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	UpdateAnnotation(ctx context.Context, resourceType, objectName string, annotations map[string]string, opts ...executables.KubectlOpt) error
	RemoveAnnotation(ctx context.Context, resourceType, objectName string, key string, opts ...executables.KubectlOpt) error
	DeleteSecret(ctx context.Context, managementCluster *types.Cluster, secretName, namespace string) error
}

type fluxClient struct {
	flux FluxClient
	kube KubeClient
	*retrier.Retrier
}

func newFluxClient(flux FluxClient, kube KubeClient) *fluxClient {
	return &fluxClient{
		flux:    flux,
		kube:    kube,
		Retrier: retrier.NewWithMaxRetries(maxRetries, backOffPeriod),
	}
}

func (c *fluxClient) BootstrapGithub(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error {
	return c.Retry(
		func() error {
			return c.flux.BootstrapGithub(ctx, cluster, fluxConfig)
		},
	)
}

func (c *fluxClient) BootstrapGit(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig, cliConfig *config.CliConfig) error {
	return c.Retry(
		func() error {
			return c.flux.BootstrapGit(ctx, cluster, fluxConfig, cliConfig)
		},
	)
}

func (c *fluxClient) Uninstall(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error {
	return c.Retry(
		func() error {
			return c.flux.Uninstall(ctx, cluster, fluxConfig)
		},
	)
}

func (c *fluxClient) Reconcile(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error {
	return c.Retry(
		func() error {
			return c.flux.Reconcile(ctx, cluster, fluxConfig)
		},
	)
}

func (c *fluxClient) ForceReconcile(ctx context.Context, cluster *types.Cluster, namespace string) error {
	annotations := map[string]string{
		"reconcile.fluxcd.io/requestedAt": strconv.FormatInt(time.Now().Unix(), 10),
	}

	return c.Retry(
		func() error {
			return c.kube.UpdateAnnotation(ctx, "gitrepositories", namespace, annotations, executables.WithOverwrite(), executables.WithCluster(cluster), executables.WithNamespace(namespace))
		},
	)
}

func (c *fluxClient) DisableResourceReconcile(ctx context.Context, cluster *types.Cluster, resourceType, objectName, namespace string) error {
	annotations := map[string]string{
		reconcileAnnotation: "disabled",
	}

	return c.Retry(
		func() error {
			return c.kube.UpdateAnnotation(ctx, resourceType, objectName, annotations, executables.WithOverwrite(), executables.WithCluster(cluster), executables.WithNamespace(namespace))
		},
	)
}

func (c *fluxClient) EnableResourceReconcile(ctx context.Context, cluster *types.Cluster, resourceType, objectName, namespace string) error {
	return c.Retry(
		func() error {
			return c.kube.RemoveAnnotation(ctx, resourceType, objectName, reconcileAnnotation, executables.WithOverwrite(), executables.WithCluster(cluster), executables.WithNamespace(namespace))
		},
	)
}

func (c *fluxClient) DeleteSystemSecret(ctx context.Context, cluster *types.Cluster, namespace string) error {
	return c.Retry(
		func() error {
			return c.kube.DeleteSecret(ctx, cluster, "flux-system", namespace)
		},
	)
}

func (c *fluxClient) GetCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (eksaCluster *v1alpha1.Cluster, err error) {
	err = c.Retry(
		func() error {
			eksaCluster, err = c.kube.GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name)
			return err
		},
	)
	return eksaCluster, err
}
