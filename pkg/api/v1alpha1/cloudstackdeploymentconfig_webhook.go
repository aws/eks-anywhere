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
var cloudstackdeploymentconfiglog = logf.Log.WithName("cloudstackdeploymentconfig-resource")

func (r *CloudStackDeploymentConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-cloudstackdeploymentconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=cloudstackdeploymentconfigs,verbs=create;update,versions=v1alpha1,name=validation.cloudstackdeploymentconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &CloudStackDeploymentConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CloudStackDeploymentConfig) ValidateCreate() error {
	cloudstackdeploymentconfiglog.Info("validate create", "name", r.Name)
	if r.IsReconcilePaused() {
		cloudstackdeploymentconfiglog.Info("CloudStackDeploymentConfig is paused, so allowing create", "name", r.Name)
		return nil
	}
	return apierrors.NewBadRequest("Creating new CloudStackDeploymentConfig on existing cluster is not supported")
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CloudStackDeploymentConfig) ValidateUpdate(old runtime.Object) error {
	cloudstackdeploymentconfiglog.Info("validate update", "name", r.Name)

	oldDatacenterConfig, ok := old.(*CloudStackDeploymentConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a CloudStackDataCenterConfig but got a %T", old))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableFieldsCloudStackCluster(r, oldDatacenterConfig)...)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(GroupVersion.WithKind(CloudStackDeploymentKind).GroupKind(), r.Name, allErrs)
}

func validateImmutableFieldsCloudStackCluster(new, old *CloudStackDeploymentConfig) field.ErrorList {
	if old.IsReconcilePaused() {
		cloudstackdeploymentconfiglog.Info("Reconciliation is paused")
		return nil
	}

	var allErrs field.ErrorList

	if old.Spec.Domain != new.Spec.Domain {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "domain"), new.Spec.Domain, "field is immutable"),
		)
	}

	if old.Spec.Zone != new.Spec.Zone {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "zone"), new.Spec.Zone, "field is immutable"),
		)
	}

	if old.Spec.Account != new.Spec.Account {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "account"), new.Spec.Account, "field is immutable"),
		)
	}

	if old.Spec.Network != new.Spec.Network {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "network"), new.Spec.Network, "field is immutable"),
		)
	}

	if old.Spec.ManagementApiEndpoint != new.Spec.ManagementApiEndpoint {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "control_plane_endpoint"), new.Spec.ManagementApiEndpoint, "field is immutable"),
		)
	}

	if old.Spec.Insecure != new.Spec.Insecure {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "insecure"), new.Spec.Insecure, "field is immutable"),
		)
	}

	return allErrs
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CloudStackDeploymentConfig) ValidateDelete() error {
	cloudstackdeploymentconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
