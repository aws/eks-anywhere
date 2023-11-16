package controllers

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// NodeUpgradeReconciler reconciles a NodeUpgrade object.
type NodeUpgradeReconciler struct {
	client client.Client
}

// NewNodeUpgradeReconciler returns a new instance of NodeUpgradeReconciler.
func NewNodeUpgradeReconciler(client client.Client) *NodeUpgradeReconciler {
	return &NodeUpgradeReconciler{
		client: client,
	}
}

//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=nodeupgrades,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=nodeupgrades/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=nodeupgrades/finalizers,verbs=update

// Reconcile reconciles a NodeUpgrade object.
func (r *NodeUpgradeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeUpgradeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.NodeUpgrade{}).
		Complete(r)
}
