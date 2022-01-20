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

// CloudStackMachineConfigSpec defines the desired state of CloudStackMachineConfig
type CloudStackMachineConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Template         string              `json:"template"`
	ComputeOffering  string              `json:"computeOffering"`
	DiskOffering     string              `json:"diskOffering,omitempty"`
	OSFamily         OSFamily            `json:"osFamily,omitempty"`
	Details          map[string]string   `json:"details,omitempty"`
	Users            []UserConfiguration `json:"users,omitempty"`
	AffinityGroupIds []string            `json:"affinityGroupIds,omitempty"`
}

func (c *CloudStackMachineConfig) PauseReconcile() {
	c.Annotations[pausedAnnotation] = "true"
}

func (c *CloudStackMachineConfig) IsReconcilePaused() bool {
	if s, ok := c.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *CloudStackMachineConfig) SetControlPlane() {
	c.Annotations[controlPlaneAnnotation] = "true"
}

func (c *CloudStackMachineConfig) IsControlPlane() bool {
	if s, ok := c.Annotations[controlPlaneAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *CloudStackMachineConfig) SetEtcd() {
	c.Annotations[etcdAnnotation] = "true"
}

func (c *CloudStackMachineConfig) IsEtcd() bool {
	if s, ok := c.Annotations[etcdAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *CloudStackMachineConfig) SetManagement(clusterName string) {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[managementAnnotation] = clusterName
}

func (c *CloudStackMachineConfig) IsManagement() bool {
	if s, ok := c.Annotations[managementAnnotation]; ok {
		return s != ""
	}
	return false
}

func (c *CloudStackMachineConfig) OSFamily() OSFamily {
	return c.Spec.OSFamily
}

func (c *CloudStackMachineConfig) GetNamespace() string {
	return c.Namespace
}

func (c *CloudStackMachineConfig) GetName() string {
	return c.Name
}

// CloudStackMachineConfigStatus defines the observed state of CloudStackMachineConfig
type CloudStackMachineConfigStatus struct { // INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudStackMachineConfig is the Schema for the cloudstackmachineconfigs API
type CloudStackMachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudStackMachineConfigSpec   `json:"spec,omitempty"`
	Status CloudStackMachineConfigStatus `json:"status,omitempty"`
}

func (c *CloudStackMachineConfig) ConvertConfigToConfigGenerateStruct() *CloudStackMachineConfigGenerate {
	config := &CloudStackMachineConfigGenerate{
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

func (c *CloudStackMachineConfig) Marshallable() Marshallable {
	return c.ConvertConfigToConfigGenerateStruct()
}

// +kubebuilder:object:generate=false

// Same as CloudStackMachineConfig except stripped down for generation of yaml file during generate clusterconfig
type CloudStackMachineConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec CloudStackMachineConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CloudStackMachineConfigList contains a list of CloudStackMachineConfig
type CloudStackMachineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudStackMachineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudStackMachineConfig{}, &CloudStackMachineConfigList{})
}
