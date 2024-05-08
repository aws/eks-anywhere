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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// CloudStackMachineTemplateResource defines the data needed to create a CloudstackMachine from a template
type CloudStackMachineTemplateResource struct {
	// Spec is the specification of a desired behavior of the machine
	Spec CloudStackMachineSpec `json:"spec"`
}

// CloudStackMachineTemplateSpec defines the desired state of CloudstackMachineTemplate
type CloudStackMachineTemplateSpec struct {
	Template CloudStackMachineTemplateResource `json:"template"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// CloudStackMachineTemplate is the Schema for the cloudstackmachinetemplates API
type CloudStackMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CloudStackMachineTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CloudStackMachineTemplateList contains a list of CloudStackMachineTemplate
type CloudStackMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudStackMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudStackMachineTemplate{}, &CloudStackMachineTemplateList{})
}
