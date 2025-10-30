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
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/util/version"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/aws/eks-anywhere/pkg/semver"
)

const supportedMinorVersionIncrement int64 = 1

// log is for logging in this package.
var clusterlog = logf.Log.WithName("cluster-resource")

func (r *Cluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-cluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=clusters,verbs=create;update,versions=v1alpha1,name=mutation.cluster.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &Cluster{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (r *Cluster) Default(_ context.Context, obj runtime.Object) error {
	cluster, ok := obj.(*Cluster)
	if !ok {
		return fmt.Errorf("expected a Cluster but got %T", obj)
	}

	clusterlog.Info("Setting up Cluster defaults", "name", cluster.Name, "namespace", cluster.Namespace)
	cluster.SetDefaults()

	return nil
}

// Change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-cluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=clusters,verbs=create;update,versions=v1alpha1,name=validation.cluster.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &Cluster{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *Cluster) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cluster, ok := obj.(*Cluster)
	if !ok {
		return nil, fmt.Errorf("expected a Cluster but got %T", obj)
	}

	clusterlog.Info("validate create", "name", cluster.Name)

	var allErrs field.ErrorList

	if !cluster.IsReconcilePaused() && cluster.IsSelfManaged() && !cluster.IsManagedByCLI() {
		return nil, apierrors.NewBadRequest("creating new cluster on existing cluster is not supported for self managed clusters")
	}

	if !cluster.IsReconcilePaused() && cluster.Spec.EtcdEncryption != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.etcdEncryption"), cluster.Spec.EtcdEncryption, "etcdEncryption is not supported during cluster creation"))
	}

	if err := cluster.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), cluster.Spec, err.Error()))
	}

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(ClusterKind).GroupKind(), cluster.Name, allErrs)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *Cluster) ValidateUpdate(_ context.Context, old, obj runtime.Object) (admission.Warnings, error) {
	newCluster, ok := obj.(*Cluster)
	if !ok {
		return nil, fmt.Errorf("expected a Cluster but got %T", obj)
	}

	clusterlog.Info("validate update", "name", newCluster.Name)
	oldCluster, ok := old.(*Cluster)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Cluster but got a %T", old))
	}

	var allErrs field.ErrorList

	if newCluster.Spec.DatacenterRef.Kind == TinkerbellDatacenterKind {
		allErrs = append(allErrs, validateUpgradeRequestTinkerbell(newCluster, oldCluster)...)
	}

	allErrs = append(allErrs, validateImmutableFieldsCluster(newCluster, oldCluster)...)

	allErrs = append(allErrs, validateBundlesRefCluster(newCluster, oldCluster)...)

	allErrs = append(allErrs, ValidateKubernetesVersionSkew(newCluster, oldCluster)...)

	allErrs = append(allErrs, validateEksaVersionCluster(newCluster, oldCluster)...)

	if !newCluster.EksaVersionSkewCheckDisabled() {
		allErrs = append(allErrs, ValidateEksaVersionSkew(newCluster, oldCluster)...)
	}

	allErrs = append(allErrs, ValidateWorkerKubernetesVersionSkew(newCluster, oldCluster)...)

	allErrs = append(allErrs, validateEtcdEncryptionSupport(newCluster)...)

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(ClusterKind).GroupKind(), newCluster.Name, allErrs)
	}

	if err := newCluster.Validate(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), newCluster.Spec, err.Error()))
	}

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(ClusterKind).GroupKind(), newCluster.Name, allErrs)
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *Cluster) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cluster, ok := obj.(*Cluster)
	if !ok {
		return nil, fmt.Errorf("expected a Cluster but got %T", obj)
	}

	clusterlog.Info("validate delete", "name", cluster.Name)

	return nil, nil
}

func supportsEtcdEncryption(cluster *Cluster) bool {
	return cluster.Spec.DatacenterRef.Kind == CloudStackDatacenterKind ||
		cluster.Spec.DatacenterRef.Kind == VSphereDatacenterKind ||
		cluster.Spec.DatacenterRef.Kind == NutanixDatacenterKind
}

