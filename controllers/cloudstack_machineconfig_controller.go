package controllers

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// CloudStackMachineConfigReconciler reconciles a CloudStackMachineConfig object
type CloudStackMachineConfigReconciler struct {
	log       logr.Logger
	client    client.Client
	validator *cloudstack.Validator
}

func NewCloudStackMachineConfigReconciler(client client.Client, log logr.Logger, validator *cloudstack.Validator) *CloudStackMachineConfigReconciler {
	return &CloudStackMachineConfigReconciler{
		client:    client,
		validator: validator,
		log:       log,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudStackMachineConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.CloudStackMachineConfig{}).
		Complete(r)
}

// TODO: add here kubebuilder permissions as needed
func (r *CloudStackMachineConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := r.log.WithValues("csMachineConfig", req.NamespacedName)

	// Fetch the CloudStackMachineConfig object
	csMachineConfig := &anywherev1.CloudStackMachineConfig{}
	log.Info("Reconciling cloudstackmachineconfig")
	if err := r.client.Get(ctx, req.NamespacedName, csMachineConfig); err != nil {
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(csMachineConfig, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch the object and status after each reconciliation.
		patchOpts := []patch.Option{}

		if err := patchHelper.Patch(ctx, csMachineConfig, patchOpts...); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, fmt.Errorf("patching cloudstackmachineconfig: %v", err)})
		}
	}()

	// There's no need to go any further if the CloudStackMachineConfig is marked for deletion.
	if !csMachineConfig.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, reterr
	}

	result, err := r.reconcile(ctx, csMachineConfig)
	if err != nil {
		reterr = kerrors.NewAggregate([]error{reterr, fmt.Errorf("reconciling cloudstackmachineconfig: %v", err)})
	}
	return result, reterr
}

func (r *CloudStackMachineConfigReconciler) reconcile(ctx context.Context, csMachineConfig *anywherev1.CloudStackMachineConfig) (_ ctrl.Result, reterr error) {
	// TODO: need to figure out how to load creds in controller
	var allErrs []error
	if err := r.validator.ValidateClusterMachineConfig(ctx, csMachineConfig); err != nil {
		csMachineConfig.Status.SpecValid = false
		aggregate := kerrors.NewAggregate(allErrs)
		failureMessage := aggregate.Error()
		csMachineConfig.Status.FailureMessage = &failureMessage
		return ctrl.Result{}, aggregate
	}
	csMachineConfig.Status.SpecValid = true
	csMachineConfig.Status.FailureMessage = nil
	return ctrl.Result{}, nil
}
