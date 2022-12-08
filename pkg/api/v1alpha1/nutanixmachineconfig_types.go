// Important: Run "make generate" to regenerate code after modifying this file
// json tags are required; new fields must have json tags for the fields to be serialized

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// NutanixMachineConfigSpec defines the desired state of NutanixMachineConfig.
type NutanixMachineConfigSpec struct {
	OSFamily OSFamily            `json:"osFamily"`
	Users    []UserConfiguration `json:"users,omitempty"`
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
	// image is to identify the OS image uploaded to the Prism Central (PC)
	// The image identifier (uuid or name) can be obtained from the Prism Central console
	// or using the Prism Central API.
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
	// or using the Prism Central API.
	// +kubebuilder:validation:Required
	Subnet NutanixResourceIdentifier `json:"subnet"`

	// systemDiskSize is size (in Quantity format) of the system disk of the VM
	// The minimum systemDiskSize is 20Gi bytes
	// +kubebuilder:validation:Required
	SystemDiskSize resource.Quantity `json:"systemDiskSize"`
}

func (c *NutanixMachineConfig) SetDefaults() {
	setNutanixMachineConfigDefaults(c)
}

func (in *NutanixMachineConfig) PauseReconcile() {
	in.Annotations[pausedAnnotation] = "true"
}

func (in *NutanixMachineConfig) IsReconcilePaused() bool {
	if s, ok := in.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (in *NutanixMachineConfig) SetControlPlane() {
	in.Annotations[controlPlaneAnnotation] = "true"
}

func (in *NutanixMachineConfig) IsControlPlane() bool {
	if s, ok := in.Annotations[controlPlaneAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (in *NutanixMachineConfig) SetEtcd() {
	in.Annotations[etcdAnnotation] = "true"
}

func (in *NutanixMachineConfig) IsEtcd() bool {
	if s, ok := in.Annotations[etcdAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (in *NutanixMachineConfig) SetManagedBy(clusterName string) {
	if in.Annotations == nil {
		in.Annotations = map[string]string{}
	}
	in.Annotations[managementAnnotation] = clusterName
}

func (in *NutanixMachineConfig) IsManaged() bool {
	if s, ok := in.Annotations[managementAnnotation]; ok {
		return s != ""
	}
	return false
}

func (in *NutanixMachineConfig) OSFamily() OSFamily {
	return in.Spec.OSFamily
}

func (in *NutanixMachineConfig) GetNamespace() string {
	return in.Namespace
}

func (in *NutanixMachineConfig) GetName() string {
	return in.Name
}

// NutanixMachineConfigStatus defines the observed state of NutanixMachineConfig.
type NutanixMachineConfigStatus struct {
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

// NutanixMachineConfig is the Schema for the nutanix machine configs API
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type NutanixMachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NutanixMachineConfigSpec   `json:"spec,omitempty"`
	Status NutanixMachineConfigStatus `json:"status,omitempty"`
}

func (in *NutanixMachineConfig) ConvertConfigToConfigGenerateStruct() *NutanixMachineConfigGenerate {
	namespace := defaultEksaNamespace
	if in.Namespace != "" {
		namespace = in.Namespace
	}
	config := &NutanixMachineConfigGenerate{
		TypeMeta: in.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        in.Name,
			Annotations: in.Annotations,
			Namespace:   namespace,
		},
		Spec: in.Spec,
	}

	return config
}

func (in *NutanixMachineConfig) Marshallable() Marshallable {
	return in.ConvertConfigToConfigGenerateStruct()
}

// NutanixMachineConfigGenerate is same as NutanixMachineConfig except stripped down for generation of yaml file during
// generate clusterconfig
//
// +kubebuilder:object:generate=false
type NutanixMachineConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec NutanixMachineConfigSpec `json:"spec,omitempty"`
}

// NutanixMachineConfigList contains a list of NutanixMachineConfig
//
// +kubebuilder:object:root=true
type NutanixMachineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NutanixMachineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NutanixMachineConfig{}, &NutanixMachineConfigList{})
}
