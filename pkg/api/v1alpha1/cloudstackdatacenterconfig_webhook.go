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
	"regexp"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var cloudstackdatacenterconfiglog = logf.Log.WithName("cloudstackdatacenterconfig-resource")

func (r *CloudStackDatacenterConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-cloudstackdatacenterconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=cloudstackdatacenterconfigs,verbs=create;update,versions=v1alpha1,name=mutation.cloudstackdatacenterconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &CloudStackDatacenterConfig{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (r *CloudStackDatacenterConfig) Default(_ context.Context, obj runtime.Object) error {
	cloudstackConfig, ok := obj.(*CloudStackDatacenterConfig)
	if !ok {
		return fmt.Errorf("expected a CloudStackDatacenterConfig but got %T", obj)
	}

	cloudstackdatacenterconfiglog.Info("Setting up CloudStackDatacenterConfig defaults for", "name", cloudstackConfig.Name)
	cloudstackConfig.SetDefaults()

	return nil
}

// change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-cloudstackdatacenterconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=cloudstackdatacenterconfigs,verbs=create;update,versions=v1alpha1,name=validation.cloudstackdatacenterconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &CloudStackDatacenterConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *CloudStackDatacenterConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cloudstackConfig, ok := obj.(*CloudStackDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a CloudStackDatacenterConfig but got %T", obj)
	}

	cloudstackdatacenterconfiglog.Info("validate create", "name", cloudstackConfig.Name)
	if cloudstackConfig.IsReconcilePaused() {
		cloudstackdatacenterconfiglog.Info("CloudStackDatacenterConfig is paused, so allowing create", "name", cloudstackConfig.Name)
		return nil, nil
	}
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *CloudStackDatacenterConfig) ValidateUpdate(_ context.Context, obj, old runtime.Object) (admission.Warnings, error) {
	cloudstackConfig, ok := obj.(*CloudStackDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a CloudStackDatacenterConfig but got %T", obj)
	}

	cloudstackdatacenterconfiglog.Info("validate update", "name", cloudstackConfig.Name)

	oldDatacenterConfig, ok := old.(*CloudStackDatacenterConfig)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a CloudStackDataCenterConfig but got a %T", old))
	}

	if oldDatacenterConfig.IsReconcilePaused() {
		cloudstackdatacenterconfiglog.Info("Reconciliation is paused")
		return nil, nil
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableFieldsCloudStackCluster(r, oldDatacenterConfig)...)

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(GroupVersion.WithKind(CloudStackDatacenterKind).GroupKind(), r.Name, allErrs)
}

func isValidAzConversionName(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}")
	return r.MatchString(uuid)
}

func isCapcV1beta1ToV1beta2Upgrade(new, old *CloudStackDatacenterConfigSpec) bool {
	if len(new.AvailabilityZones) != len(old.AvailabilityZones) {
		return false
	}
	for _, az := range old.AvailabilityZones {
		if !strings.HasPrefix(az.Name, DefaultCloudStackAZPrefix) {
			return false
		}
	}
	for _, az := range new.AvailabilityZones {
		if !isValidAzConversionName(az.Name) {
			return false
		}
	}

	return true
}

func validateImmutableFieldsCloudStackCluster(new, old *CloudStackDatacenterConfig) field.ErrorList {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	// Check for CAPC v1beta1 -> CAPC v1beta2 upgrade
	if isCapcV1beta1ToV1beta2Upgrade(&new.Spec, &old.Spec) {
		return allErrs
	}
	newAzMap := make(map[string]CloudStackAvailabilityZone)
	for _, az := range new.Spec.AvailabilityZones {
		newAzMap[az.Name] = az
	}
	atLeastOneAzOverlap := false
	for _, oldAz := range old.Spec.AvailabilityZones {
		if newAz, ok := newAzMap[oldAz.Name]; ok {
			atLeastOneAzOverlap = true
			if !newAz.Equal(&oldAz) {
				allErrs = append(
					allErrs,
					field.Forbidden(specPath.Child("availabilityZone", oldAz.Name), "availabilityZone is immutable"),
				)
			}
		}
	}
	if !atLeastOneAzOverlap {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "availabilityZone"), new.Spec.AvailabilityZones, "at least one AvailabilityZone must be shared between new and old CloudStackDatacenterConfig specs"),
		)
	}

	return allErrs
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *CloudStackDatacenterConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cloudstackConfig, ok := obj.(*CloudStackDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a CloudStackDatacenterConfig but got %T", obj)
	}

	cloudstackdatacenterconfiglog.Info("validate delete", "name", cloudstackConfig.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
