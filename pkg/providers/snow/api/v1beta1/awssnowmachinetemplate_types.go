/*
Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License").
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package snow

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AWSSnowMachineTemplateSpec defines the desired state of AWSSnowMachineTemplate
type AWSSnowMachineTemplateSpec struct {
	Template AWSSnowMachineTemplateResource `json:"template"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=awssnowmachinetemplates,scope=Namespaced,categories=cluster-api,shortName=awssmt
// +kubebuilder:storageversion

// AWSSnowMachineTemplate is the Schema for the awssnowmachinetemplates API
type AWSSnowMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AWSSnowMachineTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// AWSSnowMachineTemplateList contains a list of AWSSnowMachineTemplate.
type AWSSnowMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSSnowMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSSnowMachineTemplate{}, &AWSSnowMachineTemplateList{})
}