// ValidateEksaVersionSkew ensures that upgrades are sequential by CLI minor versions.
func ValidateEksaVersionSkew(new, old *Cluster) field.ErrorList {
	var allErrs field.ErrorList
	eksaVersionPath := field.NewPath("spec").Child("EksaVersion")

	if new.Spec.EksaVersion == nil || old.Spec.EksaVersion == nil {
		return nil
	}

	// allow users to update cluster if old cluster is invalid
	oldEksaVersion, err := semver.New(string(*old.Spec.EksaVersion))
	if err != nil {
		return nil
	}

	newEksaVersion, err := semver.New(string(*new.Spec.EksaVersion))
	if err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(eksaVersionPath, new.Spec.EksaVersion, "EksaVersion is not a valid semver"))
		return allErrs
	}

	devBuildVersion, _ := semver.New(DevBuildVersion)
	if newEksaVersion.SamePatch(devBuildVersion) {
		return nil
	}

	majorVersionDifference := int64(newEksaVersion.Major) - int64(oldEksaVersion.Major)
	minorVersionDifference := int64(newEksaVersion.Minor) - int64(oldEksaVersion.Minor)

	// if major different or upgrade difference greater than one minor version
	if majorVersionDifference != 0 || minorVersionDifference > supportedMinorVersionIncrement {
		allErrs = append(
			allErrs,
			field.Invalid(eksaVersionPath, new.Spec.EksaVersion, fmt.Sprintf("cannot upgrade to %v from %v: EksaVersion upgrades must have a skew of 1 minor version", newEksaVersion, oldEksaVersion)))
		return allErrs
	}

	// Allow "downgrades" if old version is greater than managment cluster. We can't check if a cluster's EksaVersion
	// is less than or equal to the mgmt cluster in the webhook. Instead we check it in the controller where the
	// EksaVersion will already be applied. However, the cluster will never begin to reconcile due to the validation.
	// We should not block users from changing EksaVersion to a lower semver in this scenario.
	failure := old.Status.FailureReason
	if failure != nil && *failure == EksaVersionInvalidReason {
		return nil
	}

	// don't allow downgrades if old version was valid
	if minorVersionDifference < 0 {
		allErrs = append(
			allErrs,
			field.Invalid(eksaVersionPath, new.Spec.EksaVersion, fmt.Sprintf("cannot downgrade from %v to %v: EksaVersion upgrades must be incremental", oldEksaVersion, newEksaVersion)))
	}

	return allErrs
}

func validateBundlesRefCluster(new, old *Cluster) field.ErrorList {
	var allErrs field.ErrorList
	bundlesRefPath := field.NewPath("spec").Child("BundlesRef")

	if old.Spec.BundlesRef != nil && new.Spec.BundlesRef == nil && new.Spec.EksaVersion == nil {
		allErrs = append(
			allErrs,
			field.Invalid(bundlesRefPath, new.Spec.BundlesRef, fmt.Sprintf("field cannot be removed after setting. Previous value %v", old.Spec.BundlesRef)))
	}

	return allErrs
}

func validateEksaVersionCluster(new, old *Cluster) field.ErrorList {
	var allErrs field.ErrorList
	eksaVersionPath := field.NewPath("spec").Child("EksaVersion")

	if old.Spec.EksaVersion != nil && new.Spec.EksaVersion == nil {
		allErrs = append(
			allErrs,
			field.Invalid(eksaVersionPath, new.Spec.EksaVersion, fmt.Sprintf("field cannot be removed after setting. Previous value %v", old.Spec.EksaVersion)))
	}

	return allErrs
}

func validateUpgradeRequestTinkerbell(new, old *Cluster) field.ErrorList {
	var allErrs field.ErrorList
	path := field.NewPath("spec")

	if old.Spec.KubernetesVersion != new.Spec.KubernetesVersion {
		if old.Spec.ControlPlaneConfiguration.Count != new.Spec.ControlPlaneConfiguration.Count {
			allErrs = append(
				allErrs,
				field.Invalid(path, new.Spec.ControlPlaneConfiguration, fmt.Sprintf("cannot perform scale up or down during rolling upgrades. Previous control plane node count %v", old.Spec.ControlPlaneConfiguration.Count)))
		}

		if len(old.Spec.WorkerNodeGroupConfigurations) != len(new.Spec.WorkerNodeGroupConfigurations) {
			allErrs = append(
				allErrs,
				field.Invalid(path, new.Spec.WorkerNodeGroupConfigurations, "cannot perform scale up or down during rolling upgrades. Please revert to the previous worker node groups."))
		}
		workerNodeGroupMap := make(map[string]*WorkerNodeGroupConfiguration)
		for i := range old.Spec.WorkerNodeGroupConfigurations {
			workerNodeGroupMap[old.Spec.WorkerNodeGroupConfigurations[i].Name] = &old.Spec.WorkerNodeGroupConfigurations[i]
		}
		for _, nodeGroupNewSpec := range new.Spec.WorkerNodeGroupConfigurations {
			workerNodeGrpOldSpec, ok := workerNodeGroupMap[nodeGroupNewSpec.Name]
			if ok && *nodeGroupNewSpec.Count != *workerNodeGrpOldSpec.Count {
				allErrs = append(
					allErrs,
					field.Invalid(path, new.Spec.WorkerNodeGroupConfigurations, fmt.Sprintf("cannot perform scale up or down during rolling upgrades. Previous worker node count %v", *workerNodeGrpOldSpec.Count)))
			}
			if !ok {
				allErrs = append(
					allErrs,
					field.Invalid(path, new.Spec.WorkerNodeGroupConfigurations, fmt.Sprintf("cannot perform scale up or down during rolling upgrades. Please remove the new worker node group %s", nodeGroupNewSpec.Name)))
			}
		}
	}

	return allErrs
}

