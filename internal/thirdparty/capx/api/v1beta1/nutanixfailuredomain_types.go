/*
Copyright 2025 Nutanix

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // suppress complaining on Deprecated package
)

const (
	// NutanixFailureDomainKind represents the Kind of NutanixFailureDomain
	NutanixFailureDomainKind = "NutanixFailureDomain"

	// NutanixFailureDomainFinalizer is the finalizer used by the NutanixFailureDomain controller to block
	// deletion of the NutanixFailureDomain object if there are references to this object by other resources.
	NutanixFailureDomainFinalizer = "infrastructure.cluster.x-k8s.io/nutanixfailuredomain"
)

// NutanixFailureDomainSpec defines the desired state of NutanixFailureDomain.
// +kubebuilder:validation:XValidation:rule="size(self.subnets) > 1 ? self.subnets.all(x, self.subnets.exists_one(y, x == y)) : true",message="each subnet must be unique"
type NutanixFailureDomainSpec struct {
	// prismElementCluster is to identify the Prism Element cluster in the Prism Central for the failure domain.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="prismElementCluster is immutable once set"
	// +kubebuilder:validation:Required
	PrismElementCluster NutanixResourceIdentifier `json:"prismElementCluster"`

	// subnets holds a list of identifiers (one or more) of the PE cluster's network subnets
	// for the Machine's VM to connect to.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="subnets is immutable once set"
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=32
	Subnets []NutanixResourceIdentifier `json:"subnets"`
}

// NutanixFailureDomainStatus defines the observed state of NutanixFailureDomain resource.
type NutanixFailureDomainStatus struct {
	// conditions represent the latest states of the failure domain.
	// +optional
	Conditions []capiv1beta1.Condition `json:"conditions,omitempty"`

	// v1beta2 groups all the fields that will be added or modified in NutanixCluster's status with the v1beta2 version.
	// +optional
	V1Beta2 *NutanixFailureDomainV1Beta2Status `json:"v1beta2,omitempty"`
}

// NutanixFailureDomainV1Beta2Status groups all the fields that will be added or modified in NutanixFailureDomainStatus with the v1beta2 version.
// See https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20240916-improve-status-in-CAPI-resources.md for more context.
type NutanixFailureDomainV1Beta2Status struct {
	// conditions represents the observations of a NutanixFailureDomain's current state.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=nutanixfailuredomains,shortName=nfd,scope=Namespaced,categories=cluster-api
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:metadata:labels=clusterctl.cluster.x-k8s.io/move=

// NutanixFailureDomain is the Schema for the NutanixFailureDomain API.
type NutanixFailureDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NutanixFailureDomainSpec   `json:"spec,omitempty"`
	Status NutanixFailureDomainStatus `json:"status,omitempty"`
}

// GetConditions returns the set of conditions for this object.
func (nfd *NutanixFailureDomain) GetConditions() capiv1beta1.Conditions {
	return nfd.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (nfd *NutanixFailureDomain) SetConditions(conditions capiv1beta1.Conditions) {
	nfd.Status.Conditions = conditions
}

// GetV1Beta2Conditions returns the set of conditions for this object.
func (ncl *NutanixFailureDomain) GetV1Beta2Conditions() []metav1.Condition {
	if ncl.Status.V1Beta2 == nil {
		return nil
	}
	return ncl.Status.V1Beta2.Conditions
}

// SetV1Beta2Conditions sets the v1beta2 conditions on this object.
func (ncl *NutanixFailureDomain) SetV1Beta2Conditions(conditions []metav1.Condition) {
	if ncl.Status.V1Beta2 == nil {
		ncl.Status.V1Beta2 = &NutanixFailureDomainV1Beta2Status{}
	}
	ncl.Status.V1Beta2.Conditions = conditions
}

// +kubebuilder:object:root=true

// NutanixFailureDomainList contains a list of NutanixFailureDomain
type NutanixFailureDomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NutanixFailureDomain `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NutanixFailureDomain{}, &NutanixFailureDomainList{})
}
