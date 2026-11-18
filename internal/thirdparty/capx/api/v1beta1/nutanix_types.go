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

import "fmt"

// NutanixIdentifierType is an enumeration of different resource identifier types.
type NutanixIdentifierType string

// NutanixBootType is an enumeration of different boot types.
type NutanixBootType string

// NutanixGPUIdentifierType is an enumeration of different resource identifier types for GPU entities.
type NutanixGPUIdentifierType string

const (
	// NutanixIdentifierUUID is a resource identifier identifying the object by UUID.
	NutanixIdentifierUUID NutanixIdentifierType = "uuid"

	// NutanixIdentifierName is a resource identifier identifying the object by Name.
	NutanixIdentifierName NutanixIdentifierType = "name"

	// NutanixBootTypeLegacy is a resource identifier identifying the legacy boot type for virtual machines.
	NutanixBootTypeLegacy NutanixBootType = "legacy"

	// NutanixBootTypeUEFI is a resource identifier identifying the UEFI boot type for virtual machines.
	NutanixBootTypeUEFI NutanixBootType = "uefi"

	// NutanixGPUIdentifierName is a resource identifier identifying a GPU by Name.
	NutanixGPUIdentifierName NutanixGPUIdentifierType = "name"

	// NutanixGPUIdentifierDeviceID is a resource identifier identifying a GPU using device ID.
	NutanixGPUIdentifierDeviceID NutanixGPUIdentifierType = "deviceID"

	// ObsoleteDefaultCAPICategoryPrefix is the obsolete default category prefix used for CAPI clusters.
	ObsoleteDefaultCAPICategoryPrefix = "kubernetes-io-cluster-"

	// DefaultCAPICategoryKeyForName is the default category key used for CAPI clusters for cluster names.
	DefaultCAPICategoryKeyForName = "KubernetesClusterName"

	// DefaultCAPICategoryDescription is the default category description used for CAPI clusters.
	DefaultCAPICategoryDescription = "Managed by CAPX"

	// ObsoleteDefaultCAPICategoryOwnedValue is the obsolete default category value used for CAPI clusters.
	ObsoleteDefaultCAPICategoryOwnedValue = "owned"
)

// NutanixResourceIdentifier holds the identity of a Nutanix PC resource (cluster, image, subnet, etc.)
// +union
// +kubebuilder:validation:XValidation:rule="self.type == 'name' ? has(self.name) : !has(self.name)",message="'name' must be set when type is 'name', and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="self.type == 'uuid' ? has(self.uuid) && self.uuid.contains('-') : !has(self.uuid)",message="'uuid' must be set when type is 'uuid', and forbidden otherwise"
type NutanixResourceIdentifier struct {
	// Type is the identifier type to use for this resource.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=uuid;name
	Type NutanixIdentifierType `json:"type"`

	// uuid is the UUID of the resource in the PC.
	// +optional
	// +kubebuilder:validation:Format=uuid
	// +kubebuilder:validation:MaxLength=36
	// +kubebuilder:validation:MinLength=36
	UUID *string `json:"uuid,omitempty"`

	// name is the resource name in the PC
	// +optional
	// +kubebuilder:validation:MinLength=1
	Name *string `json:"name,omitempty"`
}

func (nri NutanixResourceIdentifier) String() string {
	if nri.Type == NutanixIdentifierUUID && nri.UUID != nil {
		return *nri.UUID
	}
	if nri.Type == NutanixIdentifierName && nri.Name != nil {
		return *nri.Name
	}
	return ""
}

// DisplayString returns a human-readable string representation of the NutanixResourceIdentifier
// that includes both the type and value, suitable for error messages and logging.
func (nri NutanixResourceIdentifier) DisplayString() string {
	switch nri.Type {
	case NutanixIdentifierUUID:
		if nri.UUID != nil {
			return fmt.Sprintf("uuid=%q", *nri.UUID)
		}
	case NutanixIdentifierName:
		if nri.Name != nil {
			return fmt.Sprintf("name=%q", *nri.Name)
		}
	}
	return "unknown"
}

func (nri NutanixResourceIdentifier) IsUUID() bool {
	return nri.Type == NutanixIdentifierUUID && nri.UUID != nil
}

func (nri NutanixResourceIdentifier) IsName() bool {
	return nri.Type == NutanixIdentifierName && nri.Name != nil
}

// EqualTo checks if two NutanixResourceIdentifiers are equal based on their type and value.
func (nri NutanixResourceIdentifier) EqualTo(other *NutanixResourceIdentifier) bool {
	if other == nil {
		return false
	}
	if nri.Type != other.Type {
		return false
	}

	switch nri.Type {
	case NutanixIdentifierName:
		return nri.Name != nil && other.Name != nil && *nri.Name == *other.Name
	case NutanixIdentifierUUID:
		return nri.UUID != nil && other.UUID != nil && *nri.UUID == *other.UUID
	}

	return false
}

type NutanixCategoryIdentifier struct {
	// key is the Key of category in PC.
	// +optional
	Key string `json:"key,omitempty"`

	// value is the category value linked to the category key in PC
	// +optional
	Value string `json:"value,omitempty"`
}

type NutanixGPU struct {
	// Type is the identifier type to use for this resource.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=deviceID;name
	Type NutanixGPUIdentifierType `json:"type"`

	// deviceID is the id of the GPU entity.
	// +optional
	DeviceID *int64 `json:"deviceID,omitempty"`

	// name is the GPU name
	// +optional
	Name *string `json:"name,omitempty"`
}
