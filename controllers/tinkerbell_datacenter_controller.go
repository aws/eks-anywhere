package controllers

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// TinkerbellDatacenterReconciler reconciles a TinkerbellDatacenterConfig object.
type TinkerbellDatacenterReconciler struct {
	client client.Client
}

// NewTinkerbellDatacenterReconciler creates a new instance of the TinkerbellDatacenterReconciler struct.
func NewTinkerbellDatacenterReconciler(client client.Client) *TinkerbellDatacenterReconciler {
	return &TinkerbellDatacenterReconciler{
		client: client,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TinkerbellDatacenterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.TinkerbellDatacenterConfig{}).
		Complete(r)
}

// TODO: add here kubebuilder permissions as neeeded.

// Reconcile implements the reconcile.Reconciler interface.
func (r *TinkerbellDatacenterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	// TODO fetch Tinkerbell datacenter object and implement reconcile
	return ctrl.Result{}, nil
}
