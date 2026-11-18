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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // suppress complaining on Deprecated package
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// NutanixMachineKind represents the Kind of NutanixMachine
	NutanixMachineKind = "NutanixMachine"

	// NutanixMachineFinalizer allows NutanixMachineReconciler to clean up AHV
	// resources associated with NutanixMachine before removing it from the
	// API Server.
	NutanixMachineFinalizer           = "infrastructure.cluster.x-k8s.io/nutanixmachine"
	DeprecatedNutanixMachineFinalizer = "nutanixmachine.infrastructure.cluster.x-k8s.io"

	// NutanixMachineBootstrapRefKindSecret represents the Kind of Secret
	// referenced by NutanixMachine's BootstrapRef.
	NutanixMachineBootstrapRefKindSecret = "Secret"

	// NutanixMachineBootstrapRefKindImage represents the Kind of Image
	// referenced by NutanixMachine's BootstrapRef. If the BootstrapRef.Kind is set
	// to Image, the NutanixMachine will be created with the image mounted
	// as a CD-ROM.
	NutanixMachineBootstrapRefKindImage = "Image"

	// NutanixMachineDiskModeStandard represents the standard disk mode.
	NutanixMachineDiskModeStandard NutanixMachineDiskMode = "Standard"

	// NutanixMachineDiskModeFlash represents the flash disk mode.
	NutanixMachineDiskModeFlash NutanixMachineDiskMode = "Flash"

	// NutanixMachineDiskDeviceTypeDisk represents the disk device type.
	NutanixMachineDiskDeviceTypeDisk NutanixMachineDiskDeviceType = "Disk"

	// NutanixMachineDiskDeviceTypeCDRom represents the CD-ROM device type.
	NutanixMachineDiskDeviceTypeCDRom NutanixMachineDiskDeviceType = "CDRom"

	// NutanixMachineDiskAdapterTypeSCSI represents the SCSI adapter type.
	NutanixMachineDiskAdapterTypeSCSI NutanixMachineDiskAdapterType = "SCSI"

	// NutanixMachineDiskAdapterTypeIDE represents the IDE adapter type.
	NutanixMachineDiskAdapterTypeIDE NutanixMachineDiskAdapterType = "IDE"

	// NutanixMachineDiskAdapterTypePCI represents the PCI adapter type.
	NutanixMachineDiskAdapterTypePCI NutanixMachineDiskAdapterType = "PCI"

	// NutanixMachineDiskAdapterTypeSATA represents the SATA adapter type.
	NutanixMachineDiskAdapterTypeSATA NutanixMachineDiskAdapterType = "SATA"

	// NutanixMachineDiskAdapterTypeSPAPR represents the SPAPR adapter type.
	NutanixMachineDiskAdapterTypeSPAPR NutanixMachineDiskAdapterType = "SPAPR"
)

// NutanixImageLookup defines how to fetch images for the cluster
// using the fields combined.
type NutanixImageLookup struct {
	// Format is the naming format to look up the image for this
	// machine It will be ignored if an explicit image is set. Supports
	// substitutions for {{.BaseOS}} and {{.K8sVersion}} with the base OS and
	// kubernetes version, respectively. The BaseOS will be the value in
	// BaseOS and the K8sVersion is the value in the Machine .spec.version, with the v prefix removed.
	// This is effectively the defined by the packages produced by kubernetes/release without v as a
	// prefix: 1.13.0, 1.12.5-mybuild.1, or 1.17.3. For example, the default
	// image format of {{.BaseOS}}-?{{.K8sVersion}}-* and BaseOS as "rhel-8.10" will end up
	// searching for images that match the pattern rhel-8.10-1.30.5-* for a
	// Machine that is targeting kubernetes v1.30.5. See
	// also: https://golang.org/pkg/text/template/
	// +kubebuilder:default:="capx-{{.BaseOS}}-{{.K8sVersion}}-*"
	Format *string `json:"format,omitempty"`
	// BaseOS is the name of the base operating system to use for
	// image lookup.
	// +kubebuilder:validation:MinLength:=1
	BaseOS string `json:"baseOS"`
}

