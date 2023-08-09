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

/*
Package rufiounreleased contains types that never became a formal release but were included in
EKSA releases. Given we have clusters deployed containing these types it is necessary to keep
them so we can perform conversions.
*/
// nolint
package rufiounreleased

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const BaseboardManagementResourceName = "baseboardmanagements.bmc.tinkerbell.org"

// PowerState represents power state the BaseboardManagement.
type PowerState string

// BootDevice represents boot device of the BaseboardManagement.
type BootDevice string

// BaseboardManagementConditionType represents the condition of the BaseboardManagement.
type BaseboardManagementConditionType string

// ConditionStatus represents the status of a Condition.
type ConditionStatus string

const (
	On  PowerState = "on"
	Off PowerState = "off"
)

const (
	PXE   BootDevice = "pxe"
	Disk  BootDevice = "disk"
	BIOS  BootDevice = "bios"
	CDROM BootDevice = "cdrom"
	Safe  BootDevice = "safe"
)

const (
	// Contactable defines that a connection can be made to the BaseboardManagement.
	Contactable BaseboardManagementConditionType = "Contactable"
)

const (
	ConditionTrue  ConditionStatus = "True"
	ConditionFalse ConditionStatus = "False"
)

// BaseboardManagementSpec defines the desired state of BaseboardManagement.
type BaseboardManagementSpec struct {
	// Connection represents the BaseboardManagement connectivity information.
	Connection Connection `json:"connection"`
}

type Connection struct {
	// Host is the host IP address or hostname of the BaseboardManagement.
	// +kubebuilder:validation:MinLength=1
	Host string `json:"host"`

	// Port is the port number for connecting with the BaseboardManagement.
	// +kubebuilder:default:=623
	Port int `json:"port"`

	// AuthSecretRef is the SecretReference that contains authentication information of the BaseboardManagement.
	// The Secret must contain username and password keys.
	AuthSecretRef corev1.SecretReference `json:"authSecretRef"`

	// InsecureTLS specifies trusted TLS connections.
	InsecureTLS bool `json:"insecureTLS"`
}

// BaseboardManagementStatus defines the observed state of BaseboardManagement.
type BaseboardManagementStatus struct {
	// Power is the current power state of the BaseboardManagement.
	// +kubebuilder:validation:Enum=on;off
	// +optional
	Power PowerState `json:"powerState,omitempty"`

	// Conditions represents the latest available observations of an object's current state.
	// +optional
	Conditions []BaseboardManagementCondition `json:"conditions,omitempty"`
}

type BaseboardManagementCondition struct {
	// Type of the BaseboardManagement condition.
	Type BaseboardManagementConditionType `json:"type"`

	// Status is the status of the BaseboardManagement condition.
	// Can be True or False.
	Status ConditionStatus `json:"status"`

	// Last time the BaseboardManagement condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Message represents human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:generate=false
type BaseboardManagementSetConditionOption func(*BaseboardManagementCondition)

// SetCondition applies the cType condition to bm. If the condition already exists,
// it is updated.
func (bm *BaseboardManagement) SetCondition(cType BaseboardManagementConditionType, status ConditionStatus, opts ...BaseboardManagementSetConditionOption) {
	var condition *BaseboardManagementCondition

	// Check if there's an existing condition.
	for i, c := range bm.Status.Conditions {
		if c.Type == cType {
			condition = &bm.Status.Conditions[i]
			break
		}
	}

	// We didn't find an existing condition so create a new one and append it.
	if condition == nil {
		bm.Status.Conditions = append(bm.Status.Conditions, BaseboardManagementCondition{
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

// WithBaseboardManagementConditionMessage sets message m to the BaseboardManagementCondition.
func WithBaseboardManagementConditionMessage(m string) BaseboardManagementSetConditionOption {
	return func(c *BaseboardManagementCondition) {
		c.Message = m
	}
}

// BaseboardManagementRef defines the reference information to a BaseboardManagement resource.
type BaseboardManagementRef struct {
	// Name is unique within a namespace to reference a BaseboardManagement resource.
	Name string `json:"name"`

	// Namespace defines the space within which the BaseboardManagement name must be unique.
	Namespace string `json:"namespace"`
}

//+kubebuilder:subresource:status
//+kubebuilder:resource:path=baseboardmanagements,scope=Namespaced,categories=tinkerbell,singular=baseboardmanagement,shortName=bm

// BaseboardManagement is the Schema for the baseboardmanagements API.
type BaseboardManagement struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BaseboardManagementSpec   `json:"spec,omitempty"`
	Status BaseboardManagementStatus `json:"status,omitempty"`
}

// BaseboardManagementList contains a list of BaseboardManagement.
type BaseboardManagementList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BaseboardManagement `json:"items"`
}
