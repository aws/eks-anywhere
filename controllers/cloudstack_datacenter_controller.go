package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
	log := ctrl.LoggerFrom(ctx)

	cloudstackDatacenter := &anywherev1.CloudStackDatacenterConfig{}
	if err := r.client.Get(ctx, req.NamespacedName, cloudstackDatacenter); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed getting cloudstack datacenter config: %v", err)
	}

	patchHelper, err := patch.NewHelper(cloudstackDatacenter, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch the object and status after each reconciliation.
		if err := patchHelper.Patch(ctx, cloudstackDatacenter); err != nil {
			log.Error(reterr, "Failed to patch cloudstackDatacenter")
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// There's no need to go any further if the cloudstackDatacenter is marked for deletion.
	if !cloudstackDatacenter.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, reterr
	}

	result, err := r.reconcile(ctx, cloudstackDatacenter, log)
	if err != nil {
		log.Error(err, "Failed to reconcile cloudstackDatacenter")
	}
	return result, err
}

func (r *CloudStackDatacenterReconciler) reconcile(ctx context.Context, cloudstackDatacenterConfig *anywherev1.CloudStackDatacenterConfig, log logr.Logger) (_ ctrl.Result, reterr error) {
	cloudstackDatacenterConfig.SetDefaults()

	return ctrl.Result{}, nil
}
