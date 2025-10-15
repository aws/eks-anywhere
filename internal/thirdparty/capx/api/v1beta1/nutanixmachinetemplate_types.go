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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// NutanixMachineTemplateKind represents the Kind of NutanixMachineTemplate
	NutanixMachineTemplateKind = "NutanixMachineTemplate"
)

// NutanixMachineTemplateSpec defines the desired state of NutanixMachineTemplate
type NutanixMachineTemplateSpec struct {
	Template NutanixMachineTemplateResource `json:"template"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:path=nutanixmachinetemplates,shortName=nmtmpl,scope=Namespaced,categories=cluster-api
//+kubebuilder:storageversion

// NutanixMachineTemplate is the Schema for the nutanixmachinetemplates API
type NutanixMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NutanixMachineTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// NutanixMachineTemplateList contains a list of NutanixMachineTemplate
type NutanixMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NutanixMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NutanixMachineTemplate{}, &NutanixMachineTemplateList{})
}

// NutanixMachineTemplateResource describes the data needed to create a NutanixMachine from a template
type NutanixMachineTemplateResource struct {
	// Standard object metadata.
	// Ref: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta capiv1.ObjectMeta `json:"metadata,omitempty"`
	// Spec is the specification of the desired behavior of the machine.
	Spec NutanixMachineSpec `json:"spec"`
}
