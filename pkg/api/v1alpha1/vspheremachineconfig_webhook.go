package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var vspheremachineconfiglog = logf.Log.WithName("vspheremachineconfig-resource")

func (r *VSphereMachineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-vspheremachineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=vspheremachineconfigs,verbs=create;update,versions=v1alpha1,name=mutation.vspheremachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &VSphereMachineConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *VSphereMachineConfig) Default() {
	vspheremachineconfiglog.Info("Setting up VSphere Machine Config defaults for", "name", r.Name)
	r.SetDefaults()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-vspheremachineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=vspheremachineconfigs,verbs=create;update,versions=v1alpha1,name=validation.vspheremachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &VSphereMachineConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *VSphereMachineConfig) ValidateCreate() error {
	vspheremachineconfiglog.Info("validate create", "name", r.Name)

	if err := r.ValidateHasTemplate(); err != nil {
		return apierrors.NewInvalid(GroupVersion.WithKind(VSphereMachineConfigKind).GroupKind(), r.Name, field.ErrorList{
			field.Invalid(field.NewPath("spec", "template"), r.Spec, err.Error()),
		})
	}
	if err := r.ValidateUsers(); err != nil {
		return apierrors.NewInvalid(GroupVersion.WithKind(VSphereMachineConfigKind).GroupKind(), r.Name, field.ErrorList{
			field.Invalid(field.NewPath("spec", "users"), r.Spec.Users, err.Error()),
		})
	}

	if err := r.Validate(); err != nil {
		return apierrors.NewInvalid(GroupVersion.WithKind(VSphereMachineConfigKind).GroupKind(), r.Name, field.ErrorList{
			field.Invalid(field.NewPath("spec", "users"), r.Spec.Users, err.Error()),
		})
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *VSphereMachineConfig) ValidateUpdate(old runtime.Object) error {
	vspheremachineconfiglog.Info("validate update", "name", r.Name)

	oldVSphereMachineConfig, ok := old.(*VSphereMachineConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a VSphereMachineConfig but got a %T", old))
	}

	if err := r.ValidateUsers(); err != nil {
		return apierrors.NewInvalid(GroupVersion.WithKind(VSphereMachineConfigKind).GroupKind(), r.Name, field.ErrorList{
			field.Invalid(field.NewPath("spec", "users"), r.Spec.Users, err.Error()),
		})
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableFieldsVSphereMachineConfig(r, oldVSphereMachineConfig)...)

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind(VSphereMachineConfigKind).GroupKind(), r.Name, allErrs)
	}

	if err := r.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), r.Spec, err.Error()))
	}

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind(VSphereMachineConfigKind).GroupKind(), r.Name, allErrs)
	}

	return nil
}

func validateImmutableFieldsVSphereMachineConfig(new, old *VSphereMachineConfig) field.ErrorList {
	if old.IsReconcilePaused() {
		vspheremachineconfiglog.Info("Reconciliation is paused")
		return nil
	}

	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	if old.Spec.OSFamily != new.Spec.OSFamily {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("osFamily"), "field is immutable"),
		)
	}

	if old.Spec.StoragePolicyName != new.Spec.StoragePolicyName {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("storagePolicyName"), "field is immutable"),
		)
	}

	if old.IsManaged() {
		vspheremachineconfiglog.Info("Machine config is associated with workload cluster", "name", old.Name)
		return allErrs
	}

	if !old.IsEtcd() && !old.IsControlPlane() {
		vspheremachineconfiglog.Info("Machine config is associated with management cluster's worker nodes", "name", old.Name)
		return allErrs
	}

	return allErrs
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *VSphereMachineConfig) ValidateDelete() error {
	vspheremachineconfiglog.Info("validate delete", "name", r.Name)

	return nil
}
