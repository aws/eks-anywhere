package clusters

import (
	"context"

	"github.com/aws/eks-anywhere/controllers/controllers/reconciler"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type VSphereReconciler struct {
	*providerClusterReconciler
}

func NewVSphereReconciler() *VSphereReconciler {
	return &VSphereReconciler{providerClusterReconciler: &providerClusterReconciler{}}
}

func (v *VSphereReconciler) Reconcile(ctx context.Context, cluster *anywherev1.Cluster) (reconciler.Result, error) {
	return reconciler.Result{}, nil
}
