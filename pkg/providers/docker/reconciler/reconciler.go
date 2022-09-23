package reconciler

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
)

type Reconciler struct {
	client client.Client
	*serverside.ObjectApplier
}

func New() *Reconciler {
	return &Reconciler{}
}

func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, c *anywherev1.Cluster) (controller.Result, error) {
	clusterSpec, err := cluster.BuildSpec(ctx, clientutil.NewKubeClient(r.client), c)
	if err != nil {
		return controller.Result{}, err
	}

	return r.ReconcileControlPlane(ctx, log, clusterSpec)
	// more reconciliation steps to be added here
}

func (r *Reconciler) ReconcileControlPlane(ctx context.Context, log logr.Logger, spec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileControlPlane")
	log.Info("Applying control plane CAPI objects")

	return r.Apply(ctx, func() ([]kubernetes.Object, error) {
		controlPlane, err := docker.ControlPlaneSpec(ctx, log, clientutil.NewKubeClient(r.client), spec)
		if err != nil {
			return nil, err
		}

		return controlPlane.Objects(), nil
	})
}
