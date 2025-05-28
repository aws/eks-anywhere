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
var snowippoollog = logf.Log.WithName("snowippool-resource")

// SetupWebhookWithManager sets up the webhook manager for SnowIPPool.
func (r *SnowIPPool) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-snowippool,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=snowippools,verbs=create;update,versions=v1alpha1,name=validation.snowippool.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &SnowIPPool{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *SnowIPPool) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	snowIPPool, ok := obj.(*SnowIPPool)
	if !ok {
		return nil, fmt.Errorf("expected a SnowIPPool but got %T", obj)
	}

	snowippoollog.Info("validate create", "name", snowIPPool.Name)

	return nil, snowIPPool.Validate()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *SnowIPPool) ValidateUpdate(_ context.Context, old, obj runtime.Object) (admission.Warnings, error) {
	snowIPPool, ok := obj.(*SnowIPPool)
	if !ok {
		return nil, fmt.Errorf("expected a SnowIPPool but got %T", obj)
	}

	snowippoollog.Info("validate update", "name", snowIPPool.Name)

	oldPool, ok := old.(*SnowIPPool)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a SnowIPPool but got a %T", old))
	}

	if allErrs := validateImmutableFieldsSnowIPPool(snowIPPool, oldPool); len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(SnowIPPoolKind).GroupKind(), snowIPPool.Name, allErrs)
	}

	return nil, snowIPPool.Validate()
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *SnowIPPool) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	snowIPPool, ok := obj.(*SnowIPPool)
	if !ok {
		return nil, fmt.Errorf("expected a SnowIPPool but got %T", obj)
	}

	snowippoollog.Info("validate delete", "name", snowIPPool.Name)

	return nil, nil
}

func validateImmutableFieldsSnowIPPool(new, old *SnowIPPool) field.ErrorList {
	var allErrs field.ErrorList

	if !SnowIPPoolsSliceEqual(new.Spec.Pools, old.Spec.Pools) {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath("spec").Child("pools"), "field is immutable"),
		)
	}
	return allErrs
}
