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

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CloudStackMachineConfig) ValidateCreate() error {
	cloudstackmachineconfiglog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CloudStackMachineConfig) ValidateUpdate(old runtime.Object) error {
	cloudstackmachineconfiglog.Info("validate update", "name", r.Name)

	oldCloudStackMachineConfig, ok := old.(*CloudStackMachineConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a CloudStackMachineConfig but got a %T", old))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableFieldsCloudStackMachineConfig(r, oldCloudStackMachineConfig)...)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(GroupVersion.WithKind(CloudStackMachineConfigKind).GroupKind(), r.Name, allErrs)
}

func validateImmutableFieldsCloudStackMachineConfig(new, old *CloudStackMachineConfig) field.ErrorList {
	if old.IsReconcilePaused() {
		cloudstackmachineconfiglog.Info("Reconciliation is paused")
		return nil
	}

	var allErrs field.ErrorList

	if old.Spec.OSFamily != new.Spec.OSFamily {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "osFamily"), new.Spec.OSFamily, "field is immutable"),
		)
	}

	if !reflect.DeepEqual(old.Spec.Users, new.Spec.Users) {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "users"), new.Spec.Users, "field is immutable"),
		)
	}

	if old.Spec.Template != new.Spec.Template {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "template"), new.Spec.Template, "field is immutable"),
		)
	}

	cloudstackmachineconfiglog.Info("Machine config is associated with control plane or etcd")

	if old.Spec.ComputeOffering != new.Spec.ComputeOffering {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "computeOffering"), new.Spec.ComputeOffering, "field is immutable"),
		)
	}

	if old.Spec.DiskOffering != new.Spec.DiskOffering {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "diskOffering"), new.Spec.DiskOffering, "field is immutable"),
		)
	}

	if !reflect.DeepEqual(old.Spec.Details, new.Spec.Details) {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "details"), new.Spec.Details, "field is immutable"),
		)
	}

	return allErrs
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CloudStackMachineConfig) ValidateDelete() error {
	cloudstackmachineconfiglog.Info("validate delete", "name", r.Name)

	return nil
}
