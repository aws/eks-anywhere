package fluxclient

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/types"
)

type FluxKubectl struct {
	*executables.Flux
	*executables.Kubectl
}

func (f *FluxKubectl) ForceReconcileGitRepo(ctx context.Context, cluster *types.Cluster, namespace string) error {
	a := map[string]string{
		"reconcile.fluxcd.io/requestedAt": strconv.FormatInt(time.Now().Unix(), 10),
	}
	return f.UpdateAnnotation(ctx, "gitrepositories", namespace, a,
		executables.WithOverwrite(),
		executables.WithCluster(cluster),
		executables.WithNamespace(namespace),
	)
}

func (f *FluxKubectl) DeleteFluxSystemSecret(ctx context.Context, cluster *types.Cluster, namespace string) error {
	return f.DeleteSecret(ctx, cluster, "flux-system", namespace)
}
