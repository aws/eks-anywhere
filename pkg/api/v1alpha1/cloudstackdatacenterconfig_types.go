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

// CloudstackDatacenterConfigSpec defines the desired state of CloudstackDatacenterConfig
type CloudstackDatacenterConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CloudstackDatacenterConfig. Edit cloudstackdatacenterconfig_types.go to remove/update
	Domain string `json:"domain"`
	Zone string `json:"zone"`
	Project string `json:"project,omitempty"`
	Account string `json:"account,omitempty"`
	Network    string `json:"network"`
	ControlPlaneEndpoint string `json:"control_plane_endpoint"`
	//Server     string `json:"server"`
	//Thumbprint string `json:"thumbprint"`
	Insecure   bool   `json:"insecure"`
}

// CloudstackDatacenterConfigStatus defines the observed state of CloudstackDatacenterConfig
type CloudstackDatacenterConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudstackDatacenterConfig is the Schema for the cloudstackdatacenterconfigs API
type CloudstackDatacenterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudstackDatacenterConfigSpec   `json:"spec,omitempty"`
	Status CloudstackDatacenterConfigStatus `json:"status,omitempty"`
}

func (v *CloudstackDatacenterConfig) Kind() string {
	return v.TypeMeta.Kind
}

func (v *CloudstackDatacenterConfig) ExpectedKind() string {
	return CloudstackDatacenterKind
}

func (v *CloudstackDatacenterConfig) PauseReconcile() {
	if v.Annotations == nil {
		v.Annotations = map[string]string{}
	}
	v.Annotations[pausedAnnotation] = "true"
}

func (v *CloudstackDatacenterConfig) IsReconcilePaused() bool {
	if s, ok := v.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (v *CloudstackDatacenterConfig) ClearPauseAnnotation() {
	if v.Annotations != nil {
		delete(v.Annotations, pausedAnnotation)
	}
}

func (v *CloudstackDatacenterConfig) ConvertConfigToConfigGenerateStruct() *CloudstackDatacenterConfigGenerate {
	config := &CloudstackDatacenterConfigGenerate{
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

func (v *CloudstackDatacenterConfig) Marshallable() Marshallable {
	return v.ConvertConfigToConfigGenerateStruct()
}

// +kubebuilder:object:generate=false

// Same as CloudstackDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig
type CloudstackDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec CloudstackDatacenterConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CloudstackDatacenterConfigList contains a list of CloudstackDatacenterConfig
type CloudstackDatacenterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudstackDatacenterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudstackDatacenterConfig{}, &CloudstackDatacenterConfigList{})
}
