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
)

// NutanixClusterTemplateSpec defines the desired state of NutanixClusterTemplate
type NutanixClusterTemplateSpec struct {
	Template NutanixClusterTemplateResource `json:"template"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=cluster-api

// NutanixClusterTemplate is the Schema for the nutanixclustertemplates API
type NutanixClusterTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NutanixClusterTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// NutanixClusterTemplateList contains a list of NutanixClusterTemplate
type NutanixClusterTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NutanixClusterTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NutanixClusterTemplate{}, &NutanixClusterTemplateList{})
}

// NutanixClusterTemplateResource describes the data needed to create a NutanixCluster from a template.
type NutanixClusterTemplateResource struct {
	Spec NutanixClusterSpec `json:"spec"`
}