func validateImmutableFieldsCluster(new, old *Cluster) field.ErrorList {
	if old.IsReconcilePaused() {
		return nil
	}

	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	if !old.ManagementClusterEqual(new) {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("managementCluster", new.Spec.ManagementCluster.Name), fmt.Sprintf("field is immutable %v", new.Spec.ManagementCluster)))
	}

	if !new.Spec.DatacenterRef.Equal(&old.Spec.DatacenterRef) {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("datacenterRef"), fmt.Sprintf("field is immutable %v", new.Spec.DatacenterRef)))
	}

	if !new.Spec.ControlPlaneConfiguration.Endpoint.Equal(old.Spec.ControlPlaneConfiguration.Endpoint, new.Spec.DatacenterRef.Kind) {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("ControlPlaneConfiguration.endpoint"), fmt.Sprintf("field is immutable %v", new.Spec.ControlPlaneConfiguration.Endpoint)))
	}

	if !new.Spec.ClusterNetwork.Pods.Equal(&old.Spec.ClusterNetwork.Pods) {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("clusterNetwork", "pods"), "field is immutable"))
	}

	if !new.Spec.ClusterNetwork.Services.Equal(&old.Spec.ClusterNetwork.Services) {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("clusterNetwork", "services"), "field is immutable"))
	}

	if !new.Spec.ClusterNetwork.DNS.Equal(&old.Spec.ClusterNetwork.DNS) {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("clusterNetwork", "dns"), "field is immutable"))
	}

	if !new.Spec.ClusterNetwork.Nodes.Equal(old.Spec.ClusterNetwork.Nodes) {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("clusterNetwork", "nodes"), "field is immutable"))
	}

	if !new.Spec.ProxyConfiguration.Equal(old.Spec.ProxyConfiguration) {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("ProxyConfiguration"), fmt.Sprintf("field is immutable %v", new.Spec.ProxyConfiguration)))
	}

	if new.Spec.ExternalEtcdConfiguration != nil && old.Spec.ExternalEtcdConfiguration == nil {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("externalEtcdConfiguration"), "cannot switch from stacked to external etcd topology"),
		)
	}

	if new.Spec.ExternalEtcdConfiguration == nil && old.Spec.ExternalEtcdConfiguration != nil {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("externalEtcdConfiguration"), "cannot switch from external to stacked etcd topology"),
		)
	}

	if !new.Spec.GitOpsRef.Equal(old.Spec.GitOpsRef) && !old.IsSelfManaged() {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("GitOpsRef"), fmt.Sprintf("field is immutable %v", new.Spec.GitOpsRef)))
	}

	if new.Spec.DatacenterRef.Kind == TinkerbellDatacenterKind {
		if !reflect.DeepEqual(new.Spec.ControlPlaneConfiguration.Labels, old.Spec.ControlPlaneConfiguration.Labels) {
			allErrs = append(
				allErrs,
				field.Forbidden(specPath.Child("ControlPlaneConfiguration.labels"), fmt.Sprintf("field is immutable %v", new.Spec.ControlPlaneConfiguration.Labels)))
		}

		if !reflect.DeepEqual(new.Spec.ControlPlaneConfiguration.Taints, old.Spec.ControlPlaneConfiguration.Taints) {
			allErrs = append(
				allErrs,
				field.Forbidden(specPath.Child("ControlPlaneConfiguration.taints"), fmt.Sprintf("field is immutable %v", new.Spec.ControlPlaneConfiguration.Taints)))
		}

		workerNodeGroupMap := make(map[string]*WorkerNodeGroupConfiguration)
		for i := range old.Spec.WorkerNodeGroupConfigurations {
			workerNodeGroupMap[old.Spec.WorkerNodeGroupConfigurations[i].Name] = &old.Spec.WorkerNodeGroupConfigurations[i]
		}
		for _, nodeGroupNewSpec := range new.Spec.WorkerNodeGroupConfigurations {
			if workerNodeGrpOldSpec, ok := workerNodeGroupMap[nodeGroupNewSpec.Name]; ok {
				if !reflect.DeepEqual(workerNodeGrpOldSpec.Labels, nodeGroupNewSpec.Labels) {
					allErrs = append(
						allErrs,
						field.Forbidden(specPath.Child("WorkerNodeConfiguration.labels"), fmt.Sprintf("field is immutable %v", nodeGroupNewSpec.Labels)))
				}

				if !reflect.DeepEqual(workerNodeGrpOldSpec.Taints, nodeGroupNewSpec.Taints) {
					allErrs = append(
						allErrs,
						field.Forbidden(specPath.Child("WorkerNodeConfiguration.taints"), fmt.Sprintf("field is immutable %v", nodeGroupNewSpec.Taints)))
				}
			}
		}
	}

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

	if !oldAWSIamConfig.IsEmpty() && !newAWSIamConfig.IsEmpty() && !oldAWSIamConfig.Equal(newAWSIamConfig) {
		allErrs = append(
			allErrs,
			field.Forbidden(specPath.Child("identityProviderRefs", AWSIamConfigKind), fmt.Sprintf("field is immutable %v", newAWSIamConfig.Kind)))
	}

	return allErrs
}

