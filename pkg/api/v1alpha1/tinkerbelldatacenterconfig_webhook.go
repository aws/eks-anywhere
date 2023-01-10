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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var tinkerbelldatacenterconfiglog = logf.Log.WithName("tinkerbelldatacenterconfig-resource")

// SetupWebhookWithManager sets up TinkerbellDatacenterConfig webhook to controller manager.
func (r *TinkerbellDatacenterConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-tinkerbelldatacenterconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=tinkerbelldatacenterconfigs,verbs=create;update,versions=v1alpha1,name=validation.tinkerbelldatacenterconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &TinkerbellDatacenterConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *TinkerbellDatacenterConfig) ValidateCreate() error {
	tinkerbelldatacenterconfiglog.Info("validate create", "name", r.Name)

	if err := r.Validate(); err != nil {
		return apierrors.NewInvalid(
			GroupVersion.WithKind(TinkerbellDatacenterKind).GroupKind(),
			r.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec"), r.Spec, err.Error()),
			},
		)
	}

	if r.IsReconcilePaused() {
		tinkerbelldatacenterconfiglog.Info("TinkerbellDatacenterConfig is paused, so allowing create", "name", r.Name)
		return nil
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *TinkerbellDatacenterConfig) ValidateUpdate(old runtime.Object) error {
	tinkerbelldatacenterconfiglog.Info("validate update", "name", r.Name)

	oldTinkerbellDatacenterConfig, ok := old.(*TinkerbellDatacenterConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a TinkerbellDatacenterConfig but got a %T", old))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableFieldsTinkerbellDatacenterConfig(r, oldTinkerbellDatacenterConfig)...)

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind(TinkerbellDatacenterKind).GroupKind(), r.Name, allErrs)
	}

	if err := r.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), r.Spec, err.Error()))
	}

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind(TinkerbellDatacenterKind).GroupKind(), r.Name, allErrs)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *TinkerbellDatacenterConfig) ValidateDelete() error {
	tinkerbelldatacenterconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func validateImmutableFieldsTinkerbellDatacenterConfig(new, old *TinkerbellDatacenterConfig) field.ErrorList {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	if new.Spec.TinkerbellIP != old.Spec.TinkerbellIP {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("tinkerbellIP"), "field is immutable"),
		)
	}

	return allErrs
}
