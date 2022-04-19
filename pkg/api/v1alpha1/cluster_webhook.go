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

	"github.com/aws/eks-anywhere/pkg/features"
)

// log is for logging in this package.
var clusterlog = logf.Log.WithName("cluster-resource")

func (r *Cluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// Change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-cluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=clusters,verbs=create;update,versions=v1alpha1,name=validation.cluster.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Cluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateCreate() error {
	clusterlog.Info("validate create", "name", r.Name)
	if r.IsReconcilePaused() {
		clusterlog.Info("cluster is paused, so allowing create", "name", r.Name)
		return nil
	}
	if features.IsActive(features.CloudStackProvider()) && r.Spec.DatacenterRef.Kind == CloudStackDatacenterKind &&
		len(r.Spec.WorkerNodeGroupConfigurations) > 1 {
		return apierrors.NewBadRequest("Multiple worker node groups is not supported for CloudStack provider")
	}
	if !features.IsActive(features.FullLifecycleAPI()) {
		return apierrors.NewBadRequest("Creating new cluster on existing cluster is not supported")
	}
	if r.IsSelfManaged() {
		return apierrors.NewBadRequest("Creating new cluster on existing cluster is not supported")
	}

	if err := validateCNIPlugin(r.Spec.ClusterNetwork); err != nil {
		return apierrors.NewBadRequest(err.Error())
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateUpdate(old runtime.Object) error {
	clusterlog.Info("validate update", "name", r.Name)
	oldCluster, ok := old.(*Cluster)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a Cluster but got a %T", old))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableFieldsCluster(r, oldCluster)...)

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind(ClusterKind).GroupKind(), r.Name, allErrs)
	}

	// Test for both taints and labels
	if err := validateWorkerNodeGroups(r); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "workerNodeGroupConfigurations"), r.Spec.WorkerNodeGroupConfigurations, err.Error()))
	}

	// Control plane configuration is mutable if workload cluster
	if !r.IsSelfManaged() {
		if err := validateControlPlaneLabels(r); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "controlPlaneConfiguration", "labels"), r.Spec, err.Error()))
		}
	}

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind(ClusterKind).GroupKind(), r.Name, allErrs)
	}

	return nil
}

func validateImmutableFieldsCluster(new, old *Cluster) field.ErrorList {
	if old.IsReconcilePaused() {
		return nil
	}

	var allErrs field.ErrorList

	if !old.ManagementClusterEqual(new) {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "managementCluster"), new.Spec.ManagementCluster, "field is immutable"))
	}

	if !new.Spec.ControlPlaneConfiguration.Endpoint.Equal(old.Spec.ControlPlaneConfiguration.Endpoint) {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "ControlPlaneConfiguration.endpoint"), new.Spec.ControlPlaneConfiguration.Endpoint, "field is immutable"))
	}

	if !new.Spec.DatacenterRef.Equal(&old.Spec.DatacenterRef) {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "datacenterRef"), new.Spec.DatacenterRef, "field is immutable"))
	}

	if !new.Spec.ClusterNetwork.Equal(&old.Spec.ClusterNetwork) {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "ClusterNetwork"), new.Spec.ClusterNetwork, "field is immutable"))
	}

	if !new.Spec.ProxyConfiguration.Equal(old.Spec.ProxyConfiguration) {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "ProxyConfiguration"), new.Spec.ProxyConfiguration, "field is immutable"))
	}

	if new.Spec.ExternalEtcdConfiguration != nil && old.Spec.ExternalEtcdConfiguration == nil {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec.externalEtcdConfiguration"), new.Spec.ExternalEtcdConfiguration, "cannot switch from local to external etcd topology"),
		)
	}
	if new.Spec.ExternalEtcdConfiguration != nil && old.Spec.ExternalEtcdConfiguration != nil {
		if old.Spec.ExternalEtcdConfiguration.Count != new.Spec.ExternalEtcdConfiguration.Count {
			allErrs = append(
				allErrs,
				field.Invalid(field.NewPath("spec.externalEtcdConfiguration.count"), new.Spec.ExternalEtcdConfiguration.Count, "field is immutable"),
			)
		}
	}

	if !new.Spec.GitOpsRef.Equal(old.Spec.GitOpsRef) {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "GitOpsRef"), new.Spec.GitOpsRef, "field is immutable"))
	}

	if !old.IsSelfManaged() {
		clusterlog.Info("Cluster config is associated with workload cluster", "name", old.Name)

		oldAWSIamConfig, newAWSIamConfig := &Ref{}, &Ref{}
		for _, identityProvider := range new.Spec.IdentityProviderRefs {
			if identityProvider.Kind == AWSIamConfigKind {
				newAWSIamConfig = &identityProvider
			}
		}

		for _, identityProvider := range old.Spec.IdentityProviderRefs {
			if identityProvider.Kind == AWSIamConfigKind {
				oldAWSIamConfig = &identityProvider
			}
		}

		if !oldAWSIamConfig.Equal(newAWSIamConfig) {
			allErrs = append(
				allErrs,
				field.Invalid(field.NewPath("spec", "AWS Iam Config"), newAWSIamConfig.Kind, "field is immutable"))
		}
		return allErrs
	}

	clusterlog.Info("Cluster config is associated with management cluster", "name", old.Name)

	if !RefSliceEqual(new.Spec.IdentityProviderRefs, old.Spec.IdentityProviderRefs) {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "IdentityProviderRefs"), new.Spec.IdentityProviderRefs, "field is immutable"))
	}

	if old.Spec.KubernetesVersion != new.Spec.KubernetesVersion {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "kubernetesVersion"), new.Spec.KubernetesVersion, "field is immutable"),
		)
	}

	if !old.Spec.ControlPlaneConfiguration.Equal(&new.Spec.ControlPlaneConfiguration) {
		allErrs = append(
			allErrs,
			field.Invalid(field.NewPath("spec", "ControlPlaneConfiguration"), new.Spec.ControlPlaneConfiguration, "field is immutable"))
	}

	return allErrs
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateDelete() error {
	clusterlog.Info("validate delete", "name", r.Name)

	return nil
}