// ValidateKubernetesVersionSkew validates Kubernetes version skew between upgrades.
func ValidateKubernetesVersionSkew(new, old *Cluster) field.ErrorList {
	path := field.NewPath("spec")
	oldVersion := old.Spec.KubernetesVersion
	newVersion := new.Spec.KubernetesVersion

	return validateKubeVersionSkew(newVersion, oldVersion, path)
}

func validateKubeVersionSkew(newVersion, oldVersion KubernetesVersion, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	parsedOldVersion, err := version.ParseGeneric(string(oldVersion))
	if err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(path, oldVersion, fmt.Sprintf("parsing cluster version: %v", err.Error())))
		return allErrs
	}

	parsedNewVersion, err := version.ParseGeneric(string(newVersion))
	if err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(path, newVersion, fmt.Sprintf("parsing comparison version: %v", err.Error())))
		return allErrs
	}

	if parsedNewVersion.Minor() == parsedOldVersion.Minor() && parsedNewVersion.Major() == parsedOldVersion.Major() {
		return allErrs
	}

	if err := ValidateVersionSkew(parsedOldVersion, parsedNewVersion); err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(path, newVersion, err.Error()))
	}

	return allErrs
}

// ValidateWorkerKubernetesVersionSkew validates worker node group Kubernetes version skew between upgrades.
func ValidateWorkerKubernetesVersionSkew(new, old *Cluster) field.ErrorList {
	var allErrs field.ErrorList
	newClusterVersion := new.Spec.KubernetesVersion
	oldClusterVersion := old.Spec.KubernetesVersion

	workerNodeGroupMap := make(map[string]*WorkerNodeGroupConfiguration)
	for i := range old.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMap[old.Spec.WorkerNodeGroupConfigurations[i].Name] = &old.Spec.WorkerNodeGroupConfigurations[i]
	}
	for _, nodeGroupNewSpec := range new.Spec.WorkerNodeGroupConfigurations {
		newVersion := nodeGroupNewSpec.KubernetesVersion

		if workerNodeGrpOldSpec, ok := workerNodeGroupMap[nodeGroupNewSpec.Name]; ok {
			oldVersion := workerNodeGrpOldSpec.KubernetesVersion
			allErrs = append(allErrs, performWorkerKubernetesValidations(oldVersion, newVersion, oldClusterVersion, newClusterVersion)...)
		} else {
			allErrs = append(allErrs, performWorkerKubernetesValidationsNewNodeGroup(newVersion, newClusterVersion)...)
		}
	}

	return allErrs
}

func performWorkerKubernetesValidationsNewNodeGroup(newVersion *KubernetesVersion, newClusterVersion KubernetesVersion) field.ErrorList {
	var allErrs field.ErrorList

	if newVersion != nil {
		allErrs = append(allErrs, validateCPWorkerKubeSkew(newClusterVersion, *newVersion)...)
	}

	return allErrs
}

