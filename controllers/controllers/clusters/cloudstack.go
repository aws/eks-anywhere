package clusters

import (
	"context"

	"github.com/aws/eks-anywhere/controllers/controllers/reconciler"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type CloudStackReconciler struct {
	*providerClusterReconciler
}

func NewCloudStackReconciler() *CloudStackReconciler {
	return &CloudStackReconciler{providerClusterReconciler: &providerClusterReconciler{}}
}

func (v *CloudStackReconciler) Reconcile(ctx context.Context, cluster *anywherev1.Cluster) (reconciler.Result, error) {
	return reconciler.Result{}, nil
}
