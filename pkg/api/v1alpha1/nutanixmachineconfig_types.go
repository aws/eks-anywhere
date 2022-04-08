package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NutanixMachineConfigSpec defines the desired state of NutanixMachineConfig
type NutanixMachineConfigSpec struct {
	OSFamily    OSFamily            `json:"osFamily"`
	Users       []UserConfiguration `json:"users,omitempty"`
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
	// image is to identify the rhcos image uploaded to the Prism Central (PC)
	// The image identifier (uuid or name) can be obtained from the Prism Central console
	// or using the prism_central API.
	// +kubebuilder:validation:Required
	Image NutanixResourceIdentifier `json:"image"`
	// cluster is to identify the cluster (the Prism Element under management
	// of the Prism Central), in which the Machine's VM will be created.
	// The cluster identifier (uuid or name) can be obtained from the Prism Central console
	// or using the prism_central API.
	// +kubebuilder:validation:Required
	Cluster NutanixResourceIdentifier `json:"cluster"`
	// subnet is to identify the cluster's network subnet to use for the Machine's VM
	// The cluster identifier (uuid or name) can be obtained from the Prism Central console
	// or using the prism_central API.
	// +kubebuilder:validation:Required
	Subnet NutanixResourceIdentifier `json:"subnet"`

	// systemDiskSize is size (in Quantity format) of the system disk of the VM
	// The minimum systemDiskSize is 20Gi bytes
	// +kubebuilder:validation:Required
	SystemDiskSize resource.Quantity `json:"systemDiskSize"`
}

func (c *NutanixMachineConfig) PauseReconcile() {
	c.Annotations[pausedAnnotation] = "true"
}

func (c *NutanixMachineConfig) IsReconcilePaused() bool {
	if s, ok := c.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *NutanixMachineConfig) SetControlPlane() {
	c.Annotations[controlPlaneAnnotation] = "true"
}

func (c *NutanixMachineConfig) IsControlPlane() bool {
	if s, ok := c.Annotations[controlPlaneAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *NutanixMachineConfig) SetEtcd() {
	c.Annotations[etcdAnnotation] = "true"
}

func (c *NutanixMachineConfig) IsEtcd() bool {
	if s, ok := c.Annotations[etcdAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *NutanixMachineConfig) SetManagedBy(clusterName string) {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[managementAnnotation] = clusterName
}

func (c *NutanixMachineConfig) IsManaged() bool {
	if s, ok := c.Annotations[managementAnnotation]; ok {
		return s != ""
	}
	return false
}

func (c *NutanixMachineConfig) OSFamily() OSFamily {
	return c.Spec.OSFamily
}

func (c *NutanixMachineConfig) GetNamespace() string {
	return c.Namespace
}

func (c *NutanixMachineConfig) GetName() string {
	return c.Name
}

// NutanixMachineConfigStatus defines the observed state of NutanixMachineConfig
type NutanixMachineConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// Addresses contains the Nutanix VM associated addresses.
	// Address type is one of Hostname, ExternalIP, InternalIP, ExternalDNS, InternalDNS
	Addresses []capiv1.MachineAddress `json:"addresses,omitempty"`

	// The Nutanix VM's UUID
	// +optional
	VmUUID *string `json:"vmUUID,omitempty"`

	// NodeRef is a reference to the corresponding workload cluster Node if it exists.
	// +optional
	NodeRef *corev1.ObjectReference `json:"nodeRef,omitempty"`

	// Conditions defines current service state of the NutanixMachine.
	// +optional
	Conditions capiv1.Conditions `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NutanixMachineConfig is the Schema for the tinkerbellmachineconfigs API
type NutanixMachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NutanixMachineConfigSpec   `json:"spec,omitempty"`
	Status NutanixMachineConfigStatus `json:"status,omitempty"`
}

func (c *NutanixMachineConfig) ConvertConfigToConfigGenerateStruct() *NutanixMachineConfigGenerate {
	namespace := defaultEksaNamespace
	if c.Namespace != "" {
		namespace = c.Namespace
	}
	config := &NutanixMachineConfigGenerate{
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

func (c *NutanixMachineConfig) Marshallable() Marshallable {
	return c.ConvertConfigToConfigGenerateStruct()
}

// +kubebuilder:object:generate=false

// Same as NutanixMachineConfig except stripped down for generation of yaml file during generate clusterconfig
type NutanixMachineConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec NutanixMachineConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// NutanixMachineConfigList contains a list of NutanixMachineConfig
type NutanixMachineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NutanixMachineConfig `json:"items"`
}

//var _ providers.MachineConfig = (*NutanixMachineConfig)(nil)

func init() {
	SchemeBuilder.Register(&NutanixMachineConfig{}, &NutanixMachineConfigList{})
}
