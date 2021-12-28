package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// CloudStackDeploymentReconciler reconciles a CloudStackDeploymentConfig object
type CloudStackDeploymentReconciler struct {
	client client.Client
	log    logr.Logger
}

func NewCloudStackDeploymentReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *CloudStackDeploymentReconciler {
	return &CloudStackDeploymentReconciler{
		client: client,
		log:    log,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudStackDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.CloudStackDeploymentConfig{}).
		Complete(r)
}

// TODO: add here kubebuilder permissions as neeeded
func (r *CloudStackDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	_ = r.log.WithValues("cloudstackDeployment", req.NamespacedName)

	return ctrl.Result{}, nil
}
