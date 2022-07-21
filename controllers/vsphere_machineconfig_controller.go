package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// VSphereMachineConfigReconciler reconciles a VSphereDatacenterConfig object
type VSphereMachineConfigReconciler struct {
	client client.Client
	log    logr.Logger
}

func NewVSphereMachineConfigReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *VSphereMachineConfigReconciler {
	return &VSphereMachineConfigReconciler{
		client: client,
		log:    log,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *VSphereMachineConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.VSphereMachineConfig{}).
		Complete(r)
}

// TODO: add here kubebuilder permissions as neeeded
func (r *VSphereMachineConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	_ = r.log.WithValues("vsphereMachineConfig", req.NamespacedName)

	return ctrl.Result{}, nil
}
