package v1alpha1

import (
	"context"
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var nutanixmachineconfiglog = logf.Log.WithName("nutanixmachineconfig-resource")

// SetupWebhookWithManager sets up and registers the webhook with the manager.
func (in *NutanixMachineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		WithValidator(in).
		Complete()
}

var _ webhook.CustomValidator = &NutanixMachineConfig{}

//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-nutanixmachineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=nutanixmachineconfigs,verbs=create;update,versions=v1alpha1,name=validation.nutanixmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (in *NutanixMachineConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	nutanixConfig, ok := obj.(*NutanixMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a NutanixMachineConfig but got %T", obj)
	}

	nutanixmachineconfiglog.Info("validate create", "name", nutanixConfig.Name)
	if err := nutanixConfig.Validate(); err != nil {
		return nil, apierrors.NewInvalid(
			GroupVersion.WithKind(NutanixMachineConfigKind).GroupKind(),
			nutanixConfig.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec"), nutanixConfig.Spec, err.Error()),
			},
		)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (in *NutanixMachineConfig) ValidateUpdate(_ context.Context, old, obj runtime.Object) (admission.Warnings, error) {
	nutanixConfig, ok := obj.(*NutanixMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a NutanixMachineConfig but got %T", obj)
	}

	nutanixmachineconfiglog.Info("validate update", "name", nutanixConfig.Name)

	oldNutanixMachineConfig, ok := old.(*NutanixMachineConfig)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a NutanixMachineConfig but got a %T", old))
	}

	var allErrs field.ErrorList
	if err := nutanixConfig.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), nutanixConfig.Spec, err.Error()))
	}

	if oldNutanixMachineConfig.IsReconcilePaused() {
		nutanixmachineconfiglog.Info("NutanixMachineConfig is paused, so allowing update", "name", nutanixConfig.Name)
		if len(allErrs) > 0 {
			return nil, apierrors.NewInvalid(
				GroupVersion.WithKind(NutanixMachineConfigKind).GroupKind(),
				nutanixConfig.Name,
				allErrs,
			)
		}
		return nil, nil
	}

	allErrs = append(allErrs, validateImmutableFieldsNutantixMachineConfig(nutanixConfig, oldNutanixMachineConfig)...)
	if len(allErrs) > 0 {
		return nil, apierrors.NewInvalid(
			GroupVersion.WithKind(NutanixMachineConfigKind).GroupKind(),
			nutanixConfig.Name,
			allErrs,
		)
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (in *NutanixMachineConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	nutanixConfig, ok := obj.(*NutanixMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a NutanixMachineConfig but got %T", obj)
	}

	nutanixmachineconfiglog.Info("validate delete", "name", nutanixConfig.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func validateImmutableFieldsNutantixMachineConfig(new, old *NutanixMachineConfig) field.ErrorList {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")
	if !reflect.DeepEqual(new.Spec.BootType, old.Spec.BootType) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("bootType"), "field is immutable"))
	}

	if new.Spec.OSFamily != old.Spec.OSFamily {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("OSFamily"), "field is immutable"))
	}

	if !reflect.DeepEqual(new.Spec.Cluster, old.Spec.Cluster) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("Cluster"), "field is immutable"))
	}

	if !reflect.DeepEqual(new.Spec.Subnet, old.Spec.Subnet) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("Subnet"), "field is immutable"))
	}

	if old.IsManaged() {
		nutanixmachineconfiglog.Info("Machine config is associated with workload cluster", "name", old.Name)
		return allErrs
	}

	if !old.IsEtcd() && !old.IsControlPlane() {
		nutanixmachineconfiglog.Info("Machine config is associated with management cluster's worker nodes", "name", old.Name)
		return allErrs
	}

	nutanixmachineconfiglog.Info("Machine config is associated with management cluster's control plane or etcd", "name", old.Name)

	if err := validateImmutableFieldsControlPlane(new, old); err != nil {
		allErrs = append(allErrs, err...)
	}

	return allErrs
}

func validateImmutableFieldsControlPlane(new, old *NutanixMachineConfig) field.ErrorList {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")
	if !reflect.DeepEqual(new.Spec.BootType, old.Spec.BootType) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("bootType"), "field is immutable"))
	}

	if !reflect.DeepEqual(new.Spec.VCPUSockets, old.Spec.VCPUSockets) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("vCPUSockets"), "field is immutable"))
	}
	if !reflect.DeepEqual(new.Spec.VCPUsPerSocket, old.Spec.VCPUsPerSocket) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("vCPUsPerSocket"), "field is immutable"))
	}
	if !reflect.DeepEqual(new.Spec.MemorySize, old.Spec.MemorySize) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("memorySize"), "field is immutable"))
	}
	if !reflect.DeepEqual(new.Spec.SystemDiskSize, old.Spec.SystemDiskSize) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("systemDiskSize"), "field is immutable"))
	}
	if !reflect.DeepEqual(new.Spec.Users, old.Spec.Users) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("users"), "field is immutable"))
	}

	return allErrs
}