func performWorkerKubernetesValidations(oldVersion, newVersion *KubernetesVersion, oldClusterVersion, newClusterVersion KubernetesVersion) field.ErrorList {
	var allErrs field.ErrorList
	path := field.NewPath("spec").Child("WorkerNodeConfiguration.kubernetesVersion")

	if oldVersion != nil && newVersion == nil {
		allErrs = append(
			allErrs,
			validateRemoveWorkerKubernetesVersion(newClusterVersion, oldClusterVersion, oldVersion)...,
		)
	}

	if oldVersion != nil && newVersion != nil {
		allErrs = append(
			allErrs,
			validateKubeVersionSkew(*newVersion, *oldVersion, path)...,
		)
	}

	if oldVersion == nil && newVersion != nil {
		allErrs = append(
			allErrs,
			validateKubeVersionSkew(*newVersion, oldClusterVersion, path)...,
		)
	}

	if newVersion != nil {
		allErrs = append(
			allErrs,
			validateCPWorkerKubeSkew(newClusterVersion, *newVersion)...,
		)
	}

	return allErrs
}

func validateRemoveWorkerKubernetesVersion(newCPVersion, oldCPVersion KubernetesVersion, oldWorkerVersion *KubernetesVersion) field.ErrorList {
	var allErrs field.ErrorList
	path := field.NewPath("spec").Child("WorkerNodeConfiguration.kubernetesVersion")

	parsedWorkerVersion, err := version.ParseGeneric(string(*oldWorkerVersion))
	if err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(path, oldWorkerVersion, fmt.Sprintf("could not parse version: %v", err)))
		return allErrs
	}

	parsedOldClusterVersion, err := version.ParseGeneric(string(oldCPVersion))
	if err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(path, oldCPVersion, fmt.Sprintf("could not parse version: %v", err)))
		return allErrs
	}

	parsedNewClusterVersion, err := version.ParseGeneric(string(newCPVersion))
	if err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(path, newCPVersion, fmt.Sprintf("could not parse version: %v", err)))
		return allErrs
	}

	if parsedOldClusterVersion.LessThan(parsedNewClusterVersion) {
		allErrs = append(
			allErrs,
			field.Invalid(path, oldWorkerVersion, fmt.Sprintf("can't simultaneously remove worker kubernetesVersion and upgrade cluster level kubernetesVersion: %v", newCPVersion)))
		return allErrs
	}

	if err := ValidateVersionSkew(parsedWorkerVersion, parsedNewClusterVersion); err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(path, oldWorkerVersion, err.Error()))
	}

	return allErrs
}

func validKubeMinorVersionDiff(old, new KubernetesVersion) (bool, error) {
	parsedOldVersion, err := version.ParseGeneric(string(old))
	if err != nil {
		return false, fmt.Errorf("could not parse version: %v, %v", old, err)
	}

	parsedNewVersion, err := version.ParseGeneric(string(new))
	if err != nil {
		return false, fmt.Errorf("could not parse version: %v, %v", new, err)
	}

	if parsedOldVersion.Major() != parsedNewVersion.Major() {
		return false, fmt.Errorf("major versions are not the same: %v and %v", parsedOldVersion, parsedNewVersion)
	}

	oldMinor := int(parsedOldVersion.Minor())
	newMinor := int(parsedNewVersion.Minor())
	minorDiff := newMinor - oldMinor

	if minorDiff < 0 || minorDiff > 2 {
		return false, nil
	}

	return true, nil
}

func validateCPWorkerKubeSkew(cpVersion, workerVersion KubernetesVersion) field.ErrorList {
	var allErrs field.ErrorList
	workerPath := field.NewPath("spec").Child("WorkerNodeConfiguration.kubernetesVersion")
	cpPath := field.NewPath("spec").Child("kubernetesVersion")

	validSkew, err := validKubeMinorVersionDiff(workerVersion, cpVersion)
	if err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(workerPath, workerVersion, fmt.Sprintf("could not determine minor version difference: %v", err.Error())))
		return allErrs
	}

	if !validSkew {
		allErrs = append(
			allErrs,
			field.Invalid(cpPath, cpVersion, fmt.Sprintf("cluster level minor version must be within 2 versions greater than worker node group version: %v", workerVersion)))
	}

	return allErrs
}

func validateEtcdEncryptionSupport(cluster *Cluster) field.ErrorList {
	var errs field.ErrorList

	if cluster.Spec.EtcdEncryption == nil {
		return errs
	}

	if !supportsEtcdEncryption(cluster) {
		errs = append(errs, field.Invalid(
			field.NewPath("spec.etcdEncryption"),
			cluster.Spec.EtcdEncryption,
			fmt.Sprintf("etcdEncryption is currently not supported for the provider: %s", cluster.Spec.DatacenterRef.Kind),
		))
	}

	if err := ValidateEtcdEncryptionConfig(cluster.Spec.EtcdEncryption); err != nil {
		errs = append(errs, field.Invalid(field.NewPath("spec.etcdEncryption"), cluster.Spec.EtcdEncryption, err.Error()))
	}

	return errs
}
