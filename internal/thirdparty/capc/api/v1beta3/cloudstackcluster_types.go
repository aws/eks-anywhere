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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ClusterFinalizer = "cloudstackcluster.infrastructure.cluster.x-k8s.io"
)

var K8sClient client.Client

// CloudStackClusterSpec defines the desired state of CloudStackCluster.
type CloudStackClusterSpec struct {
	FailureDomains []CloudStackFailureDomainSpec `json:"failureDomains"`

	// The kubernetes control plane endpoint.
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`
}

// The status of the CloudStackCluster object.
type CloudStackClusterStatus struct {
	// CAPI recognizes failure domains as a method to spread machines.
	// CAPC sets failure domains to indicate functioning CloudStackFailureDomains.
	// +optional
	FailureDomains clusterv1.FailureDomains `json:"failureDomains,omitempty"`

	// Reflects the readiness of the CS cluster.
	Ready bool `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// CloudStackCluster is the Schema for the cloudstackclusters API
type CloudStackCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CloudStackClusterSpec `json:"spec,omitempty"`

	// The actual cluster state reported by CloudStack.
	Status CloudStackClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CloudStackClusterList contains a list of CloudStackCluster
type CloudStackClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudStackCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudStackCluster{}, &CloudStackClusterList{})
}
