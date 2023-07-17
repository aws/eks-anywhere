package reconciler

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
)

type CiliumReconciler interface {
	Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *cluster.Spec) (controller.Result, error)
	UpdateClusterStatusForCNI(ctx context.Context, client client.Client, cluster *anywherev1.Cluster) error
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

// UpdateClusterStatusForCNI updates the Cluster status for the default cni before the control plane is ready. The CNI reconciler
// handles the rest of the logic for determining the condition and updating the status based on the current state of the cluster.
func (r *Reconciler) UpdateClusterStatusForCNI(ctx context.Context, client client.Client, cluster *anywherev1.Cluster) error {
	if cluster.Spec.ClusterNetwork.CNIConfig.Cilium != nil {
		return r.ciliumReconciler.UpdateClusterStatusForCNI(ctx, client, cluster)
	}

	return errors.New("unsupported CNI, only Cilium is supported at this time")
}