// NutanixMachineSpec defines the desired state of NutanixMachine
// +kubebuilder:validation:XValidation:rule="has(self.image) != has(self.imageLookup)",message="Either 'image' or 'imageLookup' must be set, but not both"
// +kubebuilder:validation:XValidation:rule="has(self.subnet) && size(self.subnet) > 1 ? self.subnet.all(x, self.subnet.exists_one(y, x == y)) : true",message="each subnet must be unique"
type NutanixMachineSpec struct {
	// SPEC FIELDS - desired state of NutanixMachine
	// Important: Run "make" to regenerate code after modifying this file

	// ProviderID is the unique identifier as specified by the cloud provider.
	// +optional
	ProviderID string `json:"providerID,omitempty"`
	// vcpusPerSocket is the number of vCPUs per socket of the VM
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	VCPUsPerSocket int32 `json:"vcpusPerSocket"`
	// vcpuSockets is the number of vCPU sockets of the VM
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	VCPUSockets int32 `json:"vcpuSockets"`
	// memorySize is the memory size (in Quantity format) of the VM
	// The minimum memorySize is 2Gi bytes
	// +kubebuilder:validation:Required
	MemorySize resource.Quantity `json:"memorySize"`
	// image is to identify the nutanix machine image uploaded to the Prism Central (PC)
	// The image identifier (uuid or name) can be obtained from the Prism Central console
	// or using the prism_central API.
	// +kubebuilder:validation:Optional
	// +optional
	Image *NutanixResourceIdentifier `json:"image,omitempty"`
	// imageLookup is a container that holds how to look up rhcos images for the cluster.
	// +kubebuilder:validation:Optional
	// +optional
	ImageLookup *NutanixImageLookup `json:"imageLookup,omitempty"`
	// cluster is to identify the cluster (the Prism Element under management
	// of the Prism Central), in which the Machine's VM will be created.
	// The cluster identifier (uuid or name) can be obtained from the Prism Central console
	// or using the prism_central API.
	// +kubebuilder:validation:Optional
	Cluster NutanixResourceIdentifier `json:"cluster,omitzero"`
	// subnet is to identify the cluster's network subnet to use for the Machine's VM
	// The cluster identifier (uuid or name) can be obtained from the Prism Central console
	// or using the prism_central API.
	// +kubebuilder:validation:MaxItems=32
	// +kubebuilder:validation:Optional
	Subnets []NutanixResourceIdentifier `json:"subnet,omitempty"`
	// List of categories that need to be added to the machines. Categories must already exist in Prism Central
	// +kubebuilder:validation:Optional
	AdditionalCategories []NutanixCategoryIdentifier `json:"additionalCategories,omitempty"`
	// Add the machine resources to a Prism Central project
	// +optional
	Project *NutanixResourceIdentifier `json:"project,omitempty"`
	// Defines the boot type of the virtual machine. Only supports UEFI and Legacy
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:=legacy;uefi
	BootType NutanixBootType `json:"bootType,omitempty"`
	// systemDiskSize is size (in Quantity format) of the system disk of the VM
	// The minimum systemDiskSize is 20Gi bytes
	// +kubebuilder:validation:Required
	SystemDiskSize resource.Quantity `json:"systemDiskSize"`

	// dataDisks hold the list of data disks to be attached to the VM
	// +kubebuilder:validation:Optional
	DataDisks []NutanixMachineVMDisk `json:"dataDisks,omitempty"`

	// BootstrapRef is a reference to a bootstrap provider-specific resource
	// that holds configuration details.
	// +optional
	BootstrapRef *corev1.ObjectReference `json:"bootstrapRef,omitempty"`
	// List of GPU devices that need to be added to the machines.
	// +kubebuilder:validation:Optional
	GPUs []NutanixGPU `json:"gpus,omitempty"`
}

// NutanixMachineVMDisk defines the disk configuration for a NutanixMachine
type NutanixMachineVMDisk struct {
	// diskSize is the size (in Quantity format) of the disk attached to the VM.
	// See https://pkg.go.dev/k8s.io/apimachinery/pkg/api/resource#Format for the Quantity format and example documentation.
	// The minimum diskSize is 1GB.
	// +kubebuilder:validation:Required
	DiskSize resource.Quantity `json:"diskSize"`

	// deviceProperties are the properties of the disk device.
	// +optional
	// +kubebuilder:validation:Optional
	DeviceProperties *NutanixMachineVMDiskDeviceProperties `json:"deviceProperties,omitempty"`

	// storageConfig are the storage configuration parameters of the VM disks.
	// +optional
	// +kubebuilder:validation:Optional
	StorageConfig *NutanixMachineVMStorageConfig `json:"storageConfig,omitempty"`

	// dataSource refers to a data source image for the VM disk.
	// +optional
	// +kubebuilder:validation:Optional
	DataSource *NutanixResourceIdentifier `json:"dataSource,omitempty"`
}

// NutanixMachineVMDiskDeviceProperties defines the device properties for a NutanixMachineVMDisk
type NutanixMachineVMDiskDeviceProperties struct {
	// deviceType specifies the disk device type.
	// The valid values are "Disk" and "CDRom", and the default is "Disk".
	// +kubebuilder:default=Disk
	// +kubebuilder:validation:Required
	DeviceType NutanixMachineDiskDeviceType `json:"deviceType"`

	// adapterType is the adapter type of the disk address.
	// If the deviceType is "Disk", the valid adapterType can be "SCSI", "IDE", "PCI", "SATA" or "SPAPR".
	// If the deviceType is "CDRom", the valid adapterType can be "IDE" or "SATA".
	// +kubebuilder:validation:Required
	AdapterType NutanixMachineDiskAdapterType `json:"adapterType,omitempty"`

	// deviceIndex is the index of the disk address. The valid values are non-negative integers, with the default value 0.
	// For a Machine VM, the deviceIndex for the disks with the same deviceType.adapterType combination should
	// start from 0 and increase consecutively afterwards. Note that for each Machine VM, the Disk.SCSI.0
	// and CDRom.IDE.0 are reserved to be used by the VM's system. So for dataDisks of Disk.SCSI and CDRom.IDE,
	// the deviceIndex should start from 1.
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	// +optional
	// +kubebuilder:validation:Optional
	DeviceIndex int32 `json:"deviceIndex,omitempty"`
}

