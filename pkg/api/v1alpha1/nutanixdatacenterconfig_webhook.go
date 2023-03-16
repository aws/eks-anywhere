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

// nutanixdatacenterconfiglog is for logging in this package.
var nutanixdatacenterconfiglog = logf.Log.WithName("nutanixdatacenterconfig-resource")

// SetupWebhookWithManager sets up the webhook with the manager.
func (r *NutanixDatacenterConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-nutanixdatacenterconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=nutanixdatacenterconfigs,verbs=create;update,versions=v1alpha1,name=validation.nutanixdatacenterconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &NutanixDatacenterConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) ValidateCreate() error {
	nutanixdatacenterconfiglog.Info("validate create", "name", r.Name)
	if r.IsReconcilePaused() {
		nutanixdatacenterconfiglog.Info("NutanixDatacenterConfig is paused, allowing create", "name", r.Name)
		return nil
	}

	if r.Spec.CredentialRef == nil {
		return apierrors.NewInvalid(
			GroupVersion.WithKind(NutanixDatacenterKind).GroupKind(),
			r.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec"), r.Spec, "credentialRef is required to be set to create a new NutanixDatacenterConfig"),
			})
	}

	if err := r.Validate(); err != nil {
		return apierrors.NewInvalid(
			GroupVersion.WithKind(NutanixDatacenterKind).GroupKind(),
			r.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec"), r.Spec, err.Error()),
			})
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) ValidateUpdate(old runtime.Object) error {
	nutanixdatacenterconfiglog.Info("validate update", "name", r.Name)
	oldDatacenterConfig, ok := old.(*NutanixDatacenterConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a NutanixDatacenterConfig but got a %T", old))
	}

	if oldDatacenterConfig.IsReconcilePaused() {
		nutanixdatacenterconfiglog.Info("NutanixDatacenterConfig is paused, allowing update", "name", r.Name)
		return nil
	}

	var allErrs field.ErrorList
	allErrs = append(allErrs, validateImmutableFieldsNutanixDatacenterConfig(r, oldDatacenterConfig)...)

	if r.Spec.CredentialRef == nil {
		// check if the old object has a credentialRef set
		if oldDatacenterConfig.Spec.CredentialRef != nil {
			allErrs = append(allErrs, field.Forbidden(field.NewPath("spec.credentialRef"), "credentialRef cannot be removed from an existing NutanixDatacenterConfig"))
		}
	}

	if err := r.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), r.Spec, err.Error()))
	}

	if len(allErrs) > 0 {
		return apierrors.NewInvalid(
			GroupVersion.WithKind(NutanixDatacenterKind).GroupKind(),
			r.Name,
			allErrs)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) ValidateDelete() error {
	nutanixdatacenterconfiglog.Info("validate delete", "name", r.Name)

	return nil
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
