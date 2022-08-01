package reconciler

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller"
)

type DockerReconciler struct{}

func NewDockerReconciler() *DockerReconciler {
	return &DockerReconciler{}
}

func (v *DockerReconciler) Reconcile(ctx context.Context, cluster *anywherev1.Cluster) (controller.Result, error) {
	return controller.Result{}, nil
}
