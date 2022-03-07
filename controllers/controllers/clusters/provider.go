package clusters

import (
	"context"

	"github.com/aws/eks-anywhere/controllers/controllers/reconciler"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type ProviderClusterReconciler interface {
	Reconcile(ctx context.Context, cluster *anywherev1.Cluster) (reconciler.Result, error)
}

func BuildProviderReconciler(datacenterKind string) (ProviderClusterReconciler, error) {
	return nil, nil
}

type providerClusterReconciler struct{}
