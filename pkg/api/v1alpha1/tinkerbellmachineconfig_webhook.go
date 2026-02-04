// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var tinkerbellmachineconfiglog = logf.Log.WithName("tinkerbellmachineconfig-resource")

// SetupWebhookWithManager sets up TinkerbellMachineConfig webhook to controller manager.
func (r *TinkerbellMachineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-tinkerbellmachineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=tinkerbellmachineconfigs,verbs=create;update,versions=v1alpha1,name=mutation.tinkerbellmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &TinkerbellMachineConfig{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (r *TinkerbellMachineConfig) Default(_ context.Context, obj runtime.Object) error {
	tinkerbellConfig, ok := obj.(*TinkerbellMachineConfig)
	if !ok {
		return fmt.Errorf("expected a TinkerbellMachineConfig but got %T", obj)
	}

	tinkerbellmachineconfiglog.Info("Setting up Tinkerbell Machine Config defaults", klog.KObj(tinkerbellConfig))
	tinkerbellConfig.SetDefaults()
	tinkerbellmachineconfiglog.Info("Normalize SSHKeys by removing comments from the keys", klog.KObj(tinkerbellConfig))
	normalizeSSHKeys(tinkerbellConfig)

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-tinkerbellmachineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=tinkerbellmachineconfigs,verbs=create;update,versions=v1alpha1,name=validation.tinkerbellmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &TinkerbellMachineConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *TinkerbellMachineConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	tinkerbellConfig, ok := obj.(*TinkerbellMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a TinkerbellMachineConfig but got %T", obj)
	}

	tinkerbellmachineconfiglog.Info("validate create", "name", tinkerbellConfig.Name)

	var allErrs field.ErrorList

	if err := tinkerbellConfig.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), tinkerbellConfig.Spec, err.Error()))
	}

	if len(tinkerbellConfig.Spec.Users) > 0 {
		if len(tinkerbellConfig.Spec.Users[0].SshAuthorizedKeys) == 0 || tinkerbellConfig.Spec.Users[0].SshAuthorizedKeys[0] == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), tinkerbellConfig.Spec, fmt.Sprintf("TinkerbellMachineConfig: missing spec.Users[0].SshAuthorizedKeys: %s for user %s. Please specify a ssh authorized key", tinkerbellConfig.Name, tinkerbellConfig.Spec.Users[0])))
		}
	}

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(ClusterKind).GroupKind(), tinkerbellConfig.Name, allErrs)
	}

	if tinkerbellConfig.IsReconcilePaused() {
		tinkerbellmachineconfiglog.Info("TinkerbellMachineConfig is paused, so allowing create", "name", tinkerbellConfig.Name)
		return nil, nil
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *TinkerbellMachineConfig) ValidateUpdate(_ context.Context, old, obj runtime.Object) (admission.Warnings, error) {
	tinkerbellConfig, ok := obj.(*TinkerbellMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a TinkerbellMachineConfig but got %T", obj)
	}

	tinkerbellmachineconfiglog.Info("validate update", "name", tinkerbellConfig.Name)

	oldTinkerbellMachineConfig, ok := old.(*TinkerbellMachineConfig)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a TinkerbellMachineConfig but got a %T", old))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableFieldsTinkerbellMachineConfig(tinkerbellConfig, oldTinkerbellMachineConfig)...)
	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(TinkerbellMachineConfigKind).GroupKind(), tinkerbellConfig.Name, allErrs)
	}

	if err := tinkerbellConfig.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), tinkerbellConfig.Spec, err.Error()))
	}

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(TinkerbellMachineConfigKind).GroupKind(), tinkerbellConfig.Name, allErrs)
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *TinkerbellMachineConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	tinkerbellConfig, ok := obj.(*TinkerbellMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a TinkerbellMachineConfig but got %T", obj)
	}

	tinkerbellmachineconfiglog.Info("validate delete", "name", tinkerbellConfig.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func validateImmutableFieldsTinkerbellMachineConfig(new, old *TinkerbellMachineConfig) field.ErrorList {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	if new.Spec.OSFamily != old.Spec.OSFamily {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("OSFamily"), "field is immutable"))
	}

	if len(new.Spec.Users) != len(old.Spec.Users) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("Users"), "field is immutable"))
	}

	if new.Spec.Users[0].Name != old.Spec.Users[0].Name {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("Users[0].Name"), "field is immutable"))
	}

	if len(new.Spec.Users[0].SshAuthorizedKeys) != len(old.Spec.Users[0].SshAuthorizedKeys) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("Users[0].SshAuthorizedKeys"), "field is immutable"))
	}

	if len(new.Spec.Users[0].SshAuthorizedKeys) > 0 && (new.Spec.Users[0].SshAuthorizedKeys[0] != old.Spec.Users[0].SshAuthorizedKeys[0]) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("Users[0].SshAuthorizedKeys[0]"), "field is immutable"))
	}

	return allErrs
}
