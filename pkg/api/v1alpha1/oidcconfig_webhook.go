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
var oidcconfiglog = logf.Log.WithName("oidcconfig-resource")

func (r *OIDCConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-oidcconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=oidcconfigs,verbs=create;update,versions=v1alpha1,name=validation.oidcconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &OIDCConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *OIDCConfig) ValidateCreate() error {
	oidcconfiglog.Info("validate create", "name", r.Name)

	allErrs := r.Validate()

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(GroupVersion.WithKind(OIDCConfigKind).GroupKind(), r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *OIDCConfig) ValidateUpdate(old runtime.Object) error {
	oidcconfiglog.Info("validate update", "name", r.Name)

	oldOIDCConfig, ok := old.(*OIDCConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a OIDCConfig but got a %T", old))
	}

	if oldOIDCConfig.IsManaged() {
		clusterlog.Info("OIDC config is associated with workload cluster", "name", oldOIDCConfig.Name)
		return nil
	}

	clusterlog.Info("OIDC config is associated with management cluster", "name", oldOIDCConfig.Name)

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableOIDCFields(r, oldOIDCConfig)...)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(GroupVersion.WithKind(OIDCConfigKind).GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *OIDCConfig) ValidateDelete() error {
	oidcconfiglog.Info("validate delete", "name", r.Name)

	return nil
}

func validateImmutableOIDCFields(new, old *OIDCConfig) field.ErrorList {
	var allErrs field.ErrorList

	if !new.Spec.Equal(&old.Spec) {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath(OIDCConfigKind), "config is immutable"),
		)
	}

	return allErrs
}
