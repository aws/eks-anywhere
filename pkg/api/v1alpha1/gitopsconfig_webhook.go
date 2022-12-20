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
var gitopsconfiglog = logf.Log.WithName("gitopsconfig-resource")

func (r *GitOpsConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// Change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-gitopsconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=gitopsconfigs,verbs=create;update,versions=v1alpha1,name=validation.gitopsconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &GitOpsConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *GitOpsConfig) ValidateCreate() error {
	gitopsconfiglog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *GitOpsConfig) ValidateUpdate(old runtime.Object) error {
	gitopsconfiglog.Info("validate update", "name", r.Name)

	oldGitOpsConfig, ok := old.(*GitOpsConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a GitOpsConfig but got a %T", old))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableGitOpsFields(r, oldGitOpsConfig)...)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(GroupVersion.WithKind(GitOpsConfigKind).GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *GitOpsConfig) ValidateDelete() error {
	gitopsconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func validateImmutableGitOpsFields(new, old *GitOpsConfig) field.ErrorList {
	var allErrs field.ErrorList

	if !new.Spec.Equal(&old.Spec) {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath(GitOpsConfigKind), "config is immutable"),
		)
	}

	return allErrs
}
