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
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudStackMachineConfigSpec defines the desired state of CloudStackMachineConfig
type CloudStackMachineConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Template refers to a VM image template which has been previously registered in CloudStack. It can either be specified as a UUID or name
	Template CloudStackResourceIdentifier `json:"template"`
	// ComputeOffering refers to a compute offering which has been previously registered in CloudStack. It represents a VM’s instance size including number of CPU’s, memory, and CPU speed. It can either be specified as a UUID or name
	ComputeOffering CloudStackResourceIdentifier `json:"computeOffering"`
	// DiskOffering refers to a disk offering which has been previously registered in CloudStack. It represents a disk offering with pre-defined size or custom specified disk size. It can either be specified as a UUID or name
	DiskOffering CloudStackResourceDiskOffering `json:"diskOffering,omitempty"`
	// Users consists of an array of objects containing the username, as well as a list of their public keys. These users will be authorized to ssh into the machines
	Users []UserConfiguration `json:"users,omitempty"`
	// Defaults to `no`. Can be `pro` or `anti`. If set to `pro` or `anti`, will create an affinity group per machine set of the corresponding type
	Affinity string `json:"affinity,omitempty"`
	// AffinityGroupIds allows users to pass in a list of UUIDs for previously-created Affinity Groups. Any VM’s created with this spec will be added to the affinity group, which will dictate which physical host(s) they can be placed on. Affinity groups can be type “affinity” or “anti-affinity” in CloudStack. If they are type “anti-affinity”, all VM’s in the group must be on separate physical hosts for high availability. If they are type “affinity”, all VM’s in the group must be on the same physical host for improved performance
	AffinityGroupIds []string `json:"affinityGroupIds,omitempty"`
	// UserCustomDetails allows users to pass in non-standard key value inputs, outside those defined [here](https://github.com/shapeblue/cloudstack/blob/main/api/src/main/java/com/cloud/vm/VmDetailConstants.java)
	UserCustomDetails map[string]string `json:"userCustomDetails,omitempty"`
}

type CloudStackResourceDiskOffering struct {
	CloudStackResourceIdentifier `json:",inline"`
	// disk size in GB, > 0 for customized disk offering; = 0 for non-customized disk offering
	// +optional
	CustomSize int64 `json:"customSizeInGB"`
	// path the filesystem will use to mount in VM
	MountPath string `json:"mountPath"`
	// device name of the disk offering in VM, shows up in lsblk command
	Device string `json:"device"`
	// filesystem used to mkfs in disk offering partition
	Filesystem string `json:"filesystem"`
	// disk label used to label disk partition
	Label string `json:"label"`
}

func (r *CloudStackResourceDiskOffering) Equal(o *CloudStackResourceDiskOffering) bool {
	if r == o {
		return true
	}
	if r == nil || o == nil {
		return false
	}
	if r.Id != o.Id {
		return false
	}

	if r.MountPath != o.MountPath || r.Filesystem != o.Filesystem || r.Label != o.Label || r.Device != o.Device {
		return false
	}
	return r.Id == "" && o.Id == "" && r.Name == o.Name
}

func (r *CloudStackResourceDiskOffering) Validate() (err error, field string, value string) {
	if len(r.Id) > 0 || len(r.Name) > 0 {
		if len(r.MountPath) < 2 || !strings.HasPrefix(r.MountPath, "/") {
			return errors.New("must be non-empty and starts with /"), "mountPath", r.MountPath
		}
		if len(r.Filesystem) < 1 {
			return errors.New("empty filesystem"), "filesystem", r.Filesystem
		}
		if len(r.Device) < 1 {
			return errors.New("empty device"), "device", r.Device
		}
		if len(r.Label) < 1 {
			return errors.New("empty label"), "label", r.Label
		}
	} else {
		if len(r.MountPath)+len(r.Filesystem)+len(r.Device)+len(r.Label) > 0 {
			return errors.New("empty id/name"), "id or name", r.Id
		}
	}
	return nil, "", ""
}

func (c *CloudStackMachineConfig) PauseReconcile() {
	c.Annotations[pausedAnnotation] = "true"
}

func (c *CloudStackMachineConfig) IsReconcilePaused() bool {
	if s, ok := c.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *CloudStackMachineConfig) SetControlPlane() {
	c.Annotations[controlPlaneAnnotation] = "true"
}

func (c *CloudStackMachineConfig) IsControlPlane() bool {
	if s, ok := c.Annotations[controlPlaneAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *CloudStackMachineConfig) SetEtcd() {
	c.Annotations[etcdAnnotation] = "true"
}

func (c *CloudStackMachineConfig) IsEtcd() bool {
	if s, ok := c.Annotations[etcdAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *CloudStackMachineConfig) SetManagement(clusterName string) {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[managementAnnotation] = clusterName
}

func (c *CloudStackMachineConfig) IsManagement() bool {
	if s, ok := c.Annotations[managementAnnotation]; ok {
		return s != ""
	}
	return false
}

func (c *CloudStackMachineConfig) GetNamespace() string {
	return c.Namespace
}

func (c *CloudStackMachineConfig) GetName() string {
	return c.Name
}

func (c *CloudStackMachineConfig) Validate() error {
	return nil
}

// CloudStackMachineConfigStatus defines the observed state of CloudStackMachineConfig
type CloudStackMachineConfigStatus struct { // INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudStackMachineConfig is the Schema for the cloudstackmachineconfigs API
type CloudStackMachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudStackMachineConfigSpec   `json:"spec,omitempty"`
	Status CloudStackMachineConfigStatus `json:"status,omitempty"`
}

func (c *CloudStackMachineConfig) OSFamily() OSFamily {
	// This method must be defined to implement the providers.MachineConfig interface, but it's not actually used
	return ""
}

func (c *CloudStackMachineConfigSpec) Equal(o *CloudStackMachineConfigSpec) bool {
	if c == o {
		return true
	}
	if c == nil || o == nil {
		return false
	}
	if !c.Template.Equal(&o.Template) ||
		!c.ComputeOffering.Equal(&o.ComputeOffering) ||
		!c.DiskOffering.Equal(&o.DiskOffering) {
		return false
	}
	if c.Affinity != o.Affinity {
		return false
	}
	if !SliceEqual(c.AffinityGroupIds, o.AffinityGroupIds) {
		return false
	}
	if !UsersSliceEqual(c.Users, o.Users) {
		return false
	}
	if len(c.UserCustomDetails) != len(o.UserCustomDetails) {
		return false
	}
	for detail, value := range c.UserCustomDetails {
		if value != o.UserCustomDetails[detail] {
			return false
		}
	}
	return true
}

func (c *CloudStackMachineConfig) ConvertConfigToConfigGenerateStruct() *CloudStackMachineConfigGenerate {
	namespace := defaultEksaNamespace
	if c.Namespace != "" {
		namespace = c.Namespace
	}
	config := &CloudStackMachineConfigGenerate{
		TypeMeta: c.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        c.Name,
			Annotations: c.Annotations,
			Namespace:   namespace,
		},
		Spec: c.Spec,
	}

	return config
}

func (c *CloudStackMachineConfig) Marshallable() Marshallable {
	return c.ConvertConfigToConfigGenerateStruct()
}

// +kubebuilder:object:generate=false

// Same as CloudStackMachineConfig except stripped down for generation of yaml file during generate clusterconfig
type CloudStackMachineConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec CloudStackMachineConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CloudStackMachineConfigList contains a list of CloudStackMachineConfig
type CloudStackMachineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudStackMachineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudStackMachineConfig{}, &CloudStackMachineConfigList{})
}
