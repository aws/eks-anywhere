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

func (v *CloudStackDeploymentConfig) Kind() string {
	return v.TypeMeta.Kind
}

func (v *CloudStackDeploymentConfig) ExpectedKind() string {
	return CloudStackDeploymentKind
}

func (v *CloudStackDeploymentConfig) PauseReconcile() {
	if v.Annotations == nil {
		v.Annotations = map[string]string{}
	}
	v.Annotations[pausedAnnotation] = "true"
}

func (v *CloudStackDeploymentConfig) IsReconcilePaused() bool {
	if s, ok := v.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (v *CloudStackDeploymentConfig) ClearPauseAnnotation() {
	if v.Annotations != nil {
		delete(v.Annotations, pausedAnnotation)
	}
}

func (v *CloudStackDeploymentConfig) ConvertConfigToConfigGenerateStruct() *CloudStackDeploymentConfigGenerate {
	config := &CloudStackDeploymentConfigGenerate{
		TypeMeta: v.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        v.Name,
			Annotations: v.Annotations,
			Namespace:   v.Namespace,
		},
		Spec: v.Spec,
	}

	return config
}

func (v *CloudStackDeploymentConfig) Marshallable() Marshallable {
	return v.ConvertConfigToConfigGenerateStruct()
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
