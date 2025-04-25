package v1alpha1

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var vspheremachineconfiglog = logf.Log.WithName("vspheremachineconfig-resource")

func (r *VSphereMachineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-vspheremachineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=vspheremachineconfigs,verbs=create;update,versions=v1alpha1,name=mutation.vspheremachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &VSphereMachineConfig{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (r *VSphereMachineConfig) Default(_ context.Context, obj runtime.Object) error {
	vsphereConfig, ok := obj.(*VSphereMachineConfig)
	if !ok {
		return fmt.Errorf("expected a VSphereMachineConfig but got %T", obj)
	}

	vspheremachineconfiglog.Info("Setting up VSphere Machine Config defaults for", "name", vsphereConfig.Name)
	vsphereConfig.SetDefaults()

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-vspheremachineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=vspheremachineconfigs,verbs=create;update,versions=v1alpha1,name=validation.vspheremachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &VSphereMachineConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *VSphereMachineConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	vsphereConfig, ok := obj.(*VSphereMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a VSphereMachineConfig but got %T", obj)
	}

	vspheremachineconfiglog.Info("validate create", "name", vsphereConfig.Name)

	if err := vsphereConfig.ValidateHasTemplate(); err != nil {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(VSphereMachineConfigKind).GroupKind(), vsphereConfig.Name, field.ErrorList{
			field.Invalid(field.NewPath("spec", "template"), vsphereConfig.Spec, err.Error()),
		})
	}
	if err := vsphereConfig.ValidateUsers(); err != nil {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(VSphereMachineConfigKind).GroupKind(), vsphereConfig.Name, field.ErrorList{
			field.Invalid(field.NewPath("spec", "users"), vsphereConfig.Spec.Users, err.Error()),
		})
	}

	if err := vsphereConfig.Validate(); err != nil {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(VSphereMachineConfigKind).GroupKind(), vsphereConfig.Name, field.ErrorList{
			field.Invalid(field.NewPath("spec", "users"), vsphereConfig.Spec.Users, err.Error()),
		})
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *VSphereMachineConfig) ValidateUpdate(_ context.Context, obj, old runtime.Object) (admission.Warnings, error) {
	vsphereConfig, ok := obj.(*VSphereMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a VSphereMachineConfig but got %T", obj)
	}

	vspheremachineconfiglog.Info("validate update", "name", vsphereConfig.Name)

	oldVSphereMachineConfig, ok := old.(*VSphereMachineConfig)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a VSphereMachineConfig but got a %T", old))
	}

	if err := vsphereConfig.ValidateUsers(); err != nil {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(VSphereMachineConfigKind).GroupKind(), vsphereConfig.Name, field.ErrorList{
			field.Invalid(field.NewPath("spec", "users"), vsphereConfig.Spec.Users, err.Error()),
		})
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableFieldsVSphereMachineConfig(vsphereConfig, oldVSphereMachineConfig)...)

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(VSphereMachineConfigKind).GroupKind(), vsphereConfig.Name, allErrs)
	}

	if err := vsphereConfig.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), vsphereConfig.Spec, err.Error()))
	}

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(VSphereMachineConfigKind).GroupKind(), vsphereConfig.Name, allErrs)
	}

	return nil, nil
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

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *VSphereMachineConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	vsphereConfig, ok := obj.(*VSphereMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a VSphereMachineConfig but got %T", obj)
	}

	vspheremachineconfiglog.Info("validate delete", "name", vsphereConfig.Name)

	return nil, nil
}
