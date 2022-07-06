package reconciler

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller"
	clustercontrollers "github.com/aws/eks-anywhere/pkg/controller/clusters"
)

type DockerReconciler struct {
	*clustercontrollers.ProviderClusterReconciler
}

func NewDockerReconciler() *DockerReconciler {
	return &DockerReconciler{ProviderClusterReconciler: clustercontrollers.NewProviderClusterReconciler(nil)}
}

func (v *DockerReconciler) Reconcile(ctx context.Context, cluster *anywherev1.Cluster) (controller.Result, error) {
	return controller.Result{}, nil
}
