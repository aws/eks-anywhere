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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var tinkerbelldatacenterconfiglog = logf.Log.WithName("tinkerbelldatacenterconfig-resource")

// SetupWebhookWithManager sets up TinkerbellDatacenterConfig webhook to controller manager.
func (r *TinkerbellDatacenterConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithValidator(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-tinkerbelldatacenterconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=tinkerbelldatacenterconfigs,verbs=create;update,versions=v1alpha1,name=validation.tinkerbelldatacenterconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &TinkerbellDatacenterConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *TinkerbellDatacenterConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	tinkerbellConfig, ok := obj.(*TinkerbellDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a TinkerbellDatacenterConfig but got %T", obj)
	}

	tinkerbelldatacenterconfiglog.Info("validate create", "name", tinkerbellConfig.Name)

	if err := tinkerbellConfig.Validate(); err != nil {
		return nil, apierrors.NewInvalid(
			GroupVersion.WithKind(TinkerbellDatacenterKind).GroupKind(),
			tinkerbellConfig.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec"), tinkerbellConfig.Spec, err.Error()),
			},
		)
	}

	if tinkerbellConfig.IsReconcilePaused() {
		tinkerbelldatacenterconfiglog.Info("TinkerbellDatacenterConfig is paused, so allowing create", "name", tinkerbellConfig.Name)
		return nil, nil
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *TinkerbellDatacenterConfig) ValidateUpdate(_ context.Context, old, obj runtime.Object) (admission.Warnings, error) {
	tinkerbellConfig, ok := obj.(*TinkerbellDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a TinkerbellDatacenterConfig but got %T", obj)
	}

	tinkerbelldatacenterconfiglog.Info("validate update", "name", tinkerbellConfig.Name)

	oldTinkerbellDatacenterConfig, ok := old.(*TinkerbellDatacenterConfig)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a TinkerbellDatacenterConfig but got a %T", old))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableFieldsTinkerbellDatacenterConfig(tinkerbellConfig, oldTinkerbellDatacenterConfig)...)

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(TinkerbellDatacenterKind).GroupKind(), tinkerbellConfig.Name, allErrs)
	}

	if err := tinkerbellConfig.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), tinkerbellConfig.Spec, err.Error()))
	}

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(TinkerbellDatacenterKind).GroupKind(), tinkerbellConfig.Name, allErrs)
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *TinkerbellDatacenterConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	tinkerbellConfig, ok := obj.(*TinkerbellDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a TinkerbellDatacenterConfig but got %T", obj)
	}

	tinkerbelldatacenterconfiglog.Info("validate delete", "name", tinkerbellConfig.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
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

	if new.Spec.HookImagesURLPath != old.Spec.HookImagesURLPath && !metav1.HasAnnotation(new.ObjectMeta, ManagedByCLIAnnotation) {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("hookImagesURLPath"), "field is immutable"),
		)
	}

	if new.Spec.SkipLoadBalancerDeployment != old.Spec.SkipLoadBalancerDeployment && !metav1.HasAnnotation(new.ObjectMeta, ManagedByCLIAnnotation) {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("skipLoadBalancerDeployment"), "field is immutable"),
		)
	}

	return allErrs
}
