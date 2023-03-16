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

//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-nutanixmachineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=nutanixmachineconfigs,verbs=create;update,versions=v1alpha1,name=validation.nutanixmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (in *NutanixMachineConfig) ValidateCreate() error {
	nutanixmachineconfiglog.Info("validate create", "name", in.Name)
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

	var allErrs field.ErrorList
	if err := in.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), in.Spec, err.Error()))
	}

	if oldNutanixMachineConfig.IsReconcilePaused() {
		nutanixmachineconfiglog.Info("NutanixMachineConfig is paused, so allowing update", "name", in.Name)
		if len(allErrs) > 0 {
			return apierrors.NewInvalid(
				GroupVersion.WithKind(NutanixMachineConfigKind).GroupKind(),
				in.Name,
				allErrs,
			)
		}
		return nil
	}

	allErrs = append(allErrs, validateImmutableFieldsNutantixMachineConfig(in, oldNutanixMachineConfig)...)
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
