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
	// It must include the Kubernetes version(s). For example, a template used for
	// Kubernetes 1.27 could be ubuntu-2204-1.27.
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
	// Project is an optional property that specifies the Prism Central project so that machine resources
	// can be linked to it. The project identifier (uuid or name) can be obtained from the Prism Central console
	// or using the Prism Central API.
	// +optional
	Project *NutanixResourceIdentifier `json:"project,omitempty"`

	// systemDiskSize is size (in Quantity format) of the system disk of the VM
	// The minimum systemDiskSize is 20Gi bytes
	// +kubebuilder:validation:Required
	SystemDiskSize resource.Quantity `json:"systemDiskSize"`

	// additionalCategories is a list of optional categories to be added to the VM.
	// Categories must be created in Prism Central before they can be used.
	// +kubebuilder:validation:Optional
	AdditionalCategories []NutanixCategoryIdentifier `json:"additionalCategories,omitempty"`
}

// SetDefaults sets defaults to NutanixMachineConfig if user has not provided.
func (in *NutanixMachineConfig) SetDefaults() {
	setNutanixMachineConfigDefaults(in)
}

// PauseReconcile pauses the reconciliation of the NutanixMachineConfig.
func (in *NutanixMachineConfig) PauseReconcile() {
	if in.Annotations == nil {
		in.Annotations = map[string]string{}
	}
	in.Annotations[pausedAnnotation] = "true"
}

// IsReconcilePaused returns true if the NutanixMachineConfig is paused.
func (in *NutanixMachineConfig) IsReconcilePaused() bool {
	if s, ok := in.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

// SetControlPlane sets the NutanixMachineConfig as a control plane node.
func (in *NutanixMachineConfig) SetControlPlane() {
	if in.Annotations == nil {
		in.Annotations = map[string]string{}
	}
	in.Annotations[controlPlaneAnnotation] = "true"
}

// IsControlPlane returns true if the NutanixMachineConfig is a control plane node.
func (in *NutanixMachineConfig) IsControlPlane() bool {
	if s, ok := in.Annotations[controlPlaneAnnotation]; ok {
		return s == "true"
	}
	return false
}

// SetEtcd sets the NutanixMachineConfig as an etcd node.
func (in *NutanixMachineConfig) SetEtcd() {
	if in.Annotations == nil {
		in.Annotations = map[string]string{}
	}
	in.Annotations[etcdAnnotation] = "true"
}

// IsEtcd returns true if the NutanixMachineConfig is an etcd node.
func (in *NutanixMachineConfig) IsEtcd() bool {
	if s, ok := in.Annotations[etcdAnnotation]; ok {
		return s == "true"
	}
	return false
}

// SetManagedBy sets the cluster name that manages the NutanixMachineConfig.
func (in *NutanixMachineConfig) SetManagedBy(clusterName string) {
	if in.Annotations == nil {
		in.Annotations = map[string]string{}
	}
	in.Annotations[managementAnnotation] = clusterName
}

// IsManaged returns true if the NutanixMachineConfig is managed by a cluster.
func (in *NutanixMachineConfig) IsManaged() bool {
	if s, ok := in.Annotations[managementAnnotation]; ok {
		return s != ""
	}
	return false
}

// OSFamily returns the OSFamily of the NutanixMachineConfig.
func (in *NutanixMachineConfig) OSFamily() OSFamily {
	return in.Spec.OSFamily
}

// Users returns a list of configuration for OS users.
func (in *NutanixMachineConfig) Users() []UserConfiguration {
	return in.Spec.Users
}

// GetNamespace returns the namespace of the NutanixMachineConfig.
func (in *NutanixMachineConfig) GetNamespace() string {
	return in.Namespace
}

// GetName returns the name of the NutanixMachineConfig.
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

// ConvertConfigToConfigGenerateStruct converts the NutanixMachineConfig to NutanixMachineConfigGenerate.
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

// Marshallable returns a Marshallable version of the NutanixMachineConfig.
func (in *NutanixMachineConfig) Marshallable() Marshallable {
	return in.ConvertConfigToConfigGenerateStruct()
}

// Validate validates the NutanixMachineConfig.
func (in *NutanixMachineConfig) Validate() error {
	return validateNutanixMachineConfig(in)
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
