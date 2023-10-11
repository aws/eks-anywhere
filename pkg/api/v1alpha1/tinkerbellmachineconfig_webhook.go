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
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var tinkerbellmachineconfiglog = logf.Log.WithName("tinkerbellmachineconfig-resource")

// SetupWebhookWithManager sets up TinkerbellMachineConfig webhook to controller manager.
func (r *TinkerbellMachineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-tinkerbellmachineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=tinkerbellmachineconfigs,verbs=create;update,versions=v1alpha1,name=mutation.tinkerbellmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &TinkerbellMachineConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *TinkerbellMachineConfig) Default() {
	tinkerbellmachineconfiglog.Info("Setting up Tinkerbell Machine Config defaults", klog.KObj(r))
	r.SetDefaults()
	tinkerbellmachineconfiglog.Info("Normalize SSHKeys by removing comments from the keys", klog.KObj(r))
	normalizeSSHKeys(r)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-tinkerbellmachineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=tinkerbellmachineconfigs,verbs=create;update,versions=v1alpha1,name=validation.tinkerbellmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &TinkerbellMachineConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *TinkerbellMachineConfig) ValidateCreate() error {
	tinkerbellmachineconfiglog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	if err := r.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), r.Spec, err.Error()))
	}

	if len(r.Spec.Users) > 0 {
		if len(r.Spec.Users[0].SshAuthorizedKeys) == 0 || r.Spec.Users[0].SshAuthorizedKeys[0] == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), r.Spec, fmt.Sprintf("TinkerbellMachineConfig: missing spec.Users[0].SshAuthorizedKeys: %s for user %s. Please specify a ssh authorized key", r.Name, r.Spec.Users[0])))
		}
	}

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind(ClusterKind).GroupKind(), r.Name, allErrs)
	}

	if r.IsReconcilePaused() {
		tinkerbellmachineconfiglog.Info("TinkerbellMachineConfig is paused, so allowing create", "name", r.Name)
		return nil
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *TinkerbellMachineConfig) ValidateUpdate(old runtime.Object) error {
	tinkerbellmachineconfiglog.Info("validate update", "name", r.Name)

	oldTinkerbellMachineConfig, ok := old.(*TinkerbellMachineConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a TinkerbellMachineConfig but got a %T", old))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableFieldsTinkerbellMachineConfig(r, oldTinkerbellMachineConfig)...)
	if len(allErrs) != 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind(TinkerbellMachineConfigKind).GroupKind(), r.Name, allErrs)
	}

	if err := r.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), r.Spec, err.Error()))
	}

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind(TinkerbellMachineConfigKind).GroupKind(), r.Name, allErrs)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *TinkerbellMachineConfig) ValidateDelete() error {
	tinkerbellmachineconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
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

	if !reflect.DeepEqual(new.Spec.HardwareSelector, old.Spec.HardwareSelector) {
		allErrs = append(allErrs, field.Forbidden(specPath.Child("HardwareSelector"), "field is immutable"))
	}

	return allErrs
}
