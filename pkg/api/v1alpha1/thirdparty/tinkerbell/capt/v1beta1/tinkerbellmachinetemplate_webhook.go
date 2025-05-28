/*
Copyright 2022 The Tinkerbell Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"context"
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ webhook.CustomValidator = &TinkerbellMachineTemplate{}

// SetupWebhookWithManager sets up and registers the webhook with the manager.
func (m *TinkerbellMachineTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(m).
		WithValidator(m).Complete() //nolint:wrapcheck
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (m *TinkerbellMachineTemplate) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	template, ok := obj.(*TinkerbellMachineTemplate)
	if !ok {
		return nil, fmt.Errorf("expected a TinkerbellMachineTemplate but got %T", obj)
	}

	var allErrs field.ErrorList

	spec := template.Spec.Template.Spec
	fieldBasePath := field.NewPath("spec", "template", "spec")

	if spec.ProviderID != "" {
		allErrs = append(allErrs, field.Forbidden(fieldBasePath.Child("providerID"), "cannot be set in templates"))
	}

	if spec.HardwareName != "" {
		allErrs = append(allErrs, field.Forbidden(fieldBasePath.Child("hardwareName"), "cannot be set in templates"))
	}

	return nil, aggregateObjErrors(template.GroupVersionKind().GroupKind(), template.Name, allErrs)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (m *TinkerbellMachineTemplate) ValidateUpdate(_ context.Context, old, obj runtime.Object) (admission.Warnings, error) {
	template, ok := obj.(*TinkerbellMachineTemplate)
	if !ok {
		return nil, fmt.Errorf("expected a TinkerbellMachineTemplate but got %T", obj)
	}

	oldTinkerbellMachineTemplate, ok := old.(*TinkerbellMachineTemplate)
	if !ok {
		return nil, fmt.Errorf("expected a TinkerbellMachineTemplate but got %T", old)
	}

	if !reflect.DeepEqual(template.Spec, oldTinkerbellMachineTemplate.Spec) {
		return nil, apierrors.NewBadRequest("TinkerbellMachineTemplate.Spec is immutable")
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (m *TinkerbellMachineTemplate) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*TinkerbellMachineTemplate)
	if !ok {
		return nil, fmt.Errorf("expected a TinkerbellMachineTemplate but got %T", obj)
	}

	return nil, nil
}
