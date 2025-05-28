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
var awsiamconfiglog = logf.Log.WithName("awsiamconfig-resource")

func (r *AWSIamConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-awsiamconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=awsiamconfigs,verbs=create;update,versions=v1alpha1,name=mutation.awsiamconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &AWSIamConfig{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (r *AWSIamConfig) Default(_ context.Context, obj runtime.Object) error {
	awsIamConfig, ok := obj.(*AWSIamConfig)
	if !ok {
		return fmt.Errorf("expected an AWSIamConfig but got %T", obj)
	}

	awsiamconfiglog.Info("Setting up AWSIamConfig defaults for", "name", awsIamConfig.Name)
	awsIamConfig.SetDefaults()

	return nil
}

// change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-awsiamconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=awsiamconfigs,verbs=create;update,versions=v1alpha1,name=validation.awsiamconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &AWSIamConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *AWSIamConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	awsIamConfig, ok := obj.(*AWSIamConfig)
	if !ok {
		return nil, fmt.Errorf("expected an AWSIamConfig but got %T", obj)
	}

	awsiamconfiglog.Info("validate create", "name", awsIamConfig.Name)

	return nil, awsIamConfig.Validate()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *AWSIamConfig) ValidateUpdate(_ context.Context, old, obj runtime.Object) (admission.Warnings, error) {
	awsIamConfig, ok := obj.(*AWSIamConfig)
	if !ok {
		return nil, fmt.Errorf("expected an AWSIamConfig but got %T", obj)
	}

	awsiamconfiglog.Info("validate update", "name", awsIamConfig.Name)

	oldAWSIamConfig, ok := old.(*AWSIamConfig)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a AWSIamConfig but got a %T", old))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableAWSIamFields(r, oldAWSIamConfig)...)
	if err := r.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("AWSIamConfig"), r, err.Error()))
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(GroupVersion.WithKind(AWSIamConfigKind).GroupKind(), awsIamConfig.Name, allErrs)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *AWSIamConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	awsIamConfig, ok := obj.(*AWSIamConfig)
	if !ok {
		return nil, fmt.Errorf("expected an AWSIamConfig but got %T", obj)
	}

	awsiamconfiglog.Info("validate delete", "name", awsIamConfig.Name)

	return nil, nil
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
