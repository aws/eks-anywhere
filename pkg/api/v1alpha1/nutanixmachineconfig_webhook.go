package v1alpha1

import (
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var nutanixmachineconfiglog = logf.Log.WithName("nutanixmachineconfig-resource")

// SetupWebhookWithManager sets up and registers the webhook with the manager.
func (in *NutanixMachineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

var _ webhook.Validator = &NutanixMachineConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (in *NutanixMachineConfig) ValidateCreate() error {
	nutanixmachineconfiglog.Info("validate create", "name", in.Name)

	if in.IsReconcilePaused() {
		nutanixmachineconfiglog.Info("NutanixMachineConfig is paused, so allowing create", "name", in.Name)
		return nil
	}

	if err := in.Validate(); err != nil {
		return apierrors.NewInvalid(
			GroupVersion.WithKind(NutanixMachineConfigKind).GroupKind(),
			in.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec"), in.Spec, err.Error()),
			},
		)
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (in *NutanixMachineConfig) ValidateUpdate(old runtime.Object) error {
	nutanixmachineconfiglog.Info("validate update", "name", in.Name)

	oldNutanixMachineConfig, ok := old.(*NutanixMachineConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a NutanixMachineConfig but got a %T", old))
	}

	if oldNutanixMachineConfig.IsReconcilePaused() {
		nutanixmachineconfiglog.Info("NutanixMachineConfig is paused, so allowing create", "name", in.Name)
		return nil
	}

	var allErrs field.ErrorList
	allErrs = append(allErrs, validateImmutableFieldsNutantixMachineConfig(in, oldNutanixMachineConfig)...)

	if err := in.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), in.Spec, err.Error()))
	}

	if len(allErrs) > 0 {
		return apierrors.NewInvalid(
			GroupVersion.WithKind(NutanixMachineConfigKind).GroupKind(),
			in.Name,
			allErrs,
		)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (in *NutanixMachineConfig) ValidateDelete() error {
	nutanixmachineconfiglog.Info("validate delete", "name", in.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func validateImmutableFieldsNutantixMachineConfig(new, old *NutanixMachineConfig) field.ErrorList {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")
	if new.Spec.OSFamily != old.Spec.OSFamily {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("OSFamily"), "field is immutable"))
	}

	if !reflect.DeepEqual(new.Spec.Cluster, old.Spec.Cluster) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("Cluster"), "field is immutable"))
	}

	if !reflect.DeepEqual(new.Spec.Subnet, old.Spec.Subnet) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("Subnet"), "field is immutable"))
	}

	if !reflect.DeepEqual(new.Spec.Image, old.Spec.Image) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("Image"), "field is immutable"))
	}

	return allErrs
}
