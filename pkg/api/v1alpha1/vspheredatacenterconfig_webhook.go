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
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var vspheredatacenterconfiglog = logf.Log.WithName("vspheredatacenterconfig-resource")

func (r *VSphereDatacenterConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-vspheredatacenterconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=vspheredatacenterconfigs,verbs=create;update,versions=v1alpha1,name=mutation.vspheredatacenterconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &VSphereDatacenterConfig{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (r *VSphereDatacenterConfig) Default(_ context.Context, obj runtime.Object) error {
	vsphereConfig, ok := obj.(*VSphereDatacenterConfig)
	if !ok {
		return fmt.Errorf("expected a VSphereDatacenterConfig but got %T", obj)
	}

	vspheredatacenterconfiglog.Info("Setting up VSphere Datacenter Config defaults for", "name", vsphereConfig.Name)
	vsphereConfig.SetDefaults()

	return nil
}

// change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-vspheredatacenterconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=vspheredatacenterconfigs,verbs=create;update,versions=v1alpha1,name=validation.vspheredatacenterconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &VSphereDatacenterConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *VSphereDatacenterConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	vsphereConfig, ok := obj.(*VSphereDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a VSphereDatacenterConfig but got %T", obj)
	}

	vspheredatacenterconfiglog.Info("validate create", "name", vsphereConfig.Name)

	if err := vsphereConfig.Validate(); err != nil {
		return nil, apierrors.NewInvalid(
			GroupVersion.WithKind(VSphereDatacenterKind).GroupKind(),
			vsphereConfig.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec"), vsphereConfig.Spec, err.Error()),
			},
		)
	}

	if vsphereConfig.IsReconcilePaused() {
		vspheredatacenterconfiglog.Info("VSphereDatacenterConfig is paused, so allowing create", "name", vsphereConfig.Name)
		return nil, nil
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *VSphereDatacenterConfig) ValidateUpdate(_ context.Context, obj, old runtime.Object) (admission.Warnings, error) {
	vsphereConfig, ok := obj.(*VSphereDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a VSphereDatacenterConfig but got %T", obj)
	}

	vspheredatacenterconfiglog.Info("validate update", "name", vsphereConfig.Name)

	oldDatacenterConfig, ok := old.(*VSphereDatacenterConfig)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a VSphereDataCenterConfig but got a %T", old))
	}

	if err := vsphereConfig.Validate(); err != nil {
		return nil, apierrors.NewInvalid(
			GroupVersion.WithKind(VSphereDatacenterKind).GroupKind(),
			vsphereConfig.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec"), vsphereConfig.Spec, err.Error()),
			},
		)
	}

	if oldDatacenterConfig.IsReconcilePaused() {
		vspheredatacenterconfiglog.Info("Reconciliation is paused")
		return nil, nil
	}

	vsphereConfig.SetDefaults()

	if allErrs := validateImmutableFieldsVSphereCluster(vsphereConfig, oldDatacenterConfig); len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(VSphereDatacenterKind).GroupKind(), vsphereConfig.Name, allErrs)
	}

	return nil, nil
}

func validateImmutableFieldsVSphereCluster(new, old *VSphereDatacenterConfig) field.ErrorList {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	if old.Spec.Server != new.Spec.Server {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("server"), "field is immutable"),
		)
	}

	if old.Spec.Datacenter != new.Spec.Datacenter {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("datacenter"), "field is immutable"),
		)
	}

	if old.Spec.Network != new.Spec.Network {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("network"), "field is immutable"),
		)
	}

	return allErrs
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *VSphereDatacenterConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	vsphereConfig, ok := obj.(*VSphereDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a VSphereDatacenterConfig but got %T", obj)
	}

	vspheredatacenterconfiglog.Info("validate delete", "name", vsphereConfig.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
