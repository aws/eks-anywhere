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

// nutanixdatacenterconfiglog is for logging in this package.
var nutanixdatacenterconfiglog = logf.Log.WithName("nutanixdatacenterconfig-resource")

// SetupWebhookWithManager sets up the webhook with the manager.
func (r *NutanixDatacenterConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-nutanixdatacenterconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=nutanixdatacenterconfigs,verbs=create;update,versions=v1alpha1,name=validation.nutanixdatacenterconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &NutanixDatacenterConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	nutanixConfig, ok := obj.(*NutanixDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a NutanixDatacenterConfig but got %T", obj)
	}

	nutanixdatacenterconfiglog.Info("validate create", "name", nutanixConfig.Name)
	if nutanixConfig.IsReconcilePaused() {
		nutanixdatacenterconfiglog.Info("NutanixDatacenterConfig is paused, allowing create", "name", nutanixConfig.Name)
		return nil, nil
	}

	if nutanixConfig.Spec.CredentialRef == nil {
		return nil, apierrors.NewInvalid(
			GroupVersion.WithKind(NutanixDatacenterKind).GroupKind(),
			nutanixConfig.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec"), nutanixConfig.Spec, "credentialRef is required to be set to create a new NutanixDatacenterConfig"),
			})
	}

	if err := nutanixConfig.Validate(); err != nil {
		return nil, apierrors.NewInvalid(
			GroupVersion.WithKind(NutanixDatacenterKind).GroupKind(),
			nutanixConfig.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec"), nutanixConfig.Spec, err.Error()),
			})
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) ValidateUpdate(_ context.Context, obj, old runtime.Object) (admission.Warnings, error) {
	nutanixConfig, ok := obj.(*NutanixDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a NutanixDatacenterConfig but got %T", obj)
	}

	nutanixdatacenterconfiglog.Info("validate update", "name", nutanixConfig.Name)
	oldDatacenterConfig, ok := old.(*NutanixDatacenterConfig)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a NutanixDatacenterConfig but got a %T", old))
	}

	if oldDatacenterConfig.IsReconcilePaused() {
		nutanixdatacenterconfiglog.Info("NutanixDatacenterConfig is paused, allowing update", "name", nutanixConfig.Name)
		return nil, nil
	}

	var allErrs field.ErrorList
	allErrs = append(allErrs, validateImmutableFieldsNutanixDatacenterConfig(nutanixConfig, oldDatacenterConfig)...)

	if nutanixConfig.Spec.CredentialRef == nil {
		// check if the old object has a credentialRef set
		if oldDatacenterConfig.Spec.CredentialRef != nil {
			allErrs = append(allErrs, field.Forbidden(field.NewPath("spec.credentialRef"), "credentialRef cannot be removed from an existing NutanixDatacenterConfig"))
		}
	}

	if err := nutanixConfig.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), nutanixConfig.Spec, err.Error()))
	}

	if len(allErrs) > 0 {
		return nil, apierrors.NewInvalid(
			GroupVersion.WithKind(NutanixDatacenterKind).GroupKind(),
			nutanixConfig.Name,
			allErrs)
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	nutanixConfig, ok := obj.(*NutanixDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a NutanixDatacenterConfig but got %T", obj)
	}

	nutanixdatacenterconfiglog.Info("validate delete", "name", nutanixConfig.Name)

	return nil, nil
}

func validateImmutableFieldsNutanixDatacenterConfig(new, old *NutanixDatacenterConfig) field.ErrorList {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	if old.IsReconcilePaused() {
		nutanixmachineconfiglog.Info("Reconciliation is paused")
		return nil
	}

	if new.Spec.Endpoint != old.Spec.Endpoint {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("endpoint"), "field is immutable"))
	}

	return allErrs
}
