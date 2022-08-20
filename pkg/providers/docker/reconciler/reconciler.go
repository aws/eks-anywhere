package reconciler

import (
	"context"

	"github.com/go-logr/logr"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller"
)

type Reconciler struct{}

func New() *Reconciler {
	return &Reconciler{}
}

func (v *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	return controller.Result{}, nil
}
