package v1alpha1

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

const (
	SFPPlus PhysicalNetworkConnectorType = "SFP_PLUS"
	QSFP    PhysicalNetworkConnectorType = "QSFP"
	RJ45    PhysicalNetworkConnectorType = "RJ45"
)

type PhysicalNetworkConnectorType string

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make generate" to regenerate code after modifying this file

// SnowMachineConfigSpec defines the desired state of SnowMachineConfigSpec.
type SnowMachineConfigSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	// The AMI ID from which to create the machine instance.
	AMIID string `json:"amiID,omitempty"`

	// InstanceType is the type of instance to create.
	InstanceType string `json:"instanceType,omitempty"`

	// PhysicalNetworkConnector is the physical network connector type to use for creating direct network interfaces (DNI).
	// Valid values: "SFP_PLUS" (default), "QSFP" and "RJ45".
	PhysicalNetworkConnector PhysicalNetworkConnectorType `json:"physicalNetworkConnector,omitempty"`

	// SSHKeyName is the name of the ssh key defined in the aws snow key pairs, to attach to the instance.
	SshKeyName string `json:"sshKeyName,omitempty"`

	// Devices contains a device ip list assigned by the user to provision machines.
	Devices []string `json:"devices,omitempty"`

	// ContainersVolume provides the configuration options for the containers data storage volume.
	ContainersVolume *snowv1.Volume `json:"containersVolume,omitempty"`

	// NonRootVolumes provides the configuration options for the non root storage volumes.
	NonRootVolumes []*snowv1.Volume `json:"nonRootVolumes,omitempty"`

	// OSFamily is the node instance OS.
	// Valid values: "bottlerocket" and "ubuntu".
	OSFamily OSFamily `json:"osFamily,omitempty"`

	// Network provides the custom network setting for the machine.
	Network SnowNetwork `json:"network"`

	// HostOSConfiguration provides OS specific configurations for the machine
	HostOSConfiguration *HostOSConfiguration `json:"hostOSConfiguration,omitempty"`
}

// SnowNetwork specifies the network configurations for snow.
type SnowNetwork struct {
	// DirectNetworkInterfaces contains a list of direct network interface (DNI) configuration.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	DirectNetworkInterfaces []SnowDirectNetworkInterface `json:"directNetworkInterfaces,omitempty"`
}

// SnowDirectNetworkInterface defines a direct network interface (DNI) configuration.
type SnowDirectNetworkInterface struct {
	// Index is the index number of DNI used to clarify the position in the list. Usually starts with 1.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=8
	// +optional
	Index int `json:"index,omitempty"`

	// VlanID is the vlan id assigned by the user for the DNI.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4095
	// +optional
	VlanID *int32 `json:"vlanID,omitempty"`

	// DHCP defines whether DHCP is used to assign ip for the DNI.
	// +optional
	DHCP bool `json:"dhcp,omitempty"`

	// IPPool contains a reference to a snow ip pool which provides a range of ip addresses.
	// When specified, an ip address selected from the pool is allocated to this DNI.
	// +optional
	IPPoolRef *Ref `json:"ipPoolRef,omitempty"`

	// Primary indicates whether the DNI is primary or not.
	// +optional
	Primary bool `json:"primary,omitempty"`
}

func (s *SnowMachineConfig) SetManagedBy(clusterName string) {
	if s.Annotations == nil {
		s.Annotations = map[string]string{}
	}
	s.Annotations[managementAnnotation] = clusterName
}

func (s *SnowMachineConfig) OSFamily() OSFamily {
	return s.Spec.OSFamily
}

// SnowMachineConfigStatus defines the observed state of SnowMachineConfig.
type SnowMachineConfigStatus struct {
	// SpecValid is set to true if vspheredatacenterconfig is validated.
	SpecValid bool `json:"specValid,omitempty"`

	// FailureMessage indicates that there is a fatal problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SnowMachineConfig is the Schema for the SnowMachineConfigs API.
type SnowMachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnowMachineConfigSpec   `json:"spec,omitempty"`
	Status SnowMachineConfigStatus `json:"status,omitempty"`
}

func (s *SnowMachineConfig) SetDefaults() {
	setSnowMachineConfigDefaults(s)
}

func (s *SnowMachineConfig) Validate() error {
	return validateSnowMachineConfig(s)
}

// ValidateHasSSHKeyName verifies a SnowMachineConfig object must have a SshKeyName.
// This validation only runs in SnowMachineConfig validation webhook, as we support
// auto-generate and import ssh key when creating a cluster via CLI.
func (s *SnowMachineConfig) ValidateHasSSHKeyName() error {
	if len(s.Spec.SshKeyName) <= 0 {
		return errors.New("SnowMachineConfig SshKeyName must not be empty")
	}
	return nil
}

func (s *SnowMachineConfig) SetControlPlaneAnnotation() {
	if s.Annotations == nil {
		s.Annotations = map[string]string{}
	}

	s.Annotations[controlPlaneAnnotation] = "true"
}

func (s *SnowMachineConfig) SetEtcdAnnotation() {
	if s.Annotations == nil {
		s.Annotations = map[string]string{}
	}

	s.Annotations[etcdAnnotation] = "true"
}

// IPPoolRefs returns a slice of snow IP pools that belongs to a snowMachineConfig.
func (s *SnowMachineConfig) IPPoolRefs() []Ref {
	ipPoolRefMap := make(refSet, 1)

	for _, dni := range s.Spec.Network.DirectNetworkInterfaces {
		ipPoolRefMap.addIfNotNil(dni.IPPoolRef)
	}

	return ipPoolRefMap.toSlice()
}

// +kubebuilder:object:generate=false

// Same as SnowMachineConfig except stripped down for generation of yaml file during generate clusterconfig.
type SnowMachineConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec SnowMachineConfigSpec `json:"spec,omitempty"`
}

func (s *SnowMachineConfig) ConvertConfigToConfigGenerateStruct() *SnowMachineConfigGenerate {
	namespace := defaultEksaNamespace
	if s.Namespace != "" {
		namespace = s.Namespace
	}
	config := &SnowMachineConfigGenerate{
		TypeMeta: s.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        s.Name,
			Annotations: s.Annotations,
			Namespace:   namespace,
		},
		Spec: s.Spec,
	}

	return config
}

func (s *SnowMachineConfig) Marshallable() Marshallable {
	return s.ConvertConfigToConfigGenerateStruct()
}

//+kubebuilder:object:root=true

// SnowMachineConfigList contains a list of SnowMachineConfig.
type SnowMachineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnowMachineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SnowMachineConfig{}, &SnowMachineConfigList{})
}
