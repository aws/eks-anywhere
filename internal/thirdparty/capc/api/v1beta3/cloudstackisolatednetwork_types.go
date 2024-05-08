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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// The presence of a finalizer prevents CAPI from deleting the corresponding CAPI data.
const IsolatedNetworkFinalizer = "cloudstackisolatednetwork.infrastructure.cluster.x-k8s.io"

// CloudStackIsolatedNetworkSpec defines the desired state of CloudStackIsolatedNetwork
type CloudStackIsolatedNetworkSpec struct {
	// Name.
	//+optional
	Name string `json:"name,omitempty"`

	// ID.
	//+optional
	ID string `json:"id,omitempty"`

	// The kubernetes control plane endpoint.
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`

	// FailureDomainName -- the FailureDomain the network is placed in.
	FailureDomainName string `json:"failureDomainName"`
}

// CloudStackIsolatedNetworkStatus defines the observed state of CloudStackIsolatedNetwork
type CloudStackIsolatedNetworkStatus struct {
	// The CS public IP ID to use for the k8s endpoint.
	PublicIPID string `json:"publicIPID,omitempty"`

	// The ID of the lb rule used to assign VMs to the lb.
	LBRuleID string `json:"loadBalancerRuleID,omitempty"`

	// Ready indicates the readiness of this provider resource.
	Ready bool `json:"ready"`
}

func (n *CloudStackIsolatedNetwork) Network() *Network {
	return &Network{
		Name: n.Spec.Name,
		Type: "IsolatedNetwork",
		ID:   n.Spec.ID}
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// CloudStackIsolatedNetwork is the Schema for the cloudstackisolatednetworks API
type CloudStackIsolatedNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudStackIsolatedNetworkSpec   `json:"spec,omitempty"`
	Status CloudStackIsolatedNetworkStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CloudStackIsolatedNetworkList contains a list of CloudStackIsolatedNetwork
type CloudStackIsolatedNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudStackIsolatedNetwork `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudStackIsolatedNetwork{}, &CloudStackIsolatedNetworkList{})
}
