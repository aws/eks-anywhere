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
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ admission.Validator = &TinkerbellMachineTemplate{}

// SetupWebhookWithManager sets up and registers the webhook with the manager.
func (m *TinkerbellMachineTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(m).Complete() //nolint:wrapcheck
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-tinkerbellmachinetemplate,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=tinkerbellmachinetemplates,versions=v1beta1,name=validation.tinkerbellmachinetemplate.infrastructure.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (m *TinkerbellMachineTemplate) ValidateCreate() (admission.Warnings, error) {
	var allErrs field.ErrorList

	spec := m.Spec.Template.Spec
	fieldBasePath := field.NewPath("spec", "template", "spec")

	if spec.ProviderID != "" {
		allErrs = append(allErrs, field.Forbidden(fieldBasePath.Child("providerID"), "cannot be set in templates"))
	}

	if spec.HardwareName != "" {
		allErrs = append(allErrs, field.Forbidden(fieldBasePath.Child("hardwareName"), "cannot be set in templates"))
	}

	return nil, aggregateObjErrors(m.GroupVersionKind().GroupKind(), m.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (m *TinkerbellMachineTemplate) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	oldTinkerbellMachineTemplate, _ := old.(*TinkerbellMachineTemplate)

	if !reflect.DeepEqual(m.Spec, oldTinkerbellMachineTemplate.Spec) {
		return nil, apierrors.NewBadRequest("TinkerbellMachineTemplate.Spec is immutable")
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (m *TinkerbellMachineTemplate) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}
