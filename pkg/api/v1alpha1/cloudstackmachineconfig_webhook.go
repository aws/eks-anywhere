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
var cloudstackmachineconfiglog = logf.Log.WithName("cloudstackmachineconfig-resource")

func (r *CloudStackMachineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithValidator(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-cloudstackmachineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=cloudstackmachineconfigs,verbs=create;update,versions=v1alpha1,name=validation.cloudstackmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &CloudStackMachineConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *CloudStackMachineConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cloudstackConfig, ok := obj.(*CloudStackMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a CloudStackMachineConfig but got %T", obj)
	}

	cloudstackmachineconfiglog.Info("validate create", "name", cloudstackConfig.Name)
	if err, fieldName, fieldValue := cloudstackConfig.Spec.DiskOffering.Validate(); err != nil {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("disk offering %s:%v, preventing CloudStackMachineConfig resource creation: %v", fieldName, fieldValue, err))
	}
	if err, fieldName, fieldValue := cloudstackConfig.Spec.Symlinks.Validate(); err != nil {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("symlinks %s:%v, preventing CloudStackMachineConfig resource creation: %v", fieldName, fieldValue, err))
	}

	// This is only needed for the webhook, which is why it is separate from the Validate method
	if err := cloudstackConfig.ValidateUsers(); err != nil {
		return nil, err
	}
	return nil, cloudstackConfig.Validate()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *CloudStackMachineConfig) ValidateUpdate(_ context.Context, old, obj runtime.Object) (admission.Warnings, error) {
	cloudstackConfig, ok := obj.(*CloudStackMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a CloudStackMachineConfig but got %T", obj)
	}

	cloudstackmachineconfiglog.Info("validate update", "name", cloudstackConfig.Name)

	oldCloudStackMachineConfig, ok := old.(*CloudStackMachineConfig)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a CloudStackMachineConfig but got a %T", old))
	}

	if oldCloudStackMachineConfig.IsReconcilePaused() {
		cloudstackmachineconfiglog.Info("Reconciliation is paused")
		return nil, nil
	}

	// This is only needed for the webhook, which is why it is separate from the Validate method
	if err := cloudstackConfig.ValidateUsers(); err != nil {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(CloudStackMachineConfigKind).GroupKind(),
			cloudstackConfig.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec", "users"), cloudstackConfig.Spec.Users, err.Error()),
			})
	}

	var allErrs field.ErrorList
	allErrs = append(allErrs, validateImmutableFieldsCloudStackMachineConfig(cloudstackConfig, oldCloudStackMachineConfig)...)

	if err, fieldName, fieldValue := cloudstackConfig.Spec.DiskOffering.Validate(); err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "diskOffering", fieldName), fieldValue, err.Error()),
		)
	}
	if err, fieldName, fieldValue := cloudstackConfig.Spec.Symlinks.Validate(); err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "symlinks", fieldName), fieldValue, err.Error()),
		)
	}
	if err := cloudstackConfig.ValidateUsers(); err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "users"), cloudstackConfig.Spec.Users, err.Error()))
	}
	if err := cloudstackConfig.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), cloudstackConfig.Spec, err.Error()))
	}
	if len(allErrs) > 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(CloudStackMachineConfigKind).GroupKind(), cloudstackConfig.Name, allErrs)
	}

	return nil, nil
}

func validateImmutableFieldsCloudStackMachineConfig(new, old *CloudStackMachineConfig) field.ErrorList {
	var allErrs field.ErrorList

	if old.Spec.Affinity != new.Spec.Affinity {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "affinity"), new.Spec.Affinity, "field is immutable"),
		)
	}

	affinityGroupIdsMutated := false
	if len(old.Spec.AffinityGroupIds) != len(new.Spec.AffinityGroupIds) {
		affinityGroupIdsMutated = true
	} else {
		for index, id := range old.Spec.AffinityGroupIds {
			if id != new.Spec.AffinityGroupIds[index] {
				affinityGroupIdsMutated = true
				break
			}
		}
	}
	if affinityGroupIdsMutated {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "affinityGroupIdsMutated"), new.Spec.AffinityGroupIds, "field is immutable"),
		)
	}

	return allErrs
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *CloudStackMachineConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cloudstackConfig, ok := obj.(*CloudStackMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a CloudStackMachineConfig but got %T", obj)
	}

	cloudstackmachineconfiglog.Info("validate delete", "name", cloudstackConfig.Name)

	return nil, nil
}
