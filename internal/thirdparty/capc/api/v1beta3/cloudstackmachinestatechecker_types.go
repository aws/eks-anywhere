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

// CloudStackMachineStateCheckerSpec
type CloudStackMachineStateCheckerSpec struct {
	// CloudStack machine instance ID
	InstanceID string `json:"instanceID,omitempty"`
}

// CloudStackMachineStateCheckerStatus defines the observed state of CloudStackMachineStateChecker
type CloudStackMachineStateCheckerStatus struct {
	// Reflects the readiness of the Machine State Checker.
	Ready bool `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// CloudStackMachineStateChecker is the Schema for the cloudstackmachinestatecheckers API
type CloudStackMachineStateChecker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudStackMachineStateCheckerSpec   `json:"spec,omitempty"`
	Status CloudStackMachineStateCheckerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CloudStackMachineStateCheckerList contains a list of CloudStackMachineStateChecker
type CloudStackMachineStateCheckerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudStackMachineStateChecker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudStackMachineStateChecker{}, &CloudStackMachineStateCheckerList{})
}