// NutanixMachineVMStorageConfig defines the storage configuration for a NutanixMachineVMDisk
type NutanixMachineVMStorageConfig struct {
	// diskMode specifies the disk mode.
	// The valid values are Standard and Flash, and the default is Standard.
	// +kubebuilder:default=Standard
	// +kubebuilder:validation:Required
	DiskMode NutanixMachineDiskMode `json:"diskMode"`

	// storageContainer refers to the storage_container used by the VM disk.
	// +optional
	// +kubebuilder:validation:Optional
	StorageContainer *NutanixResourceIdentifier `json:"storageContainer"`
}

// NutanixMachineDiskMode is an enumeration of different disk modes.
// +kubebuilder:validation:Enum=Standard;Flash
type NutanixMachineDiskMode string

// NutanixMachineDiskDeviceType is the VM disk device type.
// +kubebuilder:validation:Enum=Disk;CDRom
type NutanixMachineDiskDeviceType string

// NutanixMachineDiskAdapterType is an enumeration of different disk device adapter types.
// +kubebuilder:validation:Enum:=SCSI;IDE;PCI;SATA;SPAPR
type NutanixMachineDiskAdapterType string

// NutanixMachineStatus defines the observed state of NutanixMachine
type NutanixMachineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// Addresses contains the Nutanix VM associated addresses.
	// Address type is one of Hostname, ExternalIP, InternalIP, ExternalDNS, InternalDNS
	Addresses []capiv1beta1.MachineAddress `json:"addresses,omitempty"`

	// The Nutanix VM's UUID
	// +optional
	VmUUID string `json:"vmUUID,omitempty"`

	// NodeRef is a reference to the corresponding workload cluster Node if it exists.
	// Deprecated: Do not use. Will be removed in a future release.
	// +optional
	NodeRef *corev1.ObjectReference `json:"nodeRef,omitempty"`

	// Conditions defines current service state of the NutanixMachine.
	// +optional
	Conditions capiv1beta1.Conditions `json:"conditions,omitempty"`

	// Will be set in case of failure of Machine instance
	// +optional
	FailureReason *string `json:"failureReason,omitempty"`

	// Will be set in case of failure of Machine instance
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// failureDomain is the name of the failure domain where this Machine has been placed in.
	// +optional
	FailureDomain *string `json:"failureDomain,omitempty"`

	// v1beta2 groups all the fields that will be added or modified in NutanixMachine's status with the v1beta2 version.
	// +optional
	V1Beta2 *NutanixMachineV1Beta2Status `json:"v1beta2,omitempty"`
}

// NutanixMachineV1Beta2Status groups all the fields that will be added or modified in NutanixMachineStatus with the v1beta2 version.
// See https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20240916-improve-status-in-CAPI-resources.md for more context.
type NutanixMachineV1Beta2Status struct {
	// conditions represents the observations of a NutanixMachine's current state.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=nutanixmachines,shortName=nma,scope=Namespaced,categories=cluster-api
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Address",type="string",JSONPath=".status.addresses[0].address",description="The VM address"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="NutanixMachine ready status"
// +kubebuilder:printcolumn:name="ProviderID",type="string",JSONPath=".spec.providerID",description="NutanixMachine instance ID"
// +kubebuilder:printcolumn:name="FailureDomain",type="string",JSONPath=".status.failureDomain",description="NutanixMachine FailureDomain"
// NutanixMachine is the Schema for the nutanixmachines API
type NutanixMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NutanixMachineSpec   `json:"spec,omitempty"`
	Status NutanixMachineStatus `json:"status,omitempty"`
}

// GetConditions returns the set of conditions for this object.
func (nm *NutanixMachine) GetConditions() capiv1beta1.Conditions {
	return nm.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (nm *NutanixMachine) SetConditions(conditions capiv1beta1.Conditions) {
	nm.Status.Conditions = conditions
}

// GetV1Beta2Conditions returns the set of conditions for this object.
func (ncl *NutanixMachine) GetV1Beta2Conditions() []metav1.Condition {
	if ncl.Status.V1Beta2 == nil {
		return nil
	}
	return ncl.Status.V1Beta2.Conditions
}

// SetV1Beta2Conditions sets the v1beta2 conditions on this object.
func (ncl *NutanixMachine) SetV1Beta2Conditions(conditions []metav1.Condition) {
	if ncl.Status.V1Beta2 == nil {
		ncl.Status.V1Beta2 = &NutanixMachineV1Beta2Status{}
	}
	ncl.Status.V1Beta2.Conditions = conditions
}

//+kubebuilder:object:root=true

// NutanixMachineList contains a list of NutanixMachine
type NutanixMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NutanixMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NutanixMachine{}, &NutanixMachineList{})
}
