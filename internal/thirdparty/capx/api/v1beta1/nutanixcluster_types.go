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
	"cmp"
	"fmt"

	credentialTypes "github.com/nutanix-cloud-native/prism-go-client/environment/credentials"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // suppress complaining on Deprecated package
)

const (
	// NutanixClusterKind represents the Kind of NutanixCluster
	NutanixClusterKind = "NutanixCluster"

	// NutanixClusterFinalizer allows NutanixClusterReconciler to clean up AHV
	// resources associated with NutanixCluster before removing it from the
	// API Server.
	NutanixClusterFinalizer           = "infrastructure.cluster.x-k8s.io/nutanixcluster"
	DeprecatedNutanixClusterFinalizer = "nutanixcluster.infrastructure.cluster.x-k8s.io"

	NutanixClusterCredentialFinalizer           = "infrastructure.cluster.x-k8s.io/nutanixclustercredential"
	DeprecatedNutanixClusterCredentialFinalizer = "nutanixcluster/infrastructure.cluster.x-k8s.io"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NutanixClusterSpec defines the desired state of NutanixCluster
// +kubebuilder:validation:XValidation:rule="!(has(self.failureDomains) && has(self.controlPlaneFailureDomains))",message="Cannot set both 'failureDomains' and 'controlPlaneFailureDomains' fields simultaneously. Use 'controlPlaneFailureDomains' as 'failureDomains' is deprecated."
type NutanixClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// host can be either DNS name or ip address
	// +optional
	ControlPlaneEndpoint capiv1beta1.APIEndpoint `json:"controlPlaneEndpoint"`

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
	//
	// Deprecated: This field is replaced by the field controlPlaneFailureDomains and will be removed in the next apiVersion.
	//
	FailureDomains []NutanixFailureDomainConfig `json:"failureDomains,omitempty"`

	// controlPlaneFailureDomains configures references to the NutanixFailureDomain objects
	// that the cluster uses to deploy its control-plane machines.
	// +listType=map
	// +listMapKey=name
	// +optional
	ControlPlaneFailureDomains []corev1.LocalObjectReference `json:"controlPlaneFailureDomains,omitempty"`
}

// NutanixClusterStatus defines the observed state of NutanixCluster
type NutanixClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Ready bool `json:"ready,omitempty"`

	// failureDomains are a list of failure domains configured in the
	// cluster's spec and validated by the cluster controller.
	FailureDomains capiv1beta1.FailureDomains `json:"failureDomains,omitempty"`

	// Conditions defines current service state of the NutanixCluster.
	// +optional
	Conditions capiv1beta1.Conditions `json:"conditions,omitempty"`

	// Will be set in case of failure of Cluster instance
	// +optional
	FailureReason *string `json:"failureReason,omitempty"`

	// Will be set in case of failure of Cluster instance
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// v1beta2 groups all the fields that will be added or modified in NutanixCluster's status with the v1beta2 version.
	// +optional
	V1Beta2 *NutanixClusterV1Beta2Status `json:"v1beta2,omitempty"`
}

// NutanixClusterV1Beta2Status groups all the fields that will be added or modified in NutanixClusterStatus with the v1beta2 version.
// See https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20240916-improve-status-in-CAPI-resources.md for more context.
type NutanixClusterV1Beta2Status struct {
	// conditions represents the observations of a NutanixCluster's current state.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=nutanixclusters,shortName=ncl,scope=Namespaced,categories=cluster-api
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="ControlplaneEndpoint",type="string",JSONPath=".spec.controlPlaneEndpoint.host",description="ControlplaneEndpoint"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="in ready status"
// +kubebuilder:printcolumn:name="FailureDomains",type="string",JSONPath=".status.failureDomains",description="NutanixCluster FailureDomains"

// NutanixCluster is the Schema for the nutanixclusters API
type NutanixCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NutanixClusterSpec   `json:"spec,omitempty"`
	Status NutanixClusterStatus `json:"status,omitempty"`
}

// NutanixFailureDomainConfig configures failure domain information for Nutanix.
//
// Deprecated: This type is replaced by the NutanixFailureDomain CRD type and will be removed in the next apiVersion.
type NutanixFailureDomainConfig struct {
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
	Subnets []NutanixResourceIdentifier `json:"subnets"`

	// indicates if a failure domain is suited for control plane nodes
	// +kubebuilder:validation:Required
	ControlPlane bool `json:"controlPlane,omitempty"`
}

// GetConditions returns the set of conditions for this object.
func (ncl *NutanixCluster) GetConditions() capiv1beta1.Conditions {
	return ncl.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (ncl *NutanixCluster) SetConditions(conditions capiv1beta1.Conditions) {
	ncl.Status.Conditions = conditions
}

// GetV1Beta2Conditions returns the set of v1beta2 conditions for this object.
func (ncl *NutanixCluster) GetV1Beta2Conditions() []metav1.Condition {
	if ncl.Status.V1Beta2 == nil {
		return nil
	}
	return ncl.Status.V1Beta2.Conditions
}

// SetV1Beta2Conditions sets the v1beta2 conditions on this object.
func (ncl *NutanixCluster) SetV1Beta2Conditions(conditions []metav1.Condition) {
	if ncl.Status.V1Beta2 == nil {
		ncl.Status.V1Beta2 = &NutanixClusterV1Beta2Status{}
	}
	ncl.Status.V1Beta2.Conditions = conditions
}

func (ncl *NutanixCluster) GetPrismCentralCredentialRef() (*credentialTypes.NutanixCredentialReference, error) {
	prismCentralInfo := ncl.Spec.PrismCentral
	if prismCentralInfo == nil {
		return nil, nil
	}
	if prismCentralInfo.CredentialRef == nil {
		return nil, fmt.Errorf("credentialRef must be set on prismCentral attribute for cluster %s in namespace %s", ncl.Name, ncl.Namespace)
	}
	if prismCentralInfo.CredentialRef.Kind != credentialTypes.SecretKind {
		return nil, nil
	}

	return prismCentralInfo.CredentialRef, nil
}

// GetPrismCentralTrustBundle returns the trust bundle reference for the Nutanix Prism Central.
func (ncl *NutanixCluster) GetPrismCentralTrustBundle() *credentialTypes.NutanixTrustBundleReference {
	prismCentralInfo := ncl.Spec.PrismCentral
	if prismCentralInfo == nil ||
		prismCentralInfo.AdditionalTrustBundle == nil ||
		prismCentralInfo.AdditionalTrustBundle.Kind == credentialTypes.NutanixTrustBundleKindString {
		return nil
	}

	return prismCentralInfo.AdditionalTrustBundle
}

// GetNamespacedName returns the namespaced name of the NutanixCluster.
func (ncl *NutanixCluster) GetNamespacedName() string {
	namespace := cmp.Or(ncl.Namespace, corev1.NamespaceDefault)
	return fmt.Sprintf("%s/%s", namespace, ncl.Name)
}

// +kubebuilder:object:root=true

// NutanixClusterList contains a list of NutanixCluster
type NutanixClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NutanixCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NutanixCluster{}, &NutanixClusterList{})
}
