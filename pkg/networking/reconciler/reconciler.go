package reconciler

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
)

type CiliumReconciler interface {
	Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *cluster.Spec) (controller.Result, error)
}

type Reconciler struct {
	ciliumReconciler CiliumReconciler
}

func New(ciliumReconciler CiliumReconciler) *Reconciler {
	return &Reconciler{
		ciliumReconciler: ciliumReconciler,
	}
}

// Reconcile takes the specified CNI in a cluster to the desired state defined in a cluster Spec
// It uses a controller.Result to indicate when requeues are needed
// Intended to be used in a kubernetes controller
// Only Cilium CNI is supported for now.
func (r *Reconciler) Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *cluster.Spec) (controller.Result, error) {
	if spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium != nil {
		return r.ciliumReconciler.Reconcile(ctx, logger, client, spec)
	} else {
		return controller.Result{}, errors.New("unsupported CNI, only Cilium is supported at this time")
	}
}
