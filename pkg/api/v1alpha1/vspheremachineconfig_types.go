package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VSphereMachineConfigSpec defines the desired state of VSphereMachineConfig
type VSphereMachineConfigSpec struct {
	DiskGiB           int                 `json:"diskGiB,omitempty"`
	Datastore         string              `json:"datastore"`
	Folder            string              `json:"folder"`
	NumCPUs           int                 `json:"numCPUs"`
	MemoryMiB         int                 `json:"memoryMiB"`
	OSFamily          OSFamily            `json:"osFamily"`
	ResourcePool      string              `json:"resourcePool"`
	StoragePolicyName string              `json:"storagePolicyName,omitempty"`
	Template          string              `json:"template,omitempty"`
	Users             []UserConfiguration `json:"users,omitempty"`
}

func (c *VSphereMachineConfig) PauseReconcile() {
	c.Annotations[pausedAnnotation] = "true"
}

func (c *VSphereMachineConfig) IsReconcilePaused() bool {
	if s, ok := c.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *VSphereMachineConfig) SetControlPlane() {
	c.Annotations[controlPlaneAnnotation] = "true"
}

func (c *VSphereMachineConfig) IsControlPlane() bool {
	if s, ok := c.Annotations[controlPlaneAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *VSphereMachineConfig) SetEtcd() {
	c.Annotations[etcdAnnotation] = "true"
}

func (c *VSphereMachineConfig) IsEtcd() bool {
	if s, ok := c.Annotations[etcdAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *VSphereMachineConfig) SetManagedBy(clusterName string) {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[managementAnnotation] = clusterName
}

// IsManaged returns true if the vspheremachineconfig is associated with a workload cluster
func (c *VSphereMachineConfig) IsManaged() bool {
	if s, ok := c.Annotations[managementAnnotation]; ok {
		return s != ""
	}
	return false
}

func (c *VSphereMachineConfig) OSFamily() OSFamily {
	return c.Spec.OSFamily
}

func (c *VSphereMachineConfig) GetNamespace() string {
	return c.Namespace
}

func (c *VSphereMachineConfig) GetName() string {
	return c.Name
}

// VSphereMachineConfigStatus defines the observed state of VSphereMachineConfig
type VSphereMachineConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VSphereMachineConfig is the Schema for the vspheremachineconfigs API
type VSphereMachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VSphereMachineConfigSpec   `json:"spec,omitempty"`
	Status VSphereMachineConfigStatus `json:"status,omitempty"`
}

func (c *VSphereMachineConfig) ConvertConfigToConfigGenerateStruct() *VSphereMachineConfigGenerate {
	namespace := defaultEksaNamespace
	if c.Namespace != "" {
		namespace = c.Namespace
	}
	config := &VSphereMachineConfigGenerate{
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

func (c *VSphereMachineConfig) Marshallable() Marshallable {
	return c.ConvertConfigToConfigGenerateStruct()
}

// +kubebuilder:object:generate=false

// Same as VSphereMachineConfig except stripped down for generation of yaml file during generate clusterconfig
type VSphereMachineConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec VSphereMachineConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// VSphereMachineConfigList contains a list of VSphereMachineConfig
type VSphereMachineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VSphereMachineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VSphereMachineConfig{}, &VSphereMachineConfigList{})
}
