// Important: Run "make generate" to regenerate code after modifying this file
// json tags are required; new fields must have json tags for the fields to be serialized

package v1alpha1

import (
	"fmt"

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

// SetDefaults sets defaults to NutanixMachineConfig if user has not provided.
func (in *NutanixMachineConfig) SetDefaults() {
	setNutanixMachineConfigDefaults(in)
}

// PauseReconcile pauses the reconciliation of the NutanixMachineConfig.
func (in *NutanixMachineConfig) PauseReconcile() {
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

func validateNutanixMachineConfig(c *NutanixMachineConfig) error {
	if err := validateObjectMeta(c.ObjectMeta); err != nil {
		return fmt.Errorf("NutanixMachineConfig: %v", err)
	}

	if err := validateNutanixReferences(c); err != nil {
		return fmt.Errorf("NutanixMachineConfig: %v", err)
	}

	if err := validateMinimumNutanixMachineSpecs(c); err != nil {
		return fmt.Errorf("NutanixMachineConfig: %v", err)
	}

	if c.Spec.OSFamily != Ubuntu && c.Spec.OSFamily != Bottlerocket {
		return fmt.Errorf(
			"NutanixMachineConfig: unsupported spec.osFamily (%v); Please use one of the following: %s, %s",
			c.Spec.OSFamily,
			Ubuntu,
			Bottlerocket,
		)
	}

	if len(c.Spec.Users) <= 0 {
		return fmt.Errorf("NutanixMachineConfig: missing spec.Users: %s", c.Name)
	}

	return nil
}

func validateMinimumNutanixMachineSpecs(c *NutanixMachineConfig) error {
	if c.Spec.VCPUSockets <= 0 {
		return fmt.Errorf("NutanixMachineConfig: vcpu sockets must be greater than 0")
	}

	if c.Spec.VCPUsPerSocket <= 0 {
		return fmt.Errorf("NutanixMachineConfig: vcpu per socket must be greater than 0")
	}

	if c.Spec.MemorySize.Value() <= 0 {
		return fmt.Errorf("NutanixMachineConfig: memory size must be greater than 0")
	}

	if c.Spec.SystemDiskSize.Value() <= 0 {
		return fmt.Errorf("NutanixMachineConfig: system disk size must be greater than 0")
	}

	return nil
}

func validateNutanixReferences(c *NutanixMachineConfig) error {
	if err := validateNutanixResourceIdentifierType(&c.Spec.Subnet); err != nil {
		return err
	}

	if c.Spec.Subnet.Name == nil && c.Spec.Subnet.UUID == nil {
		return fmt.Errorf("NutanixMachineConfig: missing subnet name or uuid: %s", c.Name)
	}

	if err := validateNutanixResourceIdentifierType(&c.Spec.Cluster); err != nil {
		return err
	}

	if c.Spec.Cluster.Name == nil && c.Spec.Cluster.UUID == nil {
		return fmt.Errorf("NutanixMachineConfig: missing cluster name or uuid: %s", c.Name)
	}

	if err := validateNutanixResourceIdentifierType(&c.Spec.Image); err != nil {
		return err
	}

	if c.Spec.Image.Name == nil && c.Spec.Image.UUID == nil {
		return fmt.Errorf("NutanixMachineConfig: missing image name or uuid: %s", c.Name)
	}

	return nil
}

func validateNutanixResourceIdentifierType(i *NutanixResourceIdentifier) error {
	if i.Type != NutanixIdentifierName && i.Type != NutanixIdentifierUUID {
		return fmt.Errorf("NutanixMachineConfig: invalid identifier type: %s", i.Type)
	}
	return nil
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
