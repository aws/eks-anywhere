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
var fluxconfiglog = logf.Log.WithName("fluxconfig-resource")

func (r *FluxConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// Change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-fluxconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=fluxconfigs,verbs=create;update,versions=v1alpha1,name=validation.fluxconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &FluxConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *FluxConfig) ValidateCreate() error {
	fluxconfiglog.Info("validate create", "name", r.Name)

	if err := r.Validate(); err != nil {
		return apierrors.NewInvalid(
			r.GroupVersionKind().GroupKind(),
			r.Name,
			field.ErrorList{field.Invalid(field.NewPath("spec"), r.Spec, err.Error())})
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *FluxConfig) ValidateUpdate(old runtime.Object) error {
	fluxconfiglog.Info("validate update", "name", r.Name)

	oldFluxConfig, ok := old.(*FluxConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a FluxConfig but got a %T", old))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableFluxFields(r, oldFluxConfig)...)

	if err := r.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), r.Spec, err.Error()))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(GroupVersion.WithKind(FluxConfigKind).GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *FluxConfig) ValidateDelete() error {
	fluxconfiglog.Info("validate delete", "name", r.Name)

	return nil
}

func validateImmutableFluxFields(new, old *FluxConfig) field.ErrorList {
	var allErrs field.ErrorList

	if !new.Spec.Equal(&old.Spec) {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath(FluxConfigKind), "config is immutable"),
		)
	}

	return allErrs
}
