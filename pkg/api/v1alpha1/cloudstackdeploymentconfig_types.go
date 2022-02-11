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

// CloudStackDeploymentConfigSpec defines the desired state of CloudStackDeploymentConfig
type CloudStackDeploymentConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CloudStackDeploymentConfig. Edit cloudstackdeploymentconfig_types.go to remove/update
	Domain                string `json:"domain"`
	Zone                  string `json:"zone"`
	Account               string `json:"account,omitempty"`
	Network               string `json:"network"`
	ManagementApiEndpoint string `json:"managementApiEndpoint"`
	Thumbprint            string `json:"thumbprint,omitempty"`
	Insecure              bool   `json:"insecure"`
}

// CloudStackDeploymentConfigStatus defines the observed state of CloudStackDeploymentConfig
type CloudStackDeploymentConfigStatus struct { // INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudStackDeploymentConfig is the Schema for the cloudstackdeploymentconfigs API
type CloudStackDeploymentConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudStackDeploymentConfigSpec   `json:"spec,omitempty"`
	Status CloudStackDeploymentConfigStatus `json:"status,omitempty"`
}

func (c *CloudStackDeploymentConfig) Kind() string {
	return c.TypeMeta.Kind
}

func (c *CloudStackDeploymentConfig) ExpectedKind() string {
	return CloudStackDeploymentKind
}

func (c *CloudStackDeploymentConfig) PauseReconcile() {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[pausedAnnotation] = "true"
}

func (c *CloudStackDeploymentConfig) IsReconcilePaused() bool {
	if s, ok := c.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *CloudStackDeploymentConfig) ClearPauseAnnotation() {
	if c.Annotations != nil {
		delete(c.Annotations, pausedAnnotation)
	}
}

func (c *CloudStackDeploymentConfig) ConvertConfigToConfigGenerateStruct() *CloudStackDeploymentConfigGenerate {
	namespace := defaultEksaNamespace
	if c.Namespace != "" {
		namespace = c.Namespace
	}
	config := &CloudStackDeploymentConfigGenerate{
		TypeMeta: c.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        c.Name,
			Annotations: c.Annotations,
			Namespace:   namespace,
		},
		Spec: c.Spec,
	}

	return config
}

func (c *CloudStackDeploymentConfig) Marshallable() Marshallable {
	return c.ConvertConfigToConfigGenerateStruct()
}

// +kubebuilder:object:generate=false

// Same as CloudStackDeploymentConfig except stripped down for generation of yaml file during generate clusterconfig
type CloudStackDeploymentConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec CloudStackDeploymentConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CloudStackDeploymentConfigList contains a list of CloudStackDeploymentConfig
type CloudStackDeploymentConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudStackDeploymentConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudStackDeploymentConfig{}, &CloudStackDeploymentConfigList{})
}
