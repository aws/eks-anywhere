/*
Copyright 2022 Nutanix

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
	credentialTypes "github.com/nutanix-cloud-native/prism-go-client/environment/credentials"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/cluster-api/errors"
)

const (
	// NutanixClusterKind represents the Kind of NutanixCluster
	NutanixClusterKind = "NutanixCluster"

	// NutanixClusterFinalizer allows NutanixClusterReconciler to clean up AHV
	// resources associated with NutanixCluster before removing it from the
	// API Server.
	NutanixClusterFinalizer           = "nutanixcluster.infrastructure.cluster.x-k8s.io"
	NutanixClusterCredentialFinalizer = "nutanixcluster/infrastructure.cluster.x-k8s.io"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NutanixClusterSpec defines the desired state of NutanixCluster
type NutanixClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// host can be either DNS name or ip address
	// +optional
	ControlPlaneEndpoint capiv1.APIEndpoint `json:"controlPlaneEndpoint"`

	// prismCentral holds the endpoint address and port to access the Nutanix Prism Central.
	// When a cluster-wide proxy is installed, by default, this endpoint will be accessed via the proxy.
	// Should you wish for communication with this endpoint not to be proxied, please add the endpoint to the
	// proxy spec.noProxy list.
	// +optional
	PrismCentral *credentialTypes.NutanixPrismEndpoint `json:"prismCentral"`

	// failureDomains configures failure domains information for the Nutanix platform.
	// When set, the failure domains defined here may be used to spread Machines across
	// prism element clusters to improve fault tolerance of the cluster.
	// +listType=map
	// +listMapKey=name
	// +optional
	FailureDomains []NutanixFailureDomain `json:"failureDomains"`
}

// NutanixClusterStatus defines the observed state of NutanixCluster
type NutanixClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Ready bool `json:"ready,omitempty"`

	FailureDomains capiv1.FailureDomains `json:"failureDomains,omitempty"`

	// Conditions defines current service state of the NutanixCluster.
	// +optional
	Conditions capiv1.Conditions `json:"conditions,omitempty"`

	// Will be set in case of failure of Cluster instance
	// +optional
	FailureReason *errors.ClusterStatusError `json:"failureReason,omitempty"`

	// Will be set in case of failure of Cluster instance
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:path=nutanixclusters,shortName=ncl,scope=Namespaced,categories=cluster-api
//+kubebuilder:subresource:status
//+kubebuilder:storageversion
//+kubebuilder:printcolumn:name="ControlplaneEndpoint",type="string",JSONPath=".spec.controlPlaneEndpoint.host",description="ControlplaneEndpoint"
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="in ready status"

// NutanixCluster is the Schema for the nutanixclusters API
type NutanixCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NutanixClusterSpec   `json:"spec,omitempty"`
	Status NutanixClusterStatus `json:"status,omitempty"`
}

// NutanixFailureDomain configures failure domain information for Nutanix.
type NutanixFailureDomain struct {
	// name defines the unique name of a failure domain.
	// Name is required and must be at most 64 characters in length.
	// It must consist of only lower case alphanumeric characters and hyphens (-).
	// It must start and end with an alphanumeric character.
	// This value is arbitrary and is used to identify the failure domain within the platform.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Pattern=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name"`

	// cluster is to identify the cluster (the Prism Element under management of the Prism Central),
	// in which the Machine's VM will be created. The cluster identifier (uuid or name) can be obtained
	// from the Prism Central console or using the prism_central API.
	// +kubebuilder:validation:Required
	Cluster NutanixResourceIdentifier `json:"cluster"`

	// subnets holds a list of identifiers (one or more) of the cluster's network subnets
	// for the Machine's VM to connect to. The subnet identifiers (uuid or name) can be
	// obtained from the Prism Central console or using the prism_central API.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +listType=map
	// +listMapKey=type
	Subnets []NutanixResourceIdentifier `json:"subnets"`

	// indicates if a failure domain is suited for control plane nodes
	// +kubebuilder:validation:Required
	ControlPlane bool `json:"controlPlane,omitempty"`
}

// GetConditions returns the set of conditions for this object.
func (ncl *NutanixCluster) GetConditions() capiv1.Conditions {
	return ncl.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (ncl *NutanixCluster) SetConditions(conditions capiv1.Conditions) {
	ncl.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// NutanixClusterList contains a list of NutanixCluster
type NutanixClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NutanixCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NutanixCluster{}, &NutanixClusterList{})
}
