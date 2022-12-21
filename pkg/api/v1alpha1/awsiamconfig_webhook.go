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
var awsiamconfiglog = logf.Log.WithName("awsiamconfig-resource")

func (r *AWSIamConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-awsiamconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=awsiamconfigs,verbs=create;update,versions=v1alpha1,name=mutation.awsiamconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &AWSIamConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *AWSIamConfig) Default() {
	awsiamconfiglog.Info("Setting up AWSIamConfig defaults for", "name", r.Name)
	r.SetDefaults()
}

// change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-awsiamconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=awsiamconfigs,verbs=create;update,versions=v1alpha1,name=validation.awsiamconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &AWSIamConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *AWSIamConfig) ValidateCreate() error {
	awsiamconfiglog.Info("validate create", "name", r.Name)

	return r.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *AWSIamConfig) ValidateUpdate(old runtime.Object) error {
	awsiamconfiglog.Info("validate update", "name", r.Name)

	oldAWSIamConfig, ok := old.(*AWSIamConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a AWSIamConfig but got a %T", old))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableAWSIamFields(r, oldAWSIamConfig)...)
	if err := r.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("AWSIamConfig"), r, err.Error()))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(GroupVersion.WithKind(AWSIamConfigKind).GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *AWSIamConfig) ValidateDelete() error {
	awsiamconfiglog.Info("validate delete", "name", r.Name)

	return nil
}

func validateImmutableAWSIamFields(new, old *AWSIamConfig) field.ErrorList {
	var allErrs field.ErrorList

	if !new.Spec.Equal(&old.Spec) {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath(AWSIamConfigKind), "config is immutable"),
		)
	}

	return allErrs
}
