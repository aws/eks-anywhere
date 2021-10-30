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

// CloudstackMachineConfigSpec defines the desired state of CloudstackMachineConfig
type CloudstackMachineConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Template          string              `json:"template,omitempty"`
	ComputeOffering	  string			  `json:"compute_offering"`
	DiskOffering	  string			  `json:"disk_offering,omitempty"`
	OSFamily          OSFamily            `json:"osFamily"`
	KeyPair 		  string			  `json:"key_pair"`
}


func (c *CloudstackMachineConfig) PauseReconcile() {
	c.Annotations[pausedAnnotation] = "true"
}

func (c *CloudstackMachineConfig) IsReconcilePaused() bool {
	if s, ok := c.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *CloudstackMachineConfig) SetControlPlane() {
	c.Annotations[controlPlaneAnnotation] = "true"
}

func (c *CloudstackMachineConfig) IsControlPlane() bool {
	if s, ok := c.Annotations[controlPlaneAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *CloudstackMachineConfig) SetEtcd() {
	c.Annotations[etcdAnnotation] = "true"
}

func (c *CloudstackMachineConfig) IsEtcd() bool {
	if s, ok := c.Annotations[etcdAnnotation]; ok {
		return s == "true"
	}
	return false
}

// CloudstackMachineConfigStatus defines the observed state of CloudstackMachineConfig
type CloudstackMachineConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudstackMachineConfig is the Schema for the cloudstackmachineconfigs API
type CloudstackMachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudstackMachineConfigSpec   `json:"spec,omitempty"`
	Status CloudstackMachineConfigStatus `json:"status,omitempty"`
}

func (c *CloudstackMachineConfig) ConvertConfigToConfigGenerateStruct() *CloudstackMachineConfigGenerate {
	config := &CloudstackMachineConfigGenerate{
		TypeMeta: c.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        c.Name,
			Annotations: c.Annotations,
			Namespace:   c.Namespace,
		},
		Spec: c.Spec,
	}

	return config
}

func (c *CloudstackMachineConfig) Marshallable() Marshallable {
	return c.ConvertConfigToConfigGenerateStruct()
}


// +kubebuilder:object:generate=false

// Same as CloudstackMachineConfig except stripped down for generation of yaml file during generate clusterconfig
type CloudstackMachineConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec CloudstackMachineConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CloudstackMachineConfigList contains a list of CloudstackMachineConfig
type CloudstackMachineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudstackMachineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudstackMachineConfig{}, &CloudstackMachineConfigList{})
}
