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

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-cluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=clusters,verbs=create;update,versions=v1alpha1,name=mutation.cluster.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Cluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Cluster) Default() {
	clusterlog.Info("Setting up Cluster defaults", "name", r.Name, "namespace", r.Namespace)
	r.SetDefaults()
}

// Change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-cluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=clusters,verbs=create;update,versions=v1alpha1,name=validation.cluster.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Cluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateCreate() error {
	clusterlog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	if !r.IsReconcilePaused() {
		if r.IsSelfManaged() {
			return apierrors.NewBadRequest("creating new cluster on existing cluster is not supported for self managed clusters")
		} else if !features.IsActive(features.FullLifecycleAPI()) {
			return apierrors.NewBadRequest("creating new managed cluster on existing cluster is not supported")
		}
	}

	if err := r.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), r.Spec, err.Error()))
	}

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind(ClusterKind).GroupKind(), r.Name, allErrs)
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

	if r.IsSelfManaged() && !r.IsReconcilePaused() && features.IsActive(features.FullLifecycleAPI()) {
		return apierrors.NewBadRequest(fmt.Sprintf("upgrading self managed clusters is not supported: %s", r.Name))
	}

	var allErrs field.ErrorList

	allErrs = append(allErrs, validateImmutableFieldsCluster(r, oldCluster)...)

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind(ClusterKind).GroupKind(), r.Name, allErrs)
	}

	if err := r.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), r.Spec, err.Error()))
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
			field.Forbidden(field.NewPath("spec", "managementCluster", new.Spec.ManagementCluster.Name), fmt.Sprintf("field is immutable %v", new.Spec.ManagementCluster)))
	}

	if !new.Spec.ControlPlaneConfiguration.Endpoint.Equal(old.Spec.ControlPlaneConfiguration.Endpoint) {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath("spec", "ControlPlaneConfiguration.endpoint"), fmt.Sprintf("field is immutable %v", new.Spec.ControlPlaneConfiguration.Endpoint)))
	}

	if !new.Spec.DatacenterRef.Equal(&old.Spec.DatacenterRef) {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath("spec", "datacenterRef"), fmt.Sprintf("field is immutable %v", new.Spec.DatacenterRef)))
	}

	if !new.Spec.ClusterNetwork.Equal(&old.Spec.ClusterNetwork) {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath("spec", "ClusterNetwork"), fmt.Sprintf("field is immutable %v", new.Spec.ClusterNetwork)))
	}

	if !new.Spec.ProxyConfiguration.Equal(old.Spec.ProxyConfiguration) {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath("spec", "ProxyConfiguration"), fmt.Sprintf("field is immutable %v", new.Spec.ProxyConfiguration)))
	}

	if new.Spec.ExternalEtcdConfiguration != nil && old.Spec.ExternalEtcdConfiguration == nil {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath("spec.externalEtcdConfiguration"), "cannot switch from local to external etcd topology"),
		)
	}
	if new.Spec.ExternalEtcdConfiguration != nil && old.Spec.ExternalEtcdConfiguration != nil {
		if old.Spec.ExternalEtcdConfiguration.Count != new.Spec.ExternalEtcdConfiguration.Count {
			allErrs = append(
				allErrs,
				field.Forbidden(field.NewPath("spec.externalEtcdConfiguration.count"), fmt.Sprintf("field is immutable %v", new.Spec.ExternalEtcdConfiguration.Count)),
			)
		}
	}

	if !new.Spec.GitOpsRef.Equal(old.Spec.GitOpsRef) {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath("spec", "GitOpsRef"), fmt.Sprintf("field is immutable %v", new.Spec.GitOpsRef)))
	}

	if !old.IsSelfManaged() {
		clusterlog.Info("Cluster config is associated with workload cluster", "name", old.Name)

		oldAWSIamConfig, newAWSIamConfig := &Ref{}, &Ref{}
		for _, identityProvider := range new.Spec.IdentityProviderRefs {
			if identityProvider.Kind == AWSIamConfigKind {
				newAWSIamConfig = &identityProvider
				break
			}
		}

		for _, identityProvider := range old.Spec.IdentityProviderRefs {
			if identityProvider.Kind == AWSIamConfigKind {
				oldAWSIamConfig = &identityProvider
				break
			}
		}

		if !oldAWSIamConfig.Equal(newAWSIamConfig) {
			allErrs = append(
				allErrs,
				field.Forbidden(field.NewPath("spec", "AWS Iam Config"), fmt.Sprintf("field is immutable %v", newAWSIamConfig.Kind)))
		}
		return allErrs
	}

	clusterlog.Info("Cluster config is associated with management cluster", "name", old.Name)

	if !RefSliceEqual(new.Spec.IdentityProviderRefs, old.Spec.IdentityProviderRefs) {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath("spec", "IdentityProviderRefs"), fmt.Sprintf("field is immutable %v", new.Spec.IdentityProviderRefs)))
	}

	if old.Spec.KubernetesVersion != new.Spec.KubernetesVersion {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath("spec", "kubernetesVersion"), fmt.Sprintf("field is immutable %v", new.Spec.KubernetesVersion)),
		)
	}

	if !old.Spec.ControlPlaneConfiguration.Equal(&new.Spec.ControlPlaneConfiguration) {
		allErrs = append(
			allErrs,
			field.Forbidden(field.NewPath("spec", "ControlPlaneConfiguration"), fmt.Sprintf("field is immutable %v", new.Spec.ControlPlaneConfiguration)))
	}

	return allErrs
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateDelete() error {
	clusterlog.Info("validate delete", "name", r.Name)

	return nil
}
