package controllers

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

// VSphereDatacenterReconciler reconciles a VSphereDatacenterConfig object
type VSphereDatacenterReconciler struct {
	client client.Client
	log    logr.Logger

	validator *vsphere.Validator
	defaulter *vsphere.Defaulter
}

func NewVSphereDatacenterReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme, govc *executables.Govc) *VSphereDatacenterReconciler {
	validator := vsphere.NewValidator(govc, &networkutils.DefaultNetClient{})
	defaulter := vsphere.NewDefaulter(govc)

	return &VSphereDatacenterReconciler{
		client:    client,
		log:       log,
		validator: validator,
		defaulter: defaulter,
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
	log := r.log.WithValues("vsphereDatacenter", req.NamespacedName)

	// Fetch the VsphereDatacenter object
	vsphereDatacenter := &anywherev1.VSphereDatacenterConfig{}
	if err := r.client.Get(ctx, req.NamespacedName, vsphereDatacenter); err != nil {
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(vsphereDatacenter, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch the object and status after each reconciliation.
		patchOpts := []patch.Option{}
		if reterr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := patchHelper.Patch(ctx, vsphereDatacenter, patchOpts...); err != nil {
			log.Error(reterr, "Failed to patch vspheredatacenterconfig")
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// There's no need to go any further if the VsphereDatacenterConfig is marked for deletion.
	if !vsphereDatacenter.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, vsphereDatacenter, log)
	}

	result, err := r.reconcile(ctx, vsphereDatacenter, log)
	if err != nil {
		log.Error(err, "Failed to reconcile VsphereDatacenterConfig")
	}
	return result, err
}

func (r *VSphereDatacenterReconciler) reconcile(ctx context.Context, vsphereDatacenter *anywherev1.VSphereDatacenterConfig, log logr.Logger) (_ ctrl.Result, reterr error) {
	// Set up envs for executing Govc cmd and default values for datacenter config
	if err := r.setupEnvsAndDefaults(ctx, vsphereDatacenter); err != nil {
		log.Error(err, "Failed to set up env vars and default values for VsphereDatacenterConfig")
		return ctrl.Result{}, err
	}
	// Determine if VsphereDatacenterConfig is valid
	if err := r.validator.ValidateVCenterConfig(ctx, vsphereDatacenter); err != nil {
		log.Error(err, "Failed to validate VsphereDatacenterConfig")
		vsphereDatacenter.Status.SpecValid = false
		return ctrl.Result{}, err
	}

	vsphereDatacenter.Status.SpecValid = true

	return ctrl.Result{}, nil
}

func (r *VSphereDatacenterReconciler) vsphereCredentials(ctx context.Context) (*apiv1.Secret, error) {
	secret := &apiv1.Secret{}
	secretKey := client.ObjectKey{
		Namespace: "eksa-system",
		Name:      vsphere.CredentialsObjectName,
	}
	if err := r.client.Get(ctx, secretKey, secret); err != nil {
		return nil, err
	}
	return secret, nil
}

func (r *VSphereDatacenterReconciler) setupEnvsAndDefaults(ctx context.Context, vsphereDatacenter *anywherev1.VSphereDatacenterConfig) error {
	secret, err := r.vsphereCredentials(ctx)
	if err != nil {
		return fmt.Errorf("failed getting vsphere credentials secret: %v", err)
	}

	vsphereUsername := secret.Data["username"]
	vspherePassword := secret.Data["password"]

	if err := os.Setenv(vsphere.EksavSphereUsernameKey, string(vsphereUsername)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", vsphere.EksavSphereUsernameKey, err)
	}

	if err := os.Setenv(vsphere.EksavSpherePasswordKey, string(vspherePassword)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", vsphere.EksavSpherePasswordKey, err)
	}

	if err := vsphere.SetupEnvVars(vsphereDatacenter); err != nil {
		return fmt.Errorf("failed setting env vars: %v", err)
	}

	if err := r.defaulter.SetDefaultsForDatacenterConfig(ctx, vsphereDatacenter); err != nil {
		return fmt.Errorf("failed setting default values for vsphere datacenter config: %v", err)
	}

	return nil
}

func (r *VSphereDatacenterReconciler) reconcileDelete(ctx context.Context, vsphereDatacenter *anywherev1.VSphereDatacenterConfig, log logr.Logger) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
