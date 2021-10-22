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
var addonawsiamconfiglog = logf.Log.WithName("addonawsiamconfig-resource")

func (r *AddOnAWSIamConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-addonawsiamconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=addonawsiamconfigs,verbs=create;update,versions=v1alpha1,name=validation.addonawsiamconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &AddOnAWSIamConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AddOnAWSIamConfig) ValidateCreate() error {
	addonawsiamconfiglog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AddOnAWSIamConfig) ValidateUpdate(old runtime.Object) error {
	addonawsiamconfiglog.Info("validate update", "name", r.Name)

	oldAWSIamConfig, ok := old.(*AddOnAWSIamConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a AddOnAWSIamConfig but got a %T", old))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableAWSIamFields(r, oldAWSIamConfig)...)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(GroupVersion.WithKind(AddOnAWSIamConfigKind).GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AddOnAWSIamConfig) ValidateDelete() error {
	addonawsiamconfiglog.Info("validate delete", "name", r.Name)

	return nil
}

func validateImmutableAWSIamFields(new, old *AddOnAWSIamConfig) field.ErrorList {
	var allErrs field.ErrorList

	if !new.Spec.Equal(&old.Spec) {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "AddOnAWSIamConfig"), new, "config is immutable"),
		)
	}

	return allErrs
}
