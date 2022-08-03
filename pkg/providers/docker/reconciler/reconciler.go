package reconciler

import (
	"context"

	"github.com/go-logr/logr"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller"
)

type DockerReconciler struct{}

func NewDockerReconciler() *DockerReconciler {
	return &DockerReconciler{}
}

func (v *DockerReconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	return controller.Result{}, nil
}
