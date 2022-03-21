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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudStackDatacenterConfigSpec defines the desired state of CloudStackDatacenterConfig
type CloudStackDatacenterConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Domain contains a grouping of accounts. Domains usually contain multiple accounts that have some logical relationship to each other and a set of delegated administrators with some authority over the domain and its subdomains
	Domain string `json:"domain"`
	// Zones is a list of one or more zones that are managed by a single CloudStack management endpoint.
	Zones []CloudStackZone `json:"zones"`
	// Account typically represents a customer of the service provider or a department in a large organization. Multiple users can exist in an account, and all CloudStack resources belong to an account. Accounts have users and users have credentials to operate on resources within that account. If an account name is provided, a domain name must also be provided.
	Account string `json:"account,omitempty"`
	// CloudStack Management API endpoint's IP. It is added to VM's noproxy list
	ManagementApiEndpoint string `json:"managementApiEndpoint"`
}

type CloudStackResourceIdentifier struct {
	// Id of a resource in the CloudStack environment. Mutually exclusive with Name
	// +optional
	Id string `json:"id,omitempty"`
	// Name of a resource in the CloudStack environment. Mutually exclusive with Id
	// +optional
	Name string `json:"name,omitempty"`
}

// CloudStackZone is an organizational construct typically used to represent a single datacenter, and all its physical and virtual resources exist inside that zone. It can either be specified as a UUID or name
type CloudStackZone struct {
	// Zone is the name or UUID of the CloudStack zone in which clusters should be created. Zones should be managed by a single CloudStack Management endpoint.
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// Network is the name or UUID of the CloudStack network in which clusters should be created. It can either be an isolated or shared network. If it doesn’t already exist in CloudStack, it’ll automatically be created by CAPC as an isolated network. It can either be specified as a UUID or name
	// In multiple-zones situation, only 'Shared' network is supported.
	Network CloudStackResourceIdentifier `json:"network"`
}

// CloudStackDatacenterConfigStatus defines the observed state of CloudStackDatacenterConfig
type CloudStackDatacenterConfigStatus struct { // INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudStackDatacenterConfig is the Schema for the cloudstackdatacenterconfigs API
type CloudStackDatacenterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudStackDatacenterConfigSpec   `json:"spec,omitempty"`
	Status CloudStackDatacenterConfigStatus `json:"status,omitempty"`
}

func (v *CloudStackDatacenterConfig) Kind() string {
	return v.TypeMeta.Kind
}

func (v *CloudStackDatacenterConfig) ExpectedKind() string {
	return CloudStackDatacenterKind
}

func (v *CloudStackDatacenterConfig) PauseReconcile() {
	if v.Annotations == nil {
		v.Annotations = map[string]string{}
	}
	v.Annotations[pausedAnnotation] = "true"
}

func (v *CloudStackDatacenterConfig) IsReconcilePaused() bool {
	if s, ok := v.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (v *CloudStackDatacenterConfig) ClearPauseAnnotation() {
	if v.Annotations != nil {
		delete(v.Annotations, pausedAnnotation)
	}
}

func (v *CloudStackDatacenterConfig) ConvertConfigToConfigGenerateStruct() *CloudStackDatacenterConfigGenerate {
	namespace := defaultEksaNamespace
	if v.Namespace != "" {
		namespace = v.Namespace
	}
	config := &CloudStackDatacenterConfigGenerate{
		TypeMeta: v.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        v.Name,
			Annotations: v.Annotations,
			Namespace:   namespace,
		},
		Spec: v.Spec,
	}

	return config
}

func (v *CloudStackDatacenterConfig) Marshallable() Marshallable {
	return v.ConvertConfigToConfigGenerateStruct()
}

func (z *CloudStackZone) Equals(o *CloudStackZone) bool {
	if z == o {
		return true
	}
	if z == nil || o == nil {
		return false
	}
	if z.Id == o.Id &&
		z.Name == o.Name &&
		z.Network.Id == o.Network.Id &&
		z.Network.Name == o.Network.Name {
		return true
	}
	return false
}

// +kubebuilder:object:generate=false

// Same as CloudStackDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig
type CloudStackDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec CloudStackDatacenterConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CloudStackDatacenterConfigList contains a list of CloudStackDatacenterConfig
type CloudStackDatacenterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudStackDatacenterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudStackDatacenterConfig{}, &CloudStackDatacenterConfigList{})
}
