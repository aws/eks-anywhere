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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ admission.Validator = &TinkerbellMachine{}

// SetupWebhookWithManager sets up and registers the webhook with the manager.
func (m *TinkerbellMachine) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(m).Complete() //nolint:wrapcheck
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-tinkerbellmachine,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=tinkerbellmachines,versions=v1beta1,name=validation.tinkerbellmachine.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (m *TinkerbellMachine) ValidateCreate() (admission.Warnings, error) {
	allErrs := m.validateSpec()

	return nil, aggregateObjErrors(m.GroupVersionKind().GroupKind(), m.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (m *TinkerbellMachine) ValidateUpdate(oldRaw runtime.Object) (admission.Warnings, error) {
	allErrs := m.validateSpec()

	old, _ := oldRaw.(*TinkerbellMachine)

	if old.Spec.HardwareName != "" && m.Spec.HardwareName != old.Spec.HardwareName {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec", "hardwareName"), "is immutable once set"))
	}

	if old.Spec.ProviderID != "" && m.Spec.ProviderID != old.Spec.ProviderID {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec", "providerID"), "is immutable once set"))
	}

	return nil, aggregateObjErrors(m.GroupVersionKind().GroupKind(), m.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (m *TinkerbellMachine) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (m *TinkerbellMachine) validateSpec() field.ErrorList {
	var allErrs field.ErrorList

	fieldBasePath := field.NewPath("spec")

	// TODO: there are probably more fields that have requirements

	if spec := m.Spec; spec.HardwareAffinity != nil {
		for i, term := range spec.HardwareAffinity.Preferred {
			if term.Weight < 1 || term.Weight > 100 {
				allErrs = append(allErrs,
					field.Invalid(fieldBasePath.Child("HardwareAffinity", "Preferred").Index(i),
						term.Weight, "must be in the range [1,100]"))
			}
		}
	}

	return allErrs
}
