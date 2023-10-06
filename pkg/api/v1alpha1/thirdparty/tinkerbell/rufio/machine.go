// +kubebuilder:object:generate=true
package rufio

/*
Copyright 2022 Tinkerbell.

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

// These types are the Rufio v1alpha1 APIs/types copied from https://github.com/tinkerbell/rufio/tree/main/api/v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "bmc.tinkerbell.org", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// PowerState represents power state of a Machine.
type PowerState string

const (
	// On represents that a Machine is powered on.
	On PowerState = "on"
	// Off represents that a Machine is powered off.
	Off PowerState = "off"
	// Unknown represents that a Machine's power state is unknown.
	Unknown PowerState = "unknown"
	// PXE is the boot device name for PXE booting a machine.
	PXE string = "pxe"
)

// MachineConditionType represents the condition of the Machine.
type MachineConditionType string

const (
	// Contactable defines that a connection can be made to the Machine.
	Contactable MachineConditionType = "Contactable"
)

// ConditionStatus represents the status of a Condition.
type ConditionStatus string

const (
	// ConditionTrue represents that a Machine is contactable.
	ConditionTrue ConditionStatus = "True"
	// ConditionFalse represents that a Machine is not contactable.
	ConditionFalse ConditionStatus = "False"
)

// MachineSpec defines desired machine state.
type MachineSpec struct {
	// Connection contains connection data for a Baseboard Management Controller.
	Connection Connection `json:"connection"`
}

// ProviderOptions contains all the provider specific options.
type ProviderOptions struct {
	// IntelAMT contains the options to customize the IntelAMT provider.
	// +optional
	IntelAMT *IntelAMTOptions `json:"intelAMT"`

	// IPMITOOL contains the options to customize the Ipmitool provider.
	// +optional
	IPMITOOL *IPMITOOLOptions `json:"ipmitool"`

	// Redfish contains the options to customize the Redfish provider.
	// +optional
	Redfish *RedfishOptions `json:"redfish"`

	// RPC contains the options to customize the RPC provider.
	// +optional
	RPC *RPCOptions `json:"rpc"`
}

// Connection contains connection data for a Baseboard Management Controller.
type Connection struct {
	// Host is the host IP address or hostname of the Machine.
	// +kubebuilder:validation:MinLength=1
	Host string `json:"host"`

	// Port is the port number for connecting with the Machine.
	// +kubebuilder:default:=623
	// +optional
	Port int `json:"port"`

	// AuthSecretRef is the SecretReference that contains authentication information of the Machine.
	// The Secret must contain username and password keys. This is optional as it is not required when using
	// the RPC provider.
	// +optional
	AuthSecretRef corev1.SecretReference `json:"authSecretRef"`

	// InsecureTLS specifies trusted TLS connections.
	InsecureTLS bool `json:"insecureTLS"`

	// ProviderOptions contains provider specific options.
	// +optional
	ProviderOptions *ProviderOptions `json:"providerOptions,omitempty"`
}

// MachineStatus defines the observed state of Machine.
type MachineStatus struct {
	// Power is the current power state of the Machine.
	// +kubebuilder:validation:Enum=on;off;unknown
	// +optional
	Power PowerState `json:"powerState,omitempty"`

	// Conditions represents the latest available observations of an object's current state.
	// +optional
	Conditions []MachineCondition `json:"conditions,omitempty"`
}

// MachineCondition defines an observed condition of a Machine.
type MachineCondition struct {
	// Type of the Machine condition.
	Type MachineConditionType `json:"type"`

	// Status of the condition.
	Status ConditionStatus `json:"status"`

	// LastUpdateTime of the condition.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Message is a human readable message indicating with details of the last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// MachineSetConditionOption is a function that manipulates a MachineCondition.
// +kubebuilder:object:generate=false
type MachineSetConditionOption func(*MachineCondition)

// SetCondition applies the cType condition to bm. If the condition already exists,
// it is updated.
func (bm *Machine) SetCondition(cType MachineConditionType, status ConditionStatus, opts ...MachineSetConditionOption) {
	var condition *MachineCondition

	// Check if there's an existing condition.
	for i, c := range bm.Status.Conditions {
		if c.Type == cType {
			condition = &bm.Status.Conditions[i]
			break
		}
	}

	// We didn't find an existing condition so create a new one and append it.
	if condition == nil {
		bm.Status.Conditions = append(bm.Status.Conditions, MachineCondition{
			Type: cType,
		})
		condition = &bm.Status.Conditions[len(bm.Status.Conditions)-1]
	}

	if condition.Status != status {
		condition.Status = status
		condition.LastUpdateTime = metav1.Now()
	}

	for _, opt := range opts {
		opt(condition)
	}
}

// WithMachineConditionMessage sets message m to the MachineCondition.
func WithMachineConditionMessage(m string) MachineSetConditionOption {
	return func(c *MachineCondition) {
		c.Message = m
	}
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=machines,scope=Namespaced,categories=tinkerbell,singular=machine

// Machine is the Schema for the machines API.
type Machine struct {
	metav1.TypeMeta   `json:""`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineSpec   `json:"spec,omitempty"`
	Status MachineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachineList contains a list of Machines.
type MachineList struct {
	metav1.TypeMeta `json:""`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Machine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Machine{}, &MachineList{})
}
