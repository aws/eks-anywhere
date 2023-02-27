package controllers

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CloudStackDatacenterReconciler reconciles a CloudStackDatacenterConfig object.
type CloudStackDatacenterReconciler struct {
	client client.Client
}

// NewCloudStackDatacenterReconciler creates a new instance of the CloudStackDatacenterReconciler struct.
func NewCloudStackDatacenterReconciler(client client.Client) *CloudStackDatacenterReconciler {
	return &CloudStackDatacenterReconciler{
		client: client,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudStackDatacenterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.CloudStackDatacenterConfig{}).
		Complete(r)
}

// Reconcile implements the reconcile.Reconciler interface.
func (r *CloudStackDatacenterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	// TODO fetch CloudStack datacenter object and implement reconcile
	return ctrl.Result{}, nil
}
