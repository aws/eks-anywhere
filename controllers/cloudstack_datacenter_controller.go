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
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/reconciler"
)

// CloudStackDatacenterReconciler reconciles a CloudStackDatacenterConfig object
type CloudStackDatacenterReconciler struct {
	log       logr.Logger
	client    client.Client
	defaulter *cloudstack.Defaulter
	validator *cloudstack.Validator
}

func NewCloudStackDatacenterReconciler(client client.Client, log logr.Logger, validator *cloudstack.Validator, defaulter *cloudstack.Defaulter) *CloudStackDatacenterReconciler {
	return &CloudStackDatacenterReconciler{
		client:    client,
		validator: validator,
		defaulter: defaulter,
		log:       log,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudStackDatacenterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.CloudStackDatacenterConfig{}).
		Complete(r)
}

// TODO: add here kubebuilder permissions as neeeded
func (r *CloudStackDatacenterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := r.log.WithValues("cloudstackDatacenter", req.NamespacedName)

	// Fetch the CloudStackDatacenter object
	cloudstackDatacenter := &anywherev1.CloudStackDatacenterConfig{}
	if err := r.client.Get(ctx, req.NamespacedName, cloudstackDatacenter); err != nil {
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(cloudstackDatacenter, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch the object and status after each reconciliation.
		patchOpts := []patch.Option{}
		if reterr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := patchHelper.Patch(ctx, cloudstackDatacenter, patchOpts...); err != nil {
			log.Error(reterr, "Failed to patch cloudstackdatacenterconfig")
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// There's no need to go any further if the CloudStackDatacenterConfig is marked for deletion.
	if !cloudstackDatacenter.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, cloudstackDatacenter, log)
	}

	result, err := r.reconcile(ctx, cloudstackDatacenter, log)
	if err != nil {
		log.Error(err, "Failed to reconcile CloudStackDatacenterConfig")
	}
	return result, err
}

func (r *CloudStackDatacenterReconciler) reconcile(ctx context.Context, cloudstackDatacenter *anywherev1.CloudStackDatacenterConfig, log logr.Logger) (_ ctrl.Result, reterr error) {
	// Set up envs for executing Govc cmd and default values for datacenter config
	if err := reconciler.SetupEnvVars(ctx, cloudstackDatacenter, r.client); err != nil {
		log.Error(err, "Failed to set up env vars and default values for CloudStackDatacenterConfig")
		return ctrl.Result{}, err
	}
	if err := r.defaulter.SetDefaultsForDatacenterConfig(ctx, cloudstackDatacenter); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed setting default values for cloudstack datacenter config: %v", err)
	}
	// Determine if CloudStackDatacenterConfig is valid
	if err := r.validator.ValidateCloudStackDatacenterConfig(ctx, cloudstackDatacenter); err != nil {
		log.Error(err, "Failed to validate CloudStackDatacenterConfig")
		return ctrl.Result{}, err
	}

	cloudstackDatacenter.Status.SpecValid = true

	return ctrl.Result{}, nil
}

func (r *CloudStackDatacenterReconciler) reconcileDelete(ctx context.Context, cloudstackDatacenter *anywherev1.CloudStackDatacenterConfig, log logr.Logger) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
