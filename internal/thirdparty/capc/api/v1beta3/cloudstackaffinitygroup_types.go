/*
Copyright 2022 The Kubernetes Authors.

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

package v1beta3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const AffinityGroupFinalizer = "affinitygroup.infrastructure.cluster.x-k8s.io"

// CloudStackAffinityGroupSpec defines the desired state of CloudStackAffinityGroup
type CloudStackAffinityGroupSpec struct {
	// Mutually exclusive parameter with AffinityGroupIDs.
	// Can be "host affinity" or "host anti-affinity". Will create an affinity group per machine set.
	Type string `json:"type,omitempty"`

	// Name.
	Name string `json:"name,omitempty"`

	// ID.
	//+optional
	ID string `json:"id,omitempty"`

	// FailureDomainName -- the name of the FailureDomain the machine is placed in.
	// +optional
	FailureDomainName string `json:"failureDomainName,omitempty"`
}

// CloudStackAffinityGroupStatus defines the observed state of CloudStackAffinityGroup
type CloudStackAffinityGroupStatus struct {
	// Reflects the readiness of the CS Affinity Group.
	Ready bool `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// CloudStackAffinityGroup is the Schema for the cloudstackaffinitygroups API
type CloudStackAffinityGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudStackAffinityGroupSpec   `json:"spec,omitempty"`
	Status CloudStackAffinityGroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CloudStackAffinityGroupList contains a list of CloudStackAffinityGroup
type CloudStackAffinityGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudStackAffinityGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudStackAffinityGroup{}, &CloudStackAffinityGroupList{})
}
