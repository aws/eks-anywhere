package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// VSphereDatacenterReconciler reconciles a VSphereDatacenterConfig object
type VSphereDatacenterReconciler struct {
	client client.Client
	log    logr.Logger
}

func NewVSphereDatacenterReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *VSphereDatacenterReconciler {
	return &VSphereDatacenterReconciler{
		client: client,
		log:    log,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *VSphereDatacenterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.VSphereDatacenterConfig{}).
		Complete(r)
}

// TODO: add here kubebuilder permissions as neeeded
func (r *VSphereDatacenterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	_ = r.log.WithValues("vsphereDatacenter", req.NamespacedName)

	return ctrl.Result{}, nil
}
