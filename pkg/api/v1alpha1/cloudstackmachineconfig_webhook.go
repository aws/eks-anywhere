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
var cloudstackmachineconfiglog = logf.Log.WithName("cloudstackmachineconfig-resource")

func (r *CloudStackMachineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-cloudstackmachineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=cloudstackmachineconfigs,verbs=create;update,versions=v1alpha1,name=validation.cloudstackmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &CloudStackMachineConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *CloudStackMachineConfig) ValidateCreate() error {
	cloudstackmachineconfiglog.Info("validate create", "name", r.Name)
	if err, fieldName, fieldValue := r.Spec.DiskOffering.Validate(); err != nil {
		return apierrors.NewBadRequest(fmt.Sprintf("disk offering %s:%v, preventing CloudStackMachineConfig resource creation: %v", fieldName, fieldValue, err))
	}
	if err, fieldName, fieldValue := r.Spec.Symlinks.Validate(); err != nil {
		return apierrors.NewBadRequest(fmt.Sprintf("symlinks %s:%v, preventing CloudStackMachineConfig resource creation: %v", fieldName, fieldValue, err))
	}

	// This is only needed for the webhook, which is why it is separate from the Validate method
	if err := r.ValidateUsers(); err != nil {
		return err
	}
	return r.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *CloudStackMachineConfig) ValidateUpdate(old runtime.Object) error {
	cloudstackmachineconfiglog.Info("validate update", "name", r.Name)

	oldCloudStackMachineConfig, ok := old.(*CloudStackMachineConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a CloudStackMachineConfig but got a %T", old))
	}

	if oldCloudStackMachineConfig.IsReconcilePaused() {
		cloudstackmachineconfiglog.Info("Reconciliation is paused")
		return nil
	}

	// This is only needed for the webhook, which is why it is separate from the Validate method
	if err := r.ValidateUsers(); err != nil {
		return apierrors.NewInvalid(GroupVersion.WithKind(CloudStackMachineConfigKind).GroupKind(),
			r.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec", "users"), r.Spec.Users, err.Error()),
			})
	}

	var allErrs field.ErrorList
	allErrs = append(allErrs, validateImmutableFieldsCloudStackMachineConfig(r, oldCloudStackMachineConfig)...)

	if err, fieldName, fieldValue := r.Spec.DiskOffering.Validate(); err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "diskOffering", fieldName), fieldValue, err.Error()),
		)
	}
	if err, fieldName, fieldValue := r.Spec.Symlinks.Validate(); err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "symlinks", fieldName), fieldValue, err.Error()),
		)
	}
	if err := r.ValidateUsers(); err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "users"), r.Spec.Users, err.Error()))
	}
	if err := r.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), r.Spec, err.Error()))
	}
	if len(allErrs) > 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind(CloudStackMachineConfigKind).GroupKind(), r.Name, allErrs)
	}

	return nil
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

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *CloudStackMachineConfig) ValidateDelete() error {
	cloudstackmachineconfiglog.Info("validate delete", "name", r.Name)

	return nil
}
